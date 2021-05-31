package auth

import (
	"context"
	"net/http"

	"smart_intercom_api/pkg/jwt"
)

var userCtxKey = &contextKey{"auth"}

type contextKey struct {
	name string
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := header
			err := jwt.ParseTokenForUser(tokenStr)

			if err != nil {
				http.Error(w, "Invalid token", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, "done")

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) bool {
	raw, _ := ctx.Value(userCtxKey).(string)
	return raw == "done"
}
