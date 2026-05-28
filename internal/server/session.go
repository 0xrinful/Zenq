package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const sessionCookieName = "z_session"

var errInvalidSession = errors.New("server: invalid session")

type Session struct {
	UserID int64 `json:"user_id"`
}

func SetSession(w http.ResponseWriter, r *http.Request, session *Session, secret string) {
	payload, _ := json.Marshal(session)
	sig := signSession(payload, secret)

	value := base64.StdEncoding.EncodeToString(payload) + "." + base64.StdEncoding.EncodeToString(sig)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func GetSession(r *http.Request, secret string) (*Session, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return nil, errInvalidSession
	}

	payload, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errInvalidSession
	}

	sig, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidSession
	}

	expected := signSession(payload, secret)
	if !hmac.Equal(sig, expected) {
		return nil, errInvalidSession
	}

	var session Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return nil, errInvalidSession
	}

	return &session, nil
}

func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func signSession(payload []byte, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return mac.Sum(nil)
}
