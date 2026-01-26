package scoli

type ErrorResponse struct {
	Error string `json:"error"`
}

type TreeNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Children []TreeNode `json:"children,omitempty"`
}

type Note struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Modified string `json:"modified"`
}

type Sheet struct {
	Path     string     `json:"path"`
	Data     [][]string `json:"data"`
	Modified string     `json:"modified"`
}

type CreateNoteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type CreateNoteResponse struct {
	Path   string `json:"path"`
	Notice string `json:"notice,omitempty"`
}

type UpdateNoteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type UpdateNoteResponse struct {
	Path string `json:"path"`
}

type RenameNoteRequest struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type RenameNoteResponse struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type CreateSheetRequest struct {
	Path string     `json:"path"`
	Data [][]string `json:"data"`
}

type CreateSheetResponse struct {
	Path string `json:"path"`
}

type UpdateSheetRequest struct {
	Path string     `json:"path"`
	Data [][]string `json:"data"`
}

type UpdateSheetResponse struct {
	Path string `json:"path"`
}

type RenameSheetRequest struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type RenameSheetResponse struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type ImportSheetRequest struct {
	Path string `json:"path"`
	CSV  string `json:"csv"`
}

type DeleteResponse struct {
	Status string `json:"status"`
}

type FolderRequest struct {
	Path string `json:"path"`
}

type RenameFolderRequest struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath"`
}

type FolderResponse struct {
	Path    string `json:"path"`
	NewPath string `json:"newPath,omitempty"`
}

type SearchResult struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TagNote struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type TagGroup struct {
	Tag   string    `json:"tag"`
	Notes []TagNote `json:"notes"`
}

type Task struct {
	ID         string   `json:"id"`
	Path       string   `json:"path"`
	LineNumber int      `json:"lineNumber"`
	LineHash   string   `json:"lineHash"`
	Text       string   `json:"text"`
	Completed  bool     `json:"completed"`
	Project    string   `json:"project"`
	Tags       []string `json:"tags"`
	Mentions   []string `json:"mentions"`
	DueDate    string   `json:"dueDate"`
	DueDateISO string   `json:"dueDateISO"`
	Priority   int      `json:"priority"`
}

type TaskList struct {
	Tasks  []Task `json:"tasks"`
	Notice string `json:"notice,omitempty"`
}

type ToggleTaskRequest struct {
	Path       string `json:"path"`
	LineNumber int    `json:"lineNumber"`
	LineHash   string `json:"lineHash"`
	Completed  bool   `json:"completed"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ArchiveTasksResponse struct {
	Archived int `json:"archived"`
	Files    int `json:"files"`
}

type SettingsResponse struct {
	Settings Settings `json:"settings"`
	Notice   string   `json:"notice,omitempty"`
}

type Settings struct {
	Version       int    `json:"version"`
	DarkMode      bool   `json:"darkMode"`
	DefaultView   string `json:"defaultView"`
	SidebarWidth  int    `json:"sidebarWidth"`
	DefaultFolder string `json:"defaultFolder"`
	ShowTemplates bool   `json:"showTemplates"`
}
