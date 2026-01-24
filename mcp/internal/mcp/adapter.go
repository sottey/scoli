package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/sottey/scoli/mcp/internal/scoli"
)

type ToolSpec struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type ResourceSpec struct {
	URI         string
	Description string
	MimeType    string
}

type Adapter struct {
	client *scoli.Client
}

func NewAdapter(client *scoli.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) Tools() []ToolSpec {
	return []ToolSpec{
		{
			Name:        "tree.get",
			Description: "Get the notes tree or a subtree for a folder path.",
			InputSchema: schemaObject(map[string]any{
				"path": schemaString("Optional folder path to scope the tree."),
			}, nil),
		},
		{
			Name:        "note.read",
			Description: "Read a note by path.",
			InputSchema: schemaObject(map[string]any{
				"path": schemaString("Note path, relative to the notes root."),
			}, []string{"path"}),
		},
		{
			Name:        "note.create",
			Description: "Create a new note (adds .md if missing).",
			InputSchema: schemaObject(map[string]any{
				"path":    schemaString("Target note path, relative to the notes root."),
				"content": schemaString("Full note content."),
			}, []string{"path", "content"}),
		},
		{
			Name:        "note.update",
			Description: "Update an existing note's content.",
			InputSchema: schemaObject(map[string]any{
				"path":    schemaString("Note path, relative to the notes root."),
				"content": schemaString("Full note content."),
			}, []string{"path", "content"}),
		},
		{
			Name:        "note.rename",
			Description: "Rename a note.",
			InputSchema: schemaObject(map[string]any{
				"path":    schemaString("Existing note path."),
				"newPath": schemaString("New note path, relative to notes root."),
			}, []string{"path", "newPath"}),
		},
		{
			Name:        "note.delete",
			Description: "Delete a note.",
			InputSchema: schemaObject(map[string]any{
				"path": schemaString("Note path, relative to the notes root."),
			}, []string{"path"}),
		},
		{
			Name:        "folder.create",
			Description: "Create a folder.",
			InputSchema: schemaObject(map[string]any{
				"path": schemaString("Folder path, relative to the notes root."),
			}, []string{"path"}),
		},
		{
			Name:        "folder.rename",
			Description: "Rename a folder.",
			InputSchema: schemaObject(map[string]any{
				"path":    schemaString("Existing folder path."),
				"newPath": schemaString("New folder path, relative to notes root."),
			}, []string{"path", "newPath"}),
		},
		{
			Name:        "folder.delete",
			Description: "Delete a folder.",
			InputSchema: schemaObject(map[string]any{
				"path": schemaString("Folder path, relative to the notes root."),
			}, []string{"path"}),
		},
		{
			Name:        "search",
			Description: "Search notes by query string.",
			InputSchema: schemaObject(map[string]any{
				"query": schemaString("Search query string."),
			}, []string{"query"}),
		},
		{
			Name:        "tags.list",
			Description: "List tags and their notes.",
			InputSchema: schemaObject(map[string]any{}, nil),
		},
		{
			Name:        "tasks.list",
			Description: "List tasks, optionally scoped to a note path.",
			InputSchema: schemaObject(map[string]any{
				"path": schemaString("Optional note path to scope tasks."),
			}, nil),
		},
		{
			Name:        "tasks.toggle",
			Description: "Toggle a task's completion state.",
			InputSchema: schemaObject(map[string]any{
				"path":       schemaString("Note path containing the task."),
				"lineNumber": schemaInteger("Task line number (1-based)."),
				"lineHash":   schemaString("Task line hash from task listing."),
				"completed":  schemaBoolean("New completed value."),
			}, []string{"path", "lineNumber", "lineHash", "completed"}),
		},
		{
			Name:        "tasks.archive",
			Description: "Archive completed tasks by prefixing them with '~ '.",
			InputSchema: schemaObject(map[string]any{}, nil),
		},
		{
			Name:        "settings.get",
			Description: "Read settings.",
			InputSchema: schemaObject(map[string]any{}, nil),
		},
		{
			Name:        "settings.update",
			Description: "Update settings (partial update).",
			InputSchema: schemaObject(map[string]any{
				"darkMode":      schemaBoolean("Enable dark mode."),
				"defaultView":   schemaString("Default view: edit, preview, split."),
				"sidebarWidth":  schemaInteger("Sidebar width in pixels."),
				"defaultFolder": schemaString("Default folder path."),
				"showTemplates": schemaBoolean("Show templates in tree."),
			}, nil),
		},
	}
}

func (a *Adapter) Resources() []ResourceSpec {
	return []ResourceSpec{
		{
			URI:         "scoli://note/{path}",
			Description: "Read a note by path.",
			MimeType:    "text/markdown",
		},
	}
}

func (a *Adapter) CallTool(ctx context.Context, name string, args any) (any, error) {
	switch name {
	case "tree.get":
		var payload struct {
			Path string `json:"path"`
		}
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if payload.Path != "" {
			if err := validatePath(payload.Path); err != nil {
				return nil, err
			}
		}
		return a.client.GetTree(ctx, payload.Path)
	case "note.read":
		payload, err := decodePath(args)
		if err != nil {
			return nil, err
		}
		return a.client.ReadNote(ctx, payload.Path)
	case "note.create":
		var payload scoli.CreateNoteRequest
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		return a.client.CreateNote(ctx, payload)
	case "note.update":
		var payload scoli.UpdateNoteRequest
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		return a.client.UpdateNote(ctx, payload)
	case "note.rename":
		var payload scoli.RenameNoteRequest
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		if err := validatePath(payload.NewPath); err != nil {
			return nil, err
		}
		return a.client.RenameNote(ctx, payload)
	case "note.delete":
		payload, err := decodePath(args)
		if err != nil {
			return nil, err
		}
		return a.client.DeleteNote(ctx, payload.Path)
	case "folder.create":
		var payload scoli.FolderRequest
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		return a.client.CreateFolder(ctx, payload)
	case "folder.rename":
		var payload scoli.RenameFolderRequest
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		if err := validatePath(payload.NewPath); err != nil {
			return nil, err
		}
		return a.client.RenameFolder(ctx, payload)
	case "folder.delete":
		payload, err := decodePath(args)
		if err != nil {
			return nil, err
		}
		return a.client.DeleteFolder(ctx, payload.Path)
	case "search":
		var payload struct {
			Query string `json:"query"`
		}
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if strings.TrimSpace(payload.Query) == "" {
			return nil, fmt.Errorf("query is required")
		}
		return a.client.Search(ctx, payload.Query)
	case "tags.list":
		return a.client.ListTags(ctx)
	case "tasks.list":
		var payload struct {
			Path string `json:"path"`
		}
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if payload.Path == "" {
			return a.client.ListTasks(ctx)
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		return a.client.ListTasksForNote(ctx, payload.Path)
	case "tasks.toggle":
		var payload scoli.ToggleTaskRequest
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if err := validatePath(payload.Path); err != nil {
			return nil, err
		}
		return a.client.ToggleTask(ctx, payload)
	case "tasks.archive":
		return a.client.ArchiveTasks(ctx)
	case "settings.get":
		return a.client.GetSettings(ctx)
	case "settings.update":
		if args == nil {
			return nil, fmt.Errorf("settings update requires at least one field")
		}
		var payload map[string]any
		if err := decodeInput(args, &payload); err != nil {
			return nil, err
		}
		if len(payload) == 0 {
			return nil, fmt.Errorf("settings update requires at least one field")
		}
		return a.client.UpdateSettings(ctx, payload)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *Adapter) ReadResource(ctx context.Context, uri string) (*scoli.Note, error) {
	if !strings.HasPrefix(uri, "scoli://") {
		return nil, fmt.Errorf("unsupported resource uri: %s", uri)
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("parse resource uri: %w", err)
	}
	if parsed.Host == "note" {
		clean := strings.TrimPrefix(parsed.Path, "/")
		if clean == "" {
			return nil, fmt.Errorf("note resource requires a path")
		}
		if err := validatePath(clean); err != nil {
			return nil, err
		}
		return a.client.ReadNote(ctx, clean)
	}

	return nil, fmt.Errorf("unsupported resource uri: %s", uri)
}

func schemaObject(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func schemaString(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
	}
}

func schemaInteger(description string) map[string]any {
	return map[string]any{
		"type":        "integer",
		"description": description,
	}
}

func schemaBoolean(description string) map[string]any {
	return map[string]any{
		"type":        "boolean",
		"description": description,
	}
}

func decodeInput(args any, dest any) error {
	if args == nil {
		return nil
	}

	raw, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("encode input: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	return nil
}

func decodePath(args any) (struct {
	Path string `json:"path"`
}, error) {
	var payload struct {
		Path string `json:"path"`
	}
	if err := decodeInput(args, &payload); err != nil {
		return payload, err
	}
	if payload.Path == "" {
		return payload, fmt.Errorf("path is required")
	}
	if err := validatePath(payload.Path); err != nil {
		return payload, err
	}
	return payload, nil
}

func validatePath(raw string) error {
	if strings.HasPrefix(raw, "/") {
		return fmt.Errorf("path must be relative")
	}
	clean := path.Clean(raw)
	if clean == "." || clean == "" {
		return fmt.Errorf("path must not be empty")
	}
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return fmt.Errorf("path traversal is not allowed")
	}
	return nil
}
