package style

import (
	"image/color"

	"charm.land/fang/v2"
	lipglossv2 "charm.land/lipgloss/v2"
)

const (
	PurpleLight = "#6F5AF2"
	PurpleDark  = "#8B7CFF"
	TealLight   = "#1AAE9F"
	TealDark    = "#39D0BF"
	GoldLight   = "#B7791F"
	GoldDark    = "#F6C453"
	InkLight    = "#1D1B29"
	InkDark     = "#F4F0FF"
	MutedLight  = "#6B7280"
	MutedDark   = "#9AA4B2"
	BorderLight = "#DDD6FE"
	BorderDark  = "#433D68"
	PanelLight  = "#F7F4FF"
	PanelDark   = "#1D1830"
	ErrorLight  = "#C2410C"
	ErrorDark   = "#FF8A5B"
)

func hex(v string) color.Color {
	return lipglossv2.Color(v)
}

func pickColor(c lipglossv2.LightDarkFunc, light, dark string) color.Color {
	return c(hex(light), hex(dark))
}

// FangColorScheme returns a terminal-safe Kagi-inspired palette for cobra help and errors.
func FangColorScheme(c lipglossv2.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           pickColor(c, InkLight, InkDark),
		Title:          pickColor(c, PurpleLight, PurpleDark),
		Description:    pickColor(c, InkLight, InkDark),
		Codeblock:      pickColor(c, PanelLight, PanelDark),
		Program:        pickColor(c, PurpleLight, PurpleDark),
		DimmedArgument: pickColor(c, MutedLight, MutedDark),
		Comment:        pickColor(c, MutedLight, MutedDark),
		Flag:           pickColor(c, TealLight, TealDark),
		FlagDefault:    pickColor(c, GoldLight, GoldDark),
		Command:        pickColor(c, PurpleLight, PurpleDark),
		QuotedString:   pickColor(c, TealLight, TealDark),
		Argument:       pickColor(c, InkLight, InkDark),
		Help:           pickColor(c, MutedLight, MutedDark),
		Dash:           pickColor(c, MutedLight, MutedDark),
		ErrorHeader: [2]color.Color{
			pickColor(c, PanelLight, PanelDark),
			pickColor(c, ErrorLight, ErrorDark),
		},
		ErrorDetails: pickColor(c, ErrorLight, ErrorDark),
	}
}
