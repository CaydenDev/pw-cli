package ui

import "fmt"

type Color string

const (
	Black   Color = "\u001b[30m"
	Red     Color = "\u001b[31m"
	Green   Color = "\u001b[32m"
	Yellow  Color = "\u001b[33m"
	Blue    Color = "\u001b[34m"
	Magenta Color = "\u001b[35m"
	Cyan    Color = "\u001b[36m"
	White   Color = "\u001b[37m"
	Reset   Color = "\u001b[0m"
)

type Style struct {
	Bold      bool
	Underline bool
	Color     Color
}

type Theme struct {
	HeaderStyle    Style
	PromptStyle    Style
	ErrorStyle     Style
	SuccessStyle   Style
	WarningStyle   Style
	HighlightStyle Style
	MenuItemStyle  Style
	SelectedStyle  Style
	SeparatorStyle Style
	PasswordStyle  Style
}

var themes = map[string]Theme{
	"default": {
		HeaderStyle:    Style{Bold: true, Color: Blue},
		PromptStyle:    Style{Color: Green},
		ErrorStyle:     Style{Bold: true, Color: Red},
		SuccessStyle:   Style{Color: Green},
		WarningStyle:   Style{Color: Yellow},
		HighlightStyle: Style{Bold: true, Color: Cyan},
		MenuItemStyle:  Style{Color: White},
		SelectedStyle:  Style{Bold: true, Color: Cyan},
		SeparatorStyle: Style{Color: White},
		PasswordStyle:  Style{Color: Magenta},
	},
	"dark": {
		HeaderStyle:    Style{Bold: true, Color: Magenta},
		PromptStyle:    Style{Color: Cyan},
		ErrorStyle:     Style{Bold: true, Color: Red},
		SuccessStyle:   Style{Color: Green},
		WarningStyle:   Style{Color: Yellow},
		HighlightStyle: Style{Bold: true, Color: Blue},
		MenuItemStyle:  Style{Color: White},
		SelectedStyle:  Style{Bold: true, Color: Magenta},
		SeparatorStyle: Style{Color: White},
		PasswordStyle:  Style{Color: Cyan},
	},
	"light": {
		HeaderStyle:    Style{Bold: true, Color: Blue},
		PromptStyle:    Style{Color: Green},
		ErrorStyle:     Style{Bold: true, Color: Red},
		SuccessStyle:   Style{Color: Green},
		WarningStyle:   Style{Color: Yellow},
		HighlightStyle: Style{Bold: true, Color: Magenta},
		MenuItemStyle:  Style{Color: Black},
		SelectedStyle:  Style{Bold: true, Color: Blue},
		SeparatorStyle: Style{Color: Black},
		PasswordStyle:  Style{Color: Blue},
	},
}

func (s Style) Apply(text string) string {
	result := string(s.Color)

	if s.Bold {
		result += "\u001b[1m"
	}
	if s.Underline {
		result += "\u001b[4m"
	}

	return fmt.Sprintf("%s%s%s", result, text, Reset)
}

func GetTheme(name string) Theme {
	if theme, exists := themes[name]; exists {
		return theme
	}
	return themes["default"]
}

func ListThemes() []string {
	themeList := make([]string, 0, len(themes))
	for name := range themes {
		themeList = append(themeList, name)
	}
	return themeList
}
