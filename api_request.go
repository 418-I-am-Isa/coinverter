package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	baseURL := "https://api.freecurrencyapi.com/v1/latest"
	apiKey := os.Getenv("API_KEY")
	baseCurrency := "EUR"
	targetCurrencies := []string{"USD", "AUD"}

	apiKeyParam := "apikey=" + apiKey
	baseCurrencyParam := "base_currency=" + baseCurrency
	currenciesParam := "currencies=" + strings.Join(targetCurrencies, ",")
	url := baseURL + "?" + strings.Join([]string{apiKeyParam, baseCurrencyParam, currenciesParam}, "&")

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", url, err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", data)
}
