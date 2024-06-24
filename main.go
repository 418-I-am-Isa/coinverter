package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	tea "github.com/charmbracelet/bubbletea"
)

 
type model struct {
	cursor   int
	choices  []string
	selected map[int]struct{}
}

type Currency struct {
    Symbol        string `json:"symbol"`
    Name          string `json:"name"`
    SymbolNative  string `json:"symbol_native"`
    DecimalDigits int    `json:"decimal_digits"`
    Rounding      int    `json:"rounding"`
    Code          string `json:"code"`
    NamePlural    string `json:"name_plural"`
}

type Response struct {
    Data map[string]Currency `json:"data"`
}

func getCurrencies()([]string){
	baseURL := "https://api.freecurrencyapi.com/v1/currencies"
	apiKey := os.Getenv("API_KEY")
	if len(apiKey) == 0 {
		fmt.Fprintln(os.Stderr, "Error. No API key found in environment valiables. Set it in the API_KEY variable.")
		os.Exit(1)
	}
	url := baseURL + "?" + "apikey=" + apiKey
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if status := resp.StatusCode; status != 200 {
		fmt.Fprintf(os.Stderr, "HTTP response code: %d\n", status)
		os.Exit(1)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", url, err)
		os.Exit(1)
	}

	var response Response
	if err = json.Unmarshal(data, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling %s: %v\n", url, err)
		os.Exit(1)
	}
	choices := make([]string, 0, len(response.Data))
	for key, currency := range response.Data {
		choices = append(choices, key + " ("+ currency.Name + ")")
    }
	sort.Strings(choices)
	return choices
}

func initialModel() model {
	return model{
		choices: getCurrencies(),

		// A map which indicates which choices are selected. We're using
		// the map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Currency List")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Select the currencies\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

