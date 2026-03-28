package api

import (
	"net/http"
	"strings"
)

// requireToken enforces Bearer token authentication.
// The token is compared in constant time to avoid timing attacks.
func (s *Server) requireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		provided := strings.TrimPrefix(auth, "Bearer ")
		if provided == "" || provided == auth || !equalConstantTime(provided, s.p.Token) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="scrt"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// equalConstantTime compares two strings in constant time to prevent timing attacks.
func equalConstantTime(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
