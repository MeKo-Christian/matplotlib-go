package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image/color"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestNewParityRenderBatchCommandRendersMultipleIDsInOneProcess(t *testing.T) {
	t.Setenv("GOCACHE", "")

	cmd := newParityRenderBatchCommand("/tmp/repo", []string{"basic_line", "scatter_basic"})

	if cmd.Dir != "/tmp/repo" {
		t.Fatalf("Dir = %q, want %q", cmd.Dir, "/tmp/repo")
	}
	requiredArgs := []string{
		"run",
		"-tags",
		goldenUpdateBuildTag,
		"./test/parity/cmd",
		"--ids",
		"basic_line,scatter_basic",
		"--output-dir",
		filepath.Join("testdata", "golden"),
	}
	for _, arg := range requiredArgs {
		if !slices.Contains(cmd.Args, arg) {
			t.Fatalf("Args missing %q: %v", arg, cmd.Args)
		}
	}
	if slices.Contains(cmd.Args, "--all") {
		t.Fatalf("batch command should target loaded ids, not --all: %v", cmd.Args)
	}
}

func TestRerenderJobManagerTracksProgressAndCompletion(t *testing.T) {
	manager := newRerenderJobManager(func(ctx context.Context, names []string, progress func(string)) error {
		for _, name := range names {
			progress(name)
		}
		return nil
	})

	job, err := manager.start([]string{"basic_line", "scatter_basic"})
	if err != nil {
		t.Fatalf("start returned error: %v", err)
	}

	var status rerenderJobStatus
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		status, _ = manager.status(job.ID)
		if status.Status == rerenderStatusDone {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if status.Status != rerenderStatusDone {
		t.Fatalf("status = %+v, want done", status)
	}
	if status.Total != 2 || status.Completed != 2 || status.Current != "scatter_basic" {
		t.Fatalf("unexpected final progress: %+v", status)
	}
}

func TestRerenderJobManagerRejectsConcurrentJobs(t *testing.T) {
	block := make(chan struct{})
	manager := newRerenderJobManager(func(ctx context.Context, names []string, progress func(string)) error {
		<-block
		return nil
	})

	if _, err := manager.start([]string{"basic_line"}); err != nil {
		t.Fatalf("first start returned error: %v", err)
	}
	if _, err := manager.start([]string{"scatter_basic"}); err == nil {
		t.Fatal("expected concurrent start to fail")
	}
	close(block)
}

func TestRerenderHandlerStartsJobAndReturnsJSON(t *testing.T) {
	manager := newRerenderJobManager(func(ctx context.Context, names []string, progress func(string)) error {
		progress(names[0])
		return nil
	})
	form := url.Values{}
	form.Add("name", "basic_line")
	form.Add("name", "scatter_basic")

	req := httptest.NewRequest(http.MethodPost, "/rerender", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handleRerender(rec, req, viewOptions{CanRerender: true}, manager)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	var body rerenderJobStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID == "" || body.Total != 2 {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

func TestRenderPageUsesLazyImageURLsInsteadOfEmbeddedBase64(t *testing.T) {
	var out bytes.Buffer
	renderPage(&out, loadResult{Cases: []caseEntry{{
		Suite:       "plots",
		Baseline:    "matplotlib_ref",
		Name:        "basic_line",
		RefWidth:    640,
		RefHeight:   360,
		ActWidth:    640,
		ActHeight:   360,
		RefImageURL: "/image?kind=baseline&name=basic_line",
		ActImageURL: "/image?kind=artifact&name=basic_line",
		RawDiffURL:  "/image?kind=diff-raw&name=basic_line",
		AmpDiffURL:  "/image?kind=diff-amp&name=basic_line",
	}}}, viewOptions{CanRerender: true})

	html := out.String()
	if strings.Contains(html, "data:image/png;base64,") {
		t.Fatalf("page should not embed base64 PNG payloads: %s", html)
	}
	requiredSnippets := []string{
		`data-src="/image?kind=baseline&amp;name=basic_line"`,
		`data-src="/image?kind=artifact&amp;name=basic_line"`,
		`data-src="/image?kind=diff-raw&amp;name=basic_line"`,
		"function loadCardImages(card)",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(html, snippet) {
			t.Fatalf("page missing lazy image snippet %q in %s", snippet, html)
		}
	}
}

func TestPageFooterUsesSingleBulkRerenderJobWithProgress(t *testing.T) {
	requiredSnippets := []string{
		`<progress id="rerender-progress" value="0" max="1"></progress>`,
		"function startRerender(names) {",
		"params.append('name', name);",
		"function pollRerenderJob(jobID) {",
		"function updateRerenderProgress(status) {",
		"return startRerender(names).then(function(job) {",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageHeader, snippet) && !strings.Contains(pageFooter, snippet) {
			t.Fatalf("page markup missing bulk progress snippet %q", snippet)
		}
	}
	if strings.Contains(pageFooter, "chain = chain.then(function() { return rerenderArtifact(name); });") {
		t.Fatalf("bulk rerender still loops one request per card")
	}
}

func TestImageHandlerServesArtifactAndGeneratedDiff(t *testing.T) {
	baseDir := t.TempDir()
	artifactDir := t.TempDir()
	writePNG(t, filepath.Join(baseDir, "basic_line.png"), solidRGBA(2, 1, color.RGBA{R: 0, G: 0, B: 0, A: 255}))
	writePNG(t, filepath.Join(artifactDir, "basic_line.png"), solidRGBA(2, 1, color.RGBA{R: 10, G: 0, B: 0, A: 255}))

	opts := cliOptions{
		BaselineDir: baseDir,
		ArtifactDir: artifactDir,
	}
	for _, kind := range []string{"artifact", "diff-raw"} {
		req := httptest.NewRequest(http.MethodGet, "/image?suite=plots&baseline="+filepath.Base(baseDir)+"&name=basic_line&kind="+kind, nil)
		rec := httptest.NewRecorder()

		handleImage(rec, req, opts)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body %s", kind, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Content-Type"); got != "image/png" {
			t.Fatalf("%s Content-Type = %q, want image/png", kind, got)
		}
		if rec.Body.Len() == 0 {
			t.Fatalf("%s response body is empty", kind)
		}
	}
}

func TestCaseEntryCacheInvalidatesWhenFileMetadataChanges(t *testing.T) {
	baseDir := t.TempDir()
	baselinePath := filepath.Join(baseDir, "baseline.png")
	artifactPath := filepath.Join(baseDir, "artifact.png")
	writePNG(t, baselinePath, solidRGBA(1, 1, color.RGBA{R: 0, G: 0, B: 0, A: 255}))
	writePNG(t, artifactPath, solidRGBA(1, 1, color.RGBA{R: 0, G: 0, B: 0, A: 255}))

	cache := newCaseEntryCache()
	key, ok := newCaseEntryCacheKey(baselinePath, artifactPath)
	if !ok {
		t.Fatal("expected cache key for existing files")
	}
	cache.store(key, cachedCaseEntry{Stats: metrics{RMSE: 1}})
	if cached, ok := cache.lookup(key); !ok || cached.Stats.RMSE != 1 {
		t.Fatalf("cache lookup = %+v, %v; want RMSE 1 hit", cached, ok)
	}

	later := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(artifactPath, later, later); err != nil {
		t.Fatalf("chtimes artifact: %v", err)
	}
	nextKey, ok := newCaseEntryCacheKey(baselinePath, artifactPath)
	if !ok {
		t.Fatal("expected cache key after metadata change")
	}
	if _, ok := cache.lookup(nextKey); ok {
		t.Fatal("cache should miss after artifact mtime changes")
	}
}
