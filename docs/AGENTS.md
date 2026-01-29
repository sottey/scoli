# Scoli (project notes)

Scoli is a Go-based web application with a separate HTTP API. It presents a
folder tree of Markdown notes from the local `Notes/` directory, with a main
pane that supports edit, preview, or split view.

This file documents the current project shape and interaction model.

## Product behavior (high level)
- Left sidebar shows the `Notes/` directory tree (folders + `.md` files) plus
  Tags and Mentions roots.
- Left sidebar shows a `Tasks` root with project groups, a "Today" group,
  a "Someday" group, a "Task Filters" group, a "No Project" group, a "Completed" group, and an "All" group (always last).
- Left sidebar shows a `Sheets` root with spreadsheet files (`.jsh`) stored under `Sheets/`.
- Left sidebar includes a `Journal` root below Tasks and above Tags that opens a
  rolling journal feed, with archive files shown as children under the Journal root.
- Left sidebar includes a `Mentions` root below Tags that aggregates `@mention`
  tokens across notes.
- Sidebar folder, task, and tag rows display immediate-child counts in the label.
- Notes are sorted globally with folders above files, using the configured sort order.
- "Created" sorting uses best-effort file creation time and may fall back to modified time.
- Clicking Notes/Tags roots or any folder shows a summary panel in the main pane.
- Clicking Tasks root or any project group shows a task list in the main pane.
- Clicking "Task Filters" shows a filter selector and task list for the selected filter.
- Clicking Journal shows a rolling feed with a compose box, inline edit, delete,
  archive actions, and an Archive All button. Archive files open read-only feeds.
- Journal entry editors use Enter to save and Opt-Enter for a newline.
- Daily notes show a read-only journal panel below the tag bar with "Edit in Journal" and a "New Journal Entry" shortcut.
- Clicking "Today" shows incomplete tasks that are due today or earlier (or have no due date),
  excluding tasks tagged `#someday`.
- Clicking "Someday" shows incomplete tasks tagged `#someday`.
- Clicking "All" shows all open tasks sorted by due date (oldest first), then priority.
- Main pane shows a Markdown editor, a rendered preview, or a split view with a
  draggable splitter.
- PWA UX includes install button (when supported), update toast/banner, and an
  offline fallback page with retry.
- Settings button in the sidebar footer opens a settings form.
- The sidebar width is adjustable with a draggable splitter.
- A view selector in the top-right provides edit, preview, and split modes.
- The main header includes a date pill that opens (and creates) today's daily note.
- A calendar icon in the main header opens a date picker to jump to a specific daily note.
- Context menus:
  - Right-click on a folder: New Folder, New Note, Edit Template, Rename, Delete, Expand/Collapse.
  - Right-click on a note: New Note, Rename, Delete.
  - Right-click empty area in sidebar: New Folder, New Note.
  - Notes and folders include a Sort Notes option to update the global sort.
  - Edit Template creates `default.template` if missing and opens it for editing.
- The sidebar footer includes a toggle-all icon that collapses or expands root
  nodes in the tree.
- Tags are aggregated into a "Tags" root, collapsed by default, and refreshed
  when the tree reloads.
- Mentions are aggregated into a "Mentions" root, collapsed by default, and
  refreshed when the tree reloads.
- Tags under the Tags root render as wrapping pills; expanded tags list notes below.
- Mentions under the Mentions root render as wrapping pills; expanded mentions
  list notes below.
- The preview pane shows a sticky tag bar with clickable tag pills for the
  current note.
- Tag pills in the preview bar wrap to multiple lines as needed.
- Opening a note expands the tree to its node and selects it.
- Root nodes default to collapsed on initial load.
- Notes and folders can be dragged onto folders or the Notes root to move them
  (dropping on a note targets its parent folder; tasks/tags are not draggable).
- Editor and preview panes scroll together (proportional sync).
- Folder and tag rows show centered chevrons indicating expanded/collapsed
  state.
- Sheets use a Jspreadsheet CE grid with column resizing, right-click context
  menus (including import/export), and automatic column/row sizing to fill the
  visible pane.
- The header includes a Scratch Pad button that opens a modal note stored as
  `scratch.md` at the notes root (hidden from the tree), with a Move To Inbox
  action that appends its contents to `Inbox.md` and clears the scratch pad.
- Root nodes include a "Change Icon" context menu action that stores the icon
  under `internal/ui/web/icons` and records the path in settings.
- Notes save automatically shortly after changes (debounced save-on-change).
- Sheets are stored as JSON (`.jsh`) under `Sheets/` and support CSV import/export.
- The command palette (Ctrl+Alt+K) supports built-in commands and optional
  external commands loaded from `externalCommandsPath` in settings. Use `>` in
  the palette to search notes by filename/content.
- Mobile behavior (max-width 720px): sidebar starts hidden as an overlay drawer,
  swipe down on the header opens it, swipe up on the sidebar closes it, and note
  view defaults to preview on open regardless of saved default view.
- PWA shortcuts (when installed): `/?shortcut=inbox`, `/?shortcut=daily`,
  `/?shortcut=tasks` deep-link to Inbox, Daily, and Tasks.

## Architecture
- **CLI**: A Cobra-based entrypoint used to run the server and any future admin
  tasks (example: `scoli serve --notes-dir ./Notes --port 8080`).
- **API**: A JSON HTTP API that handles notes, folders, and on-demand task parsing.
- **MCP server**: A Go MCP server (stdio or HTTP transport) that fronts the JSON API for tool and resource access.
- **Web app**: A UI that consumes the API and renders the editor + preview.
- **Storage**: Notes live on disk in the `Notes/` tree as Markdown files.

## API responsibilities
- List a recursive folder tree and notes beneath a provided folder.
- Read/write note contents.
- Create/rename/delete folders.
- Create/rename/delete notes.
- Provide a refresh endpoint or tree reload operation.
- List tags extracted from note contents.
- List mentions extracted from note contents.
- Parse tasks from note contents and toggle completion by editing note lines.
- Read/write settings stored in `Notes/settings.json`.
- Read/write sheets stored under `Sheets/`.

## UI responsibilities
- Render the folder tree and handle context menus.
- Render the tags root and tag groups.
- Render the mentions root and mention groups.
- Render the tasks root and project groups, and show task lists in the main pane.
- Render the sheets root and spreadsheet grid with CSV import/export (via
  context menu).
- Render the journal root and feed view.
- Render the Markdown editor, preview, and split view with draggable splitter.
- Task list rows include a checkbox toggle and open the source note on click.
- Render a settings form with sections for Display (dark mode, default view, show templates),
  Notes (sort by, sort order), and Folders (default folder).
- The date pill always opens notes in `Daily/`.
- Ensure tag labels remain legible in dark mode.
- Render a tag bar in the preview pane.
- Call API endpoints for all mutations and refresh operations.
- Provide filename/content search with a dropdown of matches.
- Provide keyboard shortcuts using Ctrl+Alt:
  - Ctrl+Alt+S: Save
  - Ctrl+Alt+E: Edit view
  - Ctrl+Alt+V: Preview view
  - Ctrl+Alt+B: Split view
  - Ctrl+Alt+D: Open today's daily note
  - Ctrl+Alt+C: Open date picker
  - Ctrl+Alt+P: Open scratch pad
  - Ctrl+Alt+J: Open journal
  - Ctrl+Alt+K: Open command palette
  - Ctrl+Alt+I: Open inbox

## API shape
- **Base path**: `/api/v1` (no auth for now).
- **Identifiers**: use a path string relative to `Notes/` for all note/folder ops.
- **Tree**:
  - `GET /api/v1/tree?path=<folder>` returns a recursive tree under `path`
    (metadata only).
  - If `path` is omitted, the full tree under `Notes/` is returned.
- **Notes**:
  - `GET /api/v1/notes?path=<file>` returns note content and metadata.
  - `POST /api/v1/notes` creates a note at `path` with content.
  - `PATCH /api/v1/notes` updates a note at `path` with content.
  - `PATCH /api/v1/notes/rename` renames a note from `path` to `newPath`.
  - `DELETE /api/v1/notes?path=<file>` removes the note.
- **Folders**:
  - `POST /api/v1/folders` creates a folder at `path`.
  - `PATCH /api/v1/folders` renames a folder from `path` to `newPath`.
  - `DELETE /api/v1/folders?path=<folder>` removes the folder.
- **Files**:
  - `GET /api/v1/files?path=<file>` serves a raw file (used for images).
- **Search**:
  - `GET /api/v1/search?query=<text>` searches note filenames + contents.
- **Tags**:
  - `GET /api/v1/tags` returns tags with the notes that contain them.
- **Mentions**:
  - `GET /api/v1/mentions` returns mentions with the notes that contain them.
- **Sheets**:
  - `GET /api/v1/sheets/tree` returns the sheets tree under `Sheets/`.
  - `GET /api/v1/sheets?path=<file>` returns sheet JSON.
  - `POST /api/v1/sheets` creates a sheet.
  - `PATCH /api/v1/sheets` updates a sheet.
  - `PATCH /api/v1/sheets/rename` renames or moves a sheet.
  - `DELETE /api/v1/sheets?path=<file>` deletes a sheet.
  - `POST /api/v1/sheets/import` imports CSV into a new sheet.
  - `GET /api/v1/sheets/export?path=<file>` exports a sheet as CSV.
- **Tasks**:
  - `GET /api/v1/tasks` returns tasks parsed from notes.
  - `PATCH /api/v1/tasks/toggle` toggles completion for a task line.
  - `PATCH /api/v1/tasks/archive` archives completed tasks by prefixing `~ `.
- **Task Filters**:
  - `GET /api/v1/tasks/filters` returns task filters from `task-sets.json`.
  - `PUT /api/v1/tasks/filters` updates task filters.
- **Settings**:
  - `GET /api/v1/settings` returns settings.
  - `PATCH /api/v1/settings` updates settings.
- **Health**:
  - `GET /api/v1/health` returns status.

## Data rules (confirmed)
- Tree responses include metadata only, never file contents.
- If a note path is missing the `.md` extension, it is appended on create.
- Only `.md` files are considered notes; other files are ignored.
- Sheet files use the `.jsh` extension and live under `Sheets/`.
- Files starting with `._` are ignored.
- Tags match `#` followed by letters, preceded by whitespace or start of line.
- Mentions match `@` followed by letters, preceded by whitespace or start of line,
  case-insensitive and normalized to lowercase.
- Tasks are parsed from note lines starting with `- [ ] ` or `- [x] ` (leading whitespace ok).
- Task markers: `#tag`, `@mention`, `+project`, `>due`, `^priority` (1-5); markers are case-insensitive and normalized to lowercase.
- Metadata inside fenced or indented code blocks, or inline code spans, is ignored.
- Only one project is used per task (first match wins).
- Settings live in `Notes/settings.json`.
- Task filters live in `Notes/task-sets.json`.
- If a folder contains `default.template`, new notes created in that folder use
  the template contents.

## Templates
- `default.template` in a folder provides the initial content for new notes
  created in that folder.
- Example: `Notes/Daily/default.template` applies to new notes under
  `Daily/`.
- Templates can include placeholders that are replaced when the note is created:
  `{{date:YYYY-MM-DD}}`, `{{time:HH:mm}}`, `{{datetime:YYYY-MM-DD HH:mm}}`,
  `{{day:ddd}}` or `{{day:dddd}}`, `{{year:YYYY}}`, `{{month:YYYY-MM}}`,
  `{{title}}`, `{{path}}`, `{{folder}}`. All date/time values use server-local
  time.
- Date/time placeholders must include the token name (for example,
  `{{date:YYYY-MM-DD}}`, not `{{YYYY-MM-DD}}`).
- Templates can conditionally include lines based on a target date derived from
  the note filename (if it matches `YYYY-MM-DD`; otherwise the current date):
  - Inline: `{{if:day=wed|sat}} - [ ] Water cacti`
  - Block:
    ```
    {{if:day=sat}}
    - [ ] Saturday review
    {{endif}}
    ```
  - Fields: `day` (`mon`..`sun`, `weekday`, `weekend`), `dom` (`1`..`31`),
    `date` (`YYYY-MM-DD`), `month` (`1`..`12` or `jan`..`dec`), all
    case-insensitive.
  - Use `|` for OR within a field (example: `day=wed|sat`) and `&` for AND across
    fields (example: `day=sat&dom=1`).
  - Invalid conditions omit the line/block and return a warning when the
    template is used.

## Open questions to confirm
- None currently.
