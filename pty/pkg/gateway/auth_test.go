package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fnos-terminal/pkg/gateway"

	"github.com/stretchr/testify/require"
)

func TestParseGatewayHeaders(t *testing.T) {
	t.Run("ValidHeaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app/terminal", nil)
		req.Header.Set(gateway.HeaderTrimUserID, "1000")
		req.Header.Set(gateway.HeaderTrimUsername, "admin")
		req.Header.Set(gateway.HeaderTrimIsAdmin, "true")

		user := gateway.ParseGatewayHeaders(req)
		require.NotNil(t, user)
		require.Equal(t, "1000", user.UserID)
		require.Equal(t, "admin", user.Username)
		require.True(t, user.IsAdmin)
	})

	t.Run("MissingHeaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app/terminal", nil)
		user := gateway.ParseGatewayHeaders(req)
		require.Nil(t, user)
	})
}

func TestMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := gateway.GetUserContext(r.Context())
		require.True(t, ok)
		require.NotNil(t, user)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(user.Username))
		require.NoError(t, err)
	})

	t.Run("UnauthorizedInProd", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app/terminal/ws", nil)
		rec := httptest.NewRecorder()

		mw := gateway.Middleware(true, false)
		mw(dummyHandler).ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("ForbiddenNonAdmin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app/terminal/ws", nil)
		req.Header.Set(gateway.HeaderTrimUserID, "1001")
		req.Header.Set(gateway.HeaderTrimUsername, "guest")
		req.Header.Set(gateway.HeaderTrimIsAdmin, "false")
		rec := httptest.NewRecorder()

		mw := gateway.Middleware(true, false)
		mw(dummyHandler).ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("AllowedAdminInProd", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app/terminal/ws", nil)
		req.Header.Set(gateway.HeaderTrimUserID, "1000")
		req.Header.Set(gateway.HeaderTrimUsername, "admin")
		req.Header.Set(gateway.HeaderTrimIsAdmin, "true")
		rec := httptest.NewRecorder()

		mw := gateway.Middleware(true, false)
		mw(dummyHandler).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "admin", rec.Body.String())
	})

	t.Run("DevModeBypass", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app/terminal/ws", nil)
		rec := httptest.NewRecorder()

		mw := gateway.Middleware(true, true)
		mw(dummyHandler).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "admin", rec.Body.String())
	})
}
