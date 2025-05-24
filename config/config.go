package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	VaultPath      string `json:"vault_path"`
	BackupDir      string `json:"backup_dir"`
	MaxBackups     int    `json:"max_backups"`
	AutoBackup     bool   `json:"auto_backup"`
	BackupInterval int    `json:"backup_interval_days"`
	PasswordLength int    `json:"default_password_length"`
	MinStrength    int    `json:"minimum_password_strength"`
	ShowStrength   bool   `json:"show_password_strength"`
	ClearScreen    bool   `json:"clear_screen"`
	HidePasswords  bool   `json:"hide_passwords"`
	InactivityLock int    `json:"inactivity_lock_minutes"`
	ExportFormat   string `json:"export_format"`
	Theme          string `json:"theme"`
}

type Manager struct {
	configPath string
	config     *Config
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".pwvault")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")

	m := &Manager{
		configPath: configPath,
		config:     getDefaultConfig(),
	}

	if err := m.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

func getDefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		VaultPath:      filepath.Join(homeDir, ".pwvault", "vault.dat"),
		BackupDir:      filepath.Join(homeDir, ".pwvault", "backups"),
		MaxBackups:     5,
		AutoBackup:     true,
		BackupInterval: 7,
		PasswordLength: 16,
		MinStrength:    60,
		ShowStrength:   true,
		ClearScreen:    true,
		HidePasswords:  false,
		InactivityLock: 15,
		ExportFormat:   "json",
		Theme:          "default",
	}
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return m.Save()
		}
		return err
	}

	return json.Unmarshal(data, m.config)
}

func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0600)
}

func (m *Manager) Get() *Config {
	return m.config
}

func (m *Manager) Update(updater func(*Config)) error {
	updater(m.config)
	return m.Save()
}

func (m *Manager) Reset() error {
	m.config = getDefaultConfig()
	return m.Save()
}
