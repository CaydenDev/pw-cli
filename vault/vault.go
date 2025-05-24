package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"pw/crypto"
)

type Entry struct {
	Service   string
	Username  string
	Password  string
	Notes     string
	CreatedAt time.Time
}

type Vault struct {
	Entries  []Entry
	mu       sync.RWMutex
	key      []byte
	filePath string
}

func NewVault(path string, key []byte) (*Vault, error) {
	v := &Vault{
		Entries:  make([]Entry, 0),
		key:      key,
		filePath: "vault.dat",
	}

	if _, err := os.Stat("vault.dat"); err == nil {
		if err := v.Load(); err != nil {
			return nil, err
		}
		if string(v.key) != string(key) {
			return nil, fmt.Errorf("invalid master key")
		}
	}

	return v, nil
}

func (v *Vault) Encrypt(data []byte) []byte {
	return crypto.Encrypt(data, v.key)
}

func (v *Vault) Decrypt(data []byte) []byte {
	return crypto.Decrypt(data, v.key)
}

func New() *Vault {
	return &Vault{
		Entries: make([]Entry, 0),
	}
}

func (v *Vault) AddEntry(entry Entry) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Entries = append(v.Entries, entry)
	return v.Save()
}

func (v *Vault) GetEntries() []Entry {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return append([]Entry{}, v.Entries...)
}

func (v *Vault) UpdateEntry(index int, entry Entry) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if index >= 0 && index < len(v.Entries) {
		v.Entries[index] = entry
		return v.Save()
	}
	return fmt.Errorf("invalid entry index")
}

func (v *Vault) DeleteEntry(index int) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if index >= 0 && index < len(v.Entries) {
		v.Entries = append(v.Entries[:index], v.Entries[index+1:]...)
		return v.Save()
	}
	return fmt.Errorf("invalid entry index")
}

func (v *Vault) SearchEntries(query string) []Entry {
	v.mu.RLock()
	defer v.mu.RUnlock()

	query = strings.ToLower(query)
	results := make([]Entry, 0)

	for _, entry := range v.Entries {
		if strings.Contains(strings.ToLower(entry.Service), query) ||
			strings.Contains(strings.ToLower(entry.Username), query) ||
			strings.Contains(strings.ToLower(entry.Notes), query) {
			results = append(results, entry)
		}
	}

	return results
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

type VaultData struct {
	Entries []Entry `json:"entries"`
	Key     []byte  `json:"key"`
}

func (v *Vault) Save() error {
	data := VaultData{
		Entries: v.Entries,
		Key:     v.key,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	encrypted := v.Encrypt(jsonData)
	return os.WriteFile(v.filePath, encrypted, 0600)
}

func (v *Vault) Load() error {
	data, err := os.ReadFile(v.filePath)
	if err != nil {
		return err
	}

	decrypted := v.Decrypt(data)

	var vaultData VaultData
	if err := json.Unmarshal(decrypted, &vaultData); err != nil {
		return err
	}

	v.Entries = vaultData.Entries
	v.key = vaultData.Key
	return nil
}

func (v *Vault) findEntry(service, username string) *Entry {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for i := range v.Entries {
		if v.Entries[i].Service == service && v.Entries[i].Username == username {
			return &v.Entries[i]
		}
	}

	return nil
}
