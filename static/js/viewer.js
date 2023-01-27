let viewerCount = document.getElementById("viewer-count");


function connectViewer() {
  viewerWs = new WebSocket(ViewerWebsocketAddr);
  console.log("測試測試")
  console.log(viewerCount)
  viewerWs.onclose = function (evt) {
    console.log("websocket has closed");
    viewerCount.innerHTML = "0";
    setTimeout(function () {
      connectViewer();
    }, 1000);
  }

  viewerWs.onmessage = function (evt) {
    d = evt.data
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
