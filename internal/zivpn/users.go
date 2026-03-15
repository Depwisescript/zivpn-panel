package zivpn

import (
	"fmt"
	"os/exec"
	"time"
)

// CreateUser adds a new user with an expiration in days
func CreateUser(password string, days int) error {
	db, err := LoadUsers()
	if err != nil {
		return fmt.Errorf("error leyendo usuarios: %v", err)
	}

	// Check for duplicates
	for _, u := range db.Users {
		if u.Password == password {
			return fmt.Errorf("el usuario '%s' ya existe", password)
		}
	}

	now := time.Now()
	entry := UserEntry{
		Password:  password,
		CreatedAt: now,
		ExpiresAt: now.AddDate(0, 0, days),
	}

	db.Users = append(db.Users, entry)

	if err := SaveUsers(db); err != nil {
		return fmt.Errorf("error guardando usuarios: %v", err)
	}

	if err := SyncPasswords(db); err != nil {
		return fmt.Errorf("error sincronizando config: %v", err)
	}

	return restartService()
}

// RemoveUser deletes a user by password
func RemoveUser(password string) error {
	db, err := LoadUsers()
	if err != nil {
		return fmt.Errorf("error leyendo usuarios: %v", err)
	}

	found := false
	newUsers := []UserEntry{}
	for _, u := range db.Users {
		if u.Password == password {
			found = true
			continue
		}
		newUsers = append(newUsers, u)
	}

	if !found {
		return fmt.Errorf("usuario '%s' no encontrado", password)
	}

	db.Users = newUsers

	if err := SaveUsers(db); err != nil {
		return fmt.Errorf("error guardando usuarios: %v", err)
	}

	if err := SyncPasswords(db); err != nil {
		return fmt.Errorf("error sincronizando config: %v", err)
	}

	return restartService()
}

// ListUsers returns all managed users
func ListUsers() ([]UserEntry, error) {
	db, err := LoadUsers()
	if err != nil {
		return nil, err
	}
	return db.Users, nil
}

// PurgeExpired removes all users whose expiration has passed
// Returns the count of removed users
func PurgeExpired() (int, error) {
	db, err := LoadUsers()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	active := []UserEntry{}
	purged := 0

	for _, u := range db.Users {
		if u.ExpiresAt.After(now) {
			active = append(active, u)
		} else {
			purged++
		}
	}

	if purged == 0 {
		return 0, nil
	}

	db.Users = active

	if err := SaveUsers(db); err != nil {
		return 0, err
	}

	if err := SyncPasswords(db); err != nil {
		return 0, err
	}

	_ = restartService()
	return purged, nil
}

// restartService restarts the zivpn systemd service
func restartService() error {
	return exec.Command("systemctl", "restart", ServiceName).Run()
}
