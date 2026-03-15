package ui

import (
	"fmt"
	"strings"
	"time"
)

// ANSI color codes
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[1;33m"
	Blue    = "\033[0;34m"
	Magenta = "\033[0;35m"
	Cyan    = "\033[0;36m"
	White   = "\033[1;37m"

	BgBlue  = "\033[44m"
	BgGreen = "\033[42m"
	BgRed   = "\033[41m"
)

// ClearScreen clears the terminal
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// StatusIcon returns a colored status indicator
func StatusIcon(status string) string {
	switch status {
	case "active":
		return Green + "● Activo" + Reset
	case "inactive":
		return Red + "● Inactivo" + Reset
	default:
		return Yellow + "● No Instalado" + Reset
	}
}

// ShowHeader displays the main panel header
func ShowHeader(version, status, port string) {
	ClearScreen()
	statusStr := StatusIcon(status)

	fmt.Println(Cyan + Bold + "╔══════════════════════════════════════════════╗" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + White + Bold + "        ⚡ ZIVPN PANEL " + version + "               " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "╠══════════════════════════════════════════════╣" + Reset)

	portInfo := ""
	if port != "?" && port != "" {
		portInfo = Dim + " (Puerto: " + port + ")" + Reset
	}
	fmt.Printf(Cyan+Bold+"║"+Reset+"  Estado: %-36s"+Cyan+Bold+"║"+Reset+"\n", statusStr+portInfo)

	fmt.Println(Cyan + Bold + "╠══════════════════════════════════════════════╣" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Green + "  1." + Reset + " 📥 Instalar ZiVPN                    " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Green + "  2." + Reset + " 🗑️  Desinstalar ZiVPN                  " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Magenta + "  3." + Reset + " 👤 Crear Usuario                      " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Magenta + "  4." + Reset + " ❌ Eliminar Usuario                   " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Magenta + "  5." + Reset + " 📋 Listar Usuarios                    " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Yellow + "  6." + Reset + " 🧹 Purgar Expirados                   " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Blue + "  7." + Reset + " ℹ️  Estado del Servicio                 " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "║" + Reset + Red + "  0." + Reset + " 🚪 Salir                              " + Cyan + Bold + "║" + Reset)
	fmt.Println(Cyan + Bold + "╚══════════════════════════════════════════════╝" + Reset)
}

// ShowSuccess prints a success message
func ShowSuccess(msg string) {
	fmt.Println(Green + Bold + "  ✅ " + msg + Reset)
}

// ShowError prints an error message
func ShowError(msg string) {
	fmt.Println(Red + Bold + "  ❌ " + msg + Reset)
}

// ShowWarning prints a warning message
func ShowWarning(msg string) {
	fmt.Println(Yellow + Bold + "  ⚠️  " + msg + Reset)
}

// ShowInfo prints an info message
func ShowInfo(msg string) {
	fmt.Println(Cyan + "  ℹ️  " + msg + Reset)
}

// UserStatus represents an individual user's expiration status
type UserStatus struct {
	Number    int
	Password  string
	CreatedAt time.Time
	ExpiresAt time.Time
	IsExpired bool
	DaysLeft  int
}

// ShowUsersTable renders a formatted table of users
func ShowUsersTable(users []UserStatus) {
	if len(users) == 0 {
		ShowInfo("No hay usuarios registrados.")
		return
	}

	fmt.Println()
	fmt.Println(White + Bold + "  ┌────┬────────────────┬─────────────┬─────────────┬──────────────┐" + Reset)
	fmt.Println(White + Bold + "  │ #  │ Password       │ Creado      │ Expira      │ Estado       │" + Reset)
	fmt.Println(White + Bold + "  ├────┼────────────────┼─────────────┼─────────────┼──────────────┤" + Reset)

	for _, u := range users {
		statusStr := Green + "Activo" + Reset
		daysInfo := ""
		if u.IsExpired {
			statusStr = Red + "Vencido" + Reset
		} else {
			daysInfo = fmt.Sprintf(" (%dd)", u.DaysLeft)
		}

		pass := u.Password
		if len(pass) > 14 {
			pass = pass[:11] + "..."
		}

		fmt.Printf("  │ %-2d │ %-14s │ %-11s │ %-11s │ %-18s│\n",
			u.Number,
			pass,
			u.CreatedAt.Format("2006-01-02"),
			u.ExpiresAt.Format("2006-01-02"),
			statusStr+daysInfo,
		)
	}

	fmt.Println(White + Bold + "  └────┴────────────────┴─────────────┴─────────────┴──────────────┘" + Reset)

	// Summary
	activeCount := 0
	expiredCount := 0
	for _, u := range users {
		if u.IsExpired {
			expiredCount++
		} else {
			activeCount++
		}
	}
	fmt.Printf("  %sTotal: %d%s  |  %sActivos: %d%s  |  %sVencidos: %d%s\n",
		Dim, len(users), Reset,
		Green, activeCount, Reset,
		Red, expiredCount, Reset,
	)
}

// ShowServiceInfo renders detailed service status
func ShowServiceInfo(status, port string) {
	fmt.Println()
	fmt.Println(Cyan + Bold + "  ╭──────────────────────────────────────╮" + Reset)
	fmt.Println(Cyan + Bold + "  │     ℹ️  ESTADO DEL SERVICIO          │" + Reset)
	fmt.Println(Cyan + Bold + "  ╰──────────────────────────────────────╯" + Reset)
	fmt.Println()
	fmt.Printf("  Servicio:    %s\n", White+"zivpn.service"+Reset)
	fmt.Printf("  Estado:      %s\n", StatusIcon(status))
	fmt.Printf("  Puerto UDP:  %s\n", White+port+Reset)
	fmt.Printf("  Rango UDP:   %s\n", Dim+"6000-19999 → "+port+Reset)
	fmt.Printf("  Config:      %s\n", Dim+"/etc/zivpn/config.json"+Reset)
	fmt.Printf("  Binario:     %s\n", Dim+"/usr/local/bin/zivpn"+Reset)
	fmt.Println()
}

// Separator prints a visual separator
func Separator() {
	fmt.Println(Dim + "  " + strings.Repeat("─", 44) + Reset)
}
