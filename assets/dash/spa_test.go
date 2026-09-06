package dash

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDashSPAIsVue3LoginRefreshLogsPage(t *testing.T) {
	html := readDashHTML(t)
	for _, needle := range []string{
		"boltcss/bolt.min.css",
		"vue@3",
		"Vue.createApp",
		"/api/auth/login",
		"/api/auth/refresh",
		"/api/logs/history",
		"/api/logs/tail",
		"Authorization",
		"Bearer",
		"refresh_token",
		"access_token",
	} {
		if !strings.Contains(html, needle) {
			t.Errorf("index.html missing %q", needle)
		}
	}
	if strings.Contains(html, "type=\"module\"") {
		t.Error("index.html should use a classic script so file:// can load it")
	}
	if _, err := os.Stat(filepath.Join("package.json")); err == nil {
		t.Error("assets/dash must not be a Node/bundler app")
	}
}

func TestDashSPAScriptExecutesInBrowserLikeContext(t *testing.T) {
	html := readDashHTML(t)
	script := lastInlineScript(t, html)
	if strings.Contains(script, "require(") || strings.Contains(script, "module.exports") {
		t.Fatal("page script must not use Node module/require")
	}

	prelude := `
globalThis.window = globalThis;
globalThis.location = { protocol: "http:", origin: "http://127.0.0.1:8800" };
globalThis.document = {
  getElementById: function () { return { textContent: "", innerHTML: "" }; }
};
globalThis.AbortController = function () {
  this.signal = {};
  this.abort = function () {};
};
globalThis.fetch = async function () {
  return {
    ok: true,
    status: 200,
    json: async function () { return { success: true, data: {} }; },
    body: null
  };
};
globalThis.Vue = {
  createApp: function (options) {
    var data = typeof options.data === "function" ? options.data() : {};
    var ctx = Object.assign(data, options.methods || {});
    if (typeof options.unmounted === "function") {
      options.unmounted.call(ctx);
    }
    return { mount: function () { return ctx; } };
  }
};
`
	cmd := exec.Command("node", "-e", prelude+"\n"+script)
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("page script failed in a browser-like context: %v\n%s", err, out)
	}
}

func readDashHTML(t *testing.T) string {
	t.Helper()
	raw, err := Assets.ReadFile("index.html")
	if err != nil {
		t.Fatalf("read embedded index.html: %v", err)
	}
	return string(raw)
}

func lastInlineScript(t *testing.T, html string) string {
	t.Helper()
	rest := html
	var last string
	for {
		start := strings.Index(strings.ToLower(rest), "<script")
		if start < 0 {
			break
		}
		rest = rest[start:]
		tagEnd := strings.Index(rest, ">")
		if tagEnd < 0 {
			break
		}
		openTag := rest[:tagEnd+1]
		rest = rest[tagEnd+1:]
		closeIdx := strings.Index(strings.ToLower(rest), "</script>")
		if closeIdx < 0 {
			break
		}
		body := rest[:closeIdx]
		rest = rest[closeIdx+len("</script>"):]
		if !strings.Contains(strings.ToLower(openTag), "src=") {
			last = body
		}
	}
	if last == "" {
		t.Fatal("index.html has no inline script")
	}
	return last
}
