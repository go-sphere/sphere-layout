package httpsrv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-sphere/httpx"
)

// registerAcceptanceRoutes is the shared route set used to prove that the gin
// and fiber engines are interchangeable, including the error path that used
// to require the framework-specific jsonErrorContext hack.
func registerAcceptanceRoutes(r httpx.Router) {
	r.GET("/ping", httpx.WithJson(func(ctx httpx.Context) (string, error) {
		return "pong", nil
	}))
	r.GET("/boom", func(ctx httpx.Context) error {
		return httpx.NewNotFoundError("resource missing")
	})
	r.GET("/users/:id", httpx.WithJson(func(ctx httpx.Context) (map[string]string, error) {
		return map[string]string{"id": ctx.Param("id")}, nil
	}))
}

func doRequest(t *testing.T, engine httpx.Engine, target string) (int, map[string]any) {
	t.Helper()
	tr, ok := httpx.AsTestRequester(engine)
	if !ok {
		t.Fatalf("engine %T does not support in-process test requests", engine)
	}
	resp, err := tr.Do(httptest.NewRequest(http.MethodGet, target, nil))
	if err != nil {
		t.Fatalf("request %s failed: %v", target, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("response %s is not JSON: %v; body=%q", target, err, body)
	}
	return resp.StatusCode, payload
}

// TestGinAndFiberServeSameRoutes is the acceptance check for the shared
// httpx error handler: the same route set (including sphere's JSON error
// rendering) must behave identically on the gin- and fiber-backed engines.
func TestGinAndFiberServeSameRoutes(t *testing.T) {
	engines := map[string]httpx.Engine{
		"gin":   NewGinServer("test", "127.0.0.1:0"),
		"fiber": NewFiberServer("test", "127.0.0.1:0"),
	}
	type result struct {
		status  int
		payload map[string]any
	}
	targets := []string{
		"http://example.com/ping",
		"http://example.com/boom",
		"http://example.com/users/42",
	}
	results := make(map[string]map[string]result)
	for name, engine := range engines {
		registerAcceptanceRoutes(engine.Group(""))
		results[name] = make(map[string]result)
		for _, target := range targets {
			status, payload := doRequest(t, engine, target)
			results[name][target] = result{status: status, payload: payload}
		}
	}

	for _, target := range targets {
		gin := results["gin"][target]
		fiber := results["fiber"][target]
		if gin.status != fiber.status {
			t.Fatalf("%s status mismatch: gin=%d fiber=%d", target, gin.status, fiber.status)
		}
		if !reflect.DeepEqual(gin.payload, fiber.payload) {
			t.Fatalf("%s payload mismatch:\n gin=%v\n fiber=%v", target, gin.payload, fiber.payload)
		}
	}

	boom := results["gin"]["http://example.com/boom"]
	if boom.status != http.StatusNotFound {
		t.Fatalf("error route status = %d, want 404", boom.status)
	}
	if boom.payload["success"] != false || boom.payload["message"] != "resource missing" {
		t.Fatalf("error route payload = %v, want sphere error shape", boom.payload)
	}
}
