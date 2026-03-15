package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Depwisescript/zivpn-panel/internal/ui"
	"github.com/Depwisescript/zivpn-panel/internal/zivpn"
)

const version = "v1.0.0"

func main() {
	// Must run as root
	if os.Geteuid() != 0 {
		ui.ShowError("Este panel debe ejecutarse como root.")
		os.Exit(1)
	}

	// Auto-purge expired users on startup
	if zivpn.IsInstalled() {
		purged, _ := zivpn.PurgeExpired()
		if purged > 0 {
			fmt.Printf("  🧹 Auto-purga: %d usuario(s) vencido(s) eliminado(s)\n", purged)
			time.Sleep(1 * time.Second)
		}
	}

	// Main loop
	for {
		status := zivpn.GetStatus()
		port := zivpn.GetPort()
		ui.ShowHeader(version, status, port)

		opt := ui.ReadOption()

		switch opt {
		case "1":
			handleInstall()
		case "2":
			handleUninstall()
		case "3":
			handleCreateUser()
		case "4":
			handleRemoveUser()
		case "5":
			handleListUsers()
		case "6":
			handlePurge()
		case "7":
			handleServiceInfo()
		case "0":
			ui.ClearScreen()
			ui.ShowInfo("¡Hasta luego! 👋")
			fmt.Println()
			os.Exit(0)
		default:
			ui.ShowError("Opción inválida.")
			time.Sleep(1 * time.Second)
		}
	}
}

func handleInstall() {
	fmt.Println()
	ui.Separator()

	if zivpn.IsInstalled() {
		ui.ShowWarning("ZiVPN ya está instalado.")
		if !ui.Confirm("¿Deseas reinstalar?") {
			return
		}
		zivpn.RemoveZivpn()
	}

	portStr := zivpn.DefaultPort
	fmt.Println()
	ui.ShowInfo("Instalando ZiVPN en el puerto " + portStr + " (automático)...")
	fmt.Println()

	if err := zivpn.InstallZivpn(portStr); err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
	} else {
		ui.ShowSuccess("ZiVPN instalado correctamente en el puerto " + portStr)
	}

	ui.Pause()
}

func handleUninstall() {
	fmt.Println()
	ui.Separator()

	if !zivpn.IsInstalled() {
		ui.ShowWarning("ZiVPN no está instalado.")
		ui.Pause()
		return
	}

	ui.ShowWarning("Esto eliminará ZiVPN, todos los usuarios y las configuraciones.")
	if !ui.Confirm("¿Estás seguro?") {
		return
	}

	if err := zivpn.RemoveZivpn(); err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
	} else {
		ui.ShowSuccess("ZiVPN desinstalado completamente.")
	}

	ui.Pause()
}

func handleCreateUser() {
	fmt.Println()
	ui.Separator()

	if !zivpn.IsInstalled() {
		ui.ShowError("ZiVPN no está instalado. Instálalo primero (opción 1).")
		ui.Pause()
		return
	}

	password := ui.ReadString("Password del usuario")
	if password == "" {
		ui.ShowError("El password no puede estar vacío.")
		ui.Pause()
		return
	}

	days := ui.ReadInt("Días de vigencia", 1, 365)

	if err := zivpn.CreateUser(password, days); err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
	} else {
		expiresAt := time.Now().AddDate(0, 0, days).Format("2006-01-02")
		ui.ShowSuccess(fmt.Sprintf("Usuario '%s' creado (expira: %s)", password, expiresAt))
	}

	ui.Pause()
}

func handleRemoveUser() {
	fmt.Println()
	ui.Separator()

	if !zivpn.IsInstalled() {
		ui.ShowError("ZiVPN no está instalado.")
		ui.Pause()
		return
	}

	// Show current users first
	users, err := zivpn.ListUsers()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Error leyendo usuarios: %v", err))
		ui.Pause()
		return
	}

	if len(users) == 0 {
		ui.ShowInfo("No hay usuarios para eliminar.")
		ui.Pause()
		return
	}

	showUsersList(users)

	password := ui.ReadString("Password del usuario a eliminar")
	if password == "" {
		return
	}

	if !ui.Confirm(fmt.Sprintf("¿Eliminar usuario '%s'?", password)) {
		return
	}

	if err := zivpn.RemoveUser(password); err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
	} else {
		ui.ShowSuccess(fmt.Sprintf("Usuario '%s' eliminado.", password))
	}

	ui.Pause()
}

func handleListUsers() {
	fmt.Println()
	ui.Separator()

	if !zivpn.IsInstalled() {
		ui.ShowError("ZiVPN no está instalado.")
		ui.Pause()
		return
	}

	users, err := zivpn.ListUsers()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
		ui.Pause()
		return
	}

	showUsersList(users)
	ui.Pause()
}

func handlePurge() {
	fmt.Println()
	ui.Separator()

	if !zivpn.IsInstalled() {
		ui.ShowError("ZiVPN no está instalado.")
		ui.Pause()
		return
	}

	purged, err := zivpn.PurgeExpired()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Error: %v", err))
	} else if purged == 0 {
		ui.ShowInfo("No hay usuarios vencidos.")
	} else {
		ui.ShowSuccess(fmt.Sprintf("%d usuario(s) vencido(s) eliminado(s).", purged))
	}

	ui.Pause()
}

func handleServiceInfo() {
	fmt.Println()
	status := zivpn.GetStatus()
	port := zivpn.GetPort()
	ui.ShowServiceInfo(status, port)

	// Show connected users count
	users, err := zivpn.ListUsers()
	if err == nil {
		now := time.Now()
		active := 0
		for _, u := range users {
			if u.ExpiresAt.After(now) {
				active++
			}
		}
		fmt.Printf("  Usuarios:    %s%d activos%s / %d total\n", ui.Green, active, ui.Reset, len(users))
	}
	fmt.Println()

	ui.Pause()
}

// showUsersList converts UserEntry to UserStatus and renders the table
func showUsersList(users []zivpn.UserEntry) {
	now := time.Now()
	statuses := make([]ui.UserStatus, len(users))
	for i, u := range users {
		isExpired := u.ExpiresAt.Before(now)
		daysLeft := 0
		if !isExpired {
			daysLeft = int(u.ExpiresAt.Sub(now).Hours() / 24)
		}
		statuses[i] = ui.UserStatus{
			Number:    i + 1,
			Password:  u.Password,
			CreatedAt: u.CreatedAt,
			ExpiresAt: u.ExpiresAt,
			IsExpired: isExpired,
			DaysLeft:  daysLeft,
		}
	}
	ui.ShowUsersTable(statuses)
}
