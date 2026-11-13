// Package ui serves hardenme's local page: the score, what could happen
// today, every finding with its reason, a "fix everything" button and
// guided fixes that tick themselves.
//
// The server can change settings, so it is locked down: it listens only on
// 127.0.0.1, every request needs the secret token of this run, the Host and
// Origin headers must be its own, and it only ever runs fixes from the
// embedded catalog. Nothing from the page becomes a command.
package ui

import (
	"context"
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
	"github.com/XexusCode/hardenme/internal/story"
)

//go:embed page.html
var page []byte

// Deps is everything the server needs. Tests replace the runners and stores.
type Deps struct {
	Platform  string
	OSVersion string
	Lang      check.Lang
	Catalog   []check.Check
	Stories   []story.Story
	Probe     runner.Runner // reads state
	Apply     runner.Runner // applies changes; waits for the password dialog
	Log       func() (fix.Log, error)
	Profiles  profile.Store
	// Fallback is the profile used until the person saves one.
	Fallback profile.Profile
	Now      func() time.Time
}

// Server holds the last report between requests.
type Server struct {
	deps  Deps
	token string
	// host is "127.0.0.1:<port>", the only Host and Origin accepted.
	host string

	mu       sync.Mutex
	report   scan.Report
	scanned  bool
	lastSeen time.Time
	quit     chan struct{}
	quitOnce sync.Once
}

// New returns a server for requests sent to host with the given token.
func New(deps Deps, host, token string) *Server {
	return &Server{deps: deps, host: host, token: token, lastSeen: deps.Now(), quit: make(chan struct{})}
}

// Done is closed when the page asks to quit.
func (s *Server) Done() <-chan struct{} { return s.quit }

// Idle is how long ago the page last talked to the server.
func (s *Server) Idle() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deps.Now().Sub(s.lastSeen)
}

// Handler is the HTTP entry point.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handlePage)
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("POST /api/fix", s.handleFix)
	mux.HandleFunc("POST /api/recheck", s.handleRecheck)
	mux.HandleFunc("POST /api/open", s.handleOpen)
	mux.HandleFunc("POST /api/undo", s.handleUndo)
	mux.HandleFunc("POST /api/profile", s.handleProfile)
	mux.HandleFunc("POST /api/quit", s.handleQuit)
	return s.guard(mux)
}

// guard rejects anything that is not this run's own page: a wrong Host
// (DNS rebinding), a foreign Origin (another site's script) or a missing
// token.
func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != s.host {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+s.host {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		token := r.Header.Get("X-Hardenme-Token")
		if token == "" && r.Method == http.MethodGet && r.URL.Path == "/" {
			token = r.URL.Query().Get("t")
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		s.mu.Lock()
		s.lastSeen = s.deps.Now()
		s.mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handlePage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	_, _ = w.Write(page)
}

// Item is one finding as the page shows it.
type Item struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Status    check.Status    `json:"status"`
	Relevance check.Relevance `json:"relevance"`
	// Section is where the page lists it: essential, recommended, optional,
	// manual, unknown, pass or na.
	Section string        `json:"section"`
	Why     string        `json:"why"`
	Steps   []string      `json:"steps"`
	Current string        `json:"current,omitempty"`
	Fix     check.FixKind `json:"fix"`
	// CanOpen is true when hardenme can open the setting for the person.
	CanOpen bool   `json:"can_open"`
	Warning string `json:"warning,omitempty"`
}

// State is everything the page draws.
type State struct {
	Lang         check.Lang      `json:"lang"`
	Platform     string          `json:"platform"`
	OSVersion    string          `json:"os_version"`
	Profile      profile.Profile `json:"profile"`
	ProfileSaved bool            `json:"profile_saved"`
	Score        *Score          `json:"score"`
	Stories      []string        `json:"stories"`
	// Safe lists what a thief could not do, thanks to the settings that are right.
	Safe      []string          `json:"safe"`
	Items     []Item            `json:"items"`
	AutoFixes int               `json:"auto_fixes"`
	CanUndo   bool              `json:"can_undo"`
	Strings   map[string]string `json:"strings"`
}

// Score is the weighted score with its label.
type Score struct {
	Value int    `json:"value"`
	Grade string `json:"grade"`
	Label string `json:"label"`
}

// Reply is what every action returns: the new state and what happened.
type Reply struct {
	State    State                  `json:"state"`
	Outcomes map[string]fix.Outcome `json:"outcomes,omitempty"`
	// Message is a plain-language note about the action, when there is one.
	Message string `json:"message,omitempty"`
}

func (s *Server) profile() (profile.Profile, bool) {
	p, err := s.deps.Profiles.Load()
	if err != nil {
		return s.deps.Fallback, false
	}
	return p, true
}

// rescan runs every check. The caller holds s.mu.
func (s *Server) rescan(ctx context.Context) {
	p, _ := s.profile()
	s.report = scan.Scan(ctx, s.deps.Platform, s.deps.Catalog, p, s.deps.Probe)
	s.report.OSVersion = s.deps.OSVersion
	s.scanned = true
}

// state builds the page's view of the last report. The caller holds s.mu.
func (s *Server) state() State {
	story.Add(&s.report, s.deps.Stories)
	p, saved := s.profile()
	st := State{
		Lang: s.deps.Lang, Platform: s.deps.Platform, OSVersion: s.deps.OSVersion,
		Profile: p, ProfileSaved: saved,
		Stories: []string{}, Safe: []string{}, Items: []Item{}, Strings: strs[s.deps.Lang],
	}
	if sc := s.report.Score(); sc.Value >= 0 {
		st.Score = &Score{Value: sc.Value, Grade: sc.Grade, Label: strs[s.deps.Lang]["label_"+sc.Grade]}
	}
	for _, sto := range s.report.Stories {
		text := sto.Text[s.deps.Lang]
		if text == "" {
			text = sto.Text[check.EN]
		}
		if sto.Safe {
			st.Safe = append(st.Safe, text)
		} else {
			st.Stories = append(st.Stories, text)
		}
	}
	for _, e := range s.report.Entries {
		text := e.Check.TextFor(s.deps.Lang)
		it := Item{
			ID: e.Check.ID, Status: e.Result.Status, Relevance: e.Relevance, Section: section(e),
			Why: strings.Join(strings.Fields(text.Why), " "), Steps: text.FixSteps, Current: e.Result.Current,
			Fix: e.Check.FixKind(), CanOpen: openable(text.DeepLink), Warning: text.FixWarning,
		}
		switch it.Section {
		case "pass":
			it.Title = text.TitlePass
		case "unknown", "na":
			it.Title = text.Title
		default:
			it.Title = text.TitleFail
		}
		st.Items = append(st.Items, it)
	}
	for _, c := range fix.Candidates(s.report) {
		if c.Kind == check.FixAuto {
			st.AutoFixes++
		}
	}
	if l, err := s.deps.Log(); err == nil {
		_, st.CanUndo, _ = l.Last()
	}
	return st
}

// section mirrors the text report's grouping.
func section(e scan.Entry) string {
	switch {
	case e.Relevance == check.NotApplicable:
		return "na"
	case e.Result.Status == check.Pass:
		return "pass"
	case e.Result.Status == check.Manual:
		return "manual"
	case e.Result.Status == check.Unknown, e.Result.Status == check.Skipped:
		return "unknown"
	}
	return string(e.Relevance)
}

// openable reports whether a deep link is a System Settings pane, the only
// thing the server will ever pass to `open`.
func openable(link string) bool {
	return strings.HasPrefix(link, "x-apple.systempreferences:")
}

func (s *Server) reply(w http.ResponseWriter, r Reply) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(r)
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.scanned || r.URL.Query().Get("fresh") == "1" {
		s.rescan(r.Context())
	}
	s.reply(w, Reply{State: s.state()})
}

// handleFix applies every automatic fix that is pending, with one password
// dialog, logs the changes and checks again.
func (s *Server) handleFix(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.scanned {
		s.rescan(r.Context())
	}
	var cs []check.Check
	var ids []string
	for _, c := range fix.Candidates(s.report) {
		if c.Kind == check.FixAuto {
			cs = append(cs, c.Entry.Check)
			ids = append(ids, c.Entry.Check.ID)
		}
	}
	msg := strs[s.deps.Lang]
	if len(cs) == 0 {
		s.reply(w, Reply{State: s.state(), Message: msg["nothing_auto"]})
		return
	}
	// The request context ends if the tab closes; the change must not be
	// cut halfway, so it runs on its own.
	ctx := context.WithoutCancel(r.Context())
	steps, _ := fix.Prepare(ctx, s.deps.Probe, cs)
	res := fix.Run(ctx, s.deps.Apply, steps)
	out := Reply{}
	switch {
	case res.Cancelled:
		out.Message = msg["dialog_cancelled"]
	case res.AdminErr != nil:
		out.Message = msg["admin_error"]
	}
	if changes := res.Changes(); len(changes) > 0 {
		l, err := s.deps.Log()
		if err == nil {
			err = l.Append(fix.Batch{Time: s.deps.Now().UTC(), Changes: changes})
		}
		if err != nil {
			out.Message = msg["log_error"]
		}
	}
	p, _ := s.profile()
	out.Outcomes = fix.Recheck(ctx, check.Env{Runner: s.deps.Probe, Profile: p}, &s.report, res, ids)
	out.State = s.state()
	s.reply(w, out)
}

type idRequest struct {
	ID string `json:"id"`
}

// entry finds a report entry by the ID sent by the page. The ID is only
// ever used to look up the catalog.
func (s *Server) entry(w http.ResponseWriter, r *http.Request) *scan.Entry {
	var req idRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return nil
	}
	if !s.scanned {
		s.rescan(r.Context())
	}
	for i := range s.report.Entries {
		if s.report.Entries[i].Check.ID == req.ID {
			return &s.report.Entries[i]
		}
	}
	http.Error(w, "unknown check", http.StatusNotFound)
	return nil
}

// handleRecheck reads one setting again. The page calls it every few
// seconds while the person changes that setting by hand.
func (s *Server) handleRecheck(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.entry(w, r)
	if e == nil {
		return
	}
	if e.Relevance != check.NotApplicable {
		p, _ := s.profile()
		e.Result = e.Check.Run(r.Context(), check.Env{Runner: s.deps.Probe, Profile: p})
	}
	s.reply(w, Reply{State: s.state()})
}

// handleOpen opens the System Settings pane of one check.
func (s *Server) handleOpen(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.entry(w, r)
	if e == nil {
		return
	}
	link := e.Check.TextFor(s.deps.Lang).DeepLink
	if !openable(link) {
		http.Error(w, "nothing to open", http.StatusBadRequest)
		return
	}
	_, _ = s.deps.Probe.Run(r.Context(), "/usr/bin/open", link)
	s.reply(w, Reply{State: s.state()})
}

// handleUndo reverts the last batch of automatic changes.
func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := strs[s.deps.Lang]
	fail := func(key string) {
		if !s.scanned {
			s.rescan(r.Context())
		}
		s.reply(w, Reply{State: s.state(), Message: msg[key]})
	}
	l, err := s.deps.Log()
	if err != nil {
		fail("log_error")
		return
	}
	batch, ok, err := l.Last()
	if err != nil || !ok {
		fail("undo_nothing")
		return
	}
	steps, err := fix.UndoSteps(batch, s.deps.Catalog)
	if err != nil {
		fail("undo_not_allowed")
		return
	}
	ctx := context.WithoutCancel(r.Context())
	res := fix.Run(ctx, s.deps.Apply, steps)
	if res.Cancelled {
		fail("dialog_cancelled")
		return
	}
	if res.AdminErr != nil {
		fail("admin_error")
		return
	}
	for _, st := range res.Steps {
		if !st.Ran || st.ExitCode != 0 {
			fail("undo_failed")
			return
		}
	}
	if err := l.DropLast(); err != nil {
		fail("log_error")
		return
	}
	s.rescan(ctx)
	s.reply(w, Reply{State: s.state(), Message: msg["undo_done"]})
}

// handleProfile saves how the person uses the device and scores again.
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	var p profile.Profile
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	risk, err := profile.ParseRisk(string(p.Risk))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	p.Risk = risk
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.deps.Profiles.Save(p); err != nil {
		http.Error(w, "could not save", http.StatusInternalServerError)
		return
	}
	s.rescan(r.Context())
	s.reply(w, Reply{State: s.state()})
}

func (s *Server) handleQuit(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	s.quitOnce.Do(func() { close(s.quit) })
}
