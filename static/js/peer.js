const exitButton = document.getElementById("exit-btn")
const shareLinkBtn = document.getElementById("share-link-btn")
// const shareScreenBtn = document.getElementById('share-btn')
const videos = document.querySelector('#videos')
let localVideo = document.querySelectorAll('.localVideo')
let eachPeer = document.querySelectorAll('.each-peer')
let viewerCountNow = document.querySelector('#viewer-count')
let userName = document.querySelector('.user-name')
let videoClosedAvatar = document.querySelector('.video-closed-avatar')
const pcpsInMeeting = document.getElementById('pcpsInMeeting')
let streamOutput = { audio: true, video: true, }
let defaultSet = false
let testStreamNow
let pcpEmail = ""
let pcpId = ""
let pcpName = ""
let handWs = ""
let streamNow
let pcNow
let usersAmount = 1
let streamDict = {};

peerSize(usersAmount, localVideo)
peerConnect(defaultSet)
getUserEmail()


exitButton.addEventListener("click", function () {
  window.location.href = "/member";
});

// ===================== 複製url =====================
shareLinkBtn.addEventListener("click", () => {
  copyURL()
})

function copyURL() {
  if (!navigator.clipboard) {
    Swal.fire({
      icon: 'error',
      title: 'Oops...',
      text: 'Your browser does not support clipboard API',
    });
    return;
  }
  let infoCopied = `${window.location.href}`;
  navigator.clipboard.writeText(infoCopied).then(() => {
    Swal.fire({
      icon: 'success',
      title: 'Link Copied Successfully',
      showConfirmButton: false,
      timer: 1500,
    });
  }).catch((err) => {
    Swal.fire({
      icon: 'error',
      title: 'Oops...',
      text: `${err}`,
    });
  });
}

// ===================== peer to peer連線 =====================
// 按下允許連線
function connect(stream) {
  let pc = new RTCPeerConnection({
    iceServers: [{
      urls: 'stun:goroom.online:3478',
    },
    {
      urls: 'turn:goroom.online:3478',
      username: TurnName,
      credential: TurnPwd,
    },
    {
      'urls': "stun:relay.metered.ca:80",
    },
    {
      urls: "turn:relay.metered.ca:80",
      username: TurnName2,
      credential: TurnPwd2,
    },
    {
      urls: "turn:relay.metered.ca:443",
      username: TurnName2,
      credential: TurnPwd2,
    },
    ]
  })

  // console.log("一開始的pc")
  // console.log(pc)
  pcNow = pc

  // 接收另一端傳遞過來的多媒體資訊(videoTrack ...等)
  // 完成連線後，透過該事件能夠在發現遠端傳輸的多媒體檔案時觸發，來處理/接收多媒體數據
  // As it turns out, MediaStreamTracks get a new ID assigned on the other side.MediaStreams however, keep their assigned IDs, so use those when doing AddTrack, and then use a DataChannel to send information about the stream based on its ID.
  pc.ontrack = function (event) {
    // console.log("ontrack裡面的event")
    // console.log(event)
    // console.log(event.streams)
    // console.log(event.streams[0])
    // event => RTCTrackEvent
    // RTCTrackEvent有個streams屬性，每個對像表示track所屬的media stream
    // console.log(event.track)

    if (event.track.kind === 'audio') {
      return
    }

    eachPeerTag = document.createElement("div")
    eachPeerTag.className = "each-peer"
    eachPeerTag.id = event.streams[0].id
    let el = document.createElement(event.track.kind)
    // event.streams[0] 為MediaStream

    el.srcObject = event.streams[0]

    el.setAttribute("autoplay", "true")
    el.setAttribute("playsinline", "true")
    el.setAttribute("id", "localVideo")
    el.setAttribute("class", "localVideo")
    eachPeerTag.appendChild(el)

    let newUserName = document.createElement("div")
    newUserName.className = "user-name"
    let newRaiseHandIcon = document.createElement("img");
    newRaiseHandIcon.src = "../img/handRaise.png";
    newRaiseHandIcon.className = "raise-yellow-hand";
    newRaiseHandIcon.id = "hand-" + event.streams[0].id;

    eachPeerTag.appendChild(newRaiseHandIcon)
    eachPeerTag.appendChild(newUserName)
    let url = window.location.href
    let segments = url.split('/')
    let uuid = segments[segments.length - 1]
    getPcpInfo(uuid, event.streams[0].id, newUserName)

    document.getElementById('videos').appendChild(eachPeerTag)

    let localVideo = document.querySelectorAll('.localVideo')
    usersAmount = localVideo.length
    peerSize(usersAmount, localVideo)

    event.track.onmute = function (event) {
      el.play()
    }

    event.streams[0].onremovetrack = ({
      track  //MediaStreamTrack
    }) => {

      if (el.parentNode) {
        el.parentNode.remove()
      }
      // 更新上線者清單相關websocket=========================================
      if (track.kind == "video") {

        let streamRemove = {}
        streamRemove["streamId"] = event.streams[0].id
        pcpsWs.send(JSON.stringify({ event: "leave", data: JSON.stringify(streamRemove) }))
      }
      // // 更新上線者清單相關websocket=========================================

      let localVideo = document.querySelectorAll('.localVideo')
      usersAmount = localVideo.length
      peerSize(usersAmount, localVideo)
    }

  }
  // 透過(addTrack)載入多媒體資訊(ex: videoTrack, audioTrack ...)
  // 將stream track 與peer connection 透過addTrack()關聯起來，之後建立連結才能進行傳輸
  // 加入MediaStream object到RTCPeerconnection (pc) 中。
  // console.log("載入資訊前知道stream")
  // console.log(stream)
  // console.log(stream.getTracks())
  stream.getTracks().forEach(track => pc.addTrack(track, stream))

  let ws = new WebSocket(RoomWebsocketAddr)
  // console.log("建立新的ws")
  // console.log(ws)
  // 當查找到相對應的遠端端口時會做onicecandidate，也就是透過callback function將icecandidate 傳輸給 remote peers。進行網路資訊的共享。
  pc.onicecandidate = e => {
    // console.log("近來onicecandidate")
    // console.log(e)
    if (!e.candidate) {
      // console.log("如果沒有就停止")
      return
    }
    // console.log("ws發送訊息")
    // console.log(e.candidate)
    ws.send(JSON.stringify({
      event: 'candidate',
      data: JSON.stringify(e.candidate)
    }))
  }

  // 更新上線者清單相關websocket=========================================
  let pcpsWs = new WebSocket(PcpsWebsocketAddr)
  handWs = pcpsWs
  pcpsWs.onopen = () => {
    let url = window.location.href
    let segments = url.split('/')
    let uuid = segments[segments.length - 1]
    getAllPcpInRoom(uuid, stream.id)
    streamDict["streamId"] = stream.id
    streamDict["pcpEmail"] = pcpEmail
    streamDict["pcpId"] = pcpId
    streamDict["pcpName"] = pcpName
    pcpsWs.send(JSON.stringify({ event: "join", data: JSON.stringify(streamDict) }))
  }

  pcpsWs.onmessage = function (e) {
    leaveInfo = e.data.split('\n')[0]
    let msg = JSON.parse(leaveInfo)
    let pcpMsg = JSON.parse(msg.data)

    switch (msg.event) {
      case 'join':
        eachPcp = document.createElement("div")
        eachPcp.className = "each-pcp"
        eachPcp.id = pcpMsg.streamId
        pcpAvatar = document.createElement("img")
        pcpAvatar.className = "pcp-avatar"
        pcpAvatar.alt = pcpMsg.pcpName
        pcpName = document.createElement("div")
        pcpName.className = "pcp-name"
        pcpName.textContent = pcpMsg.pcpName
        getPcpAvatar(pcpMsg.pcpEmail, pcpAvatar)
        eachPcp.appendChild(pcpAvatar)
        eachPcp.appendChild(pcpName)
        pcpsInMeeting.appendChild(eachPcp)
        break;
      case 'leave':
        let children = pcpsInMeeting.children;
        Array.from(children).forEach(eachpeer => {
          if (eachpeer.id == pcpMsg.streamId) {
            pcpsInMeeting.removeChild(eachpeer)
          }
        });
        break;
      case 'hand':
        let userHand = document.getElementById(pcpMsg.handId)
        if (pcpMsg.handStatus === "down") {
          userHand.style.display = "none"
        } else {
          userHand.style.display = "block"
        }
        break;
      case 'lottery':
        let winnerId = pcpMsg.winnerId
        let pcpList = document.querySelectorAll(".each-pcp")
        pcpList.forEach((pcp) => {
          if (pcp.id == winnerId) {
            winner = pcp
          }
        });
        startLottery(pcpList, winner)
        break;

    }
  }


  pcpsWs.onclose = function (e) {
    console.log("pcp lists websocket has closed")
    setTimeout(function () {
      connect(stream);
    }, 1000);
  }

  pcpsWs.onerror = function (e) {
    console.log("error: " + e.data)
  }
  // signal server相關websocket=========================================

  ws.addEventListener('error', function (event) {
    console.log('error: ', event)
  })

  ws.onclose = function (e) {
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
  // 收到 Server 發來的訊息時觸發 webrtc/peers.go 、room.go
  ws.onmessage = function (e) {
    let msg = JSON.parse(e.data)
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
      // onicecandidate
      case 'candidate':
        let candidate = JSON.parse(msg.data)
        // console.log("如果是candidate印出candidate")
        // console.log(candidate)
        if (!candidate) {
          return console.log('failed to parse candidate')
        }
        // 當 remotePeer 藉由 Signaling channel 接收到由 localPeer 傳來的 ICE candidate 時，利用addIceCandidate將其丟給瀏覽器解析與匹配，看看這個ICE candidate 所提供的連線方式適不適合。
        pc.addIceCandidate(candidate)
    }
  }

  ws.onerror = function (e) {
    console.log("error: " + e.data)
  }
}

// 這是最一開始
function peerConnect(defaultSet) {
  getUserAvatar()
  if (defaultSet == true) {
    navigator.mediaDevices.getUserMedia({
      video: {
        width: { min: 1280 },
        height: { min: 720 }
      },
      audio: true,
    })
      .then(stream => {
        document.getElementById('localVideo').srcObject = stream
        // console.log("最最一開始")
        // console.log(stream)
        // console.log(stream.getTracks())
        let eachPeer = document.querySelector('.each-peer')
        yourHand = eachPeer.children[1]
        yourHand.style.top = `30px`
        eachPeer.children[1].id = "hand-" + stream.id
        console.log(eachPeer.children[1])
        eachPeer.id = stream.id
        connect(stream)
        streamNow = stream
        audioVideoDefault(streamOutput.audio, streamOutput.video)
        getUserName()
      }).catch(err => console.log(err))
  } else {
    //  ===================== Test section =====================
    const testVideoClosedBtn = document.getElementById('test-video-closed-btn')
    const testVideoOpenedBtn = document.getElementById('test-video-opened-btn')
    const testAudioClosedBtn = document.getElementById('test-audio-closed-btn')
    const testAudioOpenedBtn = document.getElementById('test-audio-opened-btn')
    const testVideo = document.getElementById('testVideo')
    const testVideoClosedAvatar = document.getElementById('test-video-closed-avatar')

    getUserAvatar()
    getUserName()

    navigator.mediaDevices.getUserMedia({
      video: {
        width: { min: 815 },
        height: { min: 480 }
      },
      audio: true
    })
      .then(testStream => {
        testVideo.srcObject = testStream
        testVideoOpenedBtn.addEventListener("click", () => {
          streamOutput.video = false;
          testVideoClosedAvatar.style.display = "block"
          testVideoOpenedBtn.style.display = "none"
          testVideoClosedBtn.style.display = "block"
          testStream.getVideoTracks()[0].enabled = false;
        })
        testVideoClosedBtn.addEventListener("click", () => {
          streamOutput.video = true;
          testVideoClosedAvatar.style.display = "none"
          testVideoOpenedBtn.style.display = "block"
          testVideoClosedBtn.style.display = "none"
          testStream.getVideoTracks()[0].enabled = true;
        })
        testAudioClosedBtn.addEventListener("click", () => {
          streamOutput.audio = true;
          testAudioOpenedBtn.style.display = "block"
          testAudioClosedBtn.style.display = "none"
          testStream.getAudioTracks()[0].enabled = true;
        })
        testAudioOpenedBtn.addEventListener("click", () => {
          streamOutput.audio = false;
          testAudioOpenedBtn.style.display = "none"
          testAudioClosedBtn.style.display = "block"
          testStream.getAudioTracks()[0].enabled = false;
        })
        testStreamNow = testStream
      }).catch(err => console.log(err))
  }
}

const testExitBtn = document.querySelector('.test-exit-btn')
const testjoinBtn = document.querySelector('.test-join-btn')
const beforeEnterSection = document.querySelector('.beforeEnter-section')
const bodySection = document.getElementById("body-section")
const bottomSection = document.getElementById("bottom-section")

testExitBtn.addEventListener('click', () => {
  window.location.href = "/member";
})

testjoinBtn.addEventListener("click", () => {
  testStreamNow.getTracks().forEach(track => {
    track.stop()
  })
  defaultSet = true
  beforeEnterSection.style.display = "none"
  bodySection.style.display = "block"
  bottomSection.style.display = "block"
  peerConnect(defaultSet)
})

// ===================== microphone and camera default setting =====================

function audioVideoDefault(audioDefault, videoDefault) {
  if (audioDefault == true) {
    startAudio(streamNow);
    audioClosedBtn.style.display = "none"
    audioOpendBtn.style.display = "block"

  } else if (audioDefault == false) {
    console.log("audioDefault == false")
    stopAudio(streamNow);
    audioOpendBtn.style.display = "none"
    audioClosedBtn.style.display = "block"
  }
  if (videoDefault == true) {
    // stream.getVideoTracks()[0].enabled = true;
    startVideo(streamNow)
    videoClosedBtn.style.display = "none"
    videoOpenedBtn.style.display = "block"

  } else if (videoDefault == false) {
    console.log("videoDefault == false")
    stopVideo(streamNow);
    videoOpenedBtn.style.display = "none"
    videoClosedBtn.style.display = "block"
  }
}


// ===================== turn of/off microphone and camera =====================
// 開關鏡頭的函數
const videoOpenedBtn = document.getElementById('video-opened-btn')
const videoClosedBtn = document.getElementById('video-closed-btn')
videoOpenedBtn.addEventListener("click", () => {
  // streamNow.getVideoTracks()[0].stop()
  stopVideo(streamNow);
  streamOutput.video = false;
  videoOpenedBtn.style.display = "none"
  videoClosedBtn.style.display = "block"
})

videoClosedBtn.addEventListener("click", () => {
  startVideo(streamNow)
  streamOutput.video = true;
  videoClosedBtn.style.display = "none"
  videoOpenedBtn.style.display = "block"
})

const audioOpendBtn = document.getElementById('audio-opened-btn')
const audioClosedBtn = document.getElementById('audio-closed-btn')

audioOpendBtn.addEventListener("click", () => {

  stopAudio(streamNow);
  streamOutput.audio = false;
  audioOpendBtn.style.display = "none"
  audioClosedBtn.style.display = "block"
})

audioClosedBtn.addEventListener("click", () => {

  startAudio(streamNow)
  streamOutput.audio = true;
  audioClosedBtn.style.display = "none"
  audioOpendBtn.style.display = "block"
})

function stopVideo(stream) {
  stream.getVideoTracks()[0].enabled = false;

}

function startVideo(stream) {
  stream.getVideoTracks()[0].enabled = true;
}

function stopAudio(stream) {
  stream.getAudioTracks()[0].enabled = false;
}

function startAudio(stream) {
  stream.getAudioTracks()[0].enabled = true;
}

// ===================== 舉手 =====================
const raiseHand = document.querySelector('.raise-hand')

raiseHand.addEventListener("click", () => {
  if (yourHand.style.display === "block") {
    let handUpdate = {}
    handUpdate["handId"] = yourHand.id
    handUpdate["handStatus"] = "down"
    handWs.send(JSON.stringify({ event: "hand", data: JSON.stringify(handUpdate) }))
  } else {

    let handUpdate = {}
    handUpdate["handId"] = yourHand.id
    handUpdate["handStatus"] = "raise"
    handWs.send(JSON.stringify({ event: "hand", data: JSON.stringify(handUpdate) }))
  }
})

// ===================== 抽籤 =====================
const pcpRandom = document.querySelector('.pcp-random')

pcpRandom.addEventListener("click", () => {
  const pcpList = document.querySelectorAll(".each-pcp");
  const winnerIndex = Math.floor(Math.random() * pcpList.length);
  const winnerUpdate = {}
  winnerUpdate["winnerId"] = (pcpList[winnerIndex]).id
  handWs.send(JSON.stringify({ event: "lottery", data: JSON.stringify(winnerUpdate) }))
  if (pcpsListOpen == false) {
    Swal.fire({
      icon: 'success',
      title: 'Random selection is finished, you can open the participant list to see the results.',
      showConfirmButton: false,
      timer: 3000,
    });
  } else {
  }
})

function startLottery(pcpList, winner) {
  pcpList.forEach((pcp) => {
    pcp.classList.remove("gray-background");
  });

  let counter = 0;
  let intervalId = setInterval(() => {
    pcpList[counter].classList.add("gray-background");

    if (counter > 0) {
      pcpList[counter - 1].classList.remove("gray-background");
    } else {
      pcpList[pcpList.length - 1].classList.remove("gray-background");
    }

    counter++;
    if (counter >= pcpList.length) {
      counter = 0;
    }
  }, 100);

  setTimeout(() => {
    clearInterval(intervalId);
    pcpList.forEach((pcp) => pcp.classList.remove("gray-background"));
    winner.classList.add("gray-background");
  }, 3000);
}

// ===================== 錄製螢幕 =====================
const startRecord = document.querySelector('.start-record')
const stopRecord = document.querySelector('.stop-record')

startRecord.addEventListener('click', function () {
  Swal.fire({
    title: 'Start screen recording?',
    text: "Are you sure you want to start recording your screen?",
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#1158bd',
    cancelButtonColor: '#dc3545',
    confirmButtonText: 'Yes',
    cancelButtonText: 'Cancel'
  }).then((result) => {
    if (result.isConfirmed) {
      startRecord.style.display = "none"
      stopRecord.style.display = "flex"
      startRecording();
    }
  })
});


stopRecord.addEventListener('click', function () {
  startRecord.style.display = "flex"
  stopRecord.style.display = "none"
  mediaRecorder.stop();
})

// ===================== 人數切版 =====================
function peerSize(usersAmount, localVideo) {
  if (usersAmount === 1) {
    localVideo[0].style.width = "1140px"
    eachPeer[0].style.width = "1140px"
    let videoWidth = (1160 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    // videoClosedAvatar.style.bottom = "-220px"
    // userName.style.bottom = `-${videoWidth - 60}px`
    userName.style.bottom = `30px`
    videos.style.cssText = "display:flex;justify-content:center;align-items:center;"
  } else if (usersAmount === 2) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let userName = document.querySelectorAll('.user-name')
    let raiseYellowHand = document.querySelectorAll('.raise-yellow-hand')
    localVideo[0].style.width = "565px"
    localVideo[1].style.width = "565px"
    eachPeer[0].style.width = "565px"
    eachPeer[1].style.width = "565px"
    let videoWidth = (565 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    eachPeer[1].style.height = `${videoWidth}px`
    // videoClosedAvatar.style.bottom = "-75px"
    userName[0].style.bottom = `10px`
    userName[1].style.bottom = `10px`
    raiseYellowHand[0].style.top = `10px`
    raiseYellowHand[1].style.top = `10px`
    videos.style.display = "flex"
    videos.style.gap = "10px"
  } else if (usersAmount === 3) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let userName = document.querySelectorAll('.user-name')
    let raiseYellowHand = document.querySelectorAll('.raise-yellow-hand')
    localVideo[0].style.width = "565px"
    localVideo[1].style.width = "565px"
    localVideo[2].style.width = "565px"
    eachPeer[0].style.width = "565px"
    eachPeer[1].style.width = "565px"
    eachPeer[2].style.width = "565px"
    let videoWidth = (565 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    eachPeer[1].style.height = `${videoWidth}px`
    eachPeer[2].style.height = `${videoWidth}px`
    userName[0].style.bottom = `10px`
    userName[1].style.bottom = `10px`
    userName[2].style.bottom = `10px`
    raiseYellowHand[0].style.top = `10px`
    raiseYellowHand[1].style.top = `10px`
    raiseYellowHand[2].style.top = `10px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if (usersAmount === 4) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let userName = document.querySelectorAll('.user-name')
    let raiseYellowHand = document.querySelectorAll('.raise-yellow-hand')
    localVideo[0].style.width = "565px"
    localVideo[1].style.width = "565px"
    localVideo[2].style.width = "565px"
    localVideo[3].style.width = "565px"
    eachPeer[0].style.width = "565px"
    eachPeer[1].style.width = "565px"
    eachPeer[2].style.width = "565px"
    eachPeer[3].style.width = "565px"
    let videoWidth = (565 * 9) / 16
    eachPeer[0].style.height = `${videoWidth}px`
    eachPeer[1].style.height = `${videoWidth}px`
    eachPeer[2].style.height = `${videoWidth}px`
    eachPeer[3].style.height = `${videoWidth}px`
    userName[0].style.bottom = `10px`
    userName[1].style.bottom = `10px`
    userName[2].style.bottom = `10px`
    userName[3].style.bottom = `10px`
    raiseYellowHand[0].style.top = `10px`
    raiseYellowHand[1].style.top = `10px`
    raiseYellowHand[2].style.top = `10px`
    raiseYellowHand[3].style.top = `10px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if ((usersAmount > 4 && usersAmount <= 6) || usersAmount == 9) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let userName = document.querySelectorAll('.user-name')
    let raiseYellowHand = document.querySelectorAll('.raise-yellow-hand')
    let videoWidth = (373 * 9) / 16
    for (let i = 0; i < usersAmount; i++) {
      localVideo[i].style.width = "373px"
      eachPeer[i].style.width = "373px"
      eachPeer[i].style.height = `${videoWidth}px`
      userName[i].style.bottom = `-${videoWidth - 60}px`
      raiseYellowHand[i].style.top = `10px`
    }
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if ((usersAmount > 6 && usersAmount <= 8) || (usersAmount > 9 && usersAmount <= 12)) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let videoWidth = (277 * 9) / 16
    for (let i = 0; i < usersAmount; i++) {
      localVideo[i].style.width = "277px"
      eachPeer[i].style.width = "277px"
      eachPeer[i].style.height = `${videoWidth}px`
    }
    // userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if ((usersAmount > 12 && usersAmount <= 15) || (usersAmount > 18 && usersAmount <= 20)) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let videoWidth = (220 * 9) / 16
    for (let i = 0; i < usersAmount; i++) {
      localVideo[i].style.width = "220px"
      eachPeer[i].style.width = "220px"
      eachPeer[i].style.height = `${videoWidth}px`
    }
    // userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if ((usersAmount > 15 && usersAmount <= 18) || (usersAmount > 20 && usersAmount <= 24)) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let videoWidth = (181 * 9) / 16
    for (let i = 0; i < usersAmount; i++) {
      localVideo[i].style.width = "181px"
      eachPeer[i].style.width = "181px"
      eachPeer[i].style.height = `${videoWidth}px`
    }
    // userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if (usersAmount > 24 && usersAmount <= 28) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let videoWidth = (154 * 9) / 16
    for (let i = 0; i < usersAmount; i++) {
      localVideo[i].style.width = "154px"
      eachPeer[i].style.width = "154px"
      eachPeer[i].style.height = `${videoWidth}px`
    }
    // userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  } else if (usersAmount > 29) {
    let eachPeer = document.querySelectorAll('.each-peer')
    let videoWidth = (133 * 9) / 16
    for (let i = 0; i < usersAmount; i++) {
      localVideo[i].style.width = "133px"
      eachPeer[i].style.width = "133px"
      eachPeer[i].style.height = `${videoWidth}px`
    }
    // userName.style.bottom = `-${videoWidth - 60}px`
    videos.style.display = "flex"
    videos.style.flexWrap = "wrap"
    videos.style.gap = "10px"
  }
}

// ===================== Async function =====================
async function getUserAvatar() {
  const testAvatar = document.querySelector('.testAvatar')
  const testVideoClosedAvatar = document.getElementById('test-video-closed-avatar')
  let url = "/api/user/avatar"
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      if (defaultSet == false) {
        testAvatar.style.backgroundImage = `url(${result.userAvatar})`
        testVideoClosedAvatar.style.backgroundImage = `url(${result.userAvatar})`
      }
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

async function getUserName() {

  const testName = document.querySelector('.testName')
  const userName = document.querySelector('.user-name')
  let url = "/api/user/auth"
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      testName.textContent = result.data.name
      userName.textContent = result.data.name
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

async function getUserEmail() {
  let url = "/api/user/auth"
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      pcpEmail = result.data.email
      pcpId = result.data.id
      pcpName = result.data.name
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

async function getPcpAvatar(pcpEmail, pcpAvatar) {
  let url = `/api/avatar/:${pcpEmail}`
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      pcpAvatar.src = `${result.pcpAvatarUrl}`
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}

async function getAllPcpInRoom(uuid, streamId) {
  let url = `/api/allpcps/:${uuid}`
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      // console.log(result.allPcps) //會取得不同的array

      let existPcpList = []
      result.allPcps.forEach(each => {
        if (each.pcp_stream_url !== `${streamId}`) {
          existPcpList.push(each)
        }
      })

      existPcpList.forEach(each => {
        eachPcp = document.createElement("div")
        eachPcp.className = "each-pcp"
        eachPcp.id = each.pcp_stream_url
        pcpAvatar = document.createElement("img")
        pcpAvatar.className = "pcp-avatar"
        pcpAvatar.alt = each.username
        pcpName = document.createElement("div")
        pcpName.className = "pcp-name"
        pcpName.textContent = each.username
        pcpAvatar.src = each.avatar_url
        eachPcp.appendChild(pcpAvatar)
        eachPcp.appendChild(pcpName)
        pcpsInMeeting.appendChild(eachPcp)
      })

    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}


async function getPcpInfo(uuid, streamId, newUserName) {
  let url = `/api/pcp/:${uuid}/:${streamId}`
  let options = {
    method: "GET",
  }
  try {
    let response = await fetch(url, options);
    let result = await response.json();
    if (response.status === 200) {
      newUserName.textContent = result.pcpName
    }
  } catch (err) {
    console.log({ "error": err.message });
  }
}


var mediaRecorder;
var chunks = [];

async function captureScreen(
  mediaContraints = {
    video: true,
  }
) {
  const screenStream = await navigator.mediaDevices.getDisplayMedia(
    mediaContraints
  );
  return screenStream;
}
async function captureAudio(
  mediaContraints = {
    video: false,
    audio: true,
  }
) {
  const audioStream = await navigator.mediaDevices.getUserMedia(
    mediaContraints
  );
  return audioStream;
}


async function startRecording() {
  const screenStream = await captureScreen();
  const audioStream = await captureAudio();
  const stream = new MediaStream([
    ...screenStream.getTracks(),
    ...audioStream.getTracks(),
  ]);
  mediaRecorder = new MediaRecorder(stream);
  mediaRecorder.start();
  mediaRecorder.onstop = function () {
    Swal.fire({
      title: 'Enter a name for your recording',
      input: 'text',
      showCancelButton: true,
      confirmButtonColor: '#1158bd',
      cancelButtonColor: '#dc3545',
      confirmButtonText: 'Save',
      cancelButtonText: 'Cancel',
    }).then((result) => {
      if (result.isConfirmed) {
        const clipName = result.value;
        stream.getTracks().forEach((track) => track.stop());
        const blob = new Blob(chunks, {
          type: "video/mp4",
        });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.style.display = "none";
        a.href = url;
        a.download = clipName + ".mp4";
        document.body.appendChild(a);
        a.click();
        setTimeout(() => {
          document.body.removeChild(a);
          window.URL.revokeObjectURL(url);
        }, 100);
      }
    });
  };

  mediaRecorder.ondataavailable = function (e) {
    chunks.push(e.data);
  };
}