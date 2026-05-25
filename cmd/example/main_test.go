package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestExampleRequiredCapabilitiesAllowGoBasic(t *testing.T) {
	_, _, err := backends.NewRenderer(string(backends.GoBasic), backends.TestDefaultConfig(120, 80), exampleRequiredCapabilities())
	if err != nil {
		t.Fatalf("example required capabilities should allow GoBasic: %v", err)
	}
}

func hasNonBackgroundPixel(img image.Image, bg color.RGBA) bool {
	if img == nil {
		return false
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != bg {
				return true
			}
		}
	}
	return false
}

func TestShowcaseRegistryRendersWithGoBasic(t *testing.T) {
	for id, plot := range registry {
		id := id
		plot := plot
		t.Run(id, func(t *testing.T) {
			fig := plot()
			w := int(fig.SizePx.X)
			h := int(fig.SizePx.Y)
			r, _, err := backends.NewRenderer(string(backends.GoBasic), backends.Config{
				Width:      w,
				Height:     h,
				Background: render.Color{R: 1, G: 1, B: 1, A: 1},
				DPI:        fig.RC.DPI,
			}, exampleRequiredCapabilities())
			if err != nil {
				t.Fatalf("create GoBasic renderer: %v", err)
			}
			core.DrawFigure(fig, r)

			exporter, ok := r.(render.RGBAExporter)
			if !ok {
				t.Fatal("GoBasic renderer should expose RGBA output")
			}
			if !hasNonBackgroundPixel(exporter.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
				t.Fatal("GoBasic showcase output is blank")
			}
		})
	}
}

func TestShowcaseRegistryCoversCatalogShowcases(t *testing.T) {
	for _, c := range examplecatalog.Cases() {
		if !c.Showcase {
			continue
		}
		if _, ok := registry[c.ID]; !ok {
			t.Fatalf("catalog showcase %q is missing from cmd/example registry", c.ID)
		}
	}
}

func TestExampleCommandSmokeFormats(t *testing.T) {
	for _, format := range []string{"png", "svg", "pdf", "ps"} {
		format := format
		t.Run(format, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "basic_line."+format)
			stderr, ok := runExampleCommand(t, "-name", "basic_line", "-format", format, "-o", path)
			if !ok {
				t.Fatalf("cmd/example -format %s failed:\n%s", format, stderr)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("expected output %s: %v\nstderr:\n%s", path, err, stderr)
			}
			if info.Size() == 0 {
				t.Fatalf("expected non-empty %s output", format)
			}
		})
	}
}

func TestExampleCommandWritesPGF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "basic_line.pgf")
	stderr, ok := runExampleCommand(t, "-name", "basic_line", "-format", "pgf", "-o", path)
	if !ok {
		t.Fatalf("cmd/example -format pgf failed:\n%s", stderr)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("PGF output missing: %v\nstderr:\n%s", err, stderr)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty PGF output")
	}
}

func TestExampleCommandForwardsPDFOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "basic_line.pdf")
	stderr, ok := runExampleCommand(
		t,
		"-name", "basic_line",
		"-format", "pdf",
		"-o", path,
		"-pdf-title", "Shared Options",
		"-pdf-creation-date", "2024-05-06T07:08:09Z",
		"-pdf-font-policy", "path",
	)
	if !ok {
		t.Fatalf("cmd/example PDF options failed:\n%s", stderr)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read PDF output: %v", err)
	}
	if !bytes.Contains(data, []byte("/Title (Shared Options)")) {
		t.Fatalf("PDF output missing title metadata:\n%s", data)
	}
	if !bytes.Contains(data, []byte("/CreationDate (D:20240506070809Z)")) {
		t.Fatalf("PDF output missing creation date metadata:\n%s", data)
	}
}

func TestExampleCommandTestHelper(t *testing.T) {
	if os.Getenv("MATPLOTLIB_GO_EXAMPLE_MAIN") != "1" {
		return
	}
	args := os.Args
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = append([]string{os.Args[0]}, args...)
	main()
	os.Exit(0)
}

func runExampleCommand(t *testing.T, args ...string) (string, bool) {
	t.Helper()
	cmdArgs := append([]string{"-test.run=TestExampleCommandTestHelper", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(
		os.Environ(),
		"MATPLOTLIB_GO_EXAMPLE_MAIN=1",
		"MATPLOTLIB_BACKEND=",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.String(), err == nil
}
