package jobrepo

import (
	"podcast-summarizer/src/core/app/jobs"
	"podcast-summarizer/src/core/domain"
)

type JobMemoryRepo struct {
	store map[string]domain.ProcessingJob
}

func NewJobMemoryRepo() *JobMemoryRepo {
	return &JobMemoryRepo{store: make(map[string]domain.ProcessingJob)}
}

func (r *JobMemoryRepo) Create(job domain.ProcessingJob) (string, error) {
	id := "job-memory"
	job.ID = id
	r.store[id] = job
	return id, nil
}

func (r *JobMemoryRepo) UpdateStatus(id, status string, durationMs *int, errorMessage *string) error {
	job := r.store[id]
	job.Status = status
	job.DurationMs = durationMs
	job.Error = errorMessage
	r.store[id] = job
	return nil
}

func (r *JobMemoryRepo) Get(id string) (domain.ProcessingJob, error) {
	return r.store[id], nil
}

var _ jobs.JobRepository = (*JobMemoryRepo)(nil)
