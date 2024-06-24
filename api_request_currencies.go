package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"encoding/json"

)

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


func main(){
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
	choices := make([]string, len(response.Data))
	for key, currency := range response.Data {
		choices = append(choices, key + "("+ currency.Name + ")")
    }
	fmt.Printf("%s",choices)
}
