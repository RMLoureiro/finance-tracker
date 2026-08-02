package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"finance-tracker/server/internal/api"
	"finance-tracker/server/internal/testutil"
)

type client struct {
	t    *testing.T
	http *http.Client
	base string
}

func newTestClient(t *testing.T) *client {
	st := testutil.StartStore(t)
	ts := httptest.NewServer(api.New(st, false).Handler())
	t.Cleanup(ts.Close)
	jar, _ := cookiejar.New(nil)
	return &client{t: t, http: &http.Client{Jar: jar}, base: ts.URL}
}

func (c *client) do(method, path string, body any) (*http.Response, map[string]any) {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, c.base+path, &buf)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

// doList é como do, mas para endpoints que devolvem um array JSON.
func (c *client) doList(method, path string) (*http.Response, []any) {
	c.t.Helper()
	req, err := http.NewRequest(method, c.base+path, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out []any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

func (c *client) mustStatus(resp *http.Response, want int, out map[string]any) {
	c.t.Helper()
	if resp.StatusCode != want {
		c.t.Fatalf("status %d, esperado %d (resposta: %v)", resp.StatusCode, want, out)
	}
}

func TestSetupLoginLogoutFlow(t *testing.T) {
	c := newTestClient(t)

	// antes do setup
	resp, out := c.do("GET", "/api/setup", nil)
	c.mustStatus(resp, 200, out)
	if out["needsSetup"] != true {
		t.Fatalf("needsSetup devia ser true, veio %v", out)
	}
	resp, out = c.do("GET", "/api/me", nil)
	c.mustStatus(resp, 401, out)

	// password demasiado curta
	resp, out = c.do("POST", "/api/setup", map[string]string{"password": "curta"})
	c.mustStatus(resp, 400, out)

	// setup válido inicia sessão
	resp, out = c.do("POST", "/api/setup", map[string]string{"username": "loureiro", "password": "password-segura"})
	c.mustStatus(resp, 201, out)
	resp, out = c.do("GET", "/api/me", nil)
	c.mustStatus(resp, 200, out)
	if out["username"] != "loureiro" {
		t.Fatalf("username errado: %v", out)
	}

	// segundo setup é recusado
	resp, out = c.do("POST", "/api/setup", map[string]string{"password": "outra-password"})
	c.mustStatus(resp, 409, out)

	// logout termina a sessão
	resp, out = c.do("POST", "/api/logout", nil)
	c.mustStatus(resp, 200, out)
	resp, out = c.do("GET", "/api/me", nil)
	c.mustStatus(resp, 401, out)

	// login com credenciais erradas e certas
	resp, out = c.do("POST", "/api/login", map[string]string{"username": "loureiro", "password": "errada-de-todo"})
	c.mustStatus(resp, 401, out)
	resp, out = c.do("POST", "/api/login", map[string]string{"username": "loureiro", "password": "password-segura"})
	c.mustStatus(resp, 200, out)
	resp, out = c.do("GET", "/api/me", nil)
	c.mustStatus(resp, 200, out)
}
