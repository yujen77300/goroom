const exitButton = document.querySelector(".exit-button")
const copyButton = document.querySelector("#copy-button")
const switchStreamBtn = document.querySelector('.switch-stream-btn')
const videos = document.querySelector('#videos')
let localVideo = document.querySelectorAll('.localVideo')
let eachPeer = document.querySelectorAll('.each-peer')
let viewerCountNow = document.querySelector('#viewer-count')
let userName = document.querySelector('.username')
// 判斷視訊畫面是否開啟
let isStreamStarted = true
let streamResult
// 一開始有一個人
let usersAmount = 1

peerSize(usersAmount, localVideo)

exitButton.addEventListener("click", function () {
  window.location.href = "/member";
});

copyButton.addEventListener("click", () => {
  copyURL()
})

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


// 按下允許連線
function connect(stream) {
  let pc = new RTCPeerConnection({
    iceServers: [{
      'urls': 'stun:stun.l.google.com:19302',
    },
    {
      'urls': 'turn:127.0.0.1:3478',
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
    console.log("e.candidate")
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
    while (pr.childElementCount > 3) {
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
        console.log(msg.data)
        console.log(offer)
        if (!offer) {
          return console.log('failed to parse answer')
        }
        pc.setRemoteDescription(offer)
        pc.createAnswer().then(answer => {
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
    document.getElementById('localVideo').srcObject = stream
    // document.getElementById('localVideo2').srcObject = stream
    connect(stream)
    streamResult = stream
    // switchStreamBtn.addEventListener('click', () => {
    //   stopStream(stream)
    // })
  }).catch(err => console.log(err))

// 開關鏡頭的函數
switchStreamBtn.addEventListener("click", () => {
  if (isStreamStarted) {
    stopStream(streamResult);
    isStreamStarted = false;
  } else {
    navigator.mediaDevices
      .getUserMedia({
        video: {
          width: {
            max: 1280
          },
          height: {
            max: 720
          },
          aspectRatio: 4 / 3,
          frameRate: 30
        },
        audio: {
          sampleSize: 16,
          channelCount: 2,
          echoCancellation: true
        }
      })
      .then(stream => {
        document.getElementById("localVideo").srcObject = stream;
        connect(stream);
        isStreamStarted = true;
      })
      .catch(err => console.log(err));
  }
})

function stopStream(stream) {
  stream.getTracks().forEach(track => track.stop());
}

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

