package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"pw/vault"
)

type Terminal struct {
	reader *bufio.Reader
}

func New() *Terminal {
	return &Terminal{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (t *Terminal) Clear() {
	fmt.Print("\033[H\033[2J")
}

func (t *Terminal) ReadLine(prompt string) string {
	fmt.Print(prompt)
	input, _ := t.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (t *Terminal) ReadSecure(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	fmt.Print("\n")
	return input
}

func (t *Terminal) ShowMenu(title string, options []string) {
	t.Clear()
	fmt.Printf("=== %s ===\n\n", title)

	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}
	fmt.Println("\nPress Enter after your choice")
}

func (t *Terminal) ShowError(message string) {
	fmt.Printf("\nError: %s\nPress Enter to continue...", message)
	t.ReadLine("")
}

func (t *Terminal) ShowSuccess(message string) {
	fmt.Printf("\n%s\nPress Enter to continue...", message)
	t.ReadLine("")
}

func (t *Terminal) ShowPasswordEntry(entry interface{}, index int) {
	if e, ok := entry.(vault.Entry); ok {
		fmt.Printf("\n%d. Service: %s\n   Username: %s\n   Password: %s\n   Created: %s\n",
			index+1, e.Service, e.Username, e.Password, e.CreatedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("\n%d. Invalid entry format\n", index+1)
	}
}

func (t *Terminal) Confirm(prompt string) bool {
	response := t.ReadLine(fmt.Sprintf("%s (y/N): ", prompt))
	return strings.ToLower(response) == "y"
}

func (t *Terminal) WaitForEnter() {
	t.ReadLine("Press Enter to continue...")
}
