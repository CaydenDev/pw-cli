package main

import (
	"fmt"
	"os"
	"path/filepath"

	"pw/config"
	"pw/ui"
	"pw/util"
	"pw/vault"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	logDir := filepath.Join(homeDir, ".pwvault", "logs")
	logger, err := util.NewLogger(logDir, util.INFO)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	configManager, err := config.NewManager()
	if err != nil {
		logger.Error("Failed to initialize config manager: %v", err)
		os.Exit(1)
	}

	cfg := configManager.Get()

	vaultDir := filepath.Dir(cfg.VaultPath)
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		logger.Error("Failed to create vault directory: %v", err)
		os.Exit(1)
	}

	var key string
	if len(os.Args) > 1 {
		key = os.Args[1]
	} else {
		key = ui.ReadSecureInput("Enter your 32-character encryption key: ")
	}

	if len(key) != 32 {
		ui.ShowError("Key must be exactly 32 characters long")
		os.Exit(1)
	}

	v, err := vault.NewVault(cfg.VaultPath, []byte(key))
	if err != nil {
		logger.Error("Failed to initialize vault: %v", err)
		os.Exit(1)
	}

	cli := ui.NewCLI(cfg, v)

	if cfg.AutoBackup {
		backupManager, err := vault.NewBackupManager(cfg.BackupDir, cfg.MaxBackups)
		if err != nil {
			logger.Warning("Failed to initialize backup manager: %v", err)
		} else {
			if err := backupManager.CreateBackup(cfg.VaultPath, v.Encrypt); err != nil {
				logger.Warning("Auto-backup failed: %v", err)
			} else {
				logger.Info("Auto-backup created successfully")
			}
		}
	}

	if err := cli.Run(); err != nil {
		logger.Error("CLI error: %v", err)
		os.Exit(1)
	}

	if err := configManager.Save(); err != nil {
		logger.Error("Failed to save configuration: %v", err)
		os.Exit(1)
	}

	if err := v.Save(); err != nil {
		logger.Error("Failed to save vault: %v", err)
		os.Exit(1)
	}
}
