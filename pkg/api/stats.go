package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type stockInfo struct {
	Currency, Description, DisplaySymbol, Figi, Mic, Symbol, Type string
}

type stockOverview struct {
	Symbol, Name, Exchange                                               string
	Price, ChangesPercentage, Change, DayLow, DayHigh, YearHigh, YearLow float32
	MarketCap, PriceAvg50, PriceAvg100, Open, PreviousClose, Eps, Pe     float32
	Volume, AvgVolume, SharesOutstanding, Timestamp                      int
	// earningsAnnouncement                                                 time.Time
}

//GetSymbolsList returns list of all registered stocks in wanted exchange.
func GetSymbolsList(exchange, apiKey string) ([]string, error) {
	reqFullPath := fmt.Sprintf("https://finnhub.io/api/v1/stock/symbol?exchange=%s&token=%s", exchange, apiKey)

	client := http.Client{
		Timeout: time.Minute,
	}

	request, err := http.NewRequest("GET", reqFullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("GetSymbolsList: Error, got '%v'", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GetSymbolsList: Error, got '%v'", err)
	}

	//response takes time to arrive, we will read each object of it's array using json Decoder to decode the whole stream of data
	jsonDecoder := json.NewDecoder(resp.Body)
	// read open bracket (this API service returns an array of JSON objects, not just one - begoin with '[')
	_, err = jsonDecoder.Token()
	if err != nil {
		return nil, fmt.Errorf("GetSymbolsList: Error, got '%v'", err)
	}

	symbolsList := make([]string, 0)

	// while the array contains values
	for jsonDecoder.More() {
		var stockInfo stockInfo
		// decode an array value (Message)
		err := jsonDecoder.Decode(&stockInfo)
		if err != nil {
			return nil, fmt.Errorf("GetSymbolsList: Error, got '%v'", err)
		}

		symbolsList = append(symbolsList, stockInfo.Symbol)
	}

	return symbolsList, err
}

//250 req per day, 10 per minute
func GetStocksOverview(stocks []string, apiKey string) {
	var stocksToCheck string
	for i, s := range stocks {
		if i == 1000 {
			break
		}

		if i == 0 {
			stocksToCheck = s
		} else {
			stocksToCheck = fmt.Sprintf("%s,%s", stocksToCheck, s)

		}
	}

	reqFullPath := fmt.Sprintf("https://financialmodelingprep.com/api/v3/quote/%s?apikey=%s", stocksToCheck, apiKey)

	client := http.Client{
		Timeout: time.Minute,
	}
	resp, _ := client.Get(reqFullPath)
	//response takes time to arrive, we will read each object of it's array using json Decoder to decode the whole stream of data
	jsonDecoder := json.NewDecoder(resp.Body)
	// read open bracket (this API service returns an array of JSON objects, not just one - begoin with '[')
	jsonDecoder.Token()

	symbolsList := make([]string, 0)

	// while the array contains values
	for jsonDecoder.More() {
		var stock stockOverview
		// decode an array value (Message)
		err := jsonDecoder.Decode(&stock)
		if err != nil {
			log.Print(err)
		}
		log.Print(stock)
		symbolsList = append(symbolsList, stock.Symbol)
	}

	log.Print(resp)
}

//IsInteresting returns true if:
//1. stock's Price-to-earnings ratio <= 15
//2. stock's last price is lower than 85% of it's last 200 days average price
//3. stock's last price is lower than 90% of it's last 50 days average price
//4. stock's last price is lower than 90% of analyst's target price
func IsInteresting(symbol, apiKey string) (bool, error) {
	reqFullPath := fmt.Sprintf("https://www.alphavantage.co/query?function=OVERVIEW&apikey=%s&symbol=%s", apiKey, symbol)
	resp, err := http.Get(reqFullPath)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error, got '%v'", err)
	}

	defer resp.Body.Close()

	parsedStockOverview := make(map[string]string)

	bodyJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error, got '%v'", err)
	}

	if err := json.Unmarshal(bodyJSON, &parsedStockOverview); err != nil {
		return false, fmt.Errorf("IsInteresting: Error, got '%v'", err)
	}

	lastPrice, err := GetLastPrice(symbol, apiKey)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error, got '%v'", err)
	}

	if err := json.Unmarshal(bodyJSON, &parsedStockOverview); err != nil {
		return false, fmt.Errorf("IsInteresting: Error, got '%v'", err)
	}

	//Price-to-earnings ratio
	peRatio, err := strconv.ParseFloat(parsedStockOverview["PERatio"], 32)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error parsing 'PEratio', got '%v'", err)
	}

	last200DaysAvgPrice, err := strconv.ParseFloat(parsedStockOverview["200DayMovingAverage"], 32)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error parsing '200DayMovingAverage', got '%v'", err)
	}

	last50DaysAvgPrice, err := strconv.ParseFloat(parsedStockOverview["50DayMovingAverage"], 32)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error parsing '50DayMovingAverage', got '%v'", err)
	}

	analystTargetPrice, err := strconv.ParseFloat(parsedStockOverview["AnalystTargetPrice"], 32)
	if err != nil {
		return false, fmt.Errorf("IsInteresting: Error parsing 'AnalystTargetPrice', got '%v'", err)
	}

	if peRatio > 15 {
		return false, nil
	}

	if lastPrice <= (last200DaysAvgPrice*0.85) && lastPrice <= (last50DaysAvgPrice*0.9) && lastPrice <= (analystTargetPrice*0.9) {
		return true, nil
	}

	return false, nil
}

//GetLastPrice returns stock's last registered price
func GetLastPrice(symbol, apiKey string) (float64, error) {
	interval := "1min"
	reqFullPath := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&apikey=%s&symbol=%s&interval=%s", apiKey, symbol, interval)
	resp, err := http.Get(reqFullPath)
	if err != nil {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	bodyJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	parsedBody := make(map[string]interface{})

	if err := json.Unmarshal(bodyJSON, &parsedBody); err != nil {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	prices, ok := parsedBody["Time Series (1min)"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	times := make([]string, 0)

	for k := range prices {
		times = append(times, k)
	}

	//sort from earliest date and time to latest
	sort.Strings(times)

	lastTimeUpdate := times[len(times)-1]
	lastUpdateStats, ok := prices[lastTimeUpdate].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	lastPrice, ok := lastUpdateStats["4. close"].(string)
	if !ok {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	lastPriceFloat, err := strconv.ParseFloat(lastPrice, 32)
	if err != nil {
		return 0, fmt.Errorf("GetLastPrice: Error, got '%v'", err)
	}

	return lastPriceFloat, nil
}
