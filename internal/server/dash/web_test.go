package dash

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-sphere/sphere-layout/internal/pkg/dao"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/client"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	servicedash "github.com/go-sphere/sphere-layout/internal/service/dash"
	"github.com/go-sphere/sphere/cache/memory"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
	"github.com/go-sphere/sphere/storage"
	"github.com/go-sphere/sphere/utils/secure"
)

const (
	testAdminUsername = "admin"
	testAdminPassword = "aA1234567"
)

func TestWebAuthAndAdminEndpoints(t *testing.T) {
	t.Run("default credentials should return token", func(t *testing.T) {
		baseURL, cleanup := setupTestWeb(t)
		defer cleanup()

		status, body := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
			"username": testAdminUsername,
			"password": testAdminPassword,
		}, nil)
		if status != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body=%s", status, body)
		}

		tokens := parseAuthTokens(t, body)
		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Fatalf("expected non-empty access_token and refresh_token, body=%s", body)
		}
	})

	t.Run("wrong credentials should not return token", func(t *testing.T) {
		baseURL, cleanup := setupTestWeb(t)
		defer cleanup()

		status, body := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
			"username": "wrong-user",
			"password": "wrong-password",
		}, nil)
		if status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d, body=%s", status, http.StatusUnauthorized, body)
		}
	})

	t.Run("refresh rotates tokens and bearer accesses admin list", func(t *testing.T) {
		baseURL, cleanup := setupTestWeb(t)
		defer cleanup()

		loginStatus, loginBody := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
			"username": testAdminUsername,
			"password": testAdminPassword,
		}, nil)
		if loginStatus != http.StatusOK {
			t.Fatalf("login status = %d, want %d, body=%s", loginStatus, http.StatusOK, loginBody)
		}
		tokens := parseAuthTokens(t, loginBody)
		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Fatalf("expected login tokens, body=%s", loginBody)
		}

		refreshStatus, refreshBody := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/refresh", map[string]string{
			"refresh_token": tokens.RefreshToken,
		}, nil)
		if refreshStatus != http.StatusOK {
			t.Fatalf("refresh status = %d, want %d, body=%s", refreshStatus, http.StatusOK, refreshBody)
		}
		refreshed := parseAuthTokens(t, refreshBody)
		if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
			t.Fatalf("expected refreshed tokens, body=%s", refreshBody)
		}
		if refreshed.AccessToken == tokens.AccessToken || refreshed.RefreshToken == tokens.RefreshToken {
			t.Fatalf("refresh did not rotate tokens, body=%s", refreshBody)
		}

		status, body := doJSONRequest(t, http.MethodGet, baseURL+"/api/admin/list", nil, map[string]string{
			"Authorization": "Bearer " + refreshed.AccessToken,
		})
		if status != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body=%s", status, body)
		}

		count := parseAdminCount(t, body)
		if count == 0 {
			t.Fatalf("expected non-empty admin list, body=%s", body)
		}
	})

	t.Run("legacy pure-admin login path is gone", func(t *testing.T) {
		baseURL, cleanup := setupTestWeb(t)
		defer cleanup()

		status, body := doJSONRequest(t, http.MethodPost, baseURL+"/api/login", map[string]string{
			"username": testAdminUsername,
			"password": testAdminPassword,
		}, nil)
		if status == http.StatusOK {
			t.Fatalf("legacy /api/login still succeeded, body=%s", body)
		}
	})

	t.Run("invalid token should not get admin list", func(t *testing.T) {
		baseURL, cleanup := setupTestWeb(t)
		defer cleanup()

		status, body := doJSONRequest(t, http.MethodGet, baseURL+"/api/admin/list", nil, map[string]string{
			"Authorization": "Bearer invalid-token",
		})
		if status != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d, body=%s", status, http.StatusUnauthorized, body)
		}
	})
}

// TestWebAuthCookie covers the browser auth path: login and refresh write the
// auth_token cookie and the auth middleware accepts it on requests that carry
// no Authorization header (page loads, <img>, EventSource).
func TestWebAuthCookie(t *testing.T) {
	baseURL, cleanup := setupTestWeb(t)
	defer cleanup()

	base, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse base url failed: %v", err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("new cookie jar failed: %v", err)
	}
	client := &http.Client{Timeout: time.Second * 5, Jar: jar}

	status, header, body := doJSONRequestWithClient(t, client, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
		"username": testAdminUsername,
		"password": testAdminPassword,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, want 200, body=%s", status, body)
	}
	tokens := parseAuthTokens(t, body)

	var authCookie *http.Cookie
	for _, cookie := range (&http.Response{Header: header}).Cookies() {
		if cookie.Name == servicedash.AuthTokenCookieName {
			authCookie = cookie
		}
	}
	if authCookie == nil {
		t.Fatalf("login response missing %s cookie, headers=%v", servicedash.AuthTokenCookieName, header)
	}
	if authCookie.Value != tokens.AccessToken {
		t.Fatalf("cookie value does not match the access token, value_len=%d token_len=%d", len(authCookie.Value), len(tokens.AccessToken))
	}
	maxAge := int(servicedash.AuthTokenValidDuration.Seconds())
	if authCookie.Path != "/" || authCookie.MaxAge != maxAge || authCookie.SameSite != http.SameSiteLaxMode || !authCookie.HttpOnly || authCookie.Secure {
		t.Fatalf("cookie attrs = path %q max-age %d samesite %v http-only %v secure %v, want /, %d, Lax, true, false",
			authCookie.Path, authCookie.MaxAge, authCookie.SameSite, authCookie.HttpOnly, authCookie.Secure, maxAge)
	}

	// The jar replays the cookie; these requests send no Authorization header.
	status, _, body = doJSONRequestWithClient(t, client, http.MethodGet, baseURL+"/api/admin/list", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("cookie-authenticated admin list status = %d, want 200, body=%s", status, body)
	}

	// Refresh rotates the access token and must rotate the cookie with it.
	status, _, body = doJSONRequestWithClient(t, client, http.MethodPost, baseURL+"/api/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("refresh status = %d, want 200, body=%s", status, body)
	}
	refreshed := parseAuthTokens(t, body)
	var rotated string
	for _, cookie := range jar.Cookies(base) {
		if cookie.Name == servicedash.AuthTokenCookieName {
			rotated = cookie.Value
		}
	}
	if rotated != refreshed.AccessToken {
		t.Fatalf("refresh did not rotate the auth cookie, value_len=%d token_len=%d", len(rotated), len(refreshed.AccessToken))
	}

	// A bogus cookie must not authenticate, mirroring the invalid-Bearer case.
	badJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("new cookie jar failed: %v", err)
	}
	badJar.SetCookies(base, []*http.Cookie{{Name: servicedash.AuthTokenCookieName, Value: "not-a-jwt", Path: "/"}})
	status, _, body = doJSONRequestWithClient(t, &http.Client{Timeout: time.Second * 5, Jar: badJar}, http.MethodGet, baseURL+"/api/admin/list", nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("invalid cookie status = %d, want 401, body=%s", status, body)
	}

	// Behind a TLS-terminating proxy the cookie must be marked Secure; the
	// template serves plain HTTP, so the scheme comes from the proxy header.
	status, header, body = doJSONRequestWithClient(t, client, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
		"username": testAdminUsername,
		"password": testAdminPassword,
	}, map[string]string{"X-Forwarded-Proto": "https"})
	if status != http.StatusOK {
		t.Fatalf("https login status = %d, want 200, body=%s", status, body)
	}
	var secureCookie *http.Cookie
	for _, cookie := range (&http.Response{Header: header}).Cookies() {
		if cookie.Name == servicedash.AuthTokenCookieName {
			secureCookie = cookie
		}
	}
	if secureCookie == nil || !secureCookie.Secure {
		t.Fatalf("cookie behind X-Forwarded-Proto=https is not Secure, cookie=%+v", secureCookie)
	}
}

func TestWebLoginRefreshAndLogsTwice(t *testing.T) {
	for i := range 2 {
		t.Run(fmt.Sprintf("run-%d", i+1), func(t *testing.T) {
			baseURL, cleanup := setupTestWeb(t)
			defer cleanup()

			status, body := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
				"username": testAdminUsername,
				"password": testAdminPassword,
			}, nil)
			if status != http.StatusOK {
				t.Fatalf("login status = %d, want 200, body=%s", status, body)
			}
			tokens := parseAuthTokens(t, body)
			if tokens.AccessToken == "" || tokens.RefreshToken == "" {
				t.Fatalf("login missing tokens, body=%s", body)
			}
			t.Logf("login access_token_len=%d refresh_token_len=%d expires_at=%d", len(tokens.AccessToken), len(tokens.RefreshToken), tokens.ExpiresAt)

			refreshStatus, refreshBody := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/refresh", map[string]string{
				"refresh_token": tokens.RefreshToken,
			}, nil)
			if refreshStatus != http.StatusOK {
				t.Fatalf("refresh status = %d, want 200, body=%s", refreshStatus, refreshBody)
			}
			refreshed := parseAuthTokens(t, refreshBody)
			if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
				t.Fatalf("refresh missing tokens, body=%s", refreshBody)
			}
			if refreshed.AccessToken == tokens.AccessToken || refreshed.RefreshToken == tokens.RefreshToken {
				t.Fatalf("refresh did not rotate tokens, body=%s", refreshBody)
			}
			t.Logf("refresh access_token_len=%d refresh_token_len=%d expires_at=%d", len(refreshed.AccessToken), len(refreshed.RefreshToken), refreshed.ExpiresAt)

			logStatus, logBody := doJSONRequest(t, http.MethodGet, baseURL+"/api/logs/history?limit=20", nil, map[string]string{
				"Authorization": "Bearer " + refreshed.AccessToken,
			})
			if logStatus != http.StatusOK {
				t.Fatalf("logs status = %d, want 200, body=%s", logStatus, logBody)
			}
			t.Logf("log history body=%s", logBody)
			if !strings.Contains(logBody, `"entries"`) || !strings.Contains(logBody, `"stream_id"`) {
				t.Fatalf("log history body is not a sane JSON payload: %s", logBody)
			}
			if !strings.Contains(logBody, "dash web ready") {
				t.Fatalf("log history missing seeded entry, body=%s", logBody)
			}
			if !strings.Contains(logBody, "/api/auth/login") {
				t.Fatalf("log history missing access log entry, body=%s", logBody)
			}
		})
	}
}

// TestWebLogTailStreamsOverSSE pins the one response shape that depends on
// adapter-specific write plumbing: the dash log tail is the only SSE endpoint,
// and streaming needs httpx.Streamer (net/http flushing on stdx, where the
// engine itself owns the ResponseWriter). It also pins the lazy-commit
// contract: the subscription's first frame is what commits the response, so a
// client sees a 200 text/event-stream and then live entries.
func TestWebLogTailStreamsOverSSE(t *testing.T) {
	baseURL, cleanup := setupTestWeb(t)
	defer cleanup()

	status, body := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth/login", map[string]string{
		"username": testAdminUsername,
		"password": testAdminPassword,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, want 200, body=%s", status, body)
	}
	token := parseAuthTokens(t, body).AccessToken

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/logs/tail?tail_limit=50", nil)
	if err != nil {
		t.Fatalf("new tail request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open tail stream: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("tail status = %d, want 200, body=%s", resp.StatusCode, raw)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}

	frames := make(chan string, 16)
	go func() {
		defer close(frames)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			if data, ok := strings.CutPrefix(scanner.Text(), "data: "); ok {
				frames <- data
			}
		}
	}()

	// Backfill is non-empty here ("dash web ready" plus the login access log),
	// so the first frame carries it; an empty buffer would send a cursor frame
	// instead. Either way the frame commits the lazy response.
	select {
	case frame, ok := <-frames:
		if !ok {
			t.Fatal("stream closed before the first frame")
		}
		if !strings.Contains(frame, `"stream_id"`) {
			t.Fatalf("first frame = %s, want a cursor or backfill frame", frame)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no frame arrived within 5s")
	}

	// The access log writes to the same buffer this subscription reads, so a
	// completed request must show up on the open stream.
	if status, body := doJSONRequest(t, http.MethodGet, baseURL+"/api/logs/history", nil, map[string]string{
		"Authorization": "Bearer " + token,
	}); status != http.StatusOK {
		t.Fatalf("history status = %d, want 200, body=%s", status, body)
	}

	deadline := time.After(5 * time.Second)
	for {
		select {
		case frame, ok := <-frames:
			if !ok {
				t.Fatal("stream closed before the live entry arrived")
			}
			if strings.Contains(frame, "/api/logs/history") {
				return
			}
		case <-deadline:
			t.Fatal("live log entry did not arrive on the stream")
		}
	}
}

func TestDashPageIsServed(t *testing.T) {
	baseURL, cleanup := setupTestWeb(t)
	defer cleanup()

	resp, err := http.Get(baseURL + "/dash/")
	if err != nil {
		t.Fatalf("GET /dash/: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /dash/: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /dash/ status = %d, want 200, body=%s", resp.StatusCode, raw)
	}
	page := string(raw)
	if !strings.Contains(page, "vue@3") || !strings.Contains(page, "/api/auth/login") {
		t.Fatalf("dash page missing Vue 3 login app, body=%s", page)
	}
}

func TestDashSPAHeadlessScreenshot(t *testing.T) {
	out := os.Getenv("DASH_SPA_SCREENSHOT")
	if out == "" {
		t.Skip("DASH_SPA_SCREENSHOT not set")
	}
	chrome := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if _, err := os.Stat(chrome); err != nil {
		t.Skip("chrome not available")
	}
	baseURL, cleanup := setupTestWeb(t)
	defer cleanup()

	cmd := exec.Command(chrome,
		"--headless=new",
		"--disable-gpu",
		"--no-first-run",
		"--window-size=1280,900",
		"--virtual-time-budget=5000",
		"--screenshot="+out,
		baseURL+"/dash/",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chrome screenshot failed: %v\n%s", err, output)
	}
	info, err := os.Stat(out)
	if err != nil || info.Size() == 0 {
		t.Fatalf("screenshot missing or empty: %v\n%s", err, output)
	}
}

func setupTestWeb(t *testing.T) (string, func()) {
	t.Helper()

	addr := randomLocalAddress(t)
	db := newMemoryDB(t)
	insertDefaultAdmin(t, db)

	logs := logbuffer.New(32)
	logs.Log(t.Context(), log.LevelInfo, "dash web ready")
	testStorage := &noopStorage{}
	service := servicedash.NewService(dao.NewDao(db), memory.NewByteCache(), testStorage, logs)
	web := NewWebServer(Config{
		AuthJWT:    "test-auth-jwt-secret",
		RefreshJWT: "test-refresh-jwt-secret",
		HTTP: HTTPConfig{
			Address: addr,
		},
	}, testStorage, service, logs)

	startErr := make(chan error, 1)
	go func() {
		startErr <- web.Start(t.Context())
	}()

	baseURL := "http://" + addr
	waitServerReady(t, baseURL, startErr)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		_ = web.Stop(ctx)
		_ = db.Close()
		select {
		case err := <-startErr:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				t.Fatalf("web server exited with error: %v", err)
			}
		case <-time.After(time.Second):
		}
	}
	return baseURL, cleanup
}

func waitServerReady(t *testing.T, baseURL string, startErr <-chan error) {
	t.Helper()

	httpClient := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(time.Second * 5)
	for time.Now().Before(deadline) {
		select {
		case err := <-startErr:
			t.Fatalf("web server start failed: %v", err)
		default:
		}
		resp, err := httpClient.Get(baseURL + "/")
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(time.Millisecond * 50)
	}
	t.Fatalf("web server did not become ready in time")
}

func randomLocalAddress(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on random port failed: %v", err)
	}
	defer func() { _ = ln.Close() }()
	return ln.Addr().String()
}

func newMemoryDB(t *testing.T) *ent.Client {
	t.Helper()

	conf := client.Config{
		Type: "sqlite3",
		Path: fmt.Sprintf("file:dash-web-test-%d?mode=memory&cache=shared", time.Now().UnixNano()),
	}
	db, err := client.NewDataBaseClient(conf)
	if err != nil {
		t.Fatalf("create test database failed: %v", err)
	}
	return db
}

func insertDefaultAdmin(t *testing.T, db *ent.Client) {
	t.Helper()

	password, err := secure.CryptPassword(testAdminPassword)
	if err != nil {
		t.Fatalf("crypt password failed: %v", err)
	}
	_, err = db.Admin.Create().
		SetUsername(testAdminUsername).
		SetPassword(password).
		SetRoles([]string{"all"}).
		Save(t.Context())
	if err != nil {
		t.Fatalf("insert admin failed: %v", err)
	}
}

func doJSONRequest(t *testing.T, method, target string, payload any, headers map[string]string) (int, string) {
	t.Helper()

	status, _, body := doJSONRequestWithClient(t, &http.Client{Timeout: time.Second * 5}, method, target, payload, headers)
	return status, body
}

func doJSONRequestWithClient(t *testing.T, client *http.Client, method, target string, payload any, headers map[string]string) (int, http.Header, string) {
	t.Helper()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload failed: %v", err)
		}
		body = bytes.NewBuffer(raw)
	}

	req, err := http.NewRequestWithContext(t.Context(), method, target, body)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body failed: %v", err)
	}
	return resp.StatusCode, resp.Header, string(raw)
}

type authTokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func parseAuthTokens(t *testing.T, body string) authTokenData {
	t.Helper()

	var resp struct {
		Data authTokenData `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode auth response: %v, body=%s", err, body)
	}
	if strings.Contains(body, `"accessToken"`) || strings.Contains(body, `"refreshToken"`) {
		t.Fatalf("auth response still uses camelCase token fields, body=%s", body)
	}
	return resp.Data
}

func parseAdminCount(t *testing.T, body string) int {
	t.Helper()

	var resp struct {
		Data struct {
			Admins []any `json:"admins"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode admin list response: %v, body=%s", err, body)
	}
	return len(resp.Data.Admins)
}

type noopStorage struct{}

func (n *noopStorage) GenerateURL(key string, _ ...url.Values) string { return key }

func (n *noopStorage) GenerateURLs(keys []string, _ ...url.Values) []string { return keys }

func (n *noopStorage) ExtractKeyFromURL(uri string) string { return uri }

func (n *noopStorage) ExtractKeyFromURLWithMode(uri string, _ bool) (string, error) { return uri, nil }

func (n *noopStorage) GenerateUploadAuth(_ context.Context, req storage.UploadAuthRequest) (storage.UploadAuthResult, error) {
	return storage.UploadAuthResult{
		Authorization: storage.UploadAuthorization{
			Type:   storage.UploadAuthorizationTypeToken,
			Value:  "test-upload-token",
			Method: http.MethodPost,
		},
		File: storage.UploadFileInfo{
			Key: req.FileName,
			URL: req.FileName,
		},
	}, nil
}

func (n *noopStorage) UploadFile(_ context.Context, _ io.Reader, key string) (string, error) {
	return key, nil
}

func (n *noopStorage) UploadLocalFile(_ context.Context, _ string, key string) (string, error) {
	return key, nil
}

func (n *noopStorage) IsFileExists(_ context.Context, _ string) (bool, error) { return false, nil }

func (n *noopStorage) DownloadFile(_ context.Context, _ string) (storage.DownloadResult, error) {
	return storage.DownloadResult{}, errors.New("not implemented")
}

func (n *noopStorage) DeleteFile(_ context.Context, _ string) error { return nil }

func (n *noopStorage) MoveFile(_ context.Context, _, _ string, _ bool) error { return nil }

func (n *noopStorage) CopyFile(_ context.Context, _, _ string, _ bool) error { return nil }
