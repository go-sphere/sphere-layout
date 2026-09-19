package httpsrv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sphere/httpx"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
	"github.com/go-sphere/sphere/server/httpz"
)

// decodeErrorBody parses the standard error envelope so a test can assert on
// the fields a client sees rather than on the exact string.
func decodeErrorBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("error response is not JSON: %v; body=%q", err, body)
	}
	return payload
}

// rawRequest serves one in-process request and returns status, body and headers
// without assuming the body parses as JSON.
func rawRequest(t *testing.T, engine httpx.Engine, method, target string, headers map[string]string) (int, string, http.Header) {
	t.Helper()
	tr, ok := httpx.AsTestRequester(engine)
	if !ok {
		t.Fatalf("engine %T does not support in-process test requests", engine)
	}
	req := httptest.NewRequest(method, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := tr.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, string(body), resp.Header
}

// TestServerErrorResponsesUseTheJSONEnvelope pins the shape of the responses the
// layout does not register a route for. The stdx adapter renders 404 and 405
// through the httpx ErrorHandler — httpz.AbortWithJsonError here — where gin
// answered with its own plain-text body, so these are the cases a client sees
// when a path or method is wrong.
func TestServerErrorResponsesUseTheJSONEnvelope(t *testing.T) {
	engine := NewServer("test", "127.0.0.1:0", nil)
	registerAcceptanceRoutes(engine.Group(""))

	t.Run("unknown path is a JSON 404", func(t *testing.T) {
		status, body, header := rawRequest(t, engine, http.MethodGet, "http://example.com/missing", nil)
		if status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404; body=%s", status, body)
		}
		if ct := header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want JSON", ct)
		}
		if got := decodeErrorBody(t, body); got["message"] != http.StatusText(http.StatusNotFound) || got["success"] != false {
			t.Fatalf("body = %v, want the standard error envelope", got)
		}
	})

	t.Run("wrong method is a JSON 405 with Allow", func(t *testing.T) {
		status, body, header := rawRequest(t, engine, http.MethodDelete, "http://example.com/ping", nil)
		if status != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405; body=%s", status, body)
		}
		if allow := header.Get("Allow"); allow != "GET" {
			t.Fatalf("Allow = %q, want %q", allow, "GET")
		}
		if got := decodeErrorBody(t, body); got["message"] != http.StatusText(http.StatusMethodNotAllowed) {
			t.Fatalf("body = %v, want the standard error envelope", got)
		}
	})
}

// TestServerAccessLogCoversUnmatchedPaths pins that engine-scope middleware
// runs for a path no route matched. It has to: the access log and panic
// recovery are registered on the engine precisely so a 404 is still logged,
// while a group's middleware reaches only the routes registered under it.
func TestServerAccessLogCoversUnmatchedPaths(t *testing.T) {
	buf := logbuffer.New(32)
	engine := NewServer("test", "127.0.0.1:0", buf)
	registerAcceptanceRoutes(engine.Group(""))

	if status, body, _ := rawRequest(t, engine, http.MethodGet, "http://example.com/missing", nil); status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", status, body)
	}

	entries, _ := buf.History(0, 8, log.LevelDebug)
	for _, entry := range entries {
		if entry.Message == "/missing" && entry.Attrs["method"] == "GET" && entry.Attrs["status"] == int64(http.StatusNotFound) {
			return
		}
	}
	t.Fatalf("unmatched request missing from the access log: %+v", entries)
}

// TestServerCORSPreflightIsEngineWide pins that the CORS middleware answers a
// preflight for a path the router does not know, and for a path registered
// under a different method. Both are unmatched paths, which only engine-scope
// middleware covers; UseCORS registers on the engine for that reason.
func TestServerCORSPreflightIsEngineWide(t *testing.T) {
	engine := NewServer("test", "127.0.0.1:0", nil)
	registerAcceptanceRoutes(engine.Group(""))
	if err := UseCORS(engine, []string{"https://example.com"}); err != nil {
		t.Fatalf("UseCORS: %v", err)
	}

	preflight := map[string]string{
		"Origin":                        "https://example.com",
		"Access-Control-Request-Method": "GET",
	}
	for _, target := range []string{"http://example.com/not-registered", "http://example.com/ping"} {
		status, body, header := rawRequest(t, engine, http.MethodOptions, target, preflight)
		if status != http.StatusNoContent {
			t.Fatalf("OPTIONS %s status = %d, want 204; body=%s", target, status, body)
		}
		if got := header.Get("Access-Control-Allow-Origin"); got != "https://example.com" {
			t.Fatalf("OPTIONS %s Allow-Origin = %q", target, got)
		}
	}
}

// TestServerFullPathFeedsMatchOperation pins the two context reads the layout's
// authorization and rate limiting depend on: FullPath is the pattern that was
// registered — not the anonymous rewrite a framework may use internally — and a
// named wildcard resolves through Param. httpz.MatchOperation keys its route
// policy map on exactly this pair.
func TestServerFullPathFeedsMatchOperation(t *testing.T) {
	engine := NewServer("test", "127.0.0.1:0", nil)
	matcher := httpz.MatchOperation("", [][3]string{{"download", http.MethodGet, "/files/*name"}}, "download")

	engine.Group("").GET("/files/*name", func(ctx httpx.Context) error {
		if !matcher(ctx) {
			t.Errorf("MatchOperation missed %q (FullPath=%q)", ctx.Path(), ctx.FullPath())
		}
		return ctx.JSON(http.StatusOK, map[string]string{
			"fullPath": ctx.FullPath(),
			"param":    ctx.Param("name"),
		})
	})

	status, body, _ := rawRequest(t, engine, http.MethodGet, "http://example.com/files/a/b.txt", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", status, body)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("parse body: %v; body=%q", err, body)
	}
	if payload["fullPath"] != "/files/*name" || payload["param"] != "a/b.txt" {
		t.Fatalf("payload = %v, want the registered pattern and the wildcard value", payload)
	}
}

// TestServerDoesNotRedirectTrailingSlash records the one client-visible
// behavior that changed with the adapter swap. Gin answered GET /ping/ with a
// 301 to /ping (RedirectTrailingSlash is on by default); stdx deliberately has
// no trailing-slash redirect, path cleaning or case fixup, because the other
// four frameworks do not agree on them either. The request is a 404 here.
// Clients that relied on the redirect have to use the exact path.
func TestServerDoesNotRedirectTrailingSlash(t *testing.T) {
	engine := NewServer("test", "127.0.0.1:0", nil)
	registerAcceptanceRoutes(engine.Group(""))

	status, body, header := rawRequest(t, engine, http.MethodGet, "http://example.com/ping/", nil)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (stdx does not redirect); body=%s", status, body)
	}
	if location := header.Get("Location"); location != "" {
		t.Fatalf("unexpected redirect to %q", location)
	}
}
