let msg = document.getElementById("msg");
let log = document.getElementById("log");
let chat = document.getElementById('chat-content');
let slideOpen = false;
const chatButton = document.getElementById('chat-button')
const chatAlert = document.getElementById('chat-alert')
const messageHeader = document.querySelector('.message-header')
const chatBody = document.getElementById('chat-body')

messageHeader.addEventListener("click", () => {
    slideToggle()
})

connectChat();

function slideToggle() {

    if (slideOpen) {
        chat.style.display = 'none';
        slideOpen = false;
    } else {
        chat.style.display = 'block'
        chatAlert.style.display = 'none';
        // document.getElementById('msg').focus();
        slideOpen = true
    }
}

function appendLog(item) {
    // 在附加元素之前，檢查是不是可以產生卷軸
    console.log("log所在的位置")
    // 網頁被捲去的高，元素被向上滾動的高度，換句話說就是你已經走過的距離
    console.log("log.scrollTop : ", log.scrollTop)
    // 全文高，可以滾動的範圍
    console.log("log.scrollHeight : ", log.scrollHeight)
    // 可見的區域高
    console.log("log.clientHeight: ", log.clientHeight)
    // Element.scrollTop + Element.clientHeight >= Element.scrollHeight

    console.log("chatBody高度")
    console.log("chatBody.scrollTop : ", chatBody.scrollTop)
    console.log("chatBody.scrollHeight : ", chatBody.scrollHeight)
    console.log("chatBody.clientHeight : ", chatBody.clientHeight)


    // let doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    // console.log(doScroll)
    // body的overflow要從預設的visible改成auto
    log.appendChild(item);
    if (chatBody.clientHeight - log.clientHeight < 20) {
        console.log("近來這邊")
        console.log(chatBody.scrollHeight)
        console.log(chatBody.clientHeight)
        chatBody.scrollTop = chatBody.scrollHeight - chatBody.clientHeight;
        console.log("近來這邊chatBody.scrollTop : ", chatBody.scrollTop)
    }
    // if (doScroll) {
    //     log.scrollTop = log.scrollHeight - log.clientHeight;
    // }
}

function currentTime() {
    let date = new Date;
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
    console.log("近來這個聊天表格")
    chatWs.send(msg.value);
    msg.value = "";
    return false;
};

function connectChat() {
    //使用 WebSocket 的網址向 Server 開啟連結
    chatWs = new WebSocket(ChatWebsocketAddr)
    chatWs.onclose = function (evt) {
        console.log("websocket has closed")
        chatButton.disabled = true
        setTimeout(function () {
            connectChat();
        }, 1000);
    }

    chatWs.onmessage = function (evt) {
        console.log("進來onmessage")
        console.log(evt)
        console.log(evt.data)
        let messages = evt.data.split('\n');
        if (slideOpen == false) {
            chatAlert.style.display = 'block'
        }
        for (let i = 0; i < messages.length; i++) {
            let item = document.createElement("div");
            item.className = "log-item";
            item.innerText = currentTime() + " - " + messages[i];
            appendLog(item);
        }
    }

    chatWs.onerror = function (evt) {
        console.log("error: " + evt.data)
    }

    setTimeout(function () {
        if (chatWs.readyState === WebSocket.OPEN) {
            chatButton.disabled = false
        }
    }, 1000);
}

