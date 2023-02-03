let msg = document.getElementById("msg");
let log = document.getElementById("log");
let chat = document.getElementById('chat-content');
const messageHeader = document.querySelector('.message-header')

messageHeader.addEventListener("click", () => {
    slideToggle()
})

var slideOpen = false;

function slideToggle() {

    if (slideOpen) {
        chat.style.display = 'none';
        slideOpen = false;
    } else {
        chat.style.display = 'block'
        document.getElementById('chat-alert').style.display = 'none';
        document.getElementById('msg').focus();
        slideOpen = true
    }
}

function appendLog(item) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
        log.scrollTop = log.scrollHeight - log.clientHeight;
    }
}

function currentTime() {
    var date = new Date;
    hour = date.getHours();
    minute = date.getMinutes();
    if (hour < 10) {
        hour = "0" + hour
    }
    if (minute < 10) {
        minute = "0" + minute
    }
    return hour + ":" + minute
}

document.getElementById("form").onsubmit = function () {
    if (!chatWs) {
        return false;
    }
    if (!msg.value) {
        return false;
    }
    chatWs.send(msg.value);
    msg.value = "";
    return false;
};

function connectChat() {
    //使用 WebSocket 的網址向 Server 開啟連結
    chatWs = new WebSocket(ChatWebsocketAddr)
    //關閉後執行的動作，指定一個 function 會在連結中斷後執行
    chatWs.onclose = function (evt) {
        console.log("websocket has closed")
        document.getElementById('chat-button').disabled = true
        setTimeout(function () {
            connectChat();
        }, 1000);
    }

    chatWs.onmessage = function (evt) {
        var messages = evt.data.split('\n');
        if (slideOpen == false) {
            document.getElementById('chat-alert').style.display = 'block'
        }
        for (var i = 0; i < messages.length; i++) {
            var item = document.createElement("div");

            item.innerText = currentTime() + " - " + messages[i];
            appendLog(item);
        }
    }

    chatWs.onerror = function (evt) {
        console.log("error: " + evt.data)
    }

    setTimeout(function () {
        if (chatWs.readyState === WebSocket.OPEN) {
            document.getElementById('chat-button').disabled = false
        }
    }, 1000);
}

connectChat();