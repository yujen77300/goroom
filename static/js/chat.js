let msg = document.getElementById("msg");
let log = document.getElementById("log");
let chat = document.getElementById('chat-content');
let slideOpen = true;
let pcpsListOpen = false;
const chatInputButton = document.getElementById('chat-input-btn')
const chatBtn = document.getElementById('chat-btn');
const pcpsBtn = document.getElementById('pcps-btn');
const chatAlert = document.getElementById('chat-alert')
const messageHeader = document.querySelector('.message-header')
const chatBody = document.getElementById('chat-body')
const pcpsRightSection = document.getElementById('pcps-right-section')
const chatRightSection = document.getElementById('chat-right-section')
const pcpsInMeeting = document.getElementById('pcpsInMeeting')
let account = ""


// messageHeader.addEventListener("click", () => {
//     slideToggle()
// })

connectChat();

// ===================== chatroom展開與關閉 =====================
const chatroomBtn = document.getElementById('chatroom-btn')
const rightSection = document.querySelector('.right-section')
const leftSection = document.querySelector('.left-section')
const bottomLeft = document.getElementById('bottom-left')
const bottomRight = document.getElementById('bottom-right')
const videosWithChatroom = document.getElementById('videos')
chatroomBtn.addEventListener("click", () => {
    if (slideOpen == true) {
        console.log("開變成關")
        console.log(chatroomBtn)
        chatroomBtn.style.backgroundColor = "#171925"
        chatroomBtn.style.border = "2px solid #2e3231"
        rightSection.style.display = "none"
        leftSection.style.borderRight = "none"
        leftSection.style.width = "100%"
        rightSection.style.width = "0%"
        bottomLeft.style.width = "100%"
        bottomRight.style.width = "0%"
        videosWithChatroom.style.cssText = "display:flex;justify-content:center;align-items:center;gap:10px;flex-wrap:wrap;"

        slideOpen = false
    } else {
        console.log("關變成開")
        chatroomBtn.style.border = "none"
        chatroomBtn.style.backgroundColor = "#1158bd"
        pcpsBtn.style.backgroundColor = "#171925"
        rightSection.style.display = "block"
        leftSection.style.borderRight = "3px solid #242736"
        chatRightSection.style.display = "block"
        pcpsRightSection.style.display = "none"
        bottomLeft.style.width = "80%"
        bottomRight.style.width = "20%"
        leftSection.style.width = "80%"
        rightSection.style.width = "20%"
        videosWithChatroom.style.cssText = "display:flex;gap:10px;flex-wrap:wrap;justify-content:center;align-items:center;"
        chatAlert.style.display = 'none';
        slideOpen = true;
        pcpsListOpen = false;
    }
})

// ===================== 使用者清單展開與關閉 =====================
pcpsBtn.addEventListener("click", () => {
    if (pcpsListOpen == true) {
        console.log("人員清單開變成關")
        pcpsBtn.style.backgroundColor = "#171925"
        pcpsBtn.style.border = "2px solid #2e3231"
        pcpsBtn.firstElementChild.setAttribute('viewBox', '-100 -200 2300 1564');
        rightSection.style.display = "none"
        leftSection.style.borderRight = "none"
        leftSection.style.width = "100%"
        rightSection.style.width = "0%"
        bottomLeft.style.width = "100%"
        bottomRight.style.width = "0%"
        videosWithChatroom.style.cssText = "display:flex;justify-content:center;align-items:center;gap:10px;flex-wrap:wrap;"


        pcpsListOpen = false
    } else {
        console.log("人員清單關變成開")
        pcpsBtn.style.border = "none"
        pcpsBtn.style.backgroundColor = "#1158bd"
        pcpsBtn.firstElementChild.setAttribute('viewBox', '-200 -350 2300 1564');
        chatRightSection.style.display = "none"
        pcpsRightSection.style.display = "block"
        chatroomBtn.style.backgroundColor = "#171925"
        chatroomBtn.style.border = "2px solid #2e3231"
        leftSection.style.borderRight = "3px solid #242736"
        rightSection.style.display = "block"
        bottomLeft.style.width = "80%"
        bottomRight.style.width = "20%"
        leftSection.style.width = "80%"
        rightSection.style.width = "20%"
        videosWithChatroom.style.cssText = "display:flex;gap:10px;flex-wrap:wrap;justify-content:center;align-items:center;"

        pcpsListOpen = true;
        slideOpen = false;
    }
})





function appendLog(item) {
    // 在附加元素之前，檢查是不是可以產生卷軸
    console.log("log所在的位置")
    // 網頁被捲去的高，元素被向上滾動的高度，換句話說就是你已經走過的距離
    console.log("log.scrollTop : ", log.scrollTop)
    // 全文高，可以滾動的範圍
    console.log("log.scrollHeight : ", log.scrollHeight)
    // 可見的區域高
    console.log("log.clientHeight: ", log.clientHeight)




    // body的overflow要從預設的visible改成auto
    log.appendChild(item);
    if (chatBody.clientHeight - log.clientHeight < 20) {
        chatBody.scrollTop = chatBody.scrollHeight - chatBody.clientHeight;

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
    updateUserName()
    let chatInfo = {
    }
    chatInfo["account"] = account
    chatInfo["message"] = msg.value
    chatWs.send(JSON.stringify(chatInfo));
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
    chatWs = new WebSocket(ChatWebsocketAddr)
    chatWs.onclose = function (e) {
        console.log("websocket has closed")
        chatInputButton.disabled = true
        setTimeout(function () {
            connectChat();
        }, 1000);
    }

    chatWs.onopen = () => {
        console.log("進來剛開始")

        let pcps = ""
        let pcpsID = ""
        let pcpsEmail = ""
        fetch(
            "/api/user/auth"
        ).then(function (response) {
            return response.json()
        }).then(function (data) {
            let userInfo = {}
            let pcps = data.data.name
            let pcpsID = data.data.id
            let pcpsEmail = data.data.email
            userInfo["participant"] = pcps
            userInfo["participantId"] = pcpsID
            userInfo["participantEmail"] = pcpsEmail
            return userInfo
        }).then(function (userInfo) {
            console.log("我來這邊拉")
            chatWs.send(JSON.stringify(userInfo))
        })
    }

    chatWs.onmessage = function (e) {
        console.log("進來onmessage")
        console.log(e)
        console.log(e.data)
        let length = Object.keys(JSON.parse(e.data)).length
        // 如果超過三個資訊代表可能是後端的
        if (length >= 3) {
            console.log("人數測試")
            console.log(JSON.parse(e.data).participant)
            let pcpEmail = JSON.parse(e.data).participantEmail

            eachPcp = document.createElement("div")
            eachPcp.className = "each-pcp"
            eachPcp.id = JSON.parse(e.data).participantId
            pcpAvatar = document.createElement("img")
            pcpAvatar.className = "pcp-avatar"
            pcpAvatar.alt = JSON.parse(e.data).participant
            pcpName = document.createElement("div")
            pcpName.className = "pcp-name"
            pcpName.textContent = JSON.parse(e.data).participant
            getPcpAvatar(pcpEmail, pcpAvatar)
            eachPcp.appendChild(pcpAvatar)
            eachPcp.appendChild(pcpName)
            console.log("測試一下")
            console.log(pcpsInMeeting)
            pcpsInMeeting.appendChild(eachPcp)
            return

        } else {
            // if (e.data != "") {
            let accountName = JSON.parse(e.data).account
            let messages = JSON.parse(e.data).message
            console.log("聊天測試")
            console.log(messages)
            if (slideOpen == false) {
                chatAlert.style.display = 'block'
            }
            let item = document.createElement("div");
            item.className = "log-item";
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
        // }


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

async function getPcpAvatar(pcpEmail, pcpAvatar) {
    console.log("我來測試大頭貼")
    console.log(pcpEmail)
    let url = `/api/avatar/:${pcpEmail}`
    let options = {
        method: "GET",
    }
    try {
        let response = await fetch(url, options);
        let result = await response.json();
        if (response.status === 200) {
            console.log("測試url是否正確")
            console.log(result.pcpAvatarUrl)
            console.log(pcpAvatar)
            pcpAvatar.src = `${result.pcpAvatarUrl}`
        }
    } catch (err) {
        console.log({ "error": err.message });
    }
}