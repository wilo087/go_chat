package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/wilo087/go_chat/internal/stock"
	"github.com/wilo087/go_chat/package/models"
)

type Subscriber interface {
	SendMessage(msg []byte)
}

type WebSocketSubscriber struct {
	msgs chan []byte
}

func NewWebSocketSubscriber(bufferSize int) *WebSocketSubscriber {
	return &WebSocketSubscriber{
		msgs: make(chan []byte, bufferSize),
	}
}

func (sub *WebSocketSubscriber) SendMessage(msg []byte) {
	sub.msgs <- msg
}

type Publisher interface {
	PublishMsg(msg []byte)
	AddSubscriber(sub Subscriber)
}

// WebSocket suscriber
func (srv *Server) Subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	defer ws.CloseNow()

	// Create new WebSocket subscriber
	sub := NewWebSocketSubscriber(srv.subscriberMessageBuffer)
	srv.AddSubscriber(sub)

	// Initialize the ws connection
	go srv.handleIncomingMessages(ctx, ws)

	// send messages to the client
	return srv.sendMessagesToClient(ctx, ws, sub)
}

func (srv *Server) handleIncomingMessages(ctx context.Context, ws *websocket.Conn) {
	defer ws.Close(websocket.StatusNormalClosure, "Normal closure")

	for {
		msgType, msg, err := ws.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || ctx.Err() != nil {
				fmt.Println("WebSocket closed:", err)
				break
			}

			// Close the WebSocket connection if an error occurs
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				fmt.Println("Unexpected close error:", err)
				ws.Close(websocket.StatusInternalError, "A client error occurred")
			} else {
				fmt.Println("Error reading message:", err)
			}

			// Close the WebSocket connection if an error occurs
			if err := ws.Close(websocket.StatusInternalError, "Internal server error"); err != nil {
				fmt.Println("Error closing WebSocket:", err)
			}

			break
		}

		// Only process text messages
		if msgType == websocket.MessageText {
			srv.processClientMessage(msg, ws)
		}
	}
}

// Process client message
func (srv *Server) processClientMessage(msg []byte, ws *websocket.Conn) {
	var message models.Message

	if err := json.Unmarshal(msg, &message); err != nil {
		fmt.Println("Error unmarshaling message:", err)
		return
	}

	message.Time = time.Now()
	newMsg, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}

	if strings.HasPrefix(message.Content, "/stock=") {
		stockSymbol := message.Content[7:]
		go stock.FetchStockPrice(stockSymbol, ws, srv.BroadcastMsg)

	} else {
		srv.BroadcastMsg(newMsg)
	}
}

func (srv *Server) sendMessagesToClient(ctx context.Context, ws *websocket.Conn, sub *WebSocketSubscriber) error {
	for {
		select {
		case msg := <-sub.msgs:
			if err := srv.writeToWebSocket(ctx, ws, msg); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (srv *Server) writeToWebSocket(ctx context.Context, ws *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := ws.Write(ctx, websocket.MessageText, msg); err != nil {
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure || ctx.Err() != nil {
			fmt.Println("WebSocket write error:", err)
			return err
		}
		return err
	}
	return nil
}

// Broadcast a message to all subscribers
func (srv *Server) BroadcastMsg(msg []byte) {
	srv.subscribersMu.Lock()
	defer srv.subscribersMu.Unlock()

	// Unmarshal the incoming message to models.Message
	var message models.Message
	if err := json.Unmarshal(msg, &message); err != nil {
		log.Println("Error unmarshaling message:", err)
		return
	}

	// Add the new message to the slice
	srv.messages = append(srv.messages, message)

	// Limit the number of messages to the last 50
	if len(srv.messages) > srv.maxMessages {
		srv.messages = srv.messages[len(srv.messages)-srv.maxMessages:]
	}

	// Sort messages by timestamp
	sort.Slice(srv.messages, func(i, j int) bool {
		return srv.messages[i].Time.Before(srv.messages[j].Time)
	})

	// Send the last 50 messages to all subscribers
	for sub := range srv.subscribers {
		sub.SendMessage(msg)
	}
}

// Add a new suscriber to the server
func (s *Server) AddSubscriber(sub Subscriber) {
	s.subscribersMu.Lock()
	defer s.subscribersMu.Unlock()

	s.subscribers[sub] = struct{}{}
	fmt.Println("Added subscriber", sub)
}

func (srv *Server) broadcastBotMessage() {
	for {
		message := models.Message{
			Username: "system",
			Content:  "This is a system broadcast message.",
			Time:     time.Now(),
		}

		// Serializar el mensaje para enviarlo a todos los suscriptores
		msgBytes, err := json.Marshal(message)
		if err != nil {
			log.Println("Error serializing message:", err)
			continue
		}

		// broadcast the message
		srv.BroadcastMsg(msgBytes)

		// Waits for 60 seconds before sending the next message
		time.Sleep(60 * time.Second)
	}
}
