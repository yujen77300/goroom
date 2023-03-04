package webrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/pion/webrtc/v3"
)

var Testuser string

func RoomConn(c *websocket.Conn, p *Peers) {
	var config webrtc.Configuration
	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		config = turnConfig
	}
	// 建立peerconnection，然後正式區的時候建立stun伺服器
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Print(err)
		return
	}

	//建立datachannel==========================
	// type Message struct {
	// 	Email string `json:"email"`
	// }

	// //建立datachannel看看
	// peerConnection.OnDataChannel(func(datachannel *webrtc.DataChannel) {
	// 	if datachannel.Label() == "mydatachannel" {
	// 		datachannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 			var message Message
	// 			err := json.Unmarshal(msg.Data, &message)
	// 			if err != nil {
	// 				fmt.Println("解析 JSON 失敗：", err)
	// 				return
	// 			}
	// 			fmt.Println("收到的資料：", message.Email)
	// 			datachannel.SendText(message.Email)
	// 			// data := string(msg.Data)
	// 			// fmt.Println("收到的資料")
	// 			// fmt.Println(data)
	// 		})
	// 	}
	// })
	//==========================
	defer peerConnection.Close()
	// 決定接受傳入的stream類型
	for _, codecType := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		// 定義Transceiver，因後面方向定義為recvonly， transceiver 僅用於接收音頻和視頻數據
		if _, err := peerConnection.AddTransceiverFromKind(codecType, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Print(err)
			return
		}
	}
	// PeerConnection物件和WebSocket connection物件相關聯，以便將資料送到WebSocket connection。
	newPeer := PeerConnectionState{
		PeerConnection: peerConnection,
		Websocket: &ThreadSafeWriter{
			Conn:  c,
			Mutex: sync.Mutex{},
		}}

	p.ListLock.Lock()
	// 把新的PeerConnection加入PeerConnectionState的slice中
	p.Connections = append(p.Connections, newPeer)
	p.ListLock.Unlock()
	log.Println(p.Connections)

	// 如果新的ICECandidate訊息收集完成時的操作
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		fmt.Println(string(candidateString))
		if err != nil {
			log.Println(err)
			return
		}

		// 將WebRTC的 ICE candidate傳送給 WebSocket connection。peer.js onmessage
		if writeErr := newPeer.Websocket.WriteJSON(&websocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			log.Println(writeErr)
		}
	})

	// 當peerconnection的狀態改變，通常就是自global list移除。
	peerConnection.OnConnectionStateChange(func(peerConState webrtc.PeerConnectionState) {
		fmt.Println("PeerConnectionState連接狀態")
		fmt.Println(peerConState)
		switch peerConState {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Print(err)
			}
		case webrtc.PeerConnectionStateClosed:
			fmt.Println("連接失敗")
			p.SignalPeerConnections()
		case webrtc.PeerConnectionStateConnected:
			fmt.Println("連接成功")
		}
	})

	// OnTrack事件觸發，代表有新的媒體流t
	//trackremote是別人給本地的track
	//tracklocal是本地給別人的track
	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		// 創建一個新的trackLocal，並將其與 t（即收到的遠端追蹤）相關聯，這樣收到的媒體流才可以廣播給peer

		trackLocal := p.AddTrack(t)
		if trackLocal == nil {
			return
		}
		defer p.RemoveTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			if err != nil {
				return
			}
			// 追蹤t然後寫入trackLocal
			if _, err = trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})

	p.SignalPeerConnections()
	message := &websocketMessage{}
	// 用for來循環websocket的訊息
	for {
		_, raw, err := c.ReadMessage()
		// fmt.Println("近來循環ws訊息的迴圈開始")
		// fmt.Println(string(raw))
		// fmt.Println(err)
		// fmt.Println("近來循環ws訊息的迴圈結束")
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}
		// fmt.Println("for迴圈裏面印出message開始")
		// fmt.Println(message)
		// fmt.Println(message.Event)
		// fmt.Println(message.Data)
		// fmt.Println("for迴圈裏面印出message結束")

		switch message.Event {
		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			// 將message的Data字段轉換為ICECandidateInit結構體，然後使用peerConnection.AddICECandidate()方法將這個ICECandidate加入到peerConnection中
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
			// 將message的Data字段轉換為SessionDescription結構體，然後使用peerConnection.SetRemoteDescription()方法將這個SessionDescription設置為遠端Session描述。
		case "answer":
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
