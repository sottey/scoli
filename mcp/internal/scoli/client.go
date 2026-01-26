package scoli

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

const defaultTimeout = 15 * time.Second

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) GetTree(ctx context.Context, path string) (*TreeNode, error) {
	query := url.Values{}
	if path != "" {
		query.Set("path", path)
	}
	var out TreeNode
	if err := c.doJSON(ctx, http.MethodGet, "/tree", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSheetsTree(ctx context.Context) (*TreeNode, error) {
	var out TreeNode
	if err := c.doJSON(ctx, http.MethodGet, "/sheets/tree", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ReadNote(ctx context.Context, path string) (*Note, error) {
	query := url.Values{}
	query.Set("path", path)
	var out Note
	if err := c.doJSON(ctx, http.MethodGet, "/notes", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ReadSheet(ctx context.Context, path string) (*Sheet, error) {
	query := url.Values{}
	query.Set("path", path)
	var out Sheet
	if err := c.doJSON(ctx, http.MethodGet, "/sheets", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateNote(ctx context.Context, req CreateNoteRequest) (*CreateNoteResponse, error) {
	var out CreateNoteResponse
	if err := c.doJSON(ctx, http.MethodPost, "/notes", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateSheet(ctx context.Context, req CreateSheetRequest) (*CreateSheetResponse, error) {
	var out CreateSheetResponse
	if err := c.doJSON(ctx, http.MethodPost, "/sheets", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateNote(ctx context.Context, req UpdateNoteRequest) (*UpdateNoteResponse, error) {
	var out UpdateNoteResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/notes", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSheet(ctx context.Context, req UpdateSheetRequest) (*UpdateSheetResponse, error) {
	var out UpdateSheetResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/sheets", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RenameNote(ctx context.Context, req RenameNoteRequest) (*RenameNoteResponse, error) {
	var out RenameNoteResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/notes/rename", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RenameSheet(ctx context.Context, req RenameSheetRequest) (*RenameSheetResponse, error) {
	var out RenameSheetResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/sheets/rename", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteNote(ctx context.Context, path string) (*DeleteResponse, error) {
	query := url.Values{}
	query.Set("path", path)
	var out DeleteResponse
	if err := c.doJSON(ctx, http.MethodDelete, "/notes", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSheet(ctx context.Context, path string) (*DeleteResponse, error) {
	query := url.Values{}
	query.Set("path", path)
	var out DeleteResponse
	if err := c.doJSON(ctx, http.MethodDelete, "/sheets", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ImportSheet(ctx context.Context, req ImportSheetRequest) (*CreateSheetResponse, error) {
	var out CreateSheetResponse
	if err := c.doJSON(ctx, http.MethodPost, "/sheets/import", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ExportSheet(ctx context.Context, path string) (string, error) {
	query := url.Values{}
	query.Set("path", path)
	return c.doText(ctx, http.MethodGet, "/sheets/export", query, nil)
}

func (c *Client) CreateFolder(ctx context.Context, req FolderRequest) (*FolderResponse, error) {
	var out FolderResponse
	if err := c.doJSON(ctx, http.MethodPost, "/folders", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RenameFolder(ctx context.Context, req RenameFolderRequest) (*FolderResponse, error) {
	var out FolderResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/folders", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteFolder(ctx context.Context, path string) (*DeleteResponse, error) {
	query := url.Values{}
	query.Set("path", path)
	var out DeleteResponse
	if err := c.doJSON(ctx, http.MethodDelete, "/folders", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Search(ctx context.Context, queryText string) ([]SearchResult, error) {
	query := url.Values{}
	query.Set("query", queryText)
	var out []SearchResult
	if err := c.doJSON(ctx, http.MethodGet, "/search", query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListTags(ctx context.Context) ([]TagGroup, error) {
	var out []TagGroup
	if err := c.doJSON(ctx, http.MethodGet, "/tags", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListTasks(ctx context.Context) (*TaskList, error) {
	var out TaskList
	if err := c.doJSON(ctx, http.MethodGet, "/tasks", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListTasksForNote(ctx context.Context, path string) (*TaskList, error) {
	query := url.Values{}
	query.Set("path", path)
	var out TaskList
	if err := c.doJSON(ctx, http.MethodGet, "/tasks/for-note", query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ToggleTask(ctx context.Context, req ToggleTaskRequest) (*StatusResponse, error) {
	var out StatusResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/tasks/toggle", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ArchiveTasks(ctx context.Context) (*ArchiveTasksResponse, error) {
	var out ArchiveTasksResponse
	if err := c.doJSON(ctx, http.MethodPatch, "/tasks/archive", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	var out SettingsResponse
	if err := c.doJSON(ctx, http.MethodGet, "/settings", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSettings(ctx context.Context, req map[string]any) (*Settings, error) {
	var out Settings
	if err := c.doJSON(ctx, http.MethodPatch, "/settings", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) doText(ctx context.Context, method, endpoint string, query url.Values, body any) (string, error) {
	fullURL := c.BaseURL + endpoint
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", decodeAPIError(resp)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(content), nil
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, query url.Values, body any, out any) error {
	fullURL := c.BaseURL + endpoint
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp)
	}

	if out == nil {
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func decodeAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	var payload ErrorResponse
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != "" {
		return fmt.Errorf("http %d: %s", resp.StatusCode, payload.Error)
	}

	return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}
