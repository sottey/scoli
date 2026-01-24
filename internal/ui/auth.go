package ui

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envUIPassword     = "NOLDERMD_UI_PASSWORD"
	envUICookieSecret = "NOLDERMD_UI_COOKIE_SECRET"
	sessionCookieName = "scoli_session"
)

const defaultSessionTTL = 30 * 24 * time.Hour

type authConfig struct {
	enabled      bool
	password     []byte
	cookieSecret []byte
	cookieName   string
	sessionTTL   time.Duration
}

func loadAuthConfig() authConfig {
	password := os.Getenv(envUIPassword)
	if password == "" {
		return authConfig{enabled: false}
	}

	secret := os.Getenv(envUICookieSecret)
	secretBytes := []byte(secret)
	if secret == "" {
		var err error
		secretBytes, err = randomBytes(32)
		if err != nil {
			log.Printf("ui auth disabled: unable to generate cookie secret: %v", err)
			return authConfig{enabled: false}
		}
	}

	return authConfig{
		enabled:      true,
		password:     []byte(password),
		cookieSecret: secretBytes,
		cookieName:   sessionCookieName,
		sessionTTL:   defaultSessionTTL,
	}
}

func (a authConfig) middleware(next http.Handler) http.Handler {
	if !a.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAuthExemptPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if a.validSession(r) {
			next.ServeHTTP(w, r)
			return
		}

		redirectToLogin(w, r)
	})
}

func (a authConfig) validSession(r *http.Request) bool {
	cookie, err := r.Cookie(a.cookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	return verifySessionToken(cookie.Value, a.cookieSecret)
}

func (a authConfig) issueSessionCookie(w http.ResponseWriter, r *http.Request) error {
	token, err := newSessionToken(a.cookieSecret, a.sessionTTL)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
		MaxAge:   int(a.sessionTTL.Seconds()),
	})
	return nil
}

func (a authConfig) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
		MaxAge:   -1,
	})
}

func checkPassword(expected []byte, provided string) bool {
	if len(expected) == 0 {
		return false
	}
	providedBytes := []byte(provided)
	if len(expected) != len(providedBytes) {
		return false
	}
	return subtle.ConstantTimeCompare(expected, providedBytes) == 1
}

func isAuthExemptPath(path string) bool {
	switch path {
	case "/login", "/logout":
		return true
	default:
		return false
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	next := sanitizeNextPath(r.URL.RequestURI())
	target := "/login"
	if next != "/" && next != "" {
		target = "/login?next=" + url.QueryEscape(next)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func sanitizeNextPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/"
	}
	if strings.HasPrefix(path, "//") || strings.Contains(path, "\\") {
		return "/"
	}
	return path
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		return false
	}
	if idx := strings.Index(proto, ","); idx >= 0 {
		proto = proto[:idx]
	}
	return strings.EqualFold(strings.TrimSpace(proto), "https")
}

func newSessionToken(secret []byte, ttl time.Duration) (string, error) {
	if len(secret) == 0 {
		return "", errors.New("missing secret")
	}
	nonceBytes, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	expiresAt := time.Now().Add(ttl).Unix()
	payload := []byte(strconv.FormatInt(expiresAt, 10) + ":" + nonce)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	sig := mac.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return payloadB64 + "." + sigB64, nil
}

func verifySessionToken(token string, secret []byte) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expected := mac.Sum(nil)
	if subtle.ConstantTimeCompare(sig, expected) != 1 {
		return false
	}

	payloadParts := strings.SplitN(string(payload), ":", 2)
	if len(payloadParts) != 2 {
		return false
	}
	expiration, err := strconv.ParseInt(payloadParts[0], 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expiration {
		return false
	}
	return true
}

func randomBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
