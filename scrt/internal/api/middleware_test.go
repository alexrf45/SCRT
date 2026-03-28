package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEqualConstantTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"equal", "hello", "hello", true},
		{"empty both", "", "", true},
		{"differ same length", "abc", "abd", false},
		{"differ last byte", "abcdef", "abcdeg", false},
		{"different lengths", "hello", "world!", false},
		{"one empty", "", "x", false},
		{"other empty", "x", "", false},
		{"long token equal", "abc123def456ghi789jkl012mno", "abc123def456ghi789jkl012mno", true},
		{"long token single bit differ", "abc123def456ghi789jkl012mno", "abc123def456ghi789jkl012mnp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := equalConstantTime(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("equalConstantTime(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestRequireToken(t *testing.T) {
	t.Parallel()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name        string
		authHeader  string
		wantStatus  int
		wantWWWAuth bool
	}{
		{
			name:        "valid token",
			authHeader:  "Bearer " + testToken,
			wantStatus:  http.StatusOK,
			wantWWWAuth: false,
		},
		{
			name:        "no header",
			authHeader:  "",
			wantStatus:  http.StatusUnauthorized,
			wantWWWAuth: true,
		},
		{
			name:        "wrong token",
			authHeader:  "Bearer wrongtoken",
			wantStatus:  http.StatusUnauthorized,
			wantWWWAuth: true,
		},
		{
			name:        "missing bearer prefix",
			authHeader:  testToken,
			wantStatus:  http.StatusUnauthorized,
			wantWWWAuth: true,
		},
		{
			name:        "empty bearer value",
			authHeader:  "Bearer ",
			wantStatus:  http.StatusUnauthorized,
			wantWWWAuth: true,
		},
		{
			name:        "basic auth scheme",
			authHeader:  "Basic dXNlcjpwYXNz",
			wantStatus:  http.StatusUnauthorized,
			wantWWWAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := newTestServer(t, &mockBackend{})
			protected := srv.requireToken(okHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			protected.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			hasWWWAuth := w.Header().Get("WWW-Authenticate") != ""
			if hasWWWAuth != tt.wantWWWAuth {
				t.Errorf("WWW-Authenticate present = %v, want %v", hasWWWAuth, tt.wantWWWAuth)
			}
		})
	}
}
