package gateway

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const (
	GatewayUserKey contextKey = "gatewayUser"

	HeaderTrimUserID   = "X-Trim-Userid"
	HeaderTrimIsAdmin  = "X-Trim-Isadmin"
	HeaderTrimUsername = "X-Trim-Username"
)

var (
	ErrUnauthorized = errors.New("unauthorized: missing or invalid gateway authentication")
	ErrForbidden    = errors.New("forbidden: administrator privileges required")
)

type UserContext struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}

func ParseGatewayHeaders(r *http.Request) *UserContext {
	userID := strings.TrimSpace(r.Header.Get(HeaderTrimUserID))
	username := strings.TrimSpace(r.Header.Get(HeaderTrimUsername))
	isAdminStr := strings.TrimSpace(strings.ToLower(r.Header.Get(HeaderTrimIsAdmin)))
	isAdmin := isAdminStr == "true" || isAdminStr == "1"

	if userID == "" && username == "" {
		return nil
	}

	return &UserContext{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
	}
}

func GetUserContext(ctx context.Context) (*UserContext, bool) {
	val := ctx.Value(GatewayUserKey)
	if val == nil {
		return nil, false
	}
	user, ok := val.(*UserContext)
	if !ok {
		return nil, false
	}
	return user, true
}

func WithUserContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, GatewayUserKey, user)
}

func Middleware(requireAdmin bool, devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := ParseGatewayHeaders(r)

			if user == nil {
				if devMode {
					user = &UserContext{
						UserID:   "dev-1000",
						Username: "admin",
						IsAdmin:  true,
					}
				} else {
					http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
					return
				}
			}

			if requireAdmin && !user.IsAdmin {
				http.Error(w, ErrForbidden.Error(), http.StatusForbidden)
				return
			}

			ctx := WithUserContext(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
