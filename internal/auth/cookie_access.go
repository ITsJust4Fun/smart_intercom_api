package auth

import (
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type CookieAccess struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Name       string
	Token      string
	Expires    time.Time
}

func (cookieAccess *CookieAccess) SetToken() {
	http.SetCookie(cookieAccess.Writer, &http.Cookie{
		Name:     cookieAccess.Name,
		Value:    cookieAccess.Token,
		HttpOnly: true,
		Path:     "/",
		Expires:  cookieAccess.Expires,
	})
}

func (cookieAccess *CookieAccess) GetToken() error {
	if cookieAccess.Request == nil {
		return errors.New("There is no request")
	}

	c, err := cookieAccess.Request.Cookie(cookieAccess.Name)

	if err != nil {
		return errors.New("There is no token in cookies")
	}

	cookieAccess.Token = c.Value
	cookieAccess.Expires = c.Expires

	return nil
}
