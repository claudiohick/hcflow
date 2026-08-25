// Package web implements the hcflow UI HTTP server.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/hickstein/hcflow/internal/config"
	"github.com/hickstein/hcflow/internal/git"
	"github.com/hickstein/hcflow/internal/github"
	"github.com/hickstein/hcflow/internal/status"
)

//go:embed static
var staticFiles embed.FS

// Server is the hcflow UI HTTP server.
type Server struct {
	dir string
	srv *http.Server
}

// New creates a new Server rooted at dir.
func New(dir string) *Server {
	return &Server{dir: dir}
}

// Start binds to a random available port on localhost and begins serving.
// It returns the URL (e.g. "http://127.0.0.1:PORT").
func (s *Server) Start(openBrowser bool) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("binding port: %w", err)
	}
	addr := fmt.Sprintf("http://127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.srv = &http.Server{Handler: mux}
	go func() {
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("hcflow ui: %v", err)
		}
	}()

	if openBrowser {
		go openURL(addr)
	}

	return addr, nil
}

// Shutdown stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// API
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/health", s.handleHealth)

	// Static files
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Printf("web: static files error: %v", err)
		return
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	cfg, _, err := config.Load(s.dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gitSvc := git.New(s.dir)
	url := gitSvc.RemoteOrigin()
	var ghSvc *github.Service
	if url != "" {
		owner, repo, parseErr := git.ParseGitHubRemote(url)
		if parseErr == nil {
			ghSvc = github.New(s.dir, owner, repo)
		}
	}

	st := status.Gather(cfg, gitSvc, ghSvc)

	// Build response
	resp := buildStatusResponse(st, url)
	json.NewEncoder(w).Encode(resp)
}

// statusResponse is the JSON shape returned by /api/status.
type statusResponse struct {
	Project  projectInfo   `json:"project"`
	Git      gitInfo       `json:"git"`
	PR       *prInfo       `json:"pr,omitempty"`
	Release  *releaseInfo  `json:"release,omitempty"`
	Pipeline []stageInfo   `json:"pipeline"`
	HCFlow   hcflowInfo    `json:"hcflow"`
	UpdatedAt string       `json:"updatedAt"`
}

type projectInfo struct {
	Name   string `json:"name"`
	Remote string `json:"remote"`
}

type gitInfo struct {
	Branch    string `json:"branch"`
	IsClean   bool   `json:"isClean"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	LatestTag string `json:"latestTag"`
}

type prInfo struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	State     string `json:"state"`
	Mergeable string `json:"mergeable"`
	CI        string `json:"ci"`
}

type releaseInfo struct {
	Current  string  `json:"current"`
	Proposed string  `json:"proposed,omitempty"`
	PRURL    string  `json:"prUrl,omitempty"`
	PRNumber int     `json:"prNumber,omitempty"`
	CI       string  `json:"ci,omitempty"`
	Status   string  `json:"status"`
}

type stageInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type hcflowInfo struct {
	Schema          int    `json:"schema"`
	WorkflowVersion string `json:"workflowVersion"`
}

func buildStatusResponse(st *status.State, remote string) statusResponse {
	resp := statusResponse{
		Project: projectInfo{
			Name:   st.Config.Project.Name,
			Remote: remote,
		},
		Git: gitInfo{
			Branch:    st.Git.Branch,
			IsClean:   st.Git.IsClean,
			Ahead:     st.Git.Ahead,
			Behind:    st.Git.Behind,
			LatestTag: st.Git.LatestTag,
		},
		HCFlow: hcflowInfo{
			Schema:          st.HCFlow.Schema,
			WorkflowVersion: st.HCFlow.WorkflowVersion,
		},
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if st.PR != nil {
		resp.PR = &prInfo{
			Number:    st.PR.Number,
			Title:     st.PR.Title,
			URL:       st.PR.URL,
			State:     st.PR.State,
			Mergeable: st.PR.Mergeable,
			CI:        string(github.PRCIStatus(st.PR)),
		}
	}

	if st.Release != nil {
		ri := &releaseInfo{
			Current:  st.Release.Current,
			Proposed: st.Release.Proposed,
		}
		if st.Release.PR != nil {
			ri.PRURL = st.Release.PR.URL
			ri.PRNumber = st.Release.PR.Number
			ri.CI = string(st.Release.CIStatus)
			ri.Status = "ready"
		} else {
			ri.Status = "no pending release"
		}
		resp.Release = ri
	}

	// Build pipeline
	resp.Pipeline = buildPipeline(st)

	return resp
}

func buildPipeline(st *status.State) []stageInfo {
	branch := st.Git.Branch
	defaultBranch := st.Config.Git.DefaultBranch

	stages := []stageInfo{
		{Name: "branch", Status: stageBranch(branch, defaultBranch)},
		{Name: "pr", Status: stagePR(st.PR)},
		{Name: "ci", Status: stageCI(st.PR)},
		{Name: "merge", Status: stageMerge(st.PR)},
		{Name: "main", Status: stageMain(branch, defaultBranch)},
		{Name: "release-please", Status: stageReleasePlease(st.Release)},
		{Name: "release-pr", Status: stageReleasePR(st.Release)},
		{Name: "release", Status: stageRelease(st.Release)},
	}
	return stages
}

func stageBranch(branch, defaultBranch string) string {
	if branch == defaultBranch || branch == "" {
		return "not-applicable"
	}
	return "active"
}

func stagePR(pr *github.PRState) string {
	if pr == nil {
		return "pending"
	}
	return "open"
}

func stageCI(pr *github.PRState) string {
	if pr == nil {
		return "pending"
	}
	return string(github.PRCIStatus(pr))
}

func stageMerge(pr *github.PRState) string {
	if pr == nil {
		return "pending"
	}
	if pr.Mergeable == "MERGEABLE" {
		return "ready"
	}
	return "blocked"
}

func stageMain(branch, defaultBranch string) string {
	if branch == defaultBranch {
		return "active"
	}
	return "pending"
}

func stageReleasePlease(rel *status.ReleaseState) string {
	if rel == nil {
		return "pending"
	}
	if rel.PR != nil {
		return "active"
	}
	return "pending"
}

func stageReleasePR(rel *status.ReleaseState) string {
	if rel == nil || rel.PR == nil {
		return "pending"
	}
	return "open"
}

func stageRelease(rel *status.ReleaseState) string {
	if rel == nil {
		return "pending"
	}
	if rel.Current != "" && rel.Proposed == "" {
		return "released"
	}
	if rel.Proposed != "" {
		return "ready"
	}
	return "pending"
}

// openURL opens url in the default browser.
func openURL(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default: // linux, etc.
		cmd = "xdg-open"
		args = []string{url}
	}
	time.Sleep(500 * time.Millisecond)
	exec.Command(cmd, args...).Start()
}
