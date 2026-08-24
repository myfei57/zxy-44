package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store is the file-backed persistence layer of the wind control system.
// Every write goes through a temp file plus fsync so a crash cannot leave a
// half-written JSON document behind.
type Store struct {
	root string
	mu   sync.Mutex
}

// New opens or creates a store rooted at root. The root directory is created
// on demand so the control console can start against an empty data directory.
func New(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create store root %q: %w", root, err)
	}
	return &Store{root: root}, nil
}

// Root returns the absolute base directory of the store.
func (s *Store) Root() string {
	return s.root
}

// Path resolves a logical file name to a filesystem path inside the store.
func (s *Store) Path(name string) string {
	return filepath.Join(s.root, filepath.FromSlash(name))
}

// Exists reports whether a logical file is present.
func (s *Store) Exists(name string) bool {
	_, err := os.Stat(s.Path(name))
	return err == nil
}

// WriteJSON atomically persists value under name.
func (s *Store) WriteJSON(name string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", name, err)
	}
	return s.writeBytes(name, data)
}

// ReadJSON loads a JSON document previously written with WriteJSON.
func (s *Store) ReadJSON(name string, value any) error {
	data, err := os.ReadFile(s.Path(name))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("unmarshal %s: %w", name, err)
	}
	return nil
}

// AppendLine appends one line to an append-only log and fsyncs it.
func (s *Store) AppendLine(name string, line []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.Path(name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create log dir for %s: %w", name, err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log %s: %w", name, err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("append log %s: %w", name, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync log %s: %w", name, err)
	}
	return nil
}

// ReadLines returns every line of an append-only log, skipping empty lines.
func (s *Store) ReadLines(name string) ([][]byte, error) {
	data, err := os.ReadFile(s.Path(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			line := data[start:i]
			start = i + 1
			if len(line) > 0 {
				out = append(out, line)
			}
		}
	}
	return out, nil
}

// Remove deletes a logical file, ignoring missing files.
func (s *Store) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.Path(name))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Store) writeBytes(name string, data []byte) error {
	path := s.Path(name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir for %s: %w", name, err)
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create temp for %s: %w", name, err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("write temp for %s: %w", name, err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("sync temp for %s: %w", name, err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close temp for %s: %w", name, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename temp for %s: %w", name, err)
	}
	return nil
}
