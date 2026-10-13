package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/davidyoung/savers-app/backend/internal/auth"
)

type ctxKey string

const UserIDKey ctxKey = "userID"
const UsernameKey ctxKey = "username"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, `{"error":"未授权"}`, http.StatusUnauthorized)
			return
		}
		claims, err := auth.ParseToken(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			http.Error(w, `{"error":"令牌无效或已过期"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}
