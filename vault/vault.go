package vault

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
)

type Entry struct {
	Service  string
	Username string
	Password string
	Created  int64
}

type Vault struct {
	Entries []Entry
	mu      sync.RWMutex
}

func New() *Vault {
	return &Vault{
		Entries: make([]Entry, 0),
	}
}

func (v *Vault) Add(entry Entry) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Entries = append(v.Entries, entry)
}

func (v *Vault) Search(term string) []Entry {
	v.mu.RLock()
	defer v.mu.RUnlock()

	term = strings.ToLower(term)
	results := make([]Entry, 0)

	for _, entry := range v.Entries {
		if strings.Contains(strings.ToLower(entry.Service), term) ||
			strings.Contains(strings.ToLower(entry.Username), term) {
			results = append(results, entry)
		}
	}

	return results
}

func (v *Vault) Delete(service, username string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i, entry := range v.Entries {
		if entry.Service == service && entry.Username == username {
			v.Entries = append(v.Entries[:i], v.Entries[i+1:]...)
			return nil
		}
	}

	return errors.New("entry not found")
}

func (v *Vault) Update(service, username, newPassword string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i, entry := range v.Entries {
		if entry.Service == service && entry.Username == username {
			v.Entries[i].Password = newPassword
			return nil
		}
	}

	return errors.New("entry not found")
}

func (v *Vault) ToJSON() ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return json.Marshal(v.Entries)
}

func (v *Vault) FromJSON(data []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	return json.Unmarshal(data, &v.Entries)
}

func (v *Vault) SaveToFile(filename string, encrypt func([]byte) []byte) error {
	data, err := v.ToJSON()
	if err != nil {
		return err
	}

	encrypted := encrypt(data)
	return os.WriteFile(filename, encrypted, 0600)
}

func (v *Vault) LoadFromFile(filename string, decrypt func([]byte) []byte) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	decrypted := decrypt(data)
	return v.FromJSON(decrypted)
}
