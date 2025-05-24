package security

import (
	"math"
	"strings"
	"unicode"
)

type StrengthLevel int

const (
	VeryWeak StrengthLevel = iota
	Weak
	Medium
	Strong
	VeryStrong
)

type StrengthResult struct {
	Level       StrengthLevel
	Score       float64
	Suggestions []string
}

func (s StrengthLevel) String() string {
	switch s {
	case VeryWeak:
		return "Very Weak"
	case Weak:
		return "Weak"
	case Medium:
		return "Medium"
	case Strong:
		return "Strong"
	case VeryStrong:
		return "Very Strong"
	default:
		return "Unknown"
	}
}

func AnalyzePassword(password string) StrengthResult {
	var result StrengthResult
	var suggestions []string

	length := len(password)
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	hasSpace := false
	repeatCount := 0
	prevChar := rune(0)

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
		if unicode.IsSpace(char) {
			hasSpace = true
		}
		if char == prevChar {
			repeatCount++
		}
		prevChar = char
	}

	score := 0.0

	score += math.Min(float64(length)*4, 40)
	if hasUpper {
		score += 10
	} else {
		suggestions = append(suggestions, "Add uppercase letters")
	}
	if hasLower {
		score += 10
	} else {
		suggestions = append(suggestions, "Add lowercase letters")
	}
	if hasNumber {
		score += 10
	} else {
		suggestions = append(suggestions, "Add numbers")
	}
	if hasSpecial {
		score += 15
	} else {
		suggestions = append(suggestions, "Add special characters")
	}
	if hasSpace {
		score += 5
	}

	score -= float64(repeatCount) * 2

	if length < 8 {
		suggestions = append(suggestions, "Make the password longer (at least 8 characters)")
	}

	commonPatterns := []string{"123", "abc", "qwerty", "password", "admin"}
	for _, pattern := range commonPatterns {
		if strings.Contains(strings.ToLower(password), pattern) {
			score -= 10
			suggestions = append(suggestions, "Avoid common patterns like '"+pattern+"'")
		}
	}

	result.Score = math.Max(0, math.Min(100, score))
	result.Suggestions = suggestions

	switch {
	case result.Score >= 80:
		result.Level = VeryStrong
	case result.Score >= 60:
		result.Level = Strong
	case result.Score >= 40:
		result.Level = Medium
	case result.Score >= 20:
		result.Level = Weak
	default:
		result.Level = VeryWeak
	}

	return result
}
