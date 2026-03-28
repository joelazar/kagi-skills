package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ResultItem represents a single result in the result list.
type ResultItem struct {
	Title       string
	URL         string
	Description string
	Detail      string // Full detail content (markdown).
}

// FilterValue implements list.Item.
func (r ResultItem) FilterValue() string { return r.Title + " " + r.Description }

// resultDelegate renders result items.
type resultDelegate struct{}

func (d resultDelegate) Height() int                             { return 3 }
func (d resultDelegate) Spacing() int                            { return 1 }
func (d resultDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d resultDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ResultItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	var title, url, desc string
	if isSelected {
		title = SelectedItemStyle.Render("▸ " + item.Title)
		url = URLStyle.Render("  " + item.URL)
		desc = DimStyle.Render("  " + trimToLen(item.Description, 80))
	} else {
		title = NormalItemStyle.Render("  " + item.Title)
		url = DimStyle.Render("  " + item.URL)
		desc = DimStyle.Render("  " + trimToLen(item.Description, 80))
	}

	if item.URL != "" {
		fmt.Fprintf(w, "%s\n%s\n%s", title, url, desc)
	} else {
		fmt.Fprintf(w, "%s\n%s", title, desc)
	}
}

func trimToLen(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit-1]) + "…"
}

// newResultList creates a list.Model configured for displaying results.
func newResultList(title string, items []ResultItem, width, height int) list.Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	l := list.New(listItems, resultDelegate{}, width, height)
	l.Title = title
	l.Styles.Title = TitleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return l
}
