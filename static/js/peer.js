const exitButton = document.querySelector(".exit-button")
const copyButton = document.querySelector("#copy-button")
const switchStreamBtn = document.querySelector('.switch-stream-btn')
// 判斷視訊畫面是否開啟
let isStreamStarted = true
let streamResult

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

// DOM 結構被完整的讀取跟解析後就會被觸發，不須等待外部資源讀取完成
// document.addEventListener('DOMContentLoaded', () => {
//   (document.querySelectorAll('.notification .delete') || []).forEach(($delete) => {
//     const $notification = $delete.parentNode;

//     $delete.addEventListener('click', () => {
//       $notification.style.display = 'none'
//     });
//   });
// });

// 按下允許連線
function connect(stream) {
  // document.getElementById('peers').style.display = 'block'
  let pc = new RTCPeerConnection({
    iceServers: [{
      // 'urls': 'stun:turn.videochat:3478',
      'urls': 'stun:stun.l.google.com:19302',
    },
      {
        // 'urls': 'turn:turn.videochat:3478',
        'urls': 'turn:54.150.244.240:3478',
      'username': 'Dylan',
      'credential': 'Wehelp',
    }
    ]
  })
  // let pc = new RTCPeerConnection({
  //   iceServers: [
  //     {
  //       urls: 'stun:stun.l.google.com:19302'
  //     }
  //   ]
  // })
  // 接收另一端傳遞過來的多媒體資訊(videoTrack ...等)
  // 完成連線後，透過該事件能夠在發現遠端傳輸的多媒體檔案時觸發，來處理/接收多媒體數據
  pc.ontrack = function (event) {
    if (event.track.kind === 'audio') {
      return
    }

    col = document.createElement("div")
    col.className = "each-peer"
    let el = document.createElement(event.track.kind)
    el.srcObject = event.streams[0]
    el.setAttribute("controls", "true")
    el.setAttribute("autoplay", "true")
    el.setAttribute("playsinline", "true")
    el.setAttribute("id", "localVideo")
    col.appendChild(el)

    document.getElementById('videos').appendChild(col)

    event.track.onmute = function (event) {
      el.play()
    }

    event.streams[0].onremovetrack = ({
      track
    }) => {
      if (el.parentNode) {
        el.parentNode.remove()
      }
      // if (document.getElementById('videos').childElementCount <= 3) {
      //   document.getElementById('noone').style.display = 'grid'
      //   document.getElementById('noonein').style.display = 'grid'
      // }
    }
  }
  // 透過(addTrack)載入多媒體資訊(ex: videoTrack, audioTrack ...)
  // 將stream track 與peer connection 透過addTrack()關聯起來，之後建立連結才能進行傳輸
  // 加入MediaStream object到RTCPeerconnection中。
  stream.getTracks().forEach(track => pc.addTrack(track, stream))

  let ws = new WebSocket(RoomWebsocketAddr)
  // 當查找到相對應的遠端端口時會做onicecandidate，進行網路資訊的共享。
  // 當查找到相對應的遠端端口時會透過該事件來處理將 icecandidate 傳輸給 remote peers。
  pc.onicecandidate = e => {
    if (!e.candidate) {
      return
    }

    ws.send(JSON.stringify({
      event: 'candidate',
      data: JSON.stringify(e.candidate)
    }))
  }

  ws.addEventListener('error', function (event) {
    console.log('error: ', event)
  })

  ws.onclose = function (evt) {
    console.log("websocket has closed")
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

  ws.onmessage = function (evt) {
    let msg = JSON.parse(evt.data)
    if (!msg) {
      return console.log('failed to parse msg')
    }

    switch (msg.event) {
      case 'offer':
        let offer = JSON.parse(msg.data)
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

navigator.mediaDevices.getUserMedia({
  video: {
    width: {
      max: 1280
    },
    height: {
      max: 720
    },
    aspectRatio: 4 / 3,
    frameRate: 30,
  },
  audio: {
    sampleSize: 16,
    channelCount: 2,
    echoCancellation: true
  }
})
  .then(stream => {
    document.getElementById('localVideo').srcObject = stream
    connect(stream)
    streamResult = stream
    // switchStreamBtn.addEventListener('click', () => {
    //   stopStream(stream)
    // })
  }).catch(err => console.log(err))

switchStreamBtn.addEventListener("click", () => {
  if (isStreamStarted) {
    stopStream(streamResult);
    isStreamStarted = false;
  } else {
    console.log("我來了")
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

