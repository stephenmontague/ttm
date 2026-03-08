package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stephenmontague/ttm-tracker/server/internal/models"
)

// ToolExecutor dispatches tool calls to their implementations.
type ToolExecutor struct {
	lushaAPIKey   string
	httpClient    *http.Client
	workflowState *models.WorkflowState
}

// NewToolExecutor creates a tool executor with the given dependencies.
func NewToolExecutor(lushaAPIKey string, state *models.WorkflowState) *ToolExecutor {
	return &ToolExecutor{
		lushaAPIKey:   lushaAPIKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		workflowState: state,
	}
}

// ToolDefinitions returns the tool definitions to pass to Claude.
// Lusha tools are only included when an API key is configured.
func ToolDefinitions(lushaEnabled bool) []Tool {
	tools := []Tool{
		{
			Name:        "get_workflow_state",
			Description: "Returns the current outreach workflow state including company info, contacts, outreach history, and previous agent suggestions. Use this to understand what has been tried so far.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "draft_message",
			Description: "Structures a draft outreach message. Call this when you have enough context to write a message. The draft will be saved for the BDR to review and send.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"to":{"type":"string","description":"Name of the person to send the message to"},"channel":{"type":"string","description":"Recommended channel: email, linkedin, slack, or phone"},"subject":{"type":"string","description":"Message subject line (for email) or opening hook"},"body":{"type":"string","description":"The full message body"}},"required":["to","channel","body"]}`),
		},
	}

	if lushaEnabled {
		tools = append(tools, Tool{
			Name:        "lusha_contact_search",
			Description: "Search for contact information using Lusha's B2B database. Find verified email addresses, phone numbers, and professional profiles. You can search by email, LinkedIn URL, or name + company combination.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"email":{"type":"string","description":"Email address to look up"},"linkedin_url":{"type":"string","description":"LinkedIn profile URL"},"first_name":{"type":"string","description":"First name (use with last_name and company)"},"last_name":{"type":"string","description":"Last name (use with first_name and company)"},"company":{"type":"string","description":"Company name or domain (use with first_name and last_name)"}}}`),
		})
		tools = append(tools, Tool{
			Name:        "lusha_company_search",
			Description: "Get comprehensive company data from Lusha including firmographics, employee count, industry, and technology stack. Search by domain or company name.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"domain":{"type":"string","description":"Company domain (e.g., whoop.com)"},"company_name":{"type":"string","description":"Company name (if domain unknown)"}}}`),
		})
	}

	return tools
}

// Execute runs the named tool and returns the result as a string.
func (e *ToolExecutor) Execute(ctx context.Context, name string, input json.RawMessage) (string, error) {
	switch name {
	case "get_workflow_state":
		return e.getWorkflowState()
	case "draft_message":
		return e.draftMessage(input)
	case "lusha_contact_search":
		return e.lushaContactSearch(ctx, input)
	case "lusha_company_search":
		return e.lushaCompanySearch(ctx, input)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (e *ToolExecutor) getWorkflowState() (string, error) {
	s := e.workflowState
	var sb strings.Builder

	daysSince := int(time.Since(s.StartedAt).Hours() / 24)

	fmt.Fprintf(&sb, "Company: %s\n", s.CompanyName)
	fmt.Fprintf(&sb, "Status: %s\n", s.Status)
	fmt.Fprintf(&sb, "Days elapsed: %d\n", daysSince)
	fmt.Fprintf(&sb, "Started: %s\n\n", s.StartedAt.Format("2006-01-02"))

	// Contacts
	activeCount := 0
	for _, c := range s.Contacts {
		if c.Active {
			activeCount++
		}
	}
	fmt.Fprintf(&sb, "Contacts (%d active, %d total):\n", activeCount, len(s.Contacts))
	for _, c := range s.Contacts {
		status := "Active"
		if !c.Active {
			status = "Inactive"
		}
		// Count outreach to this contact
		attempts := 0
		for _, a := range s.OutreachAttempts {
			if a.Contact == c.Name {
				attempts++
			}
		}
		fmt.Fprintf(&sb, "  - %s (%s) [%s] — %d outreach attempts", c.Name, c.Role, status, attempts)
		if c.LinkedIn != "" {
			fmt.Fprintf(&sb, " — LinkedIn: %s", c.LinkedIn)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Outreach history
	fmt.Fprintf(&sb, "Outreach History (%d total):\n", len(s.OutreachAttempts))
	for i, a := range s.OutreachAttempts {
		fmt.Fprintf(&sb, "  %d. %s | %s | to %s | \"%s\"\n",
			i+1, a.Timestamp.Format("2006-01-02"), a.Channel, a.Contact, a.Notes)
	}
	sb.WriteString("\n")

	// Previous agent suggestions
	if len(s.AgentSuggestions) > 0 {
		fmt.Fprintf(&sb, "Previous Agent Suggestions (%d):\n", len(s.AgentSuggestions))
		for _, sg := range s.AgentSuggestions {
			fmt.Fprintf(&sb, "  - %s (%s): %s\n", sg.TaskType, sg.Timestamp.Format("2006-01-02"), sg.Response)
		}
	} else {
		sb.WriteString("No previous agent suggestions.\n")
	}

	return sb.String(), nil
}

func (e *ToolExecutor) draftMessage(input json.RawMessage) (string, error) {
	var draft struct {
		To      string `json:"to"`
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.Unmarshal(input, &draft); err != nil {
		return "", fmt.Errorf("parse draft_message input: %w", err)
	}
	return fmt.Sprintf("Draft saved.\nTo: %s\nChannel: %s\nSubject: %s\nBody:\n%s",
		draft.To, draft.Channel, draft.Subject, draft.Body), nil
}

// Lusha API types

type lushaContactRequest struct {
	Email      string `json:"email,omitempty"`
	LinkedInURL string `json:"linkedin_url,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Company    string `json:"company,omitempty"`
}

func (e *ToolExecutor) lushaContactSearch(ctx context.Context, input json.RawMessage) (string, error) {
	if e.lushaAPIKey == "" {
		return "", fmt.Errorf("Lusha API key not configured")
	}

	var req lushaContactRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return "", fmt.Errorf("parse lusha_contact_search input: %w", err)
	}

	// Build query params for the Lusha Person API
	params := url.Values{}
	if req.Email != "" {
		params.Set("email", req.Email)
	}
	if req.LinkedInURL != "" {
		params.Set("linkedInUrl", req.LinkedInURL)
	}
	if req.FirstName != "" {
		params.Set("firstName", req.FirstName)
	}
	if req.LastName != "" {
		params.Set("lastName", req.LastName)
	}
	if req.Company != "" {
		params.Set("company", req.Company)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.lusha.com/person?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("create lusha request: %w", err)
	}
	httpReq.Header.Set("api_key", e.lushaAPIKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("lusha API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read lusha response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Lusha API error (status %d): %s", resp.StatusCode, truncate(string(body), 300)), nil
	}

	return string(body), nil
}

func (e *ToolExecutor) lushaCompanySearch(ctx context.Context, input json.RawMessage) (string, error) {
	if e.lushaAPIKey == "" {
		return "", fmt.Errorf("Lusha API key not configured")
	}

	var req struct {
		Domain      string `json:"domain"`
		CompanyName string `json:"company_name"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return "", fmt.Errorf("parse lusha_company_search input: %w", err)
	}

	params := url.Values{}
	if req.Domain != "" {
		params.Set("domain", req.Domain)
	}
	if req.CompanyName != "" {
		params.Set("companyName", req.CompanyName)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.lusha.com/company?"+params.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("create lusha request: %w", err)
	}
	httpReq.Header.Set("api_key", e.lushaAPIKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("lusha API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read lusha response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Lusha API error (status %d): %s", resp.StatusCode, truncate(string(body), 300)), nil
	}

	return string(body), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
