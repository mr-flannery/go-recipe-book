package auth

import (
	"net/http"
	"sync"
)

// Mock user store
var (
	users = map[string]string{ // username: password
		"alice": "password1",
		"bob":   "password2",
	}
	sessions = map[string]string{} // sessionID: username
	mu       sync.Mutex
)

func Authenticate(username, password string) bool {
	mu.Lock()
	defer mu.Unlock()
	if pw, ok := users[username]; ok && pw == password {
		return true
	}
	return false
}

func SetSession(w http.ResponseWriter, username string) {
	// TODO: Use secure random session IDs
	cookie := &http.Cookie{Name: "session", Value: username, Path: "/"}
	http.SetCookie(w, cookie)
	mu.Lock()
	sessions[username] = username
	mu.Unlock()
}

func GetUser(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", false
	}
	mu.Lock()
	defer mu.Unlock()
	username, ok := sessions[cookie.Value]
	return username, ok
}

func InvalidateSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return
	}
	mu.Lock()
	delete(sessions, cookie.Value)
	mu.Unlock()
	// Clear the cookie
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
}

func IsSessionValid(r *http.Request) bool {
	_, ok := GetUser(r)
	return ok
}
