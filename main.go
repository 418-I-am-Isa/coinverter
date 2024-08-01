package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)
type model struct {
	cursor       int
	choices      []string
	selected     map[int]struct{}
	baseCurrency int
	textInput    textinput.Model
	err          error
}

type CurrencyList struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	SymbolNative  string `json:"symbol_native"`
	DecimalDigits int    `json:"decimal_digits"`
	Rounding      int    `json:"rounding"`
	Code          string `json:"code"`
	NamePlural    string `json:"name_plural"`
}

type CurrencyListResponse struct {
	Data map[string]CurrencyList `json:"data"`
}

type ConversionResponse struct {
	Data map[string]float64 `json:"data"`
}

func getCurrencies() []string {
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var response CurrencyListResponse
	if err = json.Unmarshal(data, &response); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling %s: %v\n", url, err)
		os.Exit(1)
	}
	choices := make([]string, 0, len(response.Data))
	for key, currency := range response.Data {
		choices = append(choices, key+" ("+currency.Name+")")
	}
	sort.Strings(choices)
	return choices
}

func conversion(baseCurrency string, targetCurrencies []string, baseQuantity float64) {
	baseURL := "https://api.freecurrencyapi.com/v1/latest"
	apiKey := os.Getenv("API_KEY")
	if len(apiKey) == 0 {
		fmt.Fprintln(os.Stderr, "Error. No API key found in environment valiables. Set it in the API_KEY variable.")
		os.Exit(1)
	}

	apiKeyParam := "apikey=" + apiKey
	baseCurrencyParam := "base_currency=" + baseCurrency
	currenciesParam := "currencies=" + strings.Join(targetCurrencies, ",")
	url := baseURL + "?" + strings.Join([]string{apiKeyParam, baseCurrencyParam, currenciesParam}, "&")

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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, currency := range targetCurrencies {
		var response ConversionResponse
		if err = json.Unmarshal(data, &response); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling %s: %v\n", url, err)
			os.Exit(1)
		}

		value := baseQuantity * response.Data[currency]

		fmt.Printf("%f %s = %f %s\n", baseQuantity, baseCurrency, value, currency)
	}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "1"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return model{
		choices:      getCurrencies(),
		baseCurrency: -1,
		selected:     make(map[int]struct{}),
		textInput:    ti,
		err:          nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			}
			m.baseCurrency = m.cursor
		}
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := `* Please select the base currency with the Enter key
* To select the target currencies use the Space bar
* To exit the program and process the conversion press q

`

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if m.baseCurrency == i {
			checked = "•"
		} else if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPlease enter the amound of money you want to convert: \n"

	s += m.textInput.View()

	s += "\nPress q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	if m, ok := m.(model); ok && m.baseCurrency != -1 {
		baseCurrency := m.choices[m.baseCurrency]
		baseCurrency = strings.Split(baseCurrency, " ")[0]
		targetCurrencies := make([]string, 0, len(m.selected))
		for idx := range m.selected {
			target := m.choices[idx]
			target = strings.Split(target, " ")[0]
			targetCurrencies = append(targetCurrencies, target)
		}
		sort.Strings(targetCurrencies)

		re := regexp.MustCompile(`\d+`)
		strQuantity := strings.Join(re.FindAllString(m.textInput.View(), -1), "")
		baseQuantity, err := strconv.ParseFloat(strQuantity, 64)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error Converting to float %s: %v\n", m.textInput.View(), err)
			os.Exit(1)
		}
		fmt.Printf("Text Input: %s\n", m.textInput.View())
		fmt.Printf("strQuantity: %s\n", strQuantity)
		fmt.Printf("baseQuantity: %f\n", baseQuantity)
		conversion(baseCurrency, targetCurrencies, baseQuantity)
	}
}
