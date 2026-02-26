package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.temporal.io/sdk/client"

	"github.com/stephenmontague/ttm-tracker/internal/config"
	"github.com/stephenmontague/ttm-tracker/internal/models"
	"github.com/stephenmontague/ttm-tracker/internal/repository"
)

type Handler struct {
	temporalClient client.Client
	companyRepo    *repository.CompanyRepository
}

func NewHandler(temporalClient client.Client, companyRepo *repository.CompanyRepository) *Handler {
	return &Handler{
		temporalClient: temporalClient,
		companyRepo:    companyRepo,
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// --- Public Endpoints ---

// ListCompanies returns all tracked companies from the PostgreSQL cache.
func (h *Handler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.companyRepo.ListCompanies(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list companies: "+err.Error())
		return
	}

	if companies == nil {
		companies = []repository.CompanyRow{}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"companies": companies,
		"total":     len(companies),
	})
}

// GetCompany returns a single company's public state.
func (h *Handler) GetCompany(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	company, err := h.companyRepo.GetCompanyBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, "Company not found")
		return
	}

	respondJSON(w, http.StatusOK, company)
}

// GetCompanyFeed returns the sanitized activity feed for a company.
func (h *Handler) GetCompanyFeed(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	company, err := h.companyRepo.GetCompanyBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, "Company not found")
		return
	}

	feed, err := h.companyRepo.GetActivityFeed(r.Context(), company.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get activity feed: "+err.Error())
		return
	}

	if feed == nil {
		feed = []repository.ActivityFeedRow{}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"feed":  feed,
		"total": len(feed),
	})
}

// --- Admin Endpoints ---

type CreateCompanyRequest struct {
	CompanyName string `json:"companyName"`
}

// CreateCompany starts a new outreach workflow for a company.
func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var req CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CompanyName == "" {
		respondError(w, http.StatusBadRequest, "companyName is required")
		return
	}

	slug := slugify(req.CompanyName)
	workflowID := "outreach-" + slug

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: config.GetTaskQueue(),
	}

	params := models.WorkflowParams{
		CompanyName: req.CompanyName,
		Slug:        slug,
	}

	we, err := h.temporalClient.ExecuteWorkflow(r.Context(), options, config.WorkflowName, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to start workflow: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]any{
		"workflowId": we.GetID(),
		"runId":      we.GetRunID(),
		"slug":       slug,
		"status":     "active",
	})
}

// GetAdminCompany returns full workflow state via Temporal query (including PII).
func (h *Handler) GetAdminCompany(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	resp, err := h.temporalClient.QueryWorkflow(r.Context(), workflowID, "", config.QueryGetState)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query workflow: "+err.Error())
		return
	}

	var state models.WorkflowState
	if err := resp.Get(&state); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decode workflow state: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, state)
}

type LogOutreachRequest struct {
	Channel string `json:"channel"`
	Notes   string `json:"notes"`
}

// SignalOutreach sends a LogOutreach signal to the workflow.
func (h *Handler) SignalOutreach(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	var req LogOutreachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signal := models.LogOutreachSignal{
		Channel: req.Channel,
		Notes:   req.Notes,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalLogOutreach, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "log_outreach"})
}

type UpdateContactRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	LinkedIn string `json:"linkedin"`
}

// SignalUpdateContact sends an UpdateContact signal to the workflow.
func (h *Handler) SignalUpdateContact(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	var req UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signal := models.UpdateContactSignal{
		Name:     req.Name,
		Role:     req.Role,
		LinkedIn: req.LinkedIn,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalUpdateContact, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "update_contact"})
}

type RequestAgentRequest struct {
	TaskType string `json:"taskType"`
	Context  string `json:"context"`
}

// SignalRequestAgent sends a RequestAgentHelp signal to the workflow.
func (h *Handler) SignalRequestAgent(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	var req RequestAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signal := models.RequestAgentHelpSignal{
		TaskType: req.TaskType,
		Context:  req.Context,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalRequestAgent, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "request_agent_help"})
}

type MeetingBookedRequest struct {
	Date  time.Time `json:"date"`
	Notes string    `json:"notes"`
}

// SignalMeetingBooked sends a MeetingBooked signal to the workflow.
func (h *Handler) SignalMeetingBooked(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	var req MeetingBookedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signal := models.MeetingBookedSignal{
		Date:  req.Date,
		Notes: req.Notes,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalMeetingBooked, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "meeting_booked"})
}

// SignalPause sends a Pause signal to the workflow.
func (h *Handler) SignalPause(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalPause, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "pause"})
}

// SignalResume sends a Resume signal to the workflow.
func (h *Handler) SignalResume(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalResume, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "resume"})
}

// slugify converts a company name to a URL-friendly slug.
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove anything that isn't alphanumeric or a hyphen
	var result strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}
	return result.String()
}
