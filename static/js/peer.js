const exitButton = document.querySelector(".exit-button")
// const copyButton = document.querySelector("#copy-button")
const shareScreenBtn = document.getElementById('share-btn')
const videos = document.querySelector('#videos')
let localVideo = document.querySelectorAll('.localVideo')
let eachPeer = document.querySelectorAll('.each-peer')
let viewerCountNow = document.querySelector('#viewer-count')
let userName = document.querySelector('.username')
// 判斷視訊畫面是否開啟
let streamOutput = { audio: true, video: true, }
let streamNow
let pcNow
// 一開始有一個人
let usersAmount = 1

peerSize(usersAmount, localVideo)

exitButton.addEventListener("click", function () {
  window.location.href = "/member";
});

// ===================== 複製url =====================
// copyButton.addEventListener("click", () => {
//   copyURL()
// })

function copyURL() {
  if (!navigator.clipboard) {
    alert('your broweser does not support clipboard API ');
    return;
  }
  let s = `${window.location.href}`;
  navigator.clipboard.writeText(s).then(() => {
    alert(`URL Copied!`);
  }).catch((err) => {
    alert(`${err}`);
  });
}

// ===================== peer to peer連線 =====================
// 按下允許連線
function connect(stream) {
  let pc = new RTCPeerConnection({
    iceServers: [{
      'urls': 'stun:stun.l.google.com:19302',
    },
    {
      'urls': 'turn:54.150.244.240:3478',
      'username': 'Dylan',
      'credential': 'Wehelp',
    }
    ]
  })
  // let pc = new RTCPeerConnection({
  //   iceServers: [
  //     {urls: "stun:relay.metered.ca:80",
  //   },
  //   {
  //     urls: "turn:relay.metered.ca:80",
  //     username: "",
  //     credential: "",
  //   },
  //   ],
  // })
  console.log("一開始的pc")
  console.log(pc)
  pcNow = pc
  // 接收另一端傳遞過來的多媒體資訊(videoTrack ...等)
  // 完成連線後，透過該事件能夠在發現遠端傳輸的多媒體檔案時觸發，來處理/接收多媒體數據
  pc.ontrack = function (event) {
    console.log("ontrack的event")
    console.log(event)
    // event => RTCTrackEvent
    // RTCTrackEvent有個streams屬性，每個對像表示track所屬的media stream
    console.log(event.track)
    // event.track=>MediaStreamTrack
    // 只處裡MediaStreamTrack是video的，排除音頻
    if (event.track.kind === 'audio') {
      console.log("在ontrack裡面")
      return
    }

    col = document.createElement("div")
    col.className = "each-peer"
    let el = document.createElement(event.track.kind)
    // event.streams[0] 為MediaStream
    console.log("增加的video")
    console.log(event.streams[0])
    el.srcObject = event.streams[0]
    el.setAttribute("controls", "true")
    el.setAttribute("autoplay", "true")
    el.setAttribute("playsinline", "true")
    el.setAttribute("id", "localVideo")
    el.setAttribute("class", "localVideo")
    col.appendChild(el)

    document.getElementById('videos').appendChild(col)

    let localVideo = document.querySelectorAll('.localVideo')
    usersAmount = localVideo.length
    peerSize(usersAmount, localVideo)

    event.track.onmute = function (event) {
      el.play()
    }

    // 執行removetrack方法會發生的事情，也就是移除video tag
    event.streams[0].onremovetrack = ({
      track  //MediaStreamTrack
    }) => {
      console.log("el的節點是否存在")
      console.log(track)
      if (el.parentNode) {
        el.parentNode.remove()
      }
      let localVideo = document.querySelectorAll('.localVideo')
      usersAmount = localVideo.length
      peerSize(usersAmount, localVideo)
    }

  }
  // 透過(addTrack)載入多媒體資訊(ex: videoTrack, audioTrack ...)
  // 將stream track 與peer connection 透過addTrack()關聯起來，之後建立連結才能進行傳輸
  // 加入MediaStream object到RTCPeerconnection (pc) 中。
  console.log("要載入資訊前知道stream")
  // stream是帶入函數的參數
  console.log(stream)
  console.log(stream.getTracks())
  stream.getTracks().forEach(track => pc.addTrack(track, stream))

  let ws = new WebSocket(RoomWebsocketAddr)
  console.log("建立新的ws")
  console.log(ws)
  console.log("透過(addTrack)載入多媒體後的pc")
  console.log(pc)
  // 當查找到相對應的遠端端口時會做onicecandidate，也就是透過callback function將icecandidate 傳輸給 remote peers。進行網路資訊的共享。
  pc.onicecandidate = e => {
    console.log("近來onicecandidate")
    console.log(e)
    if (!e.candidate) {
      console.log("如果沒有就停止")
      return
    }
    console.log("ws發送訊息")
    console.log(e.candidate)
    ws.send(JSON.stringify({
      event: 'candidate',
      data: JSON.stringify(e.candidate)
    }))
  }

  ws.addEventListener('error', function (event) {
    console.log('error: ', event)
  })

  ws.onclose = function (evt) {
    console.log("websocket has closed結束")
    // 關閉與 WebSocket 相關的 PeerConnection（pc.close()），並將設為null；
    pc.close();
    pc = null;
    pr = document.getElementById('videos')
    while (pr.childElementCount > 1) {
      pr.lastChild.remove()
    }
    setTimeout(function () {
      connect(stream);
    }, 1000);
  }
  // 收到 Server 發來的訊息時觸發
  ws.onmessage = function (evt) {
    let msg = JSON.parse(evt.data)
    console.log("收到 Server 發來的訊息時觸發")
    console.log(evt)
    console.log(evt.data)
    console.log(msg)
    if (!msg) {
      return console.log('failed to parse msg')
    }

    switch (msg.event) {
      case 'offer':
        let offer = JSON.parse(msg.data)
        console.log("如果是offer印出offer")
        console.log(msg)
        console.log(offer)
        if (!offer) {
          return console.log('failed to parse answer')
        }
        pc.setRemoteDescription(offer)
        pc.createAnswer().then(answer => {
          console.log("進來answer")
          console.log(answer)
          pc.setLocalDescription(answer)
          ws.send(JSON.stringify({
            event: 'answer',
            data: JSON.stringify(answer)
          }))
        })
        return

      case 'candidate':
        let candidate = JSON.parse(msg.data)
        console.log("如果是candidate印出candidate")
        console.log(candidate)
        if (!candidate) {
          return console.log('failed to parse candidate')
        }
        // 當 remotePeer 藉由 Signaling channel 接收到由 localPeer 傳來的 ICE candidate 時，利用addIceCandidate將其丟給瀏覽器解析與匹配，看看這個ICE candidate 所提供的連線方式適不適合。
        pc.addIceCandidate(candidate)
    }
  }

  ws.onerror = function (evt) {
    console.log("error: " + evt.data)
  }
}

// 這是最一開始
navigator.mediaDevices.getUserMedia({
  video: {
    width: { min: 1280 },
    height: { min: 720 }
  },
  audio: true
})
  .then(stream => {
    console.log("這是一開始的stream")
    console.log(stream)

    document.getElementById('localVideo').srcObject = stream
    // document.getElementById('localVideo2').srcObject = stream
    connect(stream)
    streamNow = stream
  }).catch(err => console.log(err))


// ===================== turn of/off microphone and camera =====================
// 開關鏡頭的函數
const videoOpenedBtn = document.getElementById('video-opened-btn')
const videoClosedBtn = document.getElementById('video-closed-btn')
videoOpenedBtn.addEventListener("click", () => {
  console.log("要進來關掉視訊了")
  stopVideo(streamNow);
  streamOutput.video = false;
  videoOpenedBtn.style.display = "none"
  videoClosedBtn.style.display = "block"
})

videoClosedBtn.addEventListener("click", () => {
  console.log("要進來開啟視訊了")
  startVideo(streamNow)
  streamOutput.video = true;
  videoClosedBtn.style.display = "none"
  videoOpenedBtn.style.display = "block"
})

const audioOpendBtn = document.getElementById('audio-opened-btn')
const audioClosedBtn = document.getElementById('audio-closed-btn')

audioOpendBtn.addEventListener("click", () => {
  console.log("要進來關掉了")
  stopAudio(streamNow);
  streamOutput.audio = false;
  audioOpendBtn.style.display = "none"
  audioClosedBtn.style.display = "block"
})

audioClosedBtn.addEventListener("click", () => {
  console.log("要進來開始了")
  startAudio(streamNow)
  streamOutput.audio = true;
  audioClosedBtn.style.display = "none"
  audioOpendBtn.style.display = "block"
})

function stopVideo(stream) {
  stream.getVideoTracks()[0].enabled = false;
  // stream.getVideoTracks()[0].stop()
  // stream.getTracks().forEach(track => pc.removeTrack(pc.addTrack(track, stream)))

  // 會讓整個黑色不見
  // document.getElementById('localVideo').srcObject = null


  // 才會關閉
  // test(stream)
  // stream.getTracks().forEach(track => {
  //   track.stop()
  // })

}

function startVideo(stream) {
  stream.getVideoTracks()[0].enabled = true;
  // navigator.mediaDevices.getUserMedia({
  //   video: {
  //     width: { min: 1280 },
  //     height: { min: 720 }
  //   },
  //   audio: true
  // })
  //   .then(stream => {
  //     streamNow = stream
  //     if (streamOutput.audio){
  //       document.getElementById('localVideo').srcObject = stream
  //       console.log("重新開始後取得的")
  //       console.log(stream)
  //       console.log(stream.getTracks())
  //     }else{
  //       stopAudio(stream)
  //       document.getElementById('localVideo').srcObject = stream
  //     }
  //   }).catch(err => console.log(err))
}

function stopAudio(stream) {
  stream.getAudioTracks()[0].enabled = false;
}

function startAudio(stream) {
  stream.getAudioTracks()[0].enabled = true;
}


// ===================== 分享螢幕 =====================


// shareScreenBtn.addEventListener("click", () => {
//   console.log(streamNow.getVideoTracks())
//   streamNow.getVideoTracks()[0].stop()
//   const constraints = {
//     frameRate: 15,
//     width: 1280,
//     height: 720,
//   }
//   navigator.mediaDevices
//     .getDisplayMedia(constraints)
//     .then(shareStream => {
//       console.log("這是真想螢幕的stream")
//       console.log(shareStream)
//       console.log(shareStream.getTracks())
//       streamNow.getTracks().forEach(track =>{
//         if (track.kind === 'video') {
//           pcNow.addTrack(track, shareStream);
//         }
//       })

//     })
//     .catch(err => console.log(err))
// })

// ===================== 人數切版 =====================
function peerSize(usersAmount, localVideo) {
  if (usersAmount === 1) {
    localVideo[0].style.width = "1160px"
    eachPeer[0].style.width = "1160px"
    let videoWidth = (1160 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    userName.style.bottom = `-${videoWidth - 60}px`
  } else if (usersAmount === 2) {
    localVideo[0].style.width = "580px"
    localVideo[1].style.width = "580px"
    let eachPeer = document.querySelectorAll('.each-peer')
    eachPeer[0].style.width = "580px"
    eachPeer[1].style.width = "580px"
    let videoWidth = (580 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    eachPeer[1].style.height = `${videoWidth}px`
    userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.gap = "10px"
  } else if (usersAmount === 3) {
    localVideo[0].style.width = "580px"
    localVideo[1].style.width = "580px"
    localVideo[2].style.width = "580px"
    let eachPeer = document.querySelectorAll('.each-peer')
    eachPeer[0].style.width = "580px"
    eachPeer[1].style.width = "580px"
    eachPeer[2].style.width = "580px"
    let videoWidth = (580 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    eachPeer[1].style.height = `${videoWidth}px`
    eachPeer[2].style.height = `${videoWidth}px`
    userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if (usersAmount === 4) {
    localVideo[0].style.width = "580px"
    localVideo[1].style.width = "580px"
    localVideo[2].style.width = "580px"
    localVideo[3].style.width = "580px"
    let eachPeer = document.querySelectorAll('.each-peer')
    eachPeer[0].style.width = "580px"
    eachPeer[1].style.width = "580px"
    eachPeer[2].style.width = "580px"
    eachPeer[3].style.width = "580px"
    let videoWidth = (580 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    eachPeer[1].style.height = `${videoWidth}px`
    eachPeer[2].style.height = `${videoWidth}px`
    eachPeer[3].style.height = `${videoWidth}px`
    userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  }
}

