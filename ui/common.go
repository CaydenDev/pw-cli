package ui

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"pw/crypto"
	"pw/vault"
)

var reader = bufio.NewReader(os.Stdin)

func ClearScreen() {
	if runtime.GOOS == "windows" {
		fmt.Print("\033[H\033[2J")
	} else {
		fmt.Print("\033c")
	}
}

func ReadInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func ReadSecureInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	fmt.Print("\n")
	return input
}

func ShowError(format string, args ...interface{}) {
	fmt.Printf("\nError: "+format+"\n", args...)
}

func ShowSuccess(format string, args ...interface{}) {
	fmt.Printf("\nSuccess: "+format+"\n", args...)
}

func ShowInfo(format string, args ...interface{}) {
	fmt.Printf("\nInfo: "+format+"\n", args...)
}

func PressEnterToContinue() {
	fmt.Print("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func ConfirmAction(prompt string) bool {
	response := ReadInput(fmt.Sprintf("%s (y/N): ", prompt))
	return strings.ToLower(response) == "y"
}

func GeneratePassword(length int) string {
	gen := crypto.NewGenerator()
	return gen.GenerateString(length, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()")
}

func ShowPasswordEntry(entry interface{}, hidePassword bool) {
	if e, ok := entry.(vault.Entry); ok {
		password := e.Password
		if hidePassword {
			password = strings.Repeat("*", len(password))
		}

		fmt.Printf("Service: %s\n", e.Service)
		fmt.Printf("Username: %s\n", e.Username)
		fmt.Printf("Password: %s\n", password)
		if e.Notes != "" {
			fmt.Printf("Notes: %s\n", e.Notes)
		}
		fmt.Printf("Created: %s\n", e.CreatedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Invalid entry format")
	}
}
