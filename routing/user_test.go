package routing

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginRoute(t *testing.T) {
	router := setupTestEnvironment()
	defer db.Close()

	var jsonStr = []byte(`{"loginName": "admin1", "password": "password"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var jsonStr2 = []byte(`{"loginName": "user1", "password": "password"}`)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr2))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var jsonStr3 = []byte(`{"loginName": "admin1", "password": "pw"}`)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr3))
	router.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)

	var jsonStr4 = []byte(`{"loginName": "admin", "password": "password"}`)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr4))
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestLogoutRoute(t *testing.T) {
	router := setupTestEnvironment()
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logout", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreateUserRoute(t *testing.T) {
	router := setupTestEnvironment()
	defer db.Close()

	var jsonStr = []byte(`{"loginName": "admin1", "password": "password"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonStr))
	router.ServeHTTP(w, req)

	var jsonStr2 = []byte(`{"name": "admin2", "password": "password2", "admin": true, "mail": "admin2@example.com"}`)
	req, _ = http.NewRequest("POST", "/user/create", bytes.NewBuffer(jsonStr2))
	req.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
}
