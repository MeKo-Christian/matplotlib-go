package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
