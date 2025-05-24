package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"pw/config"
	"pw/security"
	"pw/vault"
)

type CLI struct {
	config *config.Config
	vault  *vault.Vault
	theme  Theme
}

func NewCLI(cfg *config.Config, v *vault.Vault) *CLI {
	return &CLI{
		config: cfg,
		vault:  v,
		theme:  GetTheme(cfg.Theme),
	}
}

func (c *CLI) Run() error {
	for {
		if c.config.ClearScreen {
			ClearScreen()
		}

		c.showHeader()
		c.showMenu()

		choice := ReadInput("Enter your choice: ")

		switch strings.TrimSpace(choice) {
		case "1":
			c.handleGeneratePassword()
		case "2":
			c.handleAddPassword()
		case "3":
			c.handleViewVault()
		case "4":
			c.handleSearchEntries()
		case "5":
			c.handleUpdateEntry()
		case "6":
			c.handleDeleteEntry()
		case "7":
			c.handleExportVault()
		case "8":
			c.handleImportVault()
		case "9":
			c.handleVaultStatistics()
		case "10":
			c.handleSettings()
		case "11":
			c.handleBackup()
		case "q", "Q":
			return nil
		default:
			ShowError("Invalid choice. Please try again.")
		}

		if c.config.ClearScreen {
			PressEnterToContinue()
		}
	}
}

func (c *CLI) showHeader() {
	header := `
    ____                                     __   _    __            ____
   / __ \____ ____________    ______  _____/ /  | |  / /___ ___  __/ / /
  / /_/ / __ \/___/ ___/ /   / ___/ |/_/ _  /   | | / / __ \/  |/_/ / / 
 / ____/ /_/ /  // /__/ /___/ /  _>  <  __/ /    | |/ / /_/ />  </ / /  
/_/    \____/  / \___/____/_/  /_/|_|\__/_/     |___/\____/_/|_/_/_/   
              /_/                                                         
`
	fmt.Print(c.theme.HeaderStyle.Apply(header))
}

func (c *CLI) showMenu() {
	menu := []string{
		"1. Generate Password",
		"2. Add Password to Vault",
		"3. View Vault",
		"4. Search Entries",
		"5. Update Entry",
		"6. Delete Entry",
		"7. Export Vault",
		"8. Import Vault",
		"9. View Statistics",
		"10. Settings",
		"11. Backup Vault",
		"Q. Quit",
	}

	fmt.Println("\nMenu:")
	for _, item := range menu {
		fmt.Println(c.theme.MenuItemStyle.Apply(item))
	}
	fmt.Println()
}

func (c *CLI) handleGeneratePassword() {
	length := c.config.PasswordLength
	lengthStr := ReadInput(fmt.Sprintf("Password length (default %d): ", length))
	if lengthStr != "" {
		if n, err := strconv.Atoi(lengthStr); err == nil && n > 0 {
			length = n
		}
	}

	password := GeneratePassword(length)
	strength := security.AnalyzePassword(password)

	fmt.Printf("\nGenerated password: %s\n", c.theme.PasswordStyle.Apply(password))
	if c.config.ShowStrength {
		fmt.Printf("Password strength: %s (Score: %.0f)\n", strength.Level, strength.Score)
		if len(strength.Suggestions) > 0 {
			fmt.Println("Suggestions:")
			for _, s := range strength.Suggestions {
				fmt.Printf("- %s\n", s)
			}
		}
	}
}

func (c *CLI) handleAddPassword() {
	service := ReadInput("Enter service name: ")
	if service == "" {
		ShowError("Service name cannot be empty")
		return
	}

	username := ReadInput("Enter username: ")
	if username == "" {
		ShowError("Username cannot be empty")
		return
	}

	password := ReadSecureInput("Enter password (leave empty to generate): ")
	if password == "" {
		password = GeneratePassword(c.config.PasswordLength)
		fmt.Printf("Generated password: %s\n", c.theme.PasswordStyle.Apply(password))
	}

	if c.config.ShowStrength {
		strength := security.AnalyzePassword(password)
		if strength.Score < float64(c.config.MinStrength) {
			if !ConfirmAction(fmt.Sprintf("Password strength (%.0f) is below minimum (%d). Use anyway?",
				strength.Score, c.config.MinStrength)) {
				return
			}
		}
	}

	notes := ReadInput("Enter notes (optional): ")

	entry := vault.Entry{
		Service:   service,
		Username:  username,
		Password:  password,
		Notes:     notes,
		CreatedAt: time.Now(),
	}

	if err := c.vault.AddEntry(entry); err != nil {
		ShowError("Failed to add entry: %v", err)
		return
	}
	ShowSuccess("Password added successfully!")
}

func (c *CLI) handleViewVault() {
	entries := c.vault.GetEntries()
	if len(entries) == 0 {
		ShowInfo("No entries in vault.")
		return
	}

	for i, entry := range entries {
		fmt.Printf("\n%d. ", i+1)
		ShowPasswordEntry(entry, c.config.HidePasswords)
	}
}

func (c *CLI) handleSearchEntries() {
	query := ReadInput("Enter search term: ")
	entries := c.vault.SearchEntries(query)

	if len(entries) == 0 {
		ShowInfo("No matching entries found.")
		return
	}

	fmt.Printf("\nFound %d entries:\n", len(entries))
	for i, entry := range entries {
		fmt.Printf("\n%d. ", i+1)
		ShowPasswordEntry(entry, c.config.HidePasswords)
	}
}

func (c *CLI) handleUpdateEntry() {
	c.handleViewVault()
	entries := c.vault.GetEntries()

	idxStr := ReadInput("\nEnter entry number to update: ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(entries) {
		ShowError("Invalid entry number")
		return
	}

	entry := entries[idx-1]
	fmt.Println("\nLeave fields empty to keep current values:")

	if service := ReadInput(fmt.Sprintf("Service (%s): ", entry.Service)); service != "" {
		entry.Service = service
	}

	if username := ReadInput(fmt.Sprintf("Username (%s): ", entry.Username)); username != "" {
		entry.Username = username
	}

	if password := ReadSecureInput("New password (leave empty to keep current): "); password != "" {
		entry.Password = password
	}

	if notes := ReadInput(fmt.Sprintf("Notes (%s): ", entry.Notes)); notes != "" {
		entry.Notes = notes
	}

	if err := c.vault.UpdateEntry(idx-1, entry); err != nil {
		ShowError("Failed to update entry: %v", err)
		return
	}
	ShowSuccess("Entry updated successfully!")
}

func (c *CLI) handleDeleteEntry() {
	c.handleViewVault()

	idxStr := ReadInput("\nEnter entry number to delete: ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > len(c.vault.GetEntries()) {
		ShowError("Invalid entry number.")
		return
	}

	entry := c.vault.GetEntries()[idx-1]
	fmt.Printf("\nDeleting entry for %s (%s)\n", entry.Service, entry.Username)

	if !ConfirmAction("Are you sure you want to delete this entry?") {
		return
	}

	if err := c.vault.DeleteEntry(idx - 1); err != nil {
		ShowError("Failed to delete entry: %v", err)
		return
	}
	ShowSuccess("Entry deleted successfully!")
}

func (c *CLI) handleExportVault() {
	fmt.Println("\nExport formats:")
	fmt.Println("1. JSON")
	fmt.Println("2. CSV")
	fmt.Println("3. Text")

	choice := ReadInput("Choose format (1-3): ")

	var format vault.ExportFormat
	switch choice {
	case "1":
		format = vault.JSONFormat
	case "2":
		format = vault.CSVFormat
	case "3":
		format = vault.TextFormat
	default:
		ShowError("Invalid format choice.")
		return
	}

	filePath := ReadInput("Enter export file path: ")
	if filePath == "" {
		ShowError("File path cannot be empty.")
		return
	}

	options := vault.ExportOptions{
		Format:          format,
		IncludePassword: ConfirmAction("Include passwords in export?"),
		IncludeNotes:    ConfirmAction("Include notes in export?"),
		IncludeTime:     ConfirmAction("Include timestamps in export?"),
	}

	if err := c.vault.Export(filePath, options); err != nil {
		ShowError("Export failed: %v", err)
		return
	}

	ShowSuccess("Vault exported successfully!")
}

func (c *CLI) handleImportVault() {
	fmt.Println("\nImport formats:")
	fmt.Println("1. JSON")
	fmt.Println("2. CSV")
	fmt.Println("3. Auto-detect")

	choice := ReadInput("Choose format (1-3): ")

	var format vault.ImportFormat
	switch choice {
	case "1":
		format = vault.JSONImport
	case "2":
		format = vault.CSVImport
	case "3":
		format = vault.AutoDetect
	default:
		ShowError("Invalid format choice.")
		return
	}

	filePath := ReadInput("Enter import file path: ")
	if filePath == "" {
		ShowError("File path cannot be empty.")
		return
	}

	skipDuplicates := ConfirmAction("Skip duplicate entries?")
	options := vault.ImportOptions{
		Format:          format,
		SkipDuplicates:  skipDuplicates,
		UpdateExisting:  !skipDuplicates && ConfirmAction("Update existing entries?"),
		RequiredFields:  []string{"service", "username"},
		DefaultPassword: GeneratePassword(c.config.PasswordLength),
	}

	if err := c.vault.Import(filePath, options); err != nil {
		ShowError("Import failed: %v", err)
		return
	}

	ShowSuccess("Vault imported successfully!")
}

func (c *CLI) handleVaultStatistics() {
	stats := c.vault.CalculateStatistics()

	fmt.Println("\nVault Statistics:")
	fmt.Printf("Total Entries: %d\n", stats.TotalEntries)
	fmt.Printf("Unique Services: %d\n", stats.UniqueServices)
	fmt.Printf("Unique Usernames: %d\n", stats.UniqueUsernames)
	fmt.Printf("Average Password Length: %.1f\n", stats.AveragePasswordLen)

	fmt.Println("\nPassword Strength Distribution:")
	fmt.Printf("Strong: %d\n", stats.StrongPasswords)
	fmt.Printf("Medium: %d\n", stats.MediumPasswords)
	fmt.Printf("Weak: %d\n", stats.WeakPasswords)

	if len(stats.CommonServices) > 0 {
		fmt.Println("\nMost Used Services:")
		for _, s := range stats.CommonServices {
			fmt.Printf("%s (%d entries)\n", s.Service, s.Count)
		}
	}

	if len(stats.PasswordReuse) > 0 {
		fmt.Println("\nPassword Reuse:")
		for _, p := range stats.PasswordReuse {
			fmt.Printf("%s used in %d services: %s\n",
				p.Password, p.Count, strings.Join(p.ServicesList, ", "))
		}
	}
}

func (c *CLI) handleSettings() {
	for {
		fmt.Println("\nSettings:")
		fmt.Println("1. Change default password length")
		fmt.Println("2. Toggle password strength check")
		fmt.Println("3. Change minimum password strength")
		fmt.Println("4. Toggle password hiding")
		fmt.Println("5. Toggle screen clearing")
		fmt.Println("6. Change theme")
		fmt.Println("7. Configure backup settings")
		fmt.Println("8. Back to main menu")

		choice := ReadInput("\nEnter choice: ")

		switch choice {
		case "1":
			c.handlePasswordLengthSetting()
		case "2":
			c.config.ShowStrength = !c.config.ShowStrength
			ShowSuccess("Password strength check %s", onOff(c.config.ShowStrength))
		case "3":
			c.handleMinStrengthSetting()
		case "4":
			c.config.HidePasswords = !c.config.HidePasswords
			ShowSuccess("Password hiding %s", onOff(c.config.HidePasswords))
		case "5":
			c.config.ClearScreen = !c.config.ClearScreen
			ShowSuccess("Screen clearing %s", onOff(c.config.ClearScreen))
		case "6":
			c.handleThemeSetting()
		case "7":
			c.handleBackupSettings()
		case "8":
			return
		default:
			ShowError("Invalid choice")
		}
	}
}

func (c *CLI) handlePasswordLengthSetting() {
	lengthStr := ReadInput(fmt.Sprintf("Enter new default password length (current: %d): ",
		c.config.PasswordLength))
	if length, err := strconv.Atoi(lengthStr); err == nil && length > 0 {
		c.config.PasswordLength = length
		ShowSuccess("Default password length updated to %d", length)
	} else {
		ShowError("Invalid length")
	}
}

func (c *CLI) handleMinStrengthSetting() {
	strengthStr := ReadInput(fmt.Sprintf("Enter new minimum password strength (0-100, current: %d): ",
		c.config.MinStrength))
	if strength, err := strconv.Atoi(strengthStr); err == nil && strength >= 0 && strength <= 100 {
		c.config.MinStrength = strength
		ShowSuccess("Minimum password strength updated to %d", strength)
	} else {
		ShowError("Invalid strength value")
	}
}

func (c *CLI) handleThemeSetting() {
	fmt.Println("\nAvailable themes:")
	for i, theme := range ListThemes() {
		fmt.Printf("%d. %s\n", i+1, theme)
	}

	choiceStr := ReadInput("\nChoose theme number: ")
	if choice, err := strconv.Atoi(choiceStr); err == nil {
		themes := ListThemes()
		if choice > 0 && choice <= len(themes) {
			c.config.Theme = themes[choice-1]
			c.theme = GetTheme(c.config.Theme)
			ShowSuccess("Theme updated to %s", c.config.Theme)
			return
		}
	}
	ShowError("Invalid theme choice")
}

func (c *CLI) handleBackupSettings() {
	fmt.Println("\nBackup Settings:")
	fmt.Println("1. Toggle auto-backup")
	fmt.Println("2. Change backup interval")
	fmt.Println("3. Change max backups")
	fmt.Println("4. Change backup directory")
	fmt.Println("5. Back to settings")

	choice := ReadInput("\nEnter choice: ")

	switch choice {
	case "1":
		c.config.AutoBackup = !c.config.AutoBackup
		ShowSuccess("Auto-backup %s", onOff(c.config.AutoBackup))
	case "2":
		if days, err := strconv.Atoi(ReadInput("Enter backup interval in days: ")); err == nil && days > 0 {
			c.config.BackupInterval = days
			ShowSuccess("Backup interval updated to %d days", days)
		} else {
			ShowError("Invalid interval")
		}
	case "3":
		if max, err := strconv.Atoi(ReadInput("Enter maximum number of backups to keep: ")); err == nil && max > 0 {
			c.config.MaxBackups = max
			ShowSuccess("Maximum backups updated to %d", max)
		} else {
			ShowError("Invalid number")
		}
	case "4":
		dir := ReadInput("Enter new backup directory path: ")
		if dir != "" {
			c.config.BackupDir = dir
			ShowSuccess("Backup directory updated")
		} else {
			ShowError("Invalid directory path")
		}
	case "5":
		return
	default:
		ShowError("Invalid choice")
	}
}

func (c *CLI) handleBackup() {
	backupManager, err := vault.NewBackupManager(c.config.BackupDir, c.config.MaxBackups)
	if err != nil {
		ShowError("Failed to initialize backup manager: %v", err)
		return
	}

	fmt.Println("\nBackup Options:")
	fmt.Println("1. Create backup")
	fmt.Println("2. Restore backup")
	fmt.Println("3. List backups")
	fmt.Println("4. Back to main menu")

	choice := ReadInput("\nEnter choice: ")

	switch choice {
	case "1":
		if err := backupManager.CreateBackup(c.config.VaultPath, c.vault.Encrypt); err != nil {
			ShowError("Backup failed: %v", err)
		} else {
			ShowSuccess("Backup created successfully")
		}
	case "2":
		backups, err := backupManager.ListBackups()
		if err != nil {
			ShowError("Failed to list backups: %v", err)
			return
		}
		if len(backups) == 0 {
			ShowInfo("No backups available")
			return
		}

		fmt.Println("\nAvailable backups:")
		for i, backup := range backups {
			fmt.Printf("%d. %s\n", i+1, backup)
		}

		choiceStr := ReadInput("\nChoose backup to restore: ")
		if choice, err := strconv.Atoi(choiceStr); err == nil && choice > 0 && choice <= len(backups) {
			if ConfirmAction("This will overwrite your current vault. Continue?") {
				if err := backupManager.RestoreBackup(backups[choice-1], c.config.VaultPath, c.vault.Decrypt); err != nil {
					ShowError("Restore failed: %v", err)
				} else {
					ShowSuccess("Backup restored successfully")
				}
			}
		} else {
			ShowError("Invalid choice")
		}
	case "3":
		backups, err := backupManager.ListBackups()
		if err != nil {
			ShowError("Failed to list backups: %v", err)
			return
		}
		if len(backups) == 0 {
			ShowInfo("No backups available")
			return
		}

		fmt.Println("\nAvailable backups:")
		for _, backup := range backups {
			fmt.Println(backup)
		}
	case "4":
		return
	default:
		ShowError("Invalid choice")
	}
}

func onOff(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
