/*  src/App.jsx  – complete file  */

import { useState, useRef, useEffect } from "react";

/* ------------------------------------------------------------------ */
/*  CONFIG                                                            */
/* ------------------------------------------------------------------ */
const WS_PORT = 9002;
const WS_PATH = "/ws";
const wsURL   = `ws://${window.location.hostname}:${WS_PORT}${WS_PATH}`;

/* ------------------------------------------------------------------ */
/*  MAIN COMPONENT                                                    */
/* ------------------------------------------------------------------ */
export default function App() {
  /* ---------- form state ---------- */
  const [name,      setName]      = useState("");
  const [choice,    setChoice]    = useState("1");      // 1=create, 2=join
  const [roomData,  setRoomData]  = useState("");
  const [errorMsg,  setErrorMsg]  = useState("");

  /* ---------- connection / chat ---------- */
  const [connected, setConnected] = useState(false);
  const [messages,  setMessages]  = useState([]);

  const socketRef = useRef(null);
  const uuidRef   = useRef("");

  /* ---------- helpers ---------- */
  const addMsg = (from, text, cls = "") =>
    setMessages(m => [...m, { id: crypto.randomUUID(), from, text, cls }]);

  const reset = () => {
    if (socketRef.current) socketRef.current.close();
    socketRef.current = null;
    uuidRef.current   = "";
    setConnected(false);
    setMessages([]);
  };

  /* ------------------------------------------------------------------ */
  /*  CONNECT BUTTON HANDLER                                            */
  /* ------------------------------------------------------------------ */
  const handleConnect = () => {
    /* ---- client‑side field checks ---- */
    if (!name.trim() || !roomData.trim()) {
      setErrorMsg("Fill all fields");
      return;
    }

    if (choice === "1") {                           // create‑room path
      const cap = parseInt(roomData, 10);
      if (isNaN(cap) || cap < 1 || cap > 20) {
        setErrorMsg("Capacity must be 1‑20");
        return;
      }
    }

    setErrorMsg("");                                // clear previous

    /* ---- open WebSocket ---- */
    uuidRef.current = crypto.randomUUID();
    socketRef.current = new WebSocket(wsURL);

    socketRef.current.onopen = () => {
      socketRef.current.send(JSON.stringify({
        type: "init",
        id: uuidRef.current,
        displayName: name.trim(),
        choice,
        roomData: roomData.trim()
      }));
    };

    socketRef.current.onmessage = ev => {
      const m = JSON.parse(ev.data);

      /* ---- system messages ---- */
      if (m.from === "system") {
        switch (m.text) {
          case "room-not-found":
            setErrorMsg("Room does not exist");
            return reset();
          case "room-full":
            setErrorMsg("Room is already full");
            return reset();
          case "invalid-capacity":
            setErrorMsg("Capacity must be 1‑20");
            return reset();
          case "duplicate-uuid":
            setErrorMsg("Duplicate session detected");
            return reset();
        }

        if (m.text.startsWith("joined-room")) {
          const id = m.text.split(" ")[1];
          addMsg("system", `You joined room #${id}`, "system");
          setConnected(true);
        } else {
          addMsg("system", m.text, "system");
        }
        return;
      }

      /* ---- normal chat ---- */
      const cls = m.from === name ? "me" : "";
      addMsg(m.from, m.text, cls);
    };

    socketRef.current.onclose = () => {
      if (connected) {
        setErrorMsg("Disconnected from server");
        reset();
      }
    };
  };

  /* ------------------------------------------------------------------ */
  /*  CHAT INPUT HANDLER                                                */
  /* ------------------------------------------------------------------ */
  const handleKey = e => {
    if (e.key === "Enter" && socketRef.current?.readyState === 1) {
      const text = e.target.value.trim();
      if (text) {
        socketRef.current.send(JSON.stringify({ type: "message", text }));
        e.target.value = "";
      }
    }
  };

  /* cleanup when component unmounts */
  useEffect(() => () => reset(), []);

  /* ------------------------------------------------------------------ */
  /*  RENDER                                                            */
  /* ------------------------------------------------------------------ */
  return (
    <div id="card">
      <h1>Tchat</h1>

      {!connected ? (
        /* -------------- SETUP FORM -------------- */
        <div id="setup">
          <label>Display Name</label>
          <input
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Alice"
          />

          <label>Choose</label>
          <select value={choice} onChange={e => setChoice(e.target.value)}>
            <option value="1">Create Room</option>
            <option value="2">Join Room</option>
          </select>

          <label>
            {choice === "1" ? "Room Capacity (1‑20)" : "Room ID"}
          </label>
          <input
            value={roomData}
            onChange={e => setRoomData(e.target.value)}
            placeholder={choice === "1" ? "3" : "12345"}
          />

          <button onClick={handleConnect}>Connect</button>

          {/* inline error message */}
          {errorMsg && <p className="error">{errorMsg}</p>}
        </div>
      ) : (
        /* -------------- CHAT UI -------------- */
        <div id="chat">
          <div id="messages">
            {messages.map(m => (
              <p key={m.id} className={m.cls || undefined}>
                {m.cls === "system" ? m.text : `${m.from}: ${m.text}`}
              </p>
            ))}
          </div>
          <input
            id="msgInput"
            placeholder="Type message & hit Enter"
            onKeyDown={handleKey}
            autoFocus
          />
        </div>
      )}
    </div>
  );
}
