package file

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-sphere/sphere/server/httpz"
	spherefile "github.com/go-sphere/sphere/server/service/file"
	"github.com/go-sphere/sphere/storage"
	"github.com/go-sphere/sphere/storage/fileserver"
)

func TestWebServer_TokenUploadDownloadFlow(t *testing.T) {
	addr, baseURL := mustReserveAddress(t)

	localRoot := t.TempDir()
	fileServer, err := spherefile.NewLocalFileService(spherefile.LocalFileServiceConfig{
		RootDir:    localRoot,
		PublicBase: baseURL,
	})
	if err != nil {
		t.Fatalf("NewLocalFileService() error = %v", err)
	}

	webServer := NewWebServer(Config{Address: addr}, fileServer)

	startCtx := t.Context()

	startErrCh := make(chan error, 1)
	go func() {
		startErrCh <- webServer.Start(startCtx)
	}()

	waitForServerReady(t, baseURL)

	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = webServer.Stop(stopCtx)

		select {
		case err := <-startErrCh:
			if err != nil {
				t.Logf("web server exited with error: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Log("web server start goroutine still running after stop")
		}
	})

	content := []byte("sphere-layout upload/download e2e")

	// Step 1: get upload token.
	authData, err := fileServer.GenerateUploadAuth(context.Background(), storage.UploadAuthRequest{
		Dir:      "user",
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatalf("GenerateUploadAuth() error = %v", err)
	}
	if authData.Authorization.Value == "" {
		t.Fatal("GenerateUploadAuth() returned empty upload token url")
	}

	// Step 2: upload file with token url.
	uploadReq, err := http.NewRequest(http.MethodPut, authData.Authorization.Value, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("http.NewRequest(PUT) error = %v", err)
	}
	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		t.Fatalf("upload request error = %v", err)
	}
	defer func() {
		_ = uploadResp.Body.Close()
	}()
	if uploadResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(uploadResp.Body)
		t.Fatalf("upload status = %d, body = %s", uploadResp.StatusCode, string(body))
	}
	uploadBody, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		t.Fatalf("read upload body error = %v", err)
	}
	var uploadResult httpz.DataResponse[fileserver.UploadResult]
	if err = json.Unmarshal(uploadBody, &uploadResult); err != nil {
		t.Fatalf("decode upload response error = %v, body = %s", err, string(uploadBody))
	}
	if uploadResult.Data.Key != authData.File.Key {
		t.Fatalf("upload response key = %q, want %q", uploadResult.Data.Key, authData.File.Key)
	}
	if uploadResult.Data.URL != authData.File.URL {
		t.Fatalf("upload response url = %q, want %q", uploadResult.Data.URL, authData.File.URL)
	}

	// Step 3: download uploaded file.
	downloadResp, err := http.Get(authData.File.URL)
	if err != nil {
		t.Fatalf("download request error = %v", err)
	}
	defer func() {
		_ = downloadResp.Body.Close()
	}()
	if downloadResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(downloadResp.Body)
		t.Fatalf("download status = %d, url = %s, key = %s, body = %s", downloadResp.StatusCode, authData.File.URL, authData.File.Key, string(body))
	}
	downloaded, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll(download) error = %v", err)
	}
	if !bytes.Equal(downloaded, content) {
		t.Fatalf("download content mismatch: got = %q, want = %q", string(downloaded), string(content))
	}
}

func mustReserveAddress(t *testing.T) (string, string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("listener close error = %v", err)
	}
	return addr, fmt.Sprintf("http://%s", addr)
}

func waitForServerReady(t *testing.T, baseURL string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/__ready_check__")
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server not ready within timeout: %s", baseURL)
}
