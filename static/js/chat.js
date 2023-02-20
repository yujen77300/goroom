let msg = document.getElementById("msg");
let log = document.getElementById("log");
let chat = document.getElementById('chat-content');
let slideOpen = false;
const chatInputButton = document.getElementById('chat-input-btn')
const chatBtn = document.getElementById('chat-btn');
const chatAlert = document.getElementById('chat-alert')
const messageHeader = document.querySelector('.message-header')
const chatBody = document.getElementById('chat-body')
let account = ""

// messageHeader.addEventListener("click", () => {
//     slideToggle()
// })

connectChat();

// ===================== chatroom展開 =====================
const chatroomBtn = document.getElementById('chatroom-btn')
chatroomBtn.addEventListener("click", () => {
    if (slideOpen == false) {
        console.log("近來按鈕")
        console.log(chatroomBtn)
        chatroomBtn.style.backgroundColor = "#171925"
        chatroomBtn.style.border = "2px solid #2e3231"
        // chatroomShowed = false
        chat.style.display = 'block'
        chatAlert.style.display = 'none';
        slideOpen = true
    } else {
        console.log("近來這個按鈕")
        chatroomBtn.style.border = "none"
        chatroomBtn.style.backgroundColor = "#1158bd"
        chat.style.display = 'none';
        slideOpen = false;
    }
})


// function slideToggle() {

//     if (slideOpen) {
//         chat.style.display = 'none';
//         slideOpen = false;
//     } else {
//         chat.style.display = 'block'
//         chatAlert.style.display = 'none';
//         slideOpen = true
//     }
// }

function appendLog(item) {
    // 在附加元素之前，檢查是不是可以產生卷軸
    console.log("log所在的位置")
    // 網頁被捲去的高，元素被向上滾動的高度，換句話說就是你已經走過的距離
    console.log("log.scrollTop : ", log.scrollTop)
    // 全文高，可以滾動的範圍
    console.log("log.scrollHeight : ", log.scrollHeight)
    // 可見的區域高
    console.log("log.clientHeight: ", log.clientHeight)


    console.log("chatBody高度")
    console.log("chatBody.scrollTop : ", chatBody.scrollTop)
    console.log("chatBody.scrollHeight : ", chatBody.scrollHeight)
    console.log("chatBody.clientHeight : ", chatBody.clientHeight)


    // body的overflow要從預設的visible改成auto
    log.appendChild(item);
    if (chatBody.clientHeight - log.clientHeight < 20) {
        console.log("近來這邊")
        console.log(chatBody.scrollHeight)
        console.log(chatBody.clientHeight)
        chatBody.scrollTop = chatBody.scrollHeight - chatBody.clientHeight;
        console.log("近來這邊chatBody.scrollTop : ", chatBody.scrollTop)
    }
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

chatBtn.addEventListener("click", () => {
    chatInputButton.click()
})

document.getElementById("form").onsubmit = function () {
    if (!chatWs) {
        return false;
    }
    if (!msg.value) {
        return false;
    }
    console.log("近來這個聊天表格")
    updateUserName()
    console.log("現在的名字是")
    console.log(account)
    chatWs.send(account + "/" + msg.value);
    msg.value = "";
    return false;
};

function connectChat() {
    fetch(
        "/api/user/auth"
    ).then(function (response) {
        return response.json();
    }).then(function (data) {
        account = data.data.name
    })
    //使用 WebSocket 的網址向 Server 開啟連結
    chatWs = new WebSocket(ChatWebsocketAddr)
    chatWs.onclose = function (e) {
        console.log("websocket has closed")
        chatInputButton.disabled = true
        setTimeout(function () {
            connectChat();
        }, 1000);
    }

    chatWs.onmessage = function (e) {
        console.log("進來onmessage")
        console.log(e)
        console.log(e.data)
        let receiveMessage = e.data.split('/');
        let accountName = receiveMessage[0]
        let messages = receiveMessage[1]
        console.log(messages)
        if (slideOpen == false) {
            chatAlert.style.display = 'block'
        }
        let item = document.createElement("div");
        item.className = "log-item";
        // item.innerText = `${accountName}` + `(${currentTime()})` + " - " + messages;
        let logUser = document.createElement("div")
        logUser.className = "log-user"
        logUser.innerText = `${accountName}` + `(${currentTime()})`;
        let logMsg = document.createElement("div")
        logMsg.className = "log-msg"
        logMsg.innerText = messages

        let myLogItem = document.createElement("div")
        myLogItem.className = "my-log-item"
        if (account == accountName) {
            myLogItem.appendChild(logUser)
            myLogItem.appendChild(logMsg)
            appendLog(myLogItem);
        } else {
            item.appendChild(logUser)
            item.appendChild(logMsg)
            appendLog(item);
        }

    }

    chatWs.onerror = function (e) {
        console.log("error: " + e.data)
    }

    setTimeout(function () {
        if (chatWs.readyState === WebSocket.OPEN) {
            chatInputButton.disabled = false
        }
    }, 1000);
}


async function updateUserName() {
    let url = "/api/user/auth"
    let options = {
        method: "GET",
    }
    try {
        let response = await fetch(url, options);
        let result = await response.json();
        if (response.status === 200) {
            account = result.data.name
        }
    } catch (err) {
        console.log({ "error": err.message });
    }
}