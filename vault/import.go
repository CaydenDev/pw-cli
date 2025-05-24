package vault

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type ImportFormat string

const (
	AutoDetect ImportFormat = "auto"
	JSONImport ImportFormat = "json"
	CSVImport  ImportFormat = "csv"
)

type ImportOptions struct {
	Format          ImportFormat
	SkipDuplicates  bool
	UpdateExisting  bool
	RequiredFields  []string
	DefaultPassword string
}

func (v *Vault) Import(filePath string, options ImportOptions) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	format := options.Format
	if format == AutoDetect {
		format, err = detectFormat(file)
		if err != nil {
			return err
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			return err
		}
	}

	switch format {
	case JSONImport:
		return v.importJSON(file, options)
	case CSVImport:
		return v.importCSV(file, options)
	default:
		return fmt.Errorf("unsupported import format: %s", format)
	}
}

func detectFormat(file *os.File) (ImportFormat, error) {
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	content := string(buf[:n])
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return JSONImport, nil
	}

	if strings.Contains(content, ",") {
		return CSVImport, nil
	}

	return "", fmt.Errorf("unable to detect file format")
}

func (v *Vault) importJSON(file *os.File, options ImportOptions) error {
	var entries []Entry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		return err
	}

	return v.processImportedEntries(entries, options)
}

func (v *Vault) importCSV(file *os.File, options ImportOptions) error {
	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return err
	}

	indexMap := make(map[string]int)
	for i, field := range header {
		indexMap[strings.ToLower(strings.TrimSpace(field))] = i
	}

	var entries []Entry
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		entry := Entry{CreatedAt: time.Now()}

		if i, ok := indexMap["service"]; ok && i < len(record) {
			entry.Service = record[i]
		}
		if i, ok := indexMap["username"]; ok && i < len(record) {
			entry.Username = record[i]
		}
		if i, ok := indexMap["password"]; ok && i < len(record) {
			entry.Password = record[i]
		} else {
			entry.Password = options.DefaultPassword
		}
		if i, ok := indexMap["notes"]; ok && i < len(record) {
			entry.Notes = record[i]
		}
		if i, ok := indexMap["created_at"]; ok && i < len(record) {
			if t, err := time.Parse(time.RFC3339, record[i]); err == nil {
				entry.CreatedAt = t
			}
		}

		entries = append(entries, entry)
	}

	return v.processImportedEntries(entries, options)
}

func (v *Vault) processImportedEntries(entries []Entry, options ImportOptions) error {
	for _, entry := range entries {
		if !validateRequiredFields(entry, options.RequiredFields) {
			continue
		}

		existing := v.findEntry(entry.Service, entry.Username)
		if existing != nil {
			if options.SkipDuplicates {
				continue
			}
			if options.UpdateExisting {
				*existing = entry
				continue
			}
		}

		v.Entries = append(v.Entries, entry)
	}

	return nil
}

func validateRequiredFields(entry Entry, requiredFields []string) bool {
	for _, field := range requiredFields {
		switch strings.ToLower(field) {
		case "service":
			if entry.Service == "" {
				return false
			}
		case "username":
			if entry.Username == "" {
				return false
			}
		case "password":
			if entry.Password == "" {
				return false
			}
		}
	}
	return true
}
