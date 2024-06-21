package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"text/template"
)

type Ticker struct {
	Symbol string `json:"symbol"`
	Volume string `json:"volume"`
	Price  string `json:"lastPrice"`
}

const migrationTemplate = `-- +goose Up
{{- range . }}
INSERT INTO ticker (symbol) VALUES ('{{ .Symbol }}');
{{- end }}

-- +goose Down
{{- range . }}
DELETE FROM ticker WHERE symbol = '{{ .Symbol }}';
{{- end }}
`

func main() {
	endpoint := "https://fapi.binance.com/fapi/v1/ticker/24hr"

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Error, http status code: ", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var tickers []Ticker
	if err := json.Unmarshal(body, &tickers); err != nil {
		log.Fatal("Error unamrshaling json: ", err)
	}

	// Sort slice by volume
	sort.Slice(tickers, func(i, j int) bool {
		vol1, _ := strconv.ParseFloat(tickers[i].Volume, 64)
		vol2, _ := strconv.ParseFloat(tickers[j].Volume, 64)
		price1, _ := strconv.ParseFloat(tickers[i].Price, 64)
		price2, _ := strconv.ParseFloat(tickers[j].Price, 64)
		return vol1*price1 > vol2*price2
	})
	ticker := tickers[240]
	log.Println(ticker.Volume, ticker.Symbol)

	f, err := os.Create("db/migrations/20240614114757_insert_binance_futures_tickers.sql")
	if err != nil {
		log.Fatalf("Failed to create migration file: %v", err)
	}
	defer f.Close()

	tmpl, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	err = tmpl.Execute(f, tickers)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println("Migration file created successfully")

}
