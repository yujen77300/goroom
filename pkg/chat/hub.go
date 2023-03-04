package chat

import "fmt"

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// 定義 Factory function當作constructor
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		// 監聽 register、unregister 和 broadcast 三個 channel，然後有不同動作
		select {
		case client := <-h.register:
			fmt.Println("進來註冊")
			fmt.Println(client.Send)
			fmt.Println(*client)
			h.clients[client] = true
		case client := <-h.unregister:
			fmt.Println("進來解除註冊")
			fmt.Println(client.Send)
			fmt.Println(*client)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case message := <-h.broadcast:
			fmt.Println("進來廣播")
			fmt.Println(string(message))
			// 這裡先
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
