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

// Conn?????????????????????ws
// ?????????????????????????????????WebSocket connection?????????goroutine??????????????????????????????WebSocket connection??????????????????????????????????????????goroutine??????WebSocket connection???????????????????????????
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

	// ???????????????track
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	// fmt.Println("??????????????????")
	// fmt.Println(trackLocal)

	p.TrackLocals[t.ID()] = trackLocal
	return trackLocal
}

func (p *Peers) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	// fmt.Println("?????????remove")
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

	// ??????example-webrtc-applications/sfu-ws/main.go
	// ?????????????????????????????????
	attemptSync := func() (tryAgain bool) {
		for i := range p.Connections {
			if p.Connections[i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
				// fmt.Println("???????????????")
				// fmt.Println(p.Connections)
				log.Println("a", p.Connections)
				return true
			}
			// ?????????????????????????????????????????????????????????????????????????????????
			existingSenders := map[string]bool{}
			for _, sender := range p.Connections[i].PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}
				// fmt.Println("???existingSenders??????")
				// fmt.Println(existingSenders)
				existingSenders[sender.Track().ID()] = true

				if _, ok := p.TrackLocals[sender.Track().ID()]; !ok {
					if err := p.Connections[i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}
			// ???????????????????????????????????????????????????
			for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}
			// ????????????track??????????????????????????????????????????????????????
			for trackID := range p.TrackLocals {
				// fmt.Println("??????trackID")
				// fmt.Println(trackID)
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := p.Connections[i].PeerConnection.AddTrack(p.TrackLocals[trackID]); err != nil {
						return true
					}
				}
			}
			// ???????????? Offer????????????SetLocalDescription
			offer, err := p.Connections[i].PeerConnection.CreateOffer(nil)
			// fmt.Println("??????CreateOffer")
			if err != nil {
				return true
			}

			//??????datachannel
			// dataChannel, err := p.Connections[i].PeerConnection.CreateDataChannel("mydatachannel", nil)
			// fmt.Println("??????datachannel")
			// fmt.Println(dataChannel.Label())
			// if err != nil {
			// 	return true
			// }
			//==========================

			if err = p.Connections[i].PeerConnection.SetLocalDescription(offer); err != nil {
				// fmt.Println("??????SetLocalDescription")
				return true
			}

			offerString, err := json.Marshal(offer)
			// fmt.Println("offer????????????????????????")
			// fmt.Println(string(offerString))
			// fmt.Println("offer????????????????????????")
			if err != nil {
				return true
			}
			// ?????? WebSocket???????????????????????????
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
