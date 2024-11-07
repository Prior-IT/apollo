package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/fsnotify/fsnotify"
	"github.com/prior-it/apollo/config"
)

// Run a new file watcher on the current thread.
// This will publish new FileChanged events whenever a file in the current directory-tree changes.
//
//nolint:cyclop
func runFileWatcher(ctx context.Context, cfg *config.Config, outCh chan<- string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("cannot create watcher: %w", err)
	}
	defer watcher.Close()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not access the current working directory: %v\n", err)
	}

	var ignoreDirs []string
	for _, d := range cfg.Tools.IgnoreDirs {
		dir := cwd + "/" + d
		ignoreDirs = append(ignoreDirs, dir)
	}
	checkmark := StyleSucces.Inline(true).Render("✓")
	cross := StyleError.Inline(true).Render("✗")
	err = filepath.WalkDir(cwd, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// If this directory was ignored, skip it
			if slices.Contains(ignoreDirs, path) {
				if debug {
					outCh <- fmt.Sprintf("%v Skipping directory: %v", cross, path)
				}
				return fs.SkipDir
			}

			if debug {
				outCh <- fmt.Sprintf("%v Watching directory: %q", checkmark, path)
			}
			err := watcher.Add(path)
			if err != nil {
				return fmt.Errorf("cannot add directory %q to watcher: %w", path, err)
			}

		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("cannot walk filetree: %w", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op == fsnotify.Write {
				base := filepath.Base(event.Name)
				if base == "doc.go" {
					fmt.Printf("ignoring documentation change: %q\n", event.Name)
				} else {
					newFileChangedEvent(event.Name)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("watcher err: %v\n", err)
		case <-ctx.Done():
			return nil
		}
	}
}
