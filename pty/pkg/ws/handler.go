package ws

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"fnos-terminal/pkg/gateway"
	"fnos-terminal/pkg/pty"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // Origin checked by fnOS gateway
	},
}

type ControlMessage struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
}

type TerminalHandler struct {
	devMode bool
}

func NewTerminalHandler(devMode bool) *TerminalHandler {
	return &TerminalHandler{devMode: devMode}
}

func (h *TerminalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Failed to upgrade websocket: %v", err)
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// already closed
		}
	}()

	user, ok := gateway.GetUserContext(r.Context())
	username := ""
	if ok && user != nil {
		username = user.Username
	}

	cols := uint16(80)
	rows := uint16(24)

	q := r.URL.Query()
	if cStr := q.Get("cols"); cStr != "" {
		if c, parseErr := strconv.ParseUint(cStr, 10, 16); parseErr == nil && c > 0 {
			cols = uint16(c)
		}
	}
	if rStr := q.Get("rows"); rStr != "" {
		if rVal, parseErr := strconv.ParseUint(rStr, 10, 16); parseErr == nil && rVal > 0 {
			rows = uint16(rVal)
		}
	}

	session, err := pty.StartSession(username, cols, rows)
	if err != nil {
		log.Printf("[WS] Failed to start pty session for %q: %v", username, err)
		errPayload := []byte("\r\n\x1b[31m[Error] Failed to start terminal session: " + err.Error() + "\x1b[0m\r\n")
		if writeErr := conn.WriteMessage(websocket.TextMessage, errPayload); writeErr != nil {
			log.Printf("[WS] Failed to write error to client: %v", writeErr)
		}
		return
	}
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			log.Printf("[WS] Error closing pty session: %v", closeErr)
		}
	}()

	var closeOnce sync.Once
	closeBoth := func() {
		closeOnce.Do(func() {
			if closeErr := session.Close(); closeErr != nil {
				// handled
			}
			if closeErr := conn.Close(); closeErr != nil {
				// handled
			}
		})
	}

	var wsWriteMu sync.Mutex
	safeWrite := func(msgType int, data []byte) error {
		wsWriteMu.Lock()
		defer wsWriteMu.Unlock()
		if setDeadlineErr := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); setDeadlineErr != nil {
			return setDeadlineErr
		}
		return conn.WriteMessage(msgType, data)
	}

	// Goroutine 1: PTY -> WebSocket
	go func() {
		defer closeBoth()
		buf := make([]byte, 4096)
		for {
			n, readErr := session.Read(buf)
			if n > 0 {
				if writeErr := safeWrite(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					log.Printf("[WS] Failed writing pty output to websocket: %v", writeErr)
					return
				}
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) {
					log.Printf("[WS] PTY read finished: %v", readErr)
				}
				// Notify frontend that session exited gracefully
				exitNotice := []byte(`{"type":"exit","code":0}`)
				if writeErr := safeWrite(websocket.TextMessage, exitNotice); writeErr != nil {
					// client might already be gone
				}
				return
			}
		}
	}()

	// Goroutine 2: WebSocket -> PTY (current goroutine)
	defer closeBoth()
	for {
		msgType, payload, readErr := conn.ReadMessage()
		if readErr != nil {
			return
		}

		if msgType == websocket.TextMessage && len(payload) > 0 && payload[0] == '{' {
			var ctrl ControlMessage
			if jsonErr := json.Unmarshal(payload, &ctrl); jsonErr == nil {
				switch ctrl.Type {
				case "resize":
					if ctrl.Cols > 0 && ctrl.Rows > 0 {
						if resizeErr := session.Resize(ctrl.Cols, ctrl.Rows); resizeErr != nil {
							log.Printf("[WS] Failed to resize pty: %v", resizeErr)
						}
					}
					continue
				case "ping":
					pongData := []byte(`{"type":"pong"}`)
					if pongErr := safeWrite(websocket.TextMessage, pongData); pongErr != nil {
						log.Printf("[WS] Failed to reply pong: %v", pongErr)
					}
					continue
				}
			}
		}

		if len(payload) > 0 {
			if _, writeErr := session.Write(payload); writeErr != nil {
				log.Printf("[WS] Failed to write client input to pty: %v", writeErr)
				return
			}
		}
	}
}
