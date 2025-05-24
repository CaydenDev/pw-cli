package vault

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type ExportFormat string

const (
	JSONFormat ExportFormat = "json"
	CSVFormat  ExportFormat = "csv"
	TextFormat ExportFormat = "txt"
)

type ExportOptions struct {
	Format          ExportFormat
	IncludePassword bool
	IncludeNotes    bool
	IncludeTime     bool
}

func (v *Vault) Export(filePath string, options ExportOptions) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	switch options.Format {
	case JSONFormat:
		return v.exportJSON(file, options)
	case CSVFormat:
		return v.exportCSV(file, options)
	case TextFormat:
		return v.exportText(file, options)
	default:
		return fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

func (v *Vault) exportJSON(file *os.File, options ExportOptions) error {
	type exportEntry struct {
		Service   string     `json:"service"`
		Username  string     `json:"username"`
		Password  string     `json:"password,omitempty"`
		Notes     string     `json:"notes,omitempty"`
		CreatedAt *time.Time `json:"created_at,omitempty"`
	}

	entries := make([]exportEntry, 0, len(v.Entries))
	for _, e := range v.Entries {
		entry := exportEntry{
			Service:  e.Service,
			Username: e.Username,
		}

		if options.IncludePassword {
			entry.Password = e.Password
		}
		if options.IncludeNotes {
			entry.Notes = e.Notes
		}
		if options.IncludeTime {
			entry.CreatedAt = &e.CreatedAt
		}

		entries = append(entries, entry)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

func (v *Vault) exportCSV(file *os.File, options ExportOptions) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Service", "Username"}
	if options.IncludePassword {
		header = append(header, "Password")
	}
	if options.IncludeNotes {
		header = append(header, "Notes")
	}
	if options.IncludeTime {
		header = append(header, "Created At")
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	for _, e := range v.Entries {
		record := []string{e.Service, e.Username}
		if options.IncludePassword {
			record = append(record, e.Password)
		}
		if options.IncludeNotes {
			record = append(record, e.Notes)
		}
		if options.IncludeTime {
			record = append(record, e.CreatedAt.Format(time.RFC3339))
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func (v *Vault) exportText(file *os.File, options ExportOptions) error {
	var sb strings.Builder

	for _, e := range v.Entries {
		sb.WriteString(fmt.Sprintf("Service: %s\n", e.Service))
		sb.WriteString(fmt.Sprintf("Username: %s\n", e.Username))

		if options.IncludePassword {
			sb.WriteString(fmt.Sprintf("Password: %s\n", e.Password))
		}
		if options.IncludeNotes && e.Notes != "" {
			sb.WriteString(fmt.Sprintf("Notes: %s\n", e.Notes))
		}
		if options.IncludeTime {
			sb.WriteString(fmt.Sprintf("Created: %s\n", e.CreatedAt.Format(time.RFC3339)))
		}

		sb.WriteString("\n")
	}

	_, err := file.WriteString(sb.String())
	return err
}
