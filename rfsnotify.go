// Package rfsnotify implements recursive folder monitoring by wrapping the excellent fsnotify library
package rfsnotify

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// DirFilter ...
type DirFilter func(path string, info os.FileInfo) bool

// Watcher wraps fsnotify.Watcher. When fsnotify adds recursive watches, you should be able to switch your code to use fsnotify.Watcher
type Watcher struct {
	Events chan fsnotify.Event
	Errors chan error

	done     chan struct{}
	fsnotify *fsnotify.Watcher
	isClosed bool
}

// NewWatcher establishes a new watcher with the underlying OS and begins waiting for events.
func NewWatcher() (*Watcher, error) {
	fsWatch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{}
	w.fsnotify = fsWatch
	w.Events = make(chan fsnotify.Event)
	w.Errors = make(chan error)
	w.done = make(chan struct{})

	go w.start()

	return w, nil
}

// Add starts watching the named file or directory (non-recursively).
func (w *Watcher) Add(name string) error {
	if w.isClosed {
		return errors.New("rfsnotify instance already closed")
	}
	return w.fsnotify.Add(name)
}

// AddRecursive starts watching the named directory and all sub-directories.
func (w *Watcher) AddRecursive(name string, dirFilter DirFilter) error {
	if w.isClosed {
		return errors.New("rfsnotify instance already closed")
	}
	if err := w.watchRecursive(name, false, dirFilter); err != nil {
		return err
	}
	return nil
}

// Remove stops watching the the named file or directory (non-recursively).
func (w *Watcher) Remove(name string) error {
	return w.fsnotify.Remove(name)
}

// RemoveRecursive stops watching the named directory and all sub-directories.
func (w *Watcher) RemoveRecursive(name string) error {
	if err := w.watchRecursive(name, true, nil); err != nil {
		return err
	}
	return nil
}

// Close removes all watches and closes the events channel.
func (w *Watcher) Close() error {
	if w.isClosed {
		return nil
	}
	close(w.done)
	w.isClosed = true
	return nil
}

func (w *Watcher) start() {
	for {
		select {

		case e := <-w.fsnotify.Events:
			s, err := os.Stat(e.Name)
			if err == nil && s != nil && s.IsDir() {
				if e.Op&fsnotify.Create != 0 {
					w.watchRecursive(e.Name, false, nil)
				}
			}
			// Can't stat a deleted directory, so just pretend that it's always a directory and
			// try to remove from the watch list...  we really have no clue if it's a directory or not...
			if e.Op&fsnotify.Remove != 0 {
				w.fsnotify.Remove(e.Name)
			}
			w.Events <- e

		case e := <-w.fsnotify.Errors:
			w.Errors <- e

		case <-w.done:
			w.fsnotify.Close()
			close(w.Events)
			close(w.Errors)
			return
		}
	}
}

// watchRecursive adds all directories under the given one to the watch list.
// this is probably a very racey process. What if a file is added to a folder before we get the watch added?
func (w *Watcher) watchRecursive(path string, unWatch bool, dirFilter DirFilter) error {
	err := filepath.Walk(path, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if unWatch {
				if err = w.fsnotify.Remove(walkPath); err != nil {
					return err
				}
			} else {
				if dirFilter != nil {
					if !dirFilter(walkPath, fi) {
						return filepath.SkipDir
					}
				}
				if err = w.fsnotify.Add(walkPath); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}
