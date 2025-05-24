package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type BackupManager struct {
	backupDir  string
	maxBackups int
}

func NewBackupManager(backupDir string, maxBackups int) (*BackupManager, error) {
	if maxBackups < 1 {
		maxBackups = 5
	}

	if backupDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		backupDir = filepath.Join(homeDir, ".pwvault", "backups")
	}

	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return nil, err
	}

	return &BackupManager{
		backupDir:  backupDir,
		maxBackups: maxBackups,
	}, nil
}

func (bm *BackupManager) CreateBackup(vaultPath string, encrypt func([]byte) []byte) error {
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("vault_backup_%s.dat", timestamp)
	backupPath := filepath.Join(bm.backupDir, backupName)

	encryptedData := encrypt(vaultData)
	if err := os.WriteFile(backupPath, encryptedData, 0600); err != nil {
		return err
	}

	return bm.cleanOldBackups()
}

func (bm *BackupManager) RestoreBackup(backupFile string, vaultPath string, decrypt func([]byte) []byte) error {
	backupPath := filepath.Join(bm.backupDir, backupFile)
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	decryptedData := decrypt(backupData)
	return os.WriteFile(vaultPath, decryptedData, 0600)
}

func (bm *BackupManager) ListBackups() ([]string, error) {
	files, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return nil, err
	}

	backups := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".dat" {
			backups = append(backups, file.Name())
		}
	}

	return backups, nil
}

func (bm *BackupManager) cleanOldBackups() error {
	backups, err := bm.ListBackups()
	if err != nil {
		return err
	}

	if len(backups) <= bm.maxBackups {
		return nil
	}

	for i := 0; i < len(backups)-bm.maxBackups; i++ {
		backupPath := filepath.Join(bm.backupDir, backups[i])
		if err := os.Remove(backupPath); err != nil {
			return err
		}
	}

	return nil
}
