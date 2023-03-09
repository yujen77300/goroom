package webrtc

import (
	"encoding/json"
	// "fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"

	"github.com/yujen77300/goroom/pkg/chat"
)

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

var (
	RoomsLock sync.RWMutex
	Rooms     map[string]*Room
)

var (
	PcpRoomsLock sync.RWMutex
	PcpRooms     map[string]*PcpRoom
)

var (
	turnConfig = webrtc.Configuration{
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
		ICEServers: []webrtc.ICEServer{
			{

				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{

				URLs: []string{"turn:54.150.244.240:3478"},

				Username: "Dylan",

				Credential:     "Wehelp",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
		// ICEServers: []webrtc.ICEServer{
		// 	// {

		// 	// 	URLs: []string{"turn:relay.metered.ca:443?transport=tcp"},

		// 	// 	Username: "",

		// 	// 	Credential:     "",
		// 	// 	CredentialType: webrtc.ICECredentialTypePassword,
		// 	// },
		// },
		// ICEServers: []webrtc.ICEServer{
		// 	{
		// 		URLs: []string{"stun:goroom.online:3478"},
		// 	},
		// 	{

		// 		URLs: []string{"turn:goroom.online:3478"},

		// 		Username: "",

		// 		Credential:     "",
		// 		CredentialType: webrtc.ICECredentialTypePassword,
		// 	},
		// 	{
		// 		URLs: []string{"stun:relay.metered.ca:80"},
		// 	},
		// 	{
		// 		URLs: []string{"turn:relay.metered.ca:80"},

		// 		Username: "",

		// 		Credential:     "",
		// 		CredentialType: webrtc.ICECredentialTypePassword,
		// 	},
		// 	{

		// 		URLs: []string{"turn:relay.metered.ca:443"},

		// 		Username: "",

		// 		Credential:     "",
		// 		CredentialType: webrtc.ICECredentialTypePassword,
		// 	},
		// },
	}
)

type Room struct {
	Peers *Peers
	Hub   *chat.Hub
}

type PcpRoom struct {
	Peers  *Peers
	PcpHub *chat.PcpHub
}

type Peers struct {
	ListLock    sync.RWMutex
	Connections []PeerConnectionState
	TrackLocals map[string]*webrtc.TrackLocalStaticRTP
}

type PeerConnectionState struct {
	PeerConnection *webrtc.PeerConnection
	Websocket      *ThreadSafeWriter
}

// Conn用於寫入資料的ws
// 利用同步的方式實現了對WebSocket connection的多個goroutine的存取。目的是提供對WebSocket connection的穩定的存取，避免了因為多個goroutine存取WebSocket connection時產生的競爭問題。
type ThreadSafeWriter struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	return t.Conn.WriteJSON(v)
}

func (p *Peers) AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	p.ListLock.Lock()
	defer func() {
		p.ListLock.Unlock()
		p.SignalPeerConnections()
	}()

	// 建立本地的track
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	// fmt.Println("新增本地音軌")
	// fmt.Println(trackLocal)

	p.TrackLocals[t.ID()] = trackLocal
	return trackLocal
}

func (p *Peers) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	// fmt.Println("我進來remove")
	p.ListLock.Lock()
	defer func() {
		p.ListLock.Unlock()
		p.SignalPeerConnections()
	}()

	delete(p.TrackLocals, t.ID())
}

func (p *Peers) SignalPeerConnections() {
	p.ListLock.Lock()
	defer func() {
		p.ListLock.Unlock()
		p.DispatchKeyFrame()
	}()

	// 參考example-webrtc-applications/sfu-ws/main.go
	// 來決定是否需要再次同步
	attemptSync := func() (tryAgain bool) {
		for i := range p.Connections {
			if p.Connections[i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
				// fmt.Println("在同步裡面")
				// fmt.Println(p.Connections)
				log.Println("a", p.Connections)
				return true
			}
			// 檢查已存在的送出者，如果不存在，則從連接中刪除該送出者
			existingSenders := map[string]bool{}
			for _, sender := range p.Connections[i].PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}
				// fmt.Println("在existingSenders裡面")
				// fmt.Println(existingSenders)
				existingSenders[sender.Track().ID()] = true

				if _, ok := p.TrackLocals[sender.Track().ID()]; !ok {
					if err := p.Connections[i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}
			// 檢查已存在的接收者，標記為已存在；
			for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}
			// 檢查每個track，如果不存在於連接中，則將其加入連接
			for trackID := range p.TrackLocals {
				// fmt.Println("檢查trackID")
				// fmt.Println(trackID)
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := p.Connections[i].PeerConnection.AddTrack(p.TrackLocals[trackID]); err != nil {
						return true
					}
				}
			}
			// 創建一個 Offer，設定為SetLocalDescription
			offer, err := p.Connections[i].PeerConnection.CreateOffer(nil)
			// fmt.Println("執行CreateOffer")
			if err != nil {
				return true
			}

			//建立datachannel
			// dataChannel, err := p.Connections[i].PeerConnection.CreateDataChannel("mydatachannel", nil)
			// fmt.Println("執行datachannel")
			// fmt.Println(dataChannel.Label())
			// if err != nil {
			// 	return true
			// }
			//==========================

			if err = p.Connections[i].PeerConnection.SetLocalDescription(offer); err != nil {
				// fmt.Println("設定SetLocalDescription")
				return true
			}

			offerString, err := json.Marshal(offer)
			// fmt.Println("offer的字串是什麼開始")
			// fmt.Println(string(offerString))
			// fmt.Println("offer的字串是什麼結束")
			if err != nil {
				return true
			}
			// 透過 WebSocket將其發送到遠端連接
			if err = p.Connections[i].Websocket.WriteJSON(&websocketMessage{
				Event: "offer",
				Data:  string(offerString),
			}); err != nil {
				return true
			}
		}

		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			go func() {
				time.Sleep(time.Second * 3)
				p.SignalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

func (p *Peers) DispatchKeyFrame() {
	p.ListLock.Lock()
	defer p.ListLock.Unlock()

	for i := range p.Connections {
		for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = p.Connections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
