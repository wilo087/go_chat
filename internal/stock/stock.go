package stock

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/wilo087/go_chat/package/models"
)

func FetchStockPrice(stockCode string, ws *websocket.Conn, broadcastFunc func([]byte)) {
	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching stock price: %v", err)
		return
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		log.Printf("Error parsing CSV: %v", err)
		return
	}

	stockInfo := records[1]
	if len(stockInfo) < 4 {
		log.Printf("Invalid stock information received: %v", stockInfo)
		return
	}

	st := fmt.Sprintf("%s quote is $%s per share", stockCode, stockInfo[3])
	if stockInfo[3] == "N/D" {
		st = fmt.Sprintf("Stock code %s not found", stockCode)
	}

	message := models.Message{
		Username: "Bot",
		Content:  st,
		Time:     time.Now(),
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error serializing message:", err)
		return
	}

	// Broadcast the message
	broadcastFunc(msgBytes)
}
