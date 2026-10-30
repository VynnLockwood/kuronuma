package olympus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/modules/plutus"
)

type fakeSource struct {
	info Info
	err  error
}

func (f fakeSource) Info(context.Context) (Info, error) { return f.info, f.err }

type fakeImporter struct {
	res     plutus.ImportResult
	err     error
	calls   int
	content string
	max     int64
}

func (f *fakeImporter) ImportBackup(_ context.Context, r io.Reader, maxBytes int64) (plutus.ImportResult, error) {
	f.calls++
	f.max = maxBytes
	b, _ := io.ReadAll(r)
	f.content = string(b)
	return f.res, f.err
}

type rig struct {
	srv  *Server
	imp  *fakeImporter
	logs *bytes.Buffer
}

func newRig(t *testing.T, info Info, srcErr error, maxUpload int64) *rig {
	t.Helper()
	imp := &fakeImporter{}
	logs := &bytes.Buffer{}
	srv, err := New(Options{
		Addr:           testAddr,
		Source:         fakeSource{info: info, err: srcErr},
		Importer:       imp,
		MaxUploadBytes: maxUpload,
		Log:            slog.New(slog.NewJSONHandler(logs, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return &rig{srv: srv, imp: imp, logs: logs}
}

func (r *rig) get(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = testAddr
	rec := httptest.NewRecorder()
	r.srv.Handler().ServeHTTP(rec, req)
	return rec
}

func (r *rig) upload(t *testing.T, file string) *httptest.ResponseRecorder {
	t.Helper()
	return r.uploadWith(t, postOpts{token: r.srv.g.Token(), file: file})
}

func (r *rig) uploadWith(t *testing.T, o postOpts) *httptest.ResponseRecorder {
	t.Helper()
	o.origin = "http://127.0.0.1:8082"
	body, ct := multipartBody(t, o)
	req := httptest.NewRequest(http.MethodPost, "/import", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Origin", o.origin)
	req.Host = testAddr
	rec := httptest.NewRecorder()
	r.srv.Handler().ServeHTTP(rec, req)
	return rec
}

func sampleInfo() Info {
	last := time.Date(2026, 10, 29, 8, 30, 0, 0, time.UTC)
	return Info{
		Mode: "selfhost", Version: "0.2.0", PublicURL: "https://nemo.example.com", DatabaseOK: true, TransactionCount: 1234,
		Keys: []KeyInfo{
			{Name: "claude-desktop", Scopes: []string{"plutus:read", "plutus:write"}, CreatedAt: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC), LastUsedAt: &last},
			{Name: "old-phone", Scopes: []string{"plutus:read"}, CreatedAt: time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC), Revoked: true},
		},
	}
}

func TestIndexShowsConfig(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	rec := r.get("/")
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("status %d type %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	for _, want := range []string{
		"selfhost", "0.2.0", "https://nemo.example.com", "1234",
		"claude-desktop", "plutus:read", "plutus:write", "old-phone", "revoked", "2026-10-29", "never",
		"https://nemo.example.com/mcp/nmk_…",
		`name="csrf" value="` + r.srv.g.Token() + `"`,
		`enctype="multipart/form-data"`, `action="/import"`, `type="file"`, `name="file"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %q", want)
		}
	}
}

func TestIndexNeverShowsSecretsAndEscapesText(t *testing.T) {
	info := sampleInfo()
	info.Keys[0].Name = "<script>alert(1)</script>"
	r := newRig(t, info, nil, 0)
	body := r.get("/").Body.String()
	for _, forbidden := range []string{"DATABASE_URL", "postgres://", "key_hash", "SESSION_SECRET", "HADES_KEY", "POSTGRES_PASSWORD"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("page contains %q", forbidden)
		}
	}
	if strings.Contains(body, "<script") || !strings.Contains(body, "&lt;script&gt;") {
		t.Error("key name must be HTML-escaped")
	}
	// The CSP forbids inline script and style, so the markup must not rely on either.
	if strings.Contains(body, "style=") || strings.Contains(body, "onclick") {
		t.Error("page uses inline style or script handlers that the CSP blocks")
	}
}

func TestIndexDatabaseDown(t *testing.T) {
	info := sampleInfo()
	info.DatabaseOK = false
	rec := newRig(t, info, nil, 0).get("/")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Database unreachable") {
		t.Fatalf("status %d, want 200 with a visible notice", rec.Code)
	}
}

func TestIndexSourceErrorDoesNotLeak(t *testing.T) {
	r := newRig(t, Info{}, errors.New("dial tcp: password=hunter2 refused"), 0)
	rec := r.get("/")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Could not read the current status") {
		t.Fatalf("status %d body %q", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "hunter2") || strings.Contains(r.logs.String(), "hunter2") {
		t.Error("error text leaked into the page or the logs")
	}
}

func TestImportSuccessShowsCountsOnly(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	r.imp.res = plutus.ImportResult{Added: 2, Skipped: 1}
	rec := r.upload(t, "SECRET-FILE-TEXT")
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "Added 2 records") || !strings.Contains(body, "Skipped 1") {
		t.Fatalf("status %d body %q", rec.Code, body)
	}
	if strings.Contains(body, "SECRET-FILE-TEXT") {
		t.Error("result page echoed file content")
	}
	if r.imp.content != "SECRET-FILE-TEXT" {
		t.Errorf("importer received %q, want the uploaded file", r.imp.content)
	}
}

func TestImportRowsErrorListsFirstTwenty(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	var rows []plutus.RowError
	for i := 1; i <= 25; i++ {
		rows = append(rows, plutus.RowError{Row: i, Reason: fmt.Sprintf("reason for row %d.", i)})
	}
	r.imp.err = &plutus.RowsError{Rows: rows}
	rec := r.upload(t, "x")
	body := rec.Body.String()
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422", rec.Code)
	}
	if !strings.Contains(body, "Row 20") || strings.Contains(body, "Row 21") || !strings.Contains(body, "5 more") {
		t.Errorf("want rows 1-20 and '5 more', got %q", body)
	}
	if !strings.Contains(body, "Nothing was imported") {
		t.Error("must say nothing was written")
	}
}

func TestImportNeverEchoesFileText(t *testing.T) {
	payload := `{"merchant":"<script>alert(1)</script>"}`
	outcomes := map[string]struct {
		res plutus.ImportResult
		err error
	}{
		"success":  {res: plutus.ImportResult{Added: 1}},
		"backup":   {err: fmt.Errorf("%w: file is not a backup document", plutus.ErrBackup)},
		"rows":     {err: &plutus.RowsError{Rows: []plutus.RowError{{Row: 1, Reason: "merchant is required."}}}},
		"internal": {err: errors.New("ERROR: failing row contains (<script>alert(1)</script>)")},
	}
	for name, o := range outcomes {
		r := newRig(t, sampleInfo(), nil, 0)
		r.imp.res, r.imp.err = o.res, o.err
		rec := r.upload(t, payload)
		if strings.Contains(rec.Body.String(), "alert(1)") || strings.Contains(r.logs.String(), "alert(1)") {
			t.Errorf("%s: file text reached the page or the logs", name)
		}
	}
}

func TestImportBadFileAndInternalErrorMessages(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	r.imp.err = fmt.Errorf("%w: unknown file format", plutus.ErrBackup)
	rec := r.upload(t, "date,amount")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "not a Nemosyne backup") {
		t.Fatalf("bad file: status %d body %q", rec.Code, rec.Body.String())
	}

	r = newRig(t, sampleInfo(), nil, 0)
	r.imp.err = errors.New("boom SecretMerchant")
	rec = r.upload(t, "x")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("internal: status %d, want 500", rec.Code)
	}
	for _, leaked := range []string{"boom", "SecretMerchant"} {
		if strings.Contains(rec.Body.String(), leaked) || strings.Contains(r.logs.String(), leaked) {
			t.Errorf("%q leaked into the page or the logs", leaked)
		}
	}
}

func TestImportRefusesFileOverLimit(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 1024)
	rec := r.upload(t, strings.Repeat("x", 2000))
	if rec.Code != http.StatusRequestEntityTooLarge || r.imp.calls != 0 {
		t.Fatalf("status %d importer calls %d, want 413 and importer not called", rec.Code, r.imp.calls)
	}
}

func TestImportPassesTheConfiguredLimit(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 4096)
	r.upload(t, "x")
	if r.imp.max != 4096 {
		t.Fatalf("importer limit %d, want 4096", r.imp.max)
	}
	if def := newRig(t, sampleInfo(), nil, 0); def.srv.opts.MaxUploadBytes != 10<<20 {
		t.Fatalf("default limit %d, want 10 MiB", def.srv.opts.MaxUploadBytes)
	}
}

func TestImportMissingFile(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("csrf", r.srv.g.Token())
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Origin", "http://127.0.0.1:8082")
	req.Host = testAddr
	rec := httptest.NewRecorder()
	r.srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || r.imp.calls != 0 {
		t.Fatalf("status %d calls %d, want 400 and importer not called", rec.Code, r.imp.calls)
	}
}

func TestImportEmptyUploadIsRefused(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	rec := r.upload(t, "")
	if rec.Code != http.StatusBadRequest || r.imp.calls != 0 {
		t.Fatalf("status %d calls %d, want 400 for an empty file", rec.Code, r.imp.calls)
	}
}

func TestRequestLogsHaveNoBodies(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	r.imp.err = errors.New("ERR-SECRET")
	r.upload(t, "UPLOAD-SECRET")
	r.get("/?token=QUERY-SECRET")
	logs := r.logs.String()
	for _, leaked := range []string{"UPLOAD-SECRET", "ERR-SECRET", "QUERY-SECRET", r.srv.g.Token()} {
		if strings.Contains(logs, leaked) {
			t.Errorf("logs contain %q", leaked)
		}
	}
	if !strings.Contains(logs, `"path":"/import"`) || !strings.Contains(logs, `"status":500`) || !strings.Contains(logs, `"method":"POST"`) {
		t.Errorf("want a request line with method, path and status, got %s", logs)
	}
}

func TestStaticStylesheet(t *testing.T) {
	rec := newRig(t, sampleInfo(), nil, 0).get("/static/olympus.css")
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/css") {
		t.Fatalf("status %d type %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if strings.Contains(rec.Body.String(), "http") || strings.Contains(rec.Body.String(), "@import") {
		t.Error("stylesheet must not load anything from elsewhere")
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("static responses must also carry the guard headers")
	}
}

func TestUnknownRoutes(t *testing.T) {
	r := newRig(t, sampleInfo(), nil, 0)
	if rec := r.get("/nope"); rec.Code != http.StatusNotFound {
		t.Errorf("GET /nope: %d, want 404", rec.Code)
	}
	if rec := r.get("/import"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /import: %d, want 405", rec.Code)
	}
	if rec := r.get("/mcp/anything"); rec.Code != http.StatusNotFound {
		t.Errorf("the admin listener must not serve MCP routes, got %d", rec.Code)
	}
}

func TestNewValidatesOptions(t *testing.T) {
	good := Options{Addr: testAddr, Source: fakeSource{}, Importer: &fakeImporter{}}
	for name, mutate := range map[string]func(*Options){
		"no source":   func(o *Options) { o.Source = nil },
		"no importer": func(o *Options) { o.Importer = nil },
		"bad addr":    func(o *Options) { o.Addr = "nope" },
	} {
		o := good
		mutate(&o)
		if _, err := New(o); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	if _, err := New(good); err != nil {
		t.Fatalf("valid options rejected: %v", err)
	}
}
