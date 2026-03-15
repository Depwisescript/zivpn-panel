package zivpn

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// IsInstalled checks if the ZiVPN daemon is installed
func IsInstalled() bool {
	_, err := os.Stat(BinaryPath)
	return err == nil
}

// GetStatus returns the current service status
func GetStatus() string {
	if !IsInstalled() {
		return "no instalado"
	}
	out, err := exec.Command("systemctl", "is-active", ServiceName).Output()
	if err != nil {
		return "inactivo"
	}
	return strings.TrimSpace(string(out))
}

// GetPort reads the current listening port from config
func GetPort() string {
	cfg, err := LoadConfig()
	if err != nil {
		return "?"
	}
	// Listen format is ":PORT"
	port := strings.TrimPrefix(cfg.Listen, ":")
	if port == "" {
		return "?"
	}
	return port
}

// installLibSSL11 installs libssl1.1 if not present (required by zivpn binary)
func installLibSSL11() {
	if _, err := os.Stat("/usr/lib/x86_64-linux-gnu/libssl.so.1.1"); err == nil {
		return
	}
	if _, err := os.Stat("/usr/lib/aarch64-linux-gnu/libssl.so.1.1"); err == nil {
		return
	}

	arch := runtime.GOARCH
	var url string
	if arch == "amd64" {
		url = "http://nz2.archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb"
	} else if arch == "arm64" || arch == "aarch64" {
		url = "http://ports.ubuntu.com/ubuntu-ports/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_arm64.deb"
	}

	if url != "" {
		_ = exec.Command("wget", "-q", "-O", "/tmp/libssl1.1.deb", url).Run()
		_ = exec.Command("dpkg", "-i", "/tmp/libssl1.1.deb").Run()
		_ = os.Remove("/tmp/libssl1.1.deb")
	}
}

// InstallZivpn installs the udp-zivpn server v1.4.9 on the given port
func InstallZivpn(port string) error {
	// 0. Dependencies
	fmt.Println("  → Instalando dependencias...")
	_ = exec.Command("apt-get", "update").Run()
	_ = exec.Command("apt-get", "install", "-y", "curl", "openssl", "iptables", "libc6-i386").Run()
	installLibSSL11()

	// Enable IPv4 Forwarding
	_ = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()
	_ = exec.Command("bash", "-c", "grep -q 'net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf").Run()

	// 1. Download binary
	archRaw := runtime.GOARCH
	var binURL string

	if archRaw == "amd64" {
		binURL = "https://github.com/zahidbd2/udp-zivpn/releases/download/udp-zivpn_1.4.9/udp-zivpn-linux-amd64"
	} else if archRaw == "arm64" {
		binURL = "https://github.com/zahidbd2/udp-zivpn/releases/download/udp-zivpn_1.4.9/udp-zivpn-linux-arm64"
	} else {
		return fmt.Errorf("arquitectura no soportada: %s", archRaw)
	}

	fmt.Println("  → Descargando binario ZiVPN...")
	if _, err := os.Stat(BinaryPath); os.IsNotExist(err) {
		if err := exec.Command("curl", "-L", "-s", "-f", "-o", BinaryPath, binURL).Run(); err != nil {
			return fmt.Errorf("fallo la descarga del binario: %v", err)
		}
		os.Chmod(BinaryPath, 0755)
	}

	// 2. Config files
	fmt.Println("  → Configurando servidor...")
	os.MkdirAll(ConfigDir, 0755)
	configJSON := `{"listen": ":` + port + `", "cert": "/etc/zivpn/zivpn.crt", "key": "/etc/zivpn/zivpn.key", "max_conn": 0, "auth": {"mode": "passwords", "config": ["1"]}}`
	os.WriteFile(ConfigFile, []byte(configJSON), 0644)

	// Initialize empty users database
	if _, err := os.Stat(UsersFile); os.IsNotExist(err) {
		SaveUsers(&UsersDB{Users: []UserEntry{}})
	}

	// 3. SSL certificates
	fmt.Println("  → Generando certificados SSL...")
	exec.Command("openssl", "req", "-new", "-newkey", "rsa:4096", "-days", "3650", "-nodes", "-x509",
		"-subj", "/C=US/ST=CA/L=LA/O=Zivpn/CN=zivpn", "-keyout", "/etc/zivpn/zivpn.key", "-out", "/etc/zivpn/zivpn.crt").Run()

	// 4. Systemd service
	fmt.Println("  → Registrando servicio systemd...")
	svc := `[Unit]
Description=zivpn VPN Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/zivpn
ExecStart=/usr/local/bin/zivpn server -c /etc/zivpn/config.json
Restart=always
RestartSec=3
Environment=ZIVPN_LOG_LEVEL=info
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE CAP_NET_RAW
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE CAP_NET_RAW

[Install]
WantedBy=multi-user.target`

	os.WriteFile("/etc/systemd/system/zivpn.service", []byte(svc), 0644)
	exec.Command("systemctl", "daemon-reload").Run()
	_ = exec.Command("systemctl", "enable", "zivpn.service").Run()
	if err := exec.Command("systemctl", "restart", "zivpn.service").Run(); err != nil {
		return fmt.Errorf("fallo reiniciar zivpn.service: %v", err)
	}

	// 5. Verify
	time.Sleep(1500 * time.Millisecond)
	if err := exec.Command("systemctl", "is-active", "--quiet", "zivpn.service").Run(); err != nil {
		logCmd, _ := exec.Command("journalctl", "-u", "zivpn.service", "--no-pager", "-n", "10").Output()
		logs := string(logCmd)
		if logs == "" {
			logs = "No se pudieron obtener logs."
		}

		_ = exec.Command("systemctl", "stop", "zivpn.service").Run()
		_ = os.Remove("/etc/systemd/system/zivpn.service")
		_ = exec.Command("systemctl", "daemon-reload").Run()
		return fmt.Errorf("zivpn no pudo mantenerse activo en el puerto %s\n\nLOGS:\n%s", port, logs)
	}

	// 6. IPTables routing
	fmt.Println("  → Configurando reglas de red (iptables)...")
	setupIPTables(port)

	return nil
}

// RemoveZivpn completely uninstalls the ZiVPN daemon and cleans up
func RemoveZivpn() error {
	fmt.Println("  → Deteniendo servicio...")
	exec.Command("systemctl", "stop", "zivpn.service").Run()
	exec.Command("systemctl", "disable", "zivpn.service").Run()

	fmt.Println("  → Eliminando archivos...")
	os.Remove("/etc/systemd/system/zivpn.service")
	os.RemoveAll(ConfigDir)
	os.Remove(BinaryPath)

	fmt.Println("  → Limpiando reglas de red...")
	cleanIPTables()

	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}

// setupIPTables configures UDP port routing for gaming (6000-19999 → zivpn port)
func setupIPTables(port string) {
	devOut, _ := exec.Command("bash", "-c", "ip -4 route show default | awk '{print $5}' | head -1").Output()
	dev := strings.TrimSpace(string(devOut))
	if dev == "" {
		devOut, _ = exec.Command("bash", "-c", "ip link show up | grep -v loopback | grep -v 'lo:' | head -1 | awk '{print $2}' | cut -d':' -f1").Output()
		dev = strings.TrimSpace(string(devOut))
	}

	if dev != "" {
		// Clean existing rules
		exec.Command("bash", "-c", "iptables -t nat -S PREROUTING | grep '6000:19999' | sed 's/-A/-D/' | while read line; do iptables -t nat $line; done").Run()
		exec.Command("bash", "-c", "iptables -S INPUT | grep '6000:19999' | sed 's/-A/-D/' | while read line; do iptables $line; done").Run()
		exec.Command("bash", "-c", "iptables -S INPUT | grep -w '"+port+"' | sed 's/-A/-D/' | while read line; do iptables $line; done").Run()

		// Apply new rules
		_ = exec.Command("iptables", "-t", "nat", "-I", "PREROUTING", "1", "-i", dev, "-p", "udp", "--dport", "6000:19999", "-j", "REDIRECT", "--to-port", port).Run()
		_ = exec.Command("iptables", "-I", "INPUT", "1", "-p", "udp", "--dport", port, "-j", "ACCEPT").Run()
		_ = exec.Command("iptables", "-I", "INPUT", "1", "-p", "udp", "--dport", "6000:19999", "-j", "ACCEPT").Run()

		// MASQUERADE for return traffic
		_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run()
		_ = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run()
	}
}

// cleanIPTables removes all ZiVPN-related iptables rules
func cleanIPTables() {
	devOut, _ := exec.Command("bash", "-c", "ip -4 route ls | grep default | grep -Po '(?<=dev )(\\S+)' | head -1").Output()
	dev := strings.TrimSpace(string(devOut))
	if dev != "" {
		exec.Command("bash", "-c", "iptables -t nat -S PREROUTING | grep '6000:19999' | sed 's/-A/-D/' | while read line; do iptables -t nat $line; done").Run()
		exec.Command("bash", "-c", "iptables -S INPUT | grep '6000:19999' | sed 's/-A/-D/' | while read line; do iptables $line; done").Run()
		_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", dev, "-j", "MASQUERADE").Run()
	}
}
