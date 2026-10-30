package olympus

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/modules/plutus"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

//go:embed static/olympus.css
var stylesheet []byte

const (
	defaultMaxUpload = 10 << 20 // 10 MiB
	multipartSlack   = 64 << 10 // room for multipart framing around the file
	maxRowsShown     = 20
)

// KeyInfo describes one connector key. It never carries the key itself or its hash.
type KeyInfo struct {
	Name       string
	Scopes     []string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	Revoked    bool
}

// Info is what the config page shows: non-secret facts only.
type Info struct {
	Mode             string
	Version          string
	PublicURL        string
	DatabaseOK       bool
	TransactionCount int
	Keys             []KeyInfo
}

// Source supplies the config page's facts.
type Source interface {
	Info(ctx context.Context) (Info, error)
}

// Importer merges an uploaded backup into the owner's data. The owner is fixed by whoever builds
// the Importer; the page never chooses or learns a user id.
type Importer interface {
	ImportBackup(ctx context.Context, r io.Reader, maxBytes int64) (plutus.ImportResult, error)
}

// Options configures the admin page.
type Options struct {
	Addr           string // listen address; its port decides which Host values are accepted
	Source         Source
	Importer       Importer
	MaxUploadBytes int64 // default 10 MiB
	Log            *slog.Logger
}

// Server is the local admin page.
type Server struct {
	opts Options
	g    *guard
	tmpl *template.Template
	log  *slog.Logger
	h    http.Handler
}

// New builds the page. Handler returns it with all request checks applied.
func New(o Options) (*Server, error) {
	if o.Source == nil || o.Importer == nil {
		return nil, errors.New("olympus: Source and Importer are required")
	}
	if o.MaxUploadBytes <= 0 {
		o.MaxUploadBytes = defaultMaxUpload
	}
	g, err := newGuard(o.Addr, o.MaxUploadBytes+multipartSlack)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}
	log := o.Log
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	s := &Server{opts: o, g: g, tmpl: tmpl, log: log}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("POST /import", s.handleImport)
	mux.HandleFunc("GET /static/olympus.css", s.handleCSS)
	s.h = s.logRequests(g.Wrap(mux))
	return s, nil
}

// Handler is the page, guarded and logged.
func (s *Server) Handler() http.Handler { return s.h }

type keyView struct{ Name, Scopes, Created, LastUsed, Status string }

type indexView struct {
	Title       string
	Info        Info
	StatusError bool
	KeyRows     []keyView
	CSRF        string
	MaxMB       int64
}

func stamp(t time.Time) string { return t.UTC().Format("2006-01-02 15:04 UTC") }

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	v := indexView{Title: "Nemosyne", CSRF: s.g.Token(), MaxMB: s.opts.MaxUploadBytes >> 20}
	info, err := s.opts.Source.Info(r.Context())
	if err != nil {
		// The error text can hold connection details; only its type is logged.
		s.log.Error("status unavailable", "err_type", fmt.Sprintf("%T", err))
		v.StatusError = true
	} else {
		v.Info = info
		for _, k := range info.Keys {
			kv := keyView{Name: k.Name, Scopes: strings.Join(k.Scopes, ", "), Created: stamp(k.CreatedAt), LastUsed: "never", Status: "active"}
			if k.LastUsedAt != nil {
				kv.LastUsed = stamp(*k.LastUsedAt)
			}
			if k.Revoked {
				kv.Status = "revoked"
			}
			v.KeyRows = append(v.KeyRows, kv)
		}
	}
	s.render(w, http.StatusOK, "index.html.tmpl", v)
}

type resultView struct {
	Title          string
	OK             bool
	Heading        string
	Message        string
	ShowCounts     bool
	Added, Skipped int
	Rows           []plutus.RowError
	More           int
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	fail := func(status int, msg string) {
		s.render(w, status, "result.html.tmpl", resultView{Title: "Import failed", Heading: "Import failed", Message: msg})
	}

	file, hdr, err := r.FormFile("file")
	if err != nil || hdr.Size == 0 {
		fail(http.StatusBadRequest, "Choose a backup file first.")
		return
	}
	defer file.Close()
	if hdr.Size > s.opts.MaxUploadBytes {
		fail(http.StatusRequestEntityTooLarge, fmt.Sprintf("That file is larger than the %d MB limit. Nothing was imported.", s.opts.MaxUploadBytes>>20))
		return
	}

	res, err := s.opts.Importer.ImportBackup(r.Context(), file, s.opts.MaxUploadBytes)
	var rows *plutus.RowsError
	switch {
	case err == nil:
		s.render(w, http.StatusOK, "result.html.tmpl", resultView{
			Title: "Import finished", OK: true, Heading: "Import finished", ShowCounts: true, Added: res.Added, Skipped: res.Skipped,
		})
	case errors.Is(err, plutus.ErrBackup):
		fail(http.StatusBadRequest, "That file is not a Nemosyne backup, or it comes from a newer version. Nothing was imported.")
	case errors.As(err, &rows):
		v := resultView{Title: "Import failed", Heading: "Import failed", Message: "Some rows are not valid. Nothing was imported.", Rows: rows.Rows}
		if len(v.Rows) > maxRowsShown {
			v.More = len(v.Rows) - maxRowsShown
			v.Rows = v.Rows[:maxRowsShown]
		}
		s.render(w, http.StatusUnprocessableEntity, "result.html.tmpl", v)
	default:
		// Database errors can quote the row that failed, so only the error's type is logged.
		s.log.Error("import failed", "err_type", fmt.Sprintf("%T", err))
		fail(http.StatusInternalServerError, "Something went wrong while importing. Nothing was imported. Check the server log.")
	}
}

func (s *Server) handleCSS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(stylesheet)
}

// render executes a template into a buffer first, so a template error never sends half a page.
func (s *Server) render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		s.log.Error("render failed", "template", name)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// logRequests records method, path and status only: never the query string, the body or any header.
func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.log.Info("admin request", "method", r.Method, "path", r.URL.Path, "status", rec.status,
			"ms", time.Since(start).Milliseconds())
	})
}
