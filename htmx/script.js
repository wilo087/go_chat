let ws;
let currentUsername;

function setupWebSocket() {
    ws = new WebSocket("ws://localhost:8080/ws");

    ws.onmessage = function (event) {
        let msg = JSON.parse(event.data);
        let msgList = document.getElementById("messages");
        let newMsg = document.createElement("li");

        let avatar = document.createElement("div");
        avatar.className = "avatar";
        avatar.textContent = msg.username.slice(0, 1).toUpperCase();

        let messageContent = document.createElement("div");
        messageContent.textContent = msg.content;

        let timeSpan = document.createElement("span");
        timeSpan.className = "message-time";
        timeSpan.textContent = new Date(msg.time).toLocaleTimeString();

        if (msg.username === currentUsername) {
            newMsg.className = "message-user";
            newMsg.appendChild(timeSpan);
            newMsg.appendChild(messageContent);
        } else {
            newMsg.className = "message-others";
            newMsg.appendChild(avatar);
            newMsg.appendChild(messageContent);
            newMsg.appendChild(timeSpan);
        }

        msgList.appendChild(newMsg);
        msgList.scrollTop = msgList.scrollHeight; // Auto-scroll to the latest message
    }
}

function sendMessage() {
    let messageInput = document.getElementById("message");

    if (messageInput.value.trim() === "") {
        return;
    }

    let message = {
        username: currentUsername,
        content: messageInput.value,
        time: new Date()
    };

    ws.send(JSON.stringify(message));
    messageInput.value = '';
}

function onEnterPressSendMsg(event) {
    if (event.keyCode === 13) {
        sendMessage();
    }
}

function onEnterPressEnterChat(event) {
    if (event.keyCode === 13) {
        openChat();
    }
}

function openChat() {
    let usernameInput = document.getElementById("modal-username");

    if (usernameInput.value.trim() !== "") {
        currentUsername = usernameInput.value.trim();
        document.getElementById("username-modal").style.display = "none";
        document.getElementById("chat-container").style.display = "flex";
        setupWebSocket();
    }
}