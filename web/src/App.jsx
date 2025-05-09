// src/App.jsx
import React, { useState, useRef, useEffect } from "react";
import ChatRoom from "./ChatRoom.jsx";

const isLocal = window.location.hostname === "localhost";
// WebSocket URL
const WS_URL = isLocal
  ? "ws://localhost:9002/ws"
  : "wss://testingchat.duckdns.org/ws";
// REST API URL
const HTTP_URL = isLocal
  ? "http://localhost:9002"
  : "https://testingchat.duckdns.org";

export default function App() {
  const [displayName, setDisplayName] = useState("");
  const [choice, setChoice] = useState("1");
  const [roomData, setRoomData] = useState("");
  const [chatterCount, setChatterCount] = useState(0);

  const [joined, setJoined] = useState(false);
  const [messages, setMessages] = useState([]);
  const [errorMsg, setErrorMsg] = useState("");
  const [showError, setShowError] = useState(false);
  const [realRoomId, setRealRoomId] = useState(null);

  const wsRef = useRef(null);

  // Poll the chatter-count endpoint every 2 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      fetch(`${HTTP_URL}/chatter-count`)
        .then(res => res.json())
        .then(data => setChatterCount(data.count))
        .catch(err => console.warn("Polling count failed:", err));
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  const handleConnect = () => {
    if (!displayName || !roomData) {
      setErrorMsg("Fill all fields");
      setShowError(true);
      return;
    }
    setShowError(false);

    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: "init",
        id: crypto.randomUUID(),
        displayName,
        choice,
        roomData,
      }));
    };

    ws.onmessage = ev => {
      const msg = JSON.parse(ev.data);

      if (msg.type === "response" && msg.event === "joined") {
        setRealRoomId(msg.payload.roomID);
        setJoined(true);
        return;
      }

      if (msg.type === "response" && msg.event === "message") {
        setMessages(prev => [
          ...prev,
          { from: msg.payload.from, text: msg.payload.text }
        ]);
        return;
      }

      if (msg.type === "error") {
        setErrorMsg(msg.message);
        setShowError(true);
      }
    };

    ws.onerror = () => {
      setErrorMsg("WebSocket connection failed");
      setShowError(true);
    };
  };

  return (
    <div className="container">
      {!joined ? (
        <div className="card">
          <div>
            <h1>Chatroom</h1>
            <p style={{ fontSize: '0.9em', color: '#88c0d0' }}>
              Total Chatters: {chatterCount}
            </p>
            <p style={{ fontSize: '0.6em', marginTop: '5px' }}>
              written in GO by{" "}
              <a
                href="https://haadisaqib.github.io/"
                style={{ fontSize: 'inherit', color: '#88c0d0' }}
              >
                Haadi S.
              </a>
            </p>
          </div>

          <label>Display Name</label>
          <input
            value={displayName}
            onChange={e => setDisplayName(e.target.value)}
          />

          <label>Choose</label>
          <select
            value={choice}
            onChange={e => setChoice(e.target.value)}
          >
            <option value="1">Create Room</option>
            <option value="2">Join Room</option>
          </select>

          <label>
            {choice === "1" ? "Room Capacity (1–20)" : "Room ID"}
          </label>
          <input
            value={roomData}
            onChange={e => setRoomData(e.target.value)}
          />

          <button onClick={handleConnect}>Connect</button>
        </div>
      ) : (
        <ChatRoom
          displayName={displayName}
          roomId={realRoomId}
          messages={messages}
          ws={wsRef.current}
        />
      )}

      {showError && (
        <div className="modal">
          <div className="modal-content">
            <p>{errorMsg}</p>
            <button onClick={() => setShowError(false)}>Close</button>
          </div>
        </div>
      )}
    </div>
  );
}
