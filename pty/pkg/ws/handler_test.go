package ws_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fnos-terminal/pkg/gateway"
	"fnos-terminal/pkg/ws"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestTerminalHandlerPing(t *testing.T) {
	handler := ws.NewTerminalHandler(true)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := gateway.WithUserContext(r.Context(), &gateway.UserContext{
			UserID:   "dev-1000",
			Username: "admin",
			IsAdmin:  true,
		})
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer func() {
		closeErr := conn.Close()
		require.NoError(t, closeErr)
	}()

	// Send ping
	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
	require.NoError(t, err)

	// Wait for pong or terminal output
	receivedPong := false
	for i := 0; i < 5; i++ {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, readErr := conn.ReadMessage()
		if readErr != nil {
			break
		}
		if strings.Contains(string(msg), "pong") {
			receivedPong = true
			break
		}
	}
	require.True(t, receivedPong, "Expected to receive pong control message")
}
