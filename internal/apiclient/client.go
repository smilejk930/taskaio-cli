package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL string, token string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, query url.Values, body interface{}, out interface{}) error {
	fullURL := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	if len(query) > 0 {
		fullURL = fmt.Sprintf("%s?%s", fullURL, query.Encode())
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return &TransportError{Err: fmt.Errorf("network request failed: %w", err)}
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TransportError{Err: fmt.Errorf("failed to read response body: %w", err)}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}

		var rawErr RawErrorResponse
		if err := json.Unmarshal(respBytes, &rawErr); err == nil && rawErr.Error.Message != "" {
			apiErr.Code = rawErr.Error.Code
			apiErr.Message = rawErr.Error.Message
			apiErr.Details = rawErr.Error.Details
		} else if len(respBytes) > 0 {
			apiErr.Message = string(respBytes)
		}
		return apiErr
	}

	if out != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return &TransportError{Err: fmt.Errorf("failed to parse response JSON: %w", err)}
		}
	}

	return nil
}

// Auth
func (c *Client) GetAuthMe(ctx context.Context) (*AuthMeResponse, error) {
	// /api/v1/auth/me intentionally preserves its legacy top-level response.
	var resp AuthMeResponse
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/auth/me", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Projects
func (c *Client) ListProjects(ctx context.Context, params url.Values) (*ListResponse[Project], error) {
	var resp ListResponse[Project]
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/projects", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var resp SingleResponse[Project]
	endpoint := fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(projectID))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) CreateProject(ctx context.Context, input CreateProjectInput) (*Project, error) {
	var resp SingleResponse[Project]
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/projects", nil, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) UpdateProject(ctx context.Context, projectID string, input UpdateProjectInput) (*Project, error) {
	var resp SingleResponse[Project]
	endpoint := fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(projectID))
	if err := c.doRequest(ctx, http.MethodPatch, endpoint, nil, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	endpoint := fmt.Sprintf("/api/v1/projects/%s", url.PathEscape(projectID))
	return c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

func (c *Client) ListProjectMembers(ctx context.Context, projectID string, params url.Values) (*ListResponse[ProjectMember], error) {
	var resp ListResponse[ProjectMember]
	endpoint := fmt.Sprintf("/api/v1/projects/%s/members", url.PathEscape(projectID))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Tasks
func (c *Client) ListTasks(ctx context.Context, projectID string, params url.Values) (*ListResponse[Task], error) {
	var resp ListResponse[Task]
	endpoint := fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(projectID))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetTaskSummary(ctx context.Context, projectID string, params url.Values) (*TaskSummary, error) {
	var resp SingleResponse[TaskSummary]
	endpoint := fmt.Sprintf("/api/v1/projects/%s/tasks/summary", url.PathEscape(projectID))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var resp SingleResponse[Task]
	endpoint := fmt.Sprintf("/api/v1/tasks/%s", url.PathEscape(taskID))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) CreateTask(ctx context.Context, projectID string, input CreateTaskInput) (*Task, error) {
	var resp SingleResponse[Task]
	endpoint := fmt.Sprintf("/api/v1/projects/%s/tasks", url.PathEscape(projectID))
	if err := c.doRequest(ctx, http.MethodPost, endpoint, nil, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) UpdateTask(ctx context.Context, taskID string, input UpdateTaskInput) (*Task, error) {
	var resp SingleResponse[Task]
	endpoint := fmt.Sprintf("/api/v1/tasks/%s", url.PathEscape(taskID))
	if err := c.doRequest(ctx, http.MethodPatch, endpoint, nil, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	endpoint := fmt.Sprintf("/api/v1/tasks/%s", url.PathEscape(taskID))
	return c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

// Schedules
func (c *Client) ListSchedules(ctx context.Context, params url.Values) (*ListResponse[Schedule], error) {
	var resp ListResponse[Schedule]
	if err := c.doRequest(ctx, http.MethodGet, "/api/v1/schedules", params, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetSchedule(ctx context.Context, scheduleID string) (*Schedule, error) {
	var resp SingleResponse[Schedule]
	endpoint := fmt.Sprintf("/api/v1/schedules/%s", url.PathEscape(scheduleID))
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) CreateSchedule(ctx context.Context, input CreateScheduleInput) (*Schedule, error) {
	var resp SingleResponse[Schedule]
	if err := c.doRequest(ctx, http.MethodPost, "/api/v1/schedules", nil, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) UpdateSchedule(ctx context.Context, scheduleID string, input UpdateScheduleInput) (*Schedule, error) {
	var resp SingleResponse[Schedule]
	endpoint := fmt.Sprintf("/api/v1/schedules/%s", url.PathEscape(scheduleID))
	if err := c.doRequest(ctx, http.MethodPatch, endpoint, nil, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteSchedule(ctx context.Context, scheduleID string) error {
	endpoint := fmt.Sprintf("/api/v1/schedules/%s", url.PathEscape(scheduleID))
	return c.doRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}
