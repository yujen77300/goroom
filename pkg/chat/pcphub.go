package chat



type PcpHub struct {
	clients    map[*PcpClient]bool
	broadcast  chan []byte
	register   chan *PcpClient
	unregister chan *PcpClient
}

func NewPcpHub() *PcpHub {
	return &PcpHub{
		clients:    make(map[*PcpClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *PcpClient),
		unregister: make(chan *PcpClient),
	}
}

func (h *PcpHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case message := <-h.broadcast:
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