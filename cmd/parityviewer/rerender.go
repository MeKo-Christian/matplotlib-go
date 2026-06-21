package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	goldenUpdateBuildTag               = "freetype"
	goldenUpdateOptionalVisualTestsEnv = "RUN_OPTIONAL_VISUAL_TESTS"
	goldenUpdateFallbackGoCache        = "/tmp/mpl-parity-gocache"
	goldenUpdateRunPatternAll          = "^TestGolden$"
	goldenUpdateTimeout                = 5 * time.Minute
)

func rerenderArtifact(repoRoot, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("missing name")
	}
	return runGoParityRender(repoRoot, name)
}

func rerenderAllArtifacts(repoRoot string) error {
	return runGoGoldenUpdate(repoRoot, goldenUpdateRunPatternAll)
}

func rerenderArtifacts(repoRoot string, names []string, progress func(string)) error {
	names = normalizeCaseNames(names)
	if len(names) == 0 {
		return errors.New("missing name")
	}
	if len(names) == 1 {
		err := runGoParityRender(repoRoot, names[0])
		if err == nil && progress != nil {
			progress(names[0])
		}
		return err
	}
	return runGoParityRenderBatch(repoRoot, names, progress)
}

func runGoGoldenUpdate(repoRoot, runPattern string) error {
	ctx, cancel := context.WithTimeout(context.Background(), goldenUpdateTimeout)
	defer cancel()

	cmd := newGoldenUpdateCommandContext(ctx, repoRoot, runPattern)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("go test timed out after %s for %q: %s", goldenUpdateTimeout, runPattern, msg)
		}
		return fmt.Errorf("go test failed: %w: %s", err, msg)
	}
	if runPattern != "" {
		outText := out.String()
		if strings.Contains(outText, "testing: warning: no tests to run") {
			return fmt.Errorf("no matching test for case %q", strings.TrimSuffix(strings.TrimPrefix(runPattern, "^"), "$"))
		}
	}
	return nil
}

func runGoParityRender(repoRoot, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), goldenUpdateTimeout)
	defer cancel()

	cmd := newParityRenderCommandContext(ctx, repoRoot, name)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("go run timed out after %s for %q: %s", goldenUpdateTimeout, name, msg)
		}
		return fmt.Errorf("go run failed: %w: %s", err, msg)
	}
	return nil
}

func runGoParityRenderBatch(repoRoot string, names []string, progress func(string)) error {
	ctx, cancel := context.WithTimeout(context.Background(), goldenUpdateTimeout)
	defer cancel()

	cmd := newParityRenderBatchCommandContext(ctx, repoRoot, names)

	var out bytes.Buffer
	progressWriter := newProgressLineWriter(names, progress)
	writer := io.MultiWriter(&out, progressWriter)
	cmd.Stdout = writer
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			msg = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("go run timed out after %s for %q: %s", goldenUpdateTimeout, strings.Join(names, ","), msg)
		}
		return fmt.Errorf("go run failed: %w: %s", err, msg)
	}
	progressWriter.flush()
	return nil
}

func newGoldenUpdateCommand(repoRoot, runPattern string) *exec.Cmd {
	return newGoldenUpdateCommandContext(context.Background(), repoRoot, runPattern)
}

func newGoldenUpdateCommandContext(ctx context.Context, repoRoot, runPattern string) *exec.Cmd {
	args := []string{"test", "-tags", goldenUpdateBuildTag, "-count", "1", "-timeout", goldenUpdateTimeout.String()}
	if runPattern != "" {
		args = append(args, "-run", runPattern)
	}
	args = append(args, "./test", "-update-golden")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repoRoot

	cmd.Env = goCommandEnv()
	cmd.Env = setEnv(cmd.Env, goldenUpdateOptionalVisualTestsEnv, "true")

	return cmd
}

func newParityRenderCommand(repoRoot, name string) *exec.Cmd {
	return newParityRenderCommandContext(context.Background(), repoRoot, name)
}

func newParityRenderCommandContext(ctx context.Context, repoRoot, name string) *exec.Cmd {
	args := []string{
		"run",
		"-tags",
		goldenUpdateBuildTag,
		"./test/parity/cmd",
		"--id",
		name,
		"--output-dir",
		filepath.Join("testdata", "golden"),
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repoRoot
	cmd.Env = goCommandEnv()

	return cmd
}

func newParityRenderBatchCommand(repoRoot string, names []string) *exec.Cmd {
	return newParityRenderBatchCommandContext(context.Background(), repoRoot, names)
}

func newParityRenderBatchCommandContext(ctx context.Context, repoRoot string, names []string) *exec.Cmd {
	args := []string{
		"run",
		"-tags",
		goldenUpdateBuildTag,
		"./test/parity/cmd",
		"--ids",
		strings.Join(normalizeCaseNames(names), ","),
		"--output-dir",
		filepath.Join("testdata", "golden"),
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repoRoot
	cmd.Env = goCommandEnv()

	return cmd
}

func goCommandEnv() []string {
	env := os.Environ()
	env = setEnv(env, "CGO_ENABLED", "1")
	if cacheDir, hasCache := os.LookupEnv("GOCACHE"); !hasCache || strings.TrimSpace(cacheDir) == "" {
		env = setEnv(env, "GOCACHE", goldenUpdateDefaultGoCache())
	}
	return env
}

func goldenUpdateDefaultGoCache() string {
	out, err := exec.Command("go", "env", "GOCACHE").Output()
	if err == nil {
		if cacheDir := strings.TrimSpace(string(out)); cacheDir != "" {
			return cacheDir
		}
	}
	return goldenUpdateFallbackGoCache
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	found := false
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			found = true
		}
	}
	if !found {
		env = append(env, prefix+value)
	}
	return env
}

func testNameFromCaseName(name string) string {
	return "^TestGolden/" + name + "$"
}

func normalizeCaseNames(names []string) []string {
	out := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

type progressLineWriter struct {
	mu        sync.Mutex
	buf       []byte
	names     []string
	completed int
	progress  func(string)
}

func newProgressLineWriter(names []string, progress func(string)) *progressLineWriter {
	return &progressLineWriter{names: normalizeCaseNames(names), progress: progress}
}

func (w *progressLineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, b := range p {
		if b == '\n' {
			w.handleLine(string(w.buf))
			w.buf = w.buf[:0]
			continue
		}
		w.buf = append(w.buf, b)
	}
	return len(p), nil
}

func (w *progressLineWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buf) > 0 {
		w.handleLine(string(w.buf))
		w.buf = w.buf[:0]
	}
}

func (w *progressLineWriter) handleLine(line string) {
	if w.progress == nil || !strings.HasPrefix(strings.TrimSpace(line), "wrote ") {
		return
	}
	if w.completed >= len(w.names) {
		return
	}
	name := w.names[w.completed]
	w.completed++
	w.progress(name)
}
