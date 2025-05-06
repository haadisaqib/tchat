// src/App.jsx
import React, { useState, useRef, useEffect } from "react";
import ChatRoom from "./ChatRoom.jsx";

// 1) WS URL (unchanged)
const WS_URL = `ws://${window.location.hostname}:9002/ws`;

export default function App() {
  // Step 1: form state
  const [displayName, setDisplayName] = useState("");
  const [choice, setChoice]           = useState("1");
  const [roomData, setRoomData]       = useState("");

  // Step 2: app state
  const [joined,    setJoined]    = useState(false);
  const [messages,  setMessages]  = useState([]);      // all chat lines
  const [errorMsg,  setErrorMsg]  = useState("");
  const [showError, setShowError] = useState(false);

  // **NEW**: real room ID from server (not capacity)
  const [realRoomId, setRealRoomId] = useState(null);

  const wsRef = useRef(null);

  // --- NEW: fetch full history via HTTP ---
  const loadChatHistory = async (roomId) => {
    try {
      const res = await fetch(`${HTTP_HISTORY_URL}?roomId=${roomId}`);
      if (!res.ok) throw new Error("no history");
      const data = await res.json(); // array of {Sender,Message,...}
      // normalize into the shape your ChatRoom expects:
      const hist = data.map(m => ({ from: m.Sender, text: m.Message }));
      setMessages(hist);
    } catch (e) {
      console.warn("Could not load history:", e);
    }
  };

  // Step 4: connect / init payload
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
        type:        "init",
        id:          crypto.randomUUID(),
        displayName,
        choice,
        roomData
      }));
    };

    ws.onmessage = ev => {
      const msg = JSON.parse(ev.data);

      // CASE: joined → swap to ChatRoom
      if (msg.type === "response" && msg.event === "joined") {
        setRealRoomId(msg.payload.roomID);
        setJoined(true);
        return;
      }

      // CASE: history → load past messages
      if (msg.type === "response" && msg.event === "history") {
        loadChatHistory(msg.payload.from, msg.payload.text);
        return;
      }

      // CASE: new chat message
      if (msg.type === "response" && msg.event === "message") {
        loadChatHistory(msg.payload.from, msg.payload.text);
        return;
      }

      // CASE: error → modal
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

  // Render: before join → form; after join → ChatRoom
  return (
    <div className="container">
      {!joined ? (
        <div className="card">
          <h1>Tchat</h1>

          <label>Display Name</label>
          <input
            value={displayName}
            onChange={e => setDisplayName(e.target.value)}
            placeholder="Alice"
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
            {choice === "1"
              ? "Room Capacity (1–20)" 
              : "Room ID"}
          </label>
          <input
            value={roomData}
            onChange={e => setRoomData(e.target.value)}
            placeholder={choice === "1" ? "3" : "12345"}
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

      {/* error modal */}
      {showError && (
        <div className="modal">
          <div className="modal-content">
            <p>{errorMsg}</p>
            <button onClick={() => setShowError(false)}>
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
