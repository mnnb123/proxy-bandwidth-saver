package webapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// WebAuth provides HTTP Basic Auth middleware for the web admin panel.
type WebAuth struct {
	getCredentials func() (username, password string)
}

func NewWebAuth(getCredentials func() (username, password string)) *WebAuth {
	return &WebAuth{getCredentials: getCredentials}
}

// Wrap returns an http.Handler that requires Basic Auth before delegating to next.
func (wa *WebAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password := wa.getCredentials()
		// If no credentials configured, allow access
		if username == "" && password == "" {
			next.ServeHTTP(w, r)
			return
		}

		reqUser, reqPass, ok := r.BasicAuth()
		if !ok || !secureCompare(reqUser, username) || !secureCompare(reqPass, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Proxy Bandwidth Saver"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func secureCompare(a, b string) bool {
	ha := sha256.Sum256([]byte(a))
	hb := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
}
