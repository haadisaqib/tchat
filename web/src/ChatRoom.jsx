// src/ChatRoom.jsx
import { useState, useEffect, useRef } from "react";

export default function ChatRoom({ ws, displayName, roomId }) {
  const [messages, setMessages] = useState([]);
  const inputRef = useRef();
  const endRef = useRef();

  useEffect(() => {
    if (!ws) return;
    const onMsg = ev => {
      const msg = JSON.parse(ev.data);
      if (msg.type === "response" && (msg.event === "history" || msg.event === "message")) {
        setMessages(prev =>
          [...prev, { from: msg.payload.from, text: msg.payload.text }]
            .slice(-100) 
        );
      }
    };
    ws.addEventListener("message", onMsg);
    return () => ws.removeEventListener("message", onMsg);
  }, [ws]);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const sendMessage = () => {
    const text = inputRef.current.value.trim();
    if (!text || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ type: "message", text }));
    inputRef.current.value = "";
  };

  return (
    <div className="chat-wrapper">
      <header className="chat-header">
        Room #{roomId} — {displayName}
      </header>
      <div className="messages">
        {messages.map((m, i) => (
          <p key={i} className={m.from === displayName ? "me" : ""}>
            <strong>{m.from}:</strong> {m.text}
          </p>
        ))}
        <div ref={endRef} />
      </div>
      <div className="chat-input">
        <input
          ref={inputRef}
          onKeyDown={e => e.key === "Enter" && sendMessage()}
          placeholder="Type a message..."
        />
        <button onClick={sendMessage}>Send</button>
      </div>
    </div>
  );
}
