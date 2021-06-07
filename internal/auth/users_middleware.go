package auth

import (
	"context"
	"net/http"
	"smart_intercom_api/pkg/jwt"
	"strings"
)

type LoginContext struct {
	CookieAccess   *CookieAccess
	IsLogin        bool
}

type LoginPluginContext struct {
	Id string
}

var authCtxKey = &contextKey{"auth"}

type contextKey struct {
	name string
}

func saveLoginContext(cookieAccess *CookieAccess, r *http.Request) *http.Request {
	loginContext := LoginContext{
		CookieAccess: cookieAccess,
		IsLogin: false,
	}

	ctx := context.WithValue(r.Context(), authCtxKey, &loginContext)
	return r.WithContext(ctx)
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookieAccess := CookieAccess{
				Writer: w,
				Request: r,
				Name: "refreshToken",
			}

			_ = cookieAccess.GetToken()

			header := r.Header.Get("Authorization")

			if header == "" {
				r = saveLoginContext(&cookieAccess, r)
				next.ServeHTTP(w, r)
				return
			}

			splitToken := strings.Split(header, "Bearer ")

			if len(splitToken) != 2 {
				r = saveLoginContext(&cookieAccess, r)
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := splitToken[1]

			path := r.URL.Path

			if path == "/api" {
				err := jwt.ParseTokenForUser(tokenStr)

				if err != nil {
					http.Error(w, "Invalid token", http.StatusForbidden)
					return
				}

				loginContext := LoginContext{
					CookieAccess: &cookieAccess,
					IsLogin: true,
				}

				ctx := context.WithValue(r.Context(), authCtxKey, &loginContext)
				r = r.WithContext(ctx)
			} else {
				id, err := jwt.ParseTokenForPlugin(tokenStr)

				if err != nil {
					http.Error(w, "Invalid token", http.StatusForbidden)
					return
				}

				loginPluginContext := LoginPluginContext{
					Id: id,
				}

				ctx := context.WithValue(r.Context(), authCtxKey, &loginPluginContext)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetLoginPluginState(ctx context.Context) string {
	loginContext, _ := ctx.Value(authCtxKey).(*LoginPluginContext)

	if loginContext == nil {
		return ""
	}

	return loginContext.Id
}

func GetLoginState(ctx context.Context) bool {
	loginContext, _ := ctx.Value(authCtxKey).(*LoginContext)

	if loginContext == nil {
		return false
	}

	return loginContext.IsLogin
}

func GetCookieAccess(ctx context.Context) *CookieAccess {
	loginContext, _ := ctx.Value(authCtxKey).(*LoginContext)

	if loginContext == nil {
		return nil
	}

	return loginContext.CookieAccess
}
