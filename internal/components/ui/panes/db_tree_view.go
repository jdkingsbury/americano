package panes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jdkingsbury/americano/internal/drivers"
)

type DBTreeMsg struct {
  Notification string
  Error error
}

type ListItem struct {
	Title    string
	SubItems []ListItem
	IsOpen   bool
	Query    string
}

// FlatListItem is used for the rendering the list items
type FlatListItem struct {
	Title     string
	Level     int
	IsOpen    bool
	IsSubItem bool
}

type DBTreeModel struct {
	originalList []ListItem
	flatList     []FlatListItem
	cursor       int
}

func NewDBTreeModel(db drivers.Database) *DBTreeModel {
	var originalList []ListItem

	if db == nil {
		originalList = []ListItem{
			{Title: "No connection"},
		}
	} else {
		tables, err := db.GetTables()
		if err != nil {
			originalList = []ListItem{
				{Title: "No connection"},
			}
      // return DBTreeMsg{Error: fmt.Errorf("Error retrieving tables: %s", err.Error())
		} else {
			originalList = buildTableList(tables)
		}
	}

	flatList := flattenList(originalList, 0)

	return &DBTreeModel{
		originalList: originalList,
		flatList:     flatList,
		cursor:       0,
	}
}

func buildTableList(tables []string) []ListItem {
	var tableItems []ListItem
	for _, table := range tables {
		tableItems = append(tableItems, ListItem{Title: table})
	}

	return tableItems
}

func sampleList() []ListItem {
	return []ListItem{
		{Title: "Item 1", SubItems: []ListItem{
			{Title: "SubItem 1.1"},
			{Title: "SubItem 1.2"},
		}},
		{Title: "Item 2", SubItems: []ListItem{
			{Title: "SubItem 2.1", SubItems: []ListItem{
				{Title: "SubItem 2.1.1"},
			}},
		}},
		{Title: "Item 3"},
	}
}

func flattenList(items []ListItem, level int) []FlatListItem {
	var flatList []FlatListItem
	for _, item := range items {
		flatItem := FlatListItem{
			Title:     item.Title,
			Level:     level,
			IsOpen:    item.IsOpen,
			IsSubItem: level > 0,
		}
		flatList = append(flatList, flatItem)

		// If the item is open and has subitems, recursively flatten the subitems
		if item.IsOpen && len(item.SubItems) > 0 {
			flatList = append(flatList, flattenList(item.SubItems, level+1)...)
		}
	}
	return flatList
}

// Handle key inputs for navigation and toggling items
func handleInput(msg tea.KeyMsg, m *DBTreeModel) {
	switch msg.String() {

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.flatList)-1 {
			m.cursor++
		}

	case "enter", " ":
		m.toggleItemOpen()
	}
}

// Toggles and item's open/collapse state and rebuilds the flat list
func (m *DBTreeModel) toggleItemOpen() {
	// Find the item in the original list
	m.updateOriginalListState(m.originalList, m.flatList[m.cursor].Title, 0, m.flatList[m.cursor].Level)
	// Rebuild the flat list based on the updated original list
	m.flatList = flattenList(m.originalList, 0)
}

// Update the collapsible state and return true or false if found. Returning a bool for when making tests
func (m *DBTreeModel) updateOriginalListState(items []ListItem, title string, currentLevel, targetLevel int) bool {
	for i := range items {
		// Check if this is the correct item based on the title and level
		if items[i].Title == title && currentLevel == targetLevel {
			// Toggle open state
			items[i].IsOpen = !items[i].IsOpen
			return true // Exit after toggling
		}

		// Recursively check subitems if they exist
		if len(items[i].SubItems) > 0 {
			// If the item is found in the sublist, return true to stop recursion
			if m.updateOriginalListState(items[i].SubItems, title, currentLevel+1, targetLevel) {
				return true
			}
		}
	}

	return false // Return false if the item was not found in the branch
}

func renderFlatList(flatList []FlatListItem, cursor int) string {
	var b strings.Builder

	title := listTitleStyle.Render("Database Connection Tree")
	b.WriteString(fmt.Sprintf("%s\n", title))

	for i, item := range flatList {
		indent := strings.Repeat("  ", item.Level)
		if i == cursor {
			b.WriteString(fmt.Sprintf("%s> %s\n", indent, listSelectedItemStyle.Render(item.Title))) // Highlight the selected item
		} else {
			b.WriteString(fmt.Sprintf("%s  %s\n", indent, listItemStyle.Render(item.Title))) // Normal item
		}
	}

	return b.String()
}

func (m *DBTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		handleInput(msg, m)
	}
	return m, nil
}

func (m *DBTreeModel) Init() tea.Cmd {
	return nil
}

func (m *DBTreeModel) View() string {
	return renderFlatList(m.flatList, m.cursor)
}
