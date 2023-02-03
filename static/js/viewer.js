let viewerCount = document.getElementById("viewer-count");


function connectViewer() {
  viewerWs = new WebSocket(ViewerWebsocketAddr);
  // console.log("測試測試viewerCount")
  // console.log(viewerCount)
  // console.log(viewerWs)
  viewerWs.onclose = function (evt) {
    // console.log("websocket has closed");
    viewerCount.innerHTML = "0";
    setTimeout(function () {
      connectViewer();
    }, 1000);
  }

  viewerWs.onmessage = function (evt) {
    // console.log("測試測試evt.data")
    // console.log(evt)
    d = evt.data
    // console.log(d)
    if (d === parseInt(d, 10)) {
      return
    }
    viewerCount.innerHTML = d;
  }

  viewerWs.onerror = function (evt) {
    console.log("error: " + evt.data)
  }
}

connectViewer();
