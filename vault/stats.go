package vault

import (
	"sort"
	"strings"
	"time"
	"unicode"
)

type VaultStatistics struct {
	TotalEntries       int
	UniqueServices     int
	UniqueUsernames    int
	AveragePasswordLen float64
	WeakPasswords      int
	MediumPasswords    int
	StrongPasswords    int
	OldestEntry        time.Time
	NewestEntry        time.Time
	CommonServices     []ServiceCount
	CommonUsernames    []UsernameCount
	PasswordReuse      []PasswordReuseInfo
	EntriesPerMonth    map[string]int
}

type ServiceCount struct {
	Service string
	Count   int
}

type UsernameCount struct {
	Username string
	Count    int
}

type PasswordReuseInfo struct {
	Password     string
	Count        int
	ServicesList []string
}

func (v *Vault) CalculateStatistics() VaultStatistics {
	stats := VaultStatistics{
		TotalEntries:    len(v.Entries),
		EntriesPerMonth: make(map[string]int),
	}

	if len(v.Entries) == 0 {
		return stats
	}

	services := make(map[string]int)
	usernames := make(map[string]int)
	passwords := make(map[string][]string)
	totalPasswordLen := 0

	stats.OldestEntry = v.Entries[0].CreatedAt
	stats.NewestEntry = v.Entries[0].CreatedAt

	for _, entry := range v.Entries {
		services[entry.Service]++
		usernames[entry.Username]++
		passwords[entry.Password] = append(passwords[entry.Password], entry.Service)
		totalPasswordLen += len(entry.Password)

		if entry.CreatedAt.Before(stats.OldestEntry) {
			stats.OldestEntry = entry.CreatedAt
		}
		if entry.CreatedAt.After(stats.NewestEntry) {
			stats.NewestEntry = entry.CreatedAt
		}

		monthKey := entry.CreatedAt.Format("2006-01")
		stats.EntriesPerMonth[monthKey]++

		strength := analyzePasswordStrength(entry.Password)
		switch {
		case strength < 40:
			stats.WeakPasswords++
		case strength < 70:
			stats.MediumPasswords++
		default:
			stats.StrongPasswords++
		}
	}

	stats.UniqueServices = len(services)
	stats.UniqueUsernames = len(usernames)
	stats.AveragePasswordLen = float64(totalPasswordLen) / float64(len(v.Entries))

	stats.CommonServices = getTopCounts(services, 5)
	stats.CommonUsernames = getTopUsernameCounts(usernames, 5)
	stats.PasswordReuse = getPasswordReuse(passwords, 2)

	return stats
}

func getTopCounts(items map[string]int, limit int) []ServiceCount {
	var counts []ServiceCount
	for item, count := range items {
		counts = append(counts, ServiceCount{Service: item, Count: count})
	}

	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count == counts[j].Count {
			return counts[i].Service < counts[j].Service
		}
		return counts[i].Count > counts[j].Count
	})

	if len(counts) > limit {
		counts = counts[:limit]
	}
	return counts
}

func getTopUsernameCounts(usernames map[string]int, limit int) []UsernameCount {
	var counts []UsernameCount
	for username, count := range usernames {
		counts = append(counts, UsernameCount{Username: username, Count: count})
	}

	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count == counts[j].Count {
			return counts[i].Username < counts[j].Username
		}
		return counts[i].Count > counts[j].Count
	})

	if len(counts) > limit {
		counts = counts[:limit]
	}
	return counts
}

func getPasswordReuse(passwords map[string][]string, minCount int) []PasswordReuseInfo {
	var reuse []PasswordReuseInfo
	for password, services := range passwords {
		if len(services) >= minCount {
			reuse = append(reuse, PasswordReuseInfo{
				Password:     maskPassword(password),
				Count:        len(services),
				ServicesList: services,
			})
		}
	}

	sort.Slice(reuse, func(i, j int) bool {
		return reuse[i].Count > reuse[j].Count
	})

	return reuse
}

func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}

func analyzePasswordStrength(password string) int {
	var score int
	length := len(password)

	score += length * 4

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasUpper {
		score += 10
	}
	if hasLower {
		score += 10
	}
	if hasNumber {
		score += 10
	}
	if hasSpecial {
		score += 15
	}

	return score
}
