let viewerCount = document.querySelectorAll(".viewer-count");


function connectViewer() {
  viewerWs = new WebSocket(ViewerWebsocketAddr);
  viewerWs.onclose = function (e) {
    console.log("ViewerWebsocket has closed");
    viewerCount[0].innerHTML = "0";
    viewerCount[1].innerHTML = "0";
    setTimeout(function () {
      connectViewer();
    }, 1000);
  }

  viewerWs.onmessage = function (e) {
    amount = e.data
    if (amount === parseInt(amount, 10)) {
      return
    }
    viewerCount[0].innerHTML = amount;
    viewerCount[1].innerHTML = amount;
  }

  viewerWs.onerror = function (e) {
    console.log("error: " + e.data)
  }
}

connectViewer();
