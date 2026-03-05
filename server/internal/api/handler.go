package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	"github.com/stephenmontague/ttm-tracker/server/internal/config"
	"github.com/stephenmontague/ttm-tracker/server/internal/models"
	"github.com/stephenmontague/ttm-tracker/server/internal/repository"
)

// publicCompany is the public-safe representation of a company.
// Real names and internal IDs are replaced with the opaque public UUID.
type publicCompany struct {
	ID              string     `json:"ID"`
	CompanyName     string     `json:"CompanyName"`
	Slug            string     `json:"Slug"`
	StartedAt       time.Time  `json:"StartedAt"`
	Status          string     `json:"Status"`
	ElapsedDays     int        `json:"ElapsedDays"`
	OutreachCount   int        `json:"OutreachCount"`
	ContactCount    int        `json:"ContactCount"`
	RestartCount    int        `json:"RestartCount"`
	MeetingBookedAt *time.Time `json:"MeetingBookedAt"`
	LastSnapshotAt  *time.Time `json:"LastSnapshotAt"`
	UpdatedAt       time.Time  `json:"UpdatedAt"`
}

func obfuscateCompany(c repository.CompanyRow) publicCompany {
	return publicCompany{
		ID:              c.PublicID,
		CompanyName:     "Company " + c.PublicID[:6],
		Slug:            c.PublicID,
		StartedAt:       c.StartedAt,
		Status:          c.Status,
		ElapsedDays:     c.ElapsedDays,
		OutreachCount:   c.OutreachCount,
		ContactCount:    c.ContactCount,
		RestartCount:    c.RestartCount,
		MeetingBookedAt: c.MeetingBookedAt,
		LastSnapshotAt:  c.LastSnapshotAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

type Handler struct {
	temporalClient client.Client
	companyRepo    *repository.CompanyRepository
	authRepo       *repository.AuthRepository
}

func NewHandler(temporalClient client.Client, companyRepo *repository.CompanyRepository, authRepo *repository.AuthRepository) *Handler {
	return &Handler{
		temporalClient: temporalClient,
		companyRepo:    companyRepo,
		authRepo:       authRepo,
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

// reconcileStatuses verifies active workflows against Temporal and updates
// the DB cache if a workflow is no longer running.
func (h *Handler) reconcileStatuses(ctx context.Context, companies []repository.CompanyRow) {
	for i := range companies {
		c := &companies[i]
		if c.Status != "active" {
			continue
		}
		desc, err := h.temporalClient.DescribeWorkflowExecution(ctx, c.ID, "")
		if err != nil {
			log.Printf("Failed to describe workflow %s, marking terminated: %v", c.ID, err)
			c.Status = "terminated"
			_ = h.companyRepo.UpdateStatus(ctx, c.ID, "terminated")
			continue
		}
		status := desc.WorkflowExecutionInfo.Status
		if status != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
			newStatus := strings.ToLower(strings.TrimPrefix(status.String(), "WORKFLOW_EXECUTION_STATUS_"))
			log.Printf("Workflow %s is %s, updating cache", c.ID, newStatus)
			c.Status = newStatus
			_ = h.companyRepo.UpdateStatus(ctx, c.ID, newStatus)
		}
	}
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

	h.reconcileStatuses(r.Context(), companies)

	public := make([]publicCompany, len(companies))
	for i, c := range companies {
		public[i] = obfuscateCompany(c)
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"companies": public,
		"total":     len(public),
	})
}

// GetCompany returns a single company's public state.
func (h *Handler) GetCompany(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "slug")

	company, err := h.companyRepo.GetCompanyByPublicID(r.Context(), publicID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Company not found")
		return
	}

	respondJSON(w, http.StatusOK, obfuscateCompany(*company))
}

// GetCompanyFeed returns the sanitized activity feed for a company.
func (h *Handler) GetCompanyFeed(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "slug")

	company, err := h.companyRepo.GetCompanyByPublicID(r.Context(), publicID)
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

// ListAdminCompanies returns all companies with real (unobfuscated) data.
func (h *Handler) ListAdminCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.companyRepo.ListCompanies(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list companies: "+err.Error())
		return
	}

	if companies == nil {
		companies = []repository.CompanyRow{}
	}

	h.reconcileStatuses(r.Context(), companies)

	respondJSON(w, http.StatusOK, map[string]any{
		"companies": companies,
		"total":     len(companies),
	})
}

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

	// Seed the database so the company appears on the dashboard immediately,
	// rather than waiting for the first async workflow snapshot.
	_ = h.companyRepo.UpsertWorkflow(r.Context(), &repository.CompanyRow{
		ID:          workflowID,
		CompanyName: req.CompanyName,
		Slug:        slug,
		StartedAt:   time.Now(),
		Status:      "active",
	})

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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.temporalClient.QueryWorkflow(ctx, workflowID, "", config.QueryGetState)
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
	Channel     string `json:"channel"`
	Notes       string `json:"notes"`
	ContactName string `json:"contactName"`
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
		Channel:     req.Channel,
		Notes:       req.Notes,
		ContactName: req.ContactName,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalLogOutreach, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "log_outreach"})
}

type AddContactRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	LinkedIn string `json:"linkedin"`
}

// SignalAddContact sends an AddContact signal to the workflow.
func (h *Handler) SignalAddContact(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	var req AddContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signal := models.AddContactSignal{
		Name:     req.Name,
		Role:     req.Role,
		LinkedIn: req.LinkedIn,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalAddContact, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "add_contact"})
}

type RemoveContactRequest struct {
	Name string `json:"name"`
}

// SignalRemoveContact sends a RemoveContact signal to the workflow.
func (h *Handler) SignalRemoveContact(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	workflowID := "outreach-" + slug

	var req RemoveContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signal := models.RemoveContactSignal{
		Name: req.Name,
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", config.SignalRemoveContact, signal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to send signal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"sent": true, "signal": "remove_contact"})
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
