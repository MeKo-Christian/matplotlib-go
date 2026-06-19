package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	opts, err := parseCLIOptions()
	if err != nil {
		log.Fatalf("parse cli options: %v", err)
	}

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
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if r.FormValue("all") == "1" {
			if err := ensureRerenderSupportedWithExtra(opts.UseParity, opts.IncludeWebdemo, opts.RepoRoot, opts.ArtifactDir); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := rerenderAllArtifacts(opts.RepoRoot); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "missing name parameter", http.StatusBadRequest)
			return
		}
		if err := ensureRerenderSupportedWithExtra(opts.UseParity, opts.IncludeWebdemo, opts.RepoRoot, opts.ArtifactDir); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := rerenderArtifact(opts.RepoRoot, name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
