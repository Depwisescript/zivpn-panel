package zivpn

import (
	"encoding/json"
	"os"
	"time"
)

const (
	ConfigDir   = "/etc/zivpn"
	ConfigFile  = "/etc/zivpn/config.json"
	UsersFile   = "/etc/zivpn/users.json"
	BinaryPath  = "/usr/local/bin/zivpn"
	ServiceName = "zivpn.service"
	DefaultPort = "5667"
)

// ZivpnConfig represents the daemon's native config.json
type ZivpnConfig struct {
	Listen  string `json:"listen"`
	Cert    string `json:"cert"`
	Key     string `json:"key"`
	MaxConn int    `json:"max_conn"`
	Auth    struct {
		Mode   string   `json:"mode"`
		Config []string `json:"config"`
	} `json:"auth"`
}

// UserEntry stores user metadata with expiration
type UserEntry struct {
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UsersDB is the list of managed users
type UsersDB struct {
	Users []UserEntry `json:"users"`
}

// LoadConfig reads the ZiVPN daemon config
func LoadConfig() (*ZivpnConfig, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}
	var cfg ZivpnConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes the ZiVPN daemon config
func SaveConfig(cfg *ZivpnConfig) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

// LoadUsers reads the users database
func LoadUsers() (*UsersDB, error) {
	data, err := os.ReadFile(UsersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &UsersDB{Users: []UserEntry{}}, nil
		}
		return nil, err
	}
	var db UsersDB
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

// SaveUsers writes the users database
func SaveUsers(db *UsersDB) error {
	data, err := json.MarshalIndent(db, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(UsersFile, data, 0644)
}

// SyncPasswords updates config.json passwords from the users database
// This keeps the daemon config in sync with the managed users
func SyncPasswords(db *UsersDB) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	passwords := []string{}
	now := time.Now()
	for _, u := range db.Users {
		if u.ExpiresAt.After(now) {
			passwords = append(passwords, u.Password)
		}
	}

	// Always keep at least the default "1" password if no users
	if len(passwords) == 0 {
		passwords = []string{"1"}
	}

	cfg.Auth.Mode = "passwords"
	cfg.Auth.Config = passwords

	return SaveConfig(cfg)
}
