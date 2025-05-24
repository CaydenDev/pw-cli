package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"pw/crypto"
	"pw/ui"
	"pw/vault"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
)

func main() {
	term := ui.New()
	v := vault.New()
	gen := crypto.NewGenerator()
	var key []byte

	options := []string{
		"Generate Password",
		"Add Password to Vault",
		"View Vault",
		"Search Vault",
		"Update Password",
		"Delete Entry",
		"Exit",
	}

	for {
		term.ShowMenu("Password Generator and Vault", options)
		choice := term.ReadLine("Enter choice (1-7): ")

		switch choice {
		case "1":
			lengthStr := term.ReadLine("Enter password length: ")
			length, err := strconv.Atoi(lengthStr)
			if err != nil || length <= 0 {
				term.ShowError("Invalid length")
				continue
			}

			password := gen.GenerateString(length, charset)
			term.ShowSuccess("Generated password: " + password)

		case "2", "3", "4", "5", "6":
			if len(key) != 32 {
				keyStr := term.ReadSecure("Enter vault key (32 characters): ")
				if len(keyStr) != 32 {
					term.ShowError("Key must be 32 characters long")
					continue
				}
				key = []byte(keyStr)
			}

			if choice != "2" {
				err := v.LoadFromFile("vault.dat", func(data []byte) []byte {
					return crypto.Decrypt(data, key)
				})
				if err != nil && !os.IsNotExist(err) {
					term.ShowError("Error loading vault: " + err.Error())
					continue
				}
			}

			switch choice {
			case "2":
				entry := vault.Entry{
					Service:  term.ReadLine("Enter service name: "),
					Username: term.ReadLine("Enter username: "),
					Password: term.ReadLine("Enter password: "),
					Created:  time.Now().Unix(),
				}
				v.Add(entry)

				err := v.SaveToFile("vault.dat", func(data []byte) []byte {
					return crypto.Encrypt(data, key)
				})
				if err != nil {
					term.ShowError("Error saving vault: " + err.Error())
					continue
				}
				term.ShowSuccess("Entry saved successfully")

			case "3":
				term.Clear()
				fmt.Println("Vault contents:")
				for i, entry := range v.Search("") {
					term.ShowPasswordEntry(entry, i)
				}
				term.WaitForEnter()

			case "4":
				searchTerm := term.ReadLine("Enter search term: ")
				results := v.Search(searchTerm)

				if len(results) == 0 {
					term.ShowError("No matching entries found")
					continue
				}

				term.Clear()
				fmt.Println("Search results:")
				for i, entry := range results {
					term.ShowPasswordEntry(entry, i)
				}
				term.WaitForEnter()

			case "5":
				service := term.ReadLine("Enter service name: ")
				username := term.ReadLine("Enter username: ")
				newPassword := term.ReadLine("Enter new password: ")

				err := v.Update(service, username, newPassword)
				if err != nil {
					term.ShowError(err.Error())
					continue
				}

				err = v.SaveToFile("vault.dat", func(data []byte) []byte {
					return crypto.Encrypt(data, key)
				})
				if err != nil {
					term.ShowError("Error saving vault: " + err.Error())
					continue
				}
				term.ShowSuccess("Password updated successfully")

			case "6":
				service := term.ReadLine("Enter service name: ")
				username := term.ReadLine("Enter username: ")

				if !term.Confirm("Are you sure you want to delete this entry?") {
					continue
				}

				err := v.Delete(service, username)
				if err != nil {
					term.ShowError(err.Error())
					continue
				}

				err = v.SaveToFile("vault.dat", func(data []byte) []byte {
					return crypto.Encrypt(data, key)
				})
				if err != nil {
					term.ShowError("Error saving vault: " + err.Error())
					continue
				}
				term.ShowSuccess("Entry deleted successfully")
			}

		case "7":
			term.Clear()
			fmt.Println("Goodbye!")
			return

		default:
			term.ShowError("Invalid choice")
		}
	}
}
