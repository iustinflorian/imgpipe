package main

import (
	"encoding/json"
	"imgpipe/pkg/db"
	"imgpipe/pkg/models"
	"imgpipe/pkg/queue"
	"net/http"
)

type CreateJobRequest struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type JobHandler struct {
	repo *db.JobRepository
	mq   *queue.RabbitMQ
}

func NewJobHandler(repo *db.JobRepository, mq *queue.RabbitMQ) *JobHandler {
	return &JobHandler{
		repo: repo,
		mq:   mq,
	}
}

func (h *JobHandler) CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalid", http.StatusBadRequest)
		return
	}

	if req.URL == "" || req.Width <= 0 || req.Height <= 0 {
		http.Error(w, "invalid parameters: URL, width, height are mandatory", http.StatusBadRequest)
		return
	}

	job := models.Job{
		URL:    req.URL,
		Width:  req.Width,
		Height: req.Height,
	}

	if err := h.repo.CreateJob(r.Context(), &job); err != nil {
		http.Error(w, "error saving to database", http.StatusInternalServerError)
		return
	}

	if err := h.mq.PublishJob(r.Context(), &job); err != nil {
		http.Error(w, "error sending to message queue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "'id' parameter is mandatory", http.StatusBadRequest)
		return
	}

	job, err := h.repo.GetJobByID(r.Context(), id)
	if err != nil {
		http.Error(w, "no job found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}
