package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"thermalik-bridge/internal/domain"
)

var ErrJobNotFound = errors.New("job not found")

type JobStore interface {
	Save(context.Context, *domain.Job) error
	Get(context.Context, string) (*domain.Job, error)
}

type FileJobStore struct {
	mu   sync.RWMutex
	path string
	jobs map[string]*domain.Job
}

func NewFileJobStore(path string) (*FileJobStore, error) {
	store := &FileJobStore{
		path: path,
		jobs: make(map[string]*domain.Job),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *FileJobStore) Save(_ context.Context, job *domain.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = cloneJob(job)

	return s.persistLocked()
}

func (s *FileJobStore) Get(_ context.Context, id string) (*domain.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}

	return cloneJob(job), nil
}

func (s *FileJobStore) load() error {
	content, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	var jobs map[string]*domain.Job
	if err := json.Unmarshal(content, &jobs); err != nil {
		return err
	}

	if jobs != nil {
		s.jobs = jobs
	}

	return nil
}

func (s *FileJobStore) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(s.jobs, "", "  ")
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, content, 0o644); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}

func cloneJob(job *domain.Job) *domain.Job {
	cloned := *job

	if job.Bytes != nil {
		cloned.Bytes = append([]byte(nil), job.Bytes...)
	}

	if job.Render != nil {
		cloned.Render = make(map[string]any, len(job.Render))
		for key, value := range job.Render {
			cloned.Render[key] = value
		}
	}

	return &cloned
}
