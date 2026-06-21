package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cwbudde/matplotlib-go/test/parity"
)

func main() {
	id := flag.String("id", "", "Parity example ID to render")
	ids := flag.String("ids", "", "Comma-separated parity example IDs to render")
	outputDir := flag.String("output-dir", ".", "Directory to write PNG files")
	list := flag.Bool("list", false, "List available parity example IDs and exit")
	all := flag.Bool("all", false, "Render every parity example")
	flag.Parse()

	if *list {
		for _, c := range parity.Cases() {
			fmt.Println(c.ID)
		}
		return
	}

	renderIDs, err := renderIDsFromFlags(*id, *ids, *all)
	if err != nil {
		exitf("%v", err)
	}
	if renderIDs == nil {
		for _, c := range parity.Cases() {
			writeCase(c.ID, *outputDir)
		}
		return
	}

	for _, id := range renderIDs {
		writeCase(id, *outputDir)
	}
}

func renderIDsFromFlags(id, ids string, all bool) ([]string, error) {
	id = strings.TrimSpace(id)
	ids = strings.TrimSpace(ids)
	selectedModes := 0
	if id != "" {
		selectedModes++
	}
	if ids != "" {
		selectedModes++
	}
	if all {
		selectedModes++
	}
	if selectedModes > 1 {
		return nil, errors.New("use only one of --id, --ids, or --all")
	}
	if all {
		return nil, nil
	}
	var out []string
	if id != "" {
		out = append(out, id)
	} else {
		for _, part := range strings.Split(ids, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	if len(out) == 0 {
		return nil, errors.New("missing --id, --ids, or --all/--list")
	}
	return out, nil
}

func writeCase(id, outputDir string) {
	path, err := parity.RenderToFile(id, outputDir)
	if err != nil {
		exitf("%v", err)
	}
	fmt.Printf("wrote %s\n", path)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
