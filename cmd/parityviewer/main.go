package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	opts, err := parseCLIOptions()
	if err != nil {
		log.Fatalf("parse cli options: %v", err)
	}
	rerenderManager := newRerenderJobManager(func(ctx context.Context, names []string, progress func(string)) error {
		return rerenderArtifacts(opts.RepoRoot, names, progress)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		result, err := loadCasesUnified(opts.UseParity, opts.IncludeWebdemo, opts.ParityDir, opts.BaselineDir, opts.ArtifactDir, opts.WebBaselineDir, opts.WebArtifactDir, opts.NameFilter, opts.NamePrefix)
		if err != nil {
			http.Error(w, fmt.Sprintf("load parity cases: %v", err), http.StatusInternalServerError)
			return
		}
		renderPage(w, result, buildViewOptionsWithExtra(opts.UseParity, opts.IncludeWebdemo, opts.RepoRoot, opts.ArtifactDir))
	})
	http.HandleFunc("/rerender", func(w http.ResponseWriter, r *http.Request) {
		handleRerender(w, r, buildViewOptionsWithExtra(opts.UseParity, opts.IncludeWebdemo, opts.RepoRoot, opts.ArtifactDir), rerenderManager)
	})
	http.HandleFunc("/rerender/status", func(w http.ResponseWriter, r *http.Request) {
		handleRerenderStatus(w, r, rerenderManager)
	})
	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		handleImage(w, r, opts)
	})

	if opts.PrintOnly {
		result, err := loadCasesUnified(opts.UseParity, opts.IncludeWebdemo, opts.ParityDir, opts.BaselineDir, opts.ArtifactDir, opts.WebBaselineDir, opts.WebArtifactDir, opts.NameFilter, opts.NamePrefix)
		if err != nil {
			log.Fatalf("load parity cases: %v", err)
		}
		printCases(os.Stdout, result)
		return
	}

	addr := ":" + opts.Port
	log.Printf("Parity viewer running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
