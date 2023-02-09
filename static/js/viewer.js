let viewerCount = document.getElementById("viewer-count");


function connectViewer() {
  viewerWs = new WebSocket(ViewerWebsocketAddr);
  viewerWs.onclose = function (e) {
    console.log("ViewerWebsocket has closed");
    viewerCount.innerHTML = "0";
    setTimeout(function () {
      connectViewer();
    }, 1000);
  }

  // 接收handlers / room.go 的 roomViewerConn()
  viewerWs.onmessage = function (e) {
    d = e.data
    // let msg = JSON.parse(e.data);
    // console.log(msg)
    if (d === parseInt(d, 10)) {
      return
    }
    viewerCount.innerHTML = d;
  }

  viewerWs.onerror = function (e) {
    console.log("error: " + e.data)
  }
}

connectViewer();
