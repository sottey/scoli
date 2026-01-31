# Scoli API Deep Dive

## Overview

Scoli exposes a JSON HTTP API for notes, folders, tags, tasks, search, and
settings. Notes and sheets live on disk under the configured notes directory;
the API operates on paths relative to that root.

## Base URL

```
http://localhost:8080/api/v1
```

## UI deep links (non-API)

The web UI supports a few query params for PWA shortcuts (these are not API
endpoints):

- `/?shortcut=inbox` opens `Inbox.md`
- `/?shortcut=daily` opens today's daily note
- `/?shortcut=tasks` opens the Tasks view

## Authentication

None. The API is unauthenticated and designed for local use.

## Content types

- Requests with a body must use `Content-Type: application/json`.
- Responses use `application/json` unless otherwise noted.
- `GET /files` returns raw file content with the appropriate content type.

## Path rules

- Paths are **relative to the notes directory**.
- Absolute paths and `..` traversal are rejected.
- `.md` is appended automatically when creating notes unless already present.
- `.jsh` is appended automatically when creating sheets unless already present.

## Errors

Errors return JSON in this shape:

```json
{ "error": "message" }
```

Typical status codes:

- `400` invalid input (missing fields, invalid path, invalid payload).
- `404` not found.
- `409` conflict (already exists).
- `500` server error.

## Conventions

- JSON requests reject unknown fields and trailing data.
- Tags, projects, and mentions are normalized to lowercase.
- Task due dates are parsed into `YYYY-MM-DD` when possible and warnings are
  returned when parsing fails.
- Metadata inside fenced or indented code blocks, or inline code spans, is ignored.

## Data models

### Note

```json
{
  "path": "Daily/2026-01-06.md",
  "content": "# Title",
  "modified": "2026-01-06T10:00:00Z"
}
```

### Sheet

```json
{
  "path": "Sheets/Budget.jsh",
  "data": [["Item", "Cost"], ["Rent", "1200"]],
  "modified": "2026-01-06T10:00:00Z"
}
```

Note: the UI may append empty rows to fill the visible sheet height, so saved
data can include trailing empty rows.

### Task

```json
{
  "id": "Daily/2026-01-06.md:12",
  "path": "Daily/2026-01-06.md",
  "lineNumber": 12,
  "lineHash": "abc123...",
  "text": "Water cacti",
  "completed": false,
  "project": "home",
  "tags": ["home"],
  "mentions": [],
  "dueDate": "2026-02-01",
  "dueDateISO": "2026-02-01",
  "priority": 2
}
```

### JournalEntry

```json
{
  "id": "1737540623000000000",
  "content": "Met with the team.",
  "createdAt": "2026-01-22T15:03:43-08:00",
  "updatedAt": "2026-01-22T15:03:43-08:00"
}
```

### Settings

```json
{
  "version": 7,
  "darkMode": false,
  "defaultView": "split",
  "sidebarWidth": 300,
  "defaultFolder": "",
  "showTemplates": true,
  "showAiNode": true,
  "notesSortBy": "name",
  "notesSortOrder": "asc",
  "externalCommandsPath": "",
  "rootIcons": {
    "notes": "/icons/notes.png",
    "daily": "/icons/daily.png"
  }
}
```

### EmailSettings

```json
{
  "version": 1,
  "enabled": false,
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "you@gmail.com",
    "password": "app-password",
    "from": "you@gmail.com",
    "to": "you@gmail.com",
    "useTLS": true
  },
  "digest": {
    "enabled": true,
    "time": "08:00"
  },
  "due": {
    "enabled": true,
    "time": "07:30",
    "windowDays": 0,
    "includeOverdue": true
  },
  "templates": {
    "digest": "email/digest.template",
    "due": "email/due.template"
  }
}
```

Email settings are stored in `Notes/email-settings.json`, with templates under
`Notes/email/`.

## AI settings

AI settings are stored in `Notes/.ai/ai-settings.json`. This file is created
on server startup if missing so you can keep it out of version control. It
contains the OpenAI API key and AI tuning values.

## Templates

If a folder contains `default.template`, new notes created in that folder start
with that content.

### Placeholders

- `{{date:YYYY-MM-DD}}`
- `{{time:HH:mm}}`
- `{{datetime:YYYY-MM-DD HH:mm}}`
- `{{day:ddd}}` or `{{day:dddd}}`
- `{{year:YYYY}}`
- `{{month:YYYY-MM}}`
- `{{title}}`
- `{{path}}`
- `{{folder}}`

### Conditionals

Inline:

```
{{if:day=wed|sat}} - [ ] Water cacti #home
```

Block:

```
{{if:day=weekday}}
- [ ] Daily sites #work
{{endif}}
```

Operators:

- `&` for AND between fields
- `|` for OR within a field

Valid fields:

- `day` (`mon`..`sun`, `weekday`, `weekend`)
- `dom` (`1`..`31`)
- `date` (`YYYY-MM-DD`)
- `month` (`1`..`12` or `jan`..`dec`)

Invalid conditionals are skipped and surfaced via `notice` in the response.

## Tasks parsing

Tasks are parsed from note lines that start with:

- `- [ ] ` (open)
- `- [x] ` (complete)

Markers:

- `#tag`
- `@mention`
- `+project`
- `>due` (parsed into `YYYY-MM-DD` when possible)
- `^priority` (1-5)

Example:

```
- [ ] Call Mom +Home #family @alice >2025-01-31 ^2
```

## Endpoints

### Health

`GET /health`

Use this to check if the server is up.

Response:

```json
{ "status": "ok" }
```

### Tree

`GET /tree`

Returns the notes tree. Accepts an optional `path` query to scope to a folder.
This call ensures the Inbox note and today's Daily note exist (they may be
created if missing).

Example:

```
GET /tree?path=Projects
```

Response:

```json
{
  "name": "Notes",
  "path": "",
  "type": "folder",
  "children": [
    { "name": "Daily", "path": "Daily", "type": "folder", "children": [] },
    { "name": "Note.md", "path": "Note.md", "type": "file" }
  ]
}
```

Node types:

- `folder`
- `file` (markdown notes and templates if enabled)
- `asset` (images)
- `pdf`
- `csv`

### Notes

#### Read

`GET /notes?path=<file>`

Example:

```
GET /notes?path=Daily/2026-01-06.md
```

Response:

```json
{
  "path": "Daily/2026-01-06.md",
  "content": "# Title",
  "modified": "2026-01-06T10:00:00Z"
}
```

#### Create

`POST /notes`

Body:

```json
{
  "path": "Daily/2026-01-06",
  "content": "# Title"
}
```

Response:

```json
{ "path": "Daily/2026-01-06.md" }
```

If a template warning is detected:

```json
{ "path": "Daily/2026-01-06.md", "notice": "Invalid template condition \"...\": ..." }
```

#### Update

`PATCH /notes`

Body:

```json
{
  "path": "Daily/2026-01-06.md",
  "content": "# Updated"
}
```

Response:

```json
{ "path": "Daily/2026-01-06.md" }
```

#### Rename

`PATCH /notes/rename`

Daily notes (`Daily/YYYY-MM-DD.md`) cannot be renamed or moved.

Body:

```json
{
  "path": "Projects/Spec.md",
  "newPath": "Projects/Spec-v2"
}
```

Response:

```json
{ "path": "Projects/Spec.md", "newPath": "Projects/Spec-v2.md" }
```

#### Delete

`DELETE /notes?path=<file>`

Response:

```json
{ "status": "deleted" }
```

### Folders

#### Create

`POST /folders`

Body:

```json
{ "path": "Projects/NewFolder" }
```

Response:

```json
{ "path": "Projects/NewFolder" }
```

#### Rename

`PATCH /folders`

Body:

```json
{ "path": "Projects/NewFolder", "newPath": "Projects/RenamedFolder" }
```

Response:

```json
{ "path": "Projects/NewFolder", "newPath": "Projects/RenamedFolder" }
```

#### Delete

`DELETE /folders?path=<folder>`

Response:

```json
{ "status": "deleted" }
```

### Files

`GET /files?path=<file>`

Returns the raw file (used for images, PDFs, CSVs, etc.).

### Icons

#### Upload root icon

`POST /icons/root?root=<key>`

Uploads a root node icon via `multipart/form-data` with a file field named
`icon`. Allowed extensions: `.png`, `.svg`, `.ico`. Max size: 1MB.

Valid root keys: `notes`, `daily`, `tasks`, `tags`, `journal`, `inbox`, `ai`, `sheets`.

Response:

```json
{ "root": "notes", "path": "/icons/notes-20260123-010101-acde.png" }
```

#### Reset root icon

`DELETE /icons/root?root=<key>`

Clears the icon for the given root key.

Response:

```json
{ "root": "notes" }
```

### Search

`GET /search?query=<text>`

Searches note filenames and contents.

Response:

```json
[
  { "path": "Daily/2026-01-06.md", "name": "2026-01-06.md", "type": "note" }
]
```

### Tags

`GET /tags`

Returns tag groups with matching notes.

Response:

```json
[
  { "tag": "work", "notes": [{ "path": "Daily/2026-01-06.md", "name": "2026-01-06.md" }] }
]
```

### Mentions

`GET /mentions`

Returns mention groups with matching notes.

Response:

```json
[
  { "mention": "alice", "notes": [{ "path": "Daily/2026-01-06.md", "name": "2026-01-06.md" }] }
]
```

### Sheets

#### Tree

`GET /sheets/tree`

Returns the sheets tree under `Sheets/`.

Sheet paths are relative to `Sheets/` (the root is omitted in API calls).

Response:

```json
{
  "name": "Sheets",
  "path": "",
  "type": "folder",
  "children": [
    { "name": "Budget.jsh", "path": "Budget.jsh", "type": "sheet" }
  ]
}
```

#### Read

`GET /sheets?path=<file>`

Example:

```
GET /sheets?path=Budget.jsh
```

Response:

```json
{
  "path": "Budget.jsh",
  "data": [["Item", "Cost"], ["Rent", "1200"]],
  "modified": "2026-01-06T10:00:00Z"
}
```

#### Create

`POST /sheets`

Body:

```json
{
  "path": "Budget",
  "data": [["Item", "Cost"], ["Rent", "1200"]]
}
```

Response:

```json
{ "path": "Budget.jsh" }
```

#### Update

`PATCH /sheets`

Body:

```json
{
  "path": "Budget.jsh",
  "data": [["Item", "Cost"], ["Rent", "1250"]]
}
```

Response:

```json
{ "path": "Budget.jsh" }
```

#### Rename

`PATCH /sheets/rename`

Body:

```json
{
  "path": "Budget.jsh",
  "newPath": "Budget-2026"
}
```

Response:

```json
{ "path": "Budget.jsh", "newPath": "Budget-2026.jsh" }
```

#### Delete

`DELETE /sheets?path=<file>`

Response:

```json
{ "status": "deleted" }
```

#### Import CSV

`POST /sheets/import`

Body:

```json
{
  "path": "Budget",
  "csv": "Item,Cost\nRent,1200\n"
}
```

Response:

```json
{ "path": "Budget.jsh" }
```

#### Export CSV

`GET /sheets/export?path=<file>`

Returns `text/csv` with a `Content-Disposition` attachment filename.

### Tasks

#### List all tasks

`GET /tasks`

Response:

```json
{
  "tasks": [
    {
      "id": "Daily/2026-01-06.md:12",
      "path": "Daily/2026-01-06.md",
      "lineNumber": 12,
      "lineHash": "abc123...",
      "text": "Water cacti",
      "completed": false,
      "project": "home",
      "tags": ["home"],
      "mentions": [],
      "dueDate": "2026-02-01",
      "dueDateISO": "2026-02-01",
      "priority": 2
    }
  ],
  "notice": "Found 1 task(s) with unrecognized due dates. Examples: ..."
}
```

#### List tasks for a single note

`GET /tasks/for-note?path=<file>`

Same response shape as `/tasks`, scoped to a single note.

#### Toggle a task

`PATCH /tasks/toggle`

Body:

```json
{
  "path": "Daily/2026-01-06.md",
  "lineNumber": 12,
  "lineHash": "abc123...",
  "completed": true
}
```

Response:

```json
{ "status": "updated" }
```

#### Archive completed tasks

`PATCH /tasks/archive`

Archives completed tasks by prefixing them with `~ `.

Response:

```json
{ "archived": 12, "files": 3 }
```

### Task Filters

Task filters are stored in `Notes/task-sets.json`.

#### Get filters

`GET /tasks/filters`

Response:

```json
{
  "filters": {
    "version": 1,
    "filters": [
      {
        "id": "work-due-soon",
        "name": "Work Due Soon",
        "tags": ["work"],
        "mentions": [],
        "projects": ["work"],
        "due": { "from": "today", "to": "+7d" },
        "priority": { "min": 2, "max": 5 },
        "completed": false,
        "text": "",
        "pathPrefix": "Projects/"
      }
    ]
  }
}
```

#### Update filters

`PUT /tasks/filters`

Body:

```json
{
  "version": 1,
  "filters": []
}
```

### Journal

#### List entries

`GET /journal`

Response:

```json
{
  "entries": [
    {
      "id": "1737540623000000000",
      "content": "Met with the team.",
      "createdAt": "2026-01-22T15:03:43-08:00",
      "updatedAt": "2026-01-22T15:03:43-08:00"
    }
  ]
}
```

#### Create entry

`POST /journal`

Body:

```json
{ "content": "Met with the team." }
```

Response:

```json
{
  "id": "1737540623000000000",
  "content": "Met with the team.",
  "createdAt": "2026-01-22T15:03:43-08:00",
  "updatedAt": "2026-01-22T15:03:43-08:00"
}
```

#### Update entry

`PATCH /journal`

Body:

```json
{ "id": "1737540623000000000", "content": "Updated notes." }
```

Response:

```json
{ "status": "updated" }
```

#### Delete entry

`DELETE /journal?id=<id>`

Response:

```json
{ "status": "deleted" }
```

#### Archive entry

`POST /journal/archive`

Body:

```json
{ "id": "1737540623000000000" }
```

Response:

```json
{ "status": "archived", "archivePath": "journal/journal-archive-YYYY-MM-DD.json" }
```

#### Archive all entries

`POST /journal/archive-all`

Response:

```json
{ "status": "archived", "archivePath": "journal/journal-archive-YYYY-MM-DD.json" }
```

#### List archives

`GET /journal/archives`

Response:

```json
{
  "archives": [
    {
      "date": "2026-01-22",
      "path": "journal/journal-archive-2026-01-22.json",
      "count": 3
    }
  ],
  "activeCount": 2,
  "archiveCount": 3,
  "totalCount": 5
}
```

#### Get archive entries

`GET /journal/archive?date=YYYY-MM-DD`

Response:

```json
{
  "entries": [
    {
      "id": "1737540623000000000",
      "content": "Met with the team.",
      "createdAt": "2026-01-22T15:03:43-08:00",
      "updatedAt": "2026-01-22T15:03:43-08:00"
    }
  ]
}
```

### Settings

#### Read

`GET /settings`

Response:

```json
{
  "settings": {
    "version": 7,
    "darkMode": false,
    "defaultView": "split",
    "sidebarWidth": 300,
    "defaultFolder": "",
    "showTemplates": true,
    "showAiNode": true,
    "notesSortBy": "name",
    "notesSortOrder": "asc",
    "externalCommandsPath": "",
    "rootIcons": {
      "notes": "/icons/notes.png",
      "daily": "/icons/daily.png"
    }
  },
  "notice": "Created settings.json"
}
```

#### Update

`PATCH /settings`

Body (any subset of fields):

```json
{
  "darkMode": true,
  "defaultView": "edit",
  "sidebarWidth": 320,
  "defaultFolder": "Projects",
  "showTemplates": true,
  "showAiNode": true,
  "notesSortBy": "updated",
  "notesSortOrder": "desc",
  "externalCommandsPath": "commands.json"
}
```

### AI

#### Read AI settings

`GET /ai/settings`

Response:

```json
{
  "settings": {
    "version": 1,
    "apiKey": "",
    "chatModel": "gpt-4o-mini",
    "embedModel": "text-embedding-3-small",
    "topK": 6,
    "maxContextChunks": 6,
    "temperature": 0.2,
    "maxOutputTokens": 500,
    "chunkCharLimit": 1600,
    "sectionCharLimit": 5000
  },
  "configured": false
}
```

#### List chats

`GET /ai/chats`

#### Create chat

`POST /ai/chats`

Body:

```json
{ "title": "New Chat" }
```

#### Get chat

`GET /ai/chats/{id}`

#### Send message

`POST /ai/chats/{id}/messages`

Body:

```json
{ "content": "What was that quote from Grape of Wrath I wrote about?" }
```

### Email settings

#### Read

`GET /email/settings`

Response:

```json
{
  "settings": {
    "version": 1,
    "enabled": false,
    "smtp": {
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "you@gmail.com",
      "password": "app-password",
      "from": "you@gmail.com",
      "to": "you@gmail.com",
      "useTLS": true
    },
    "digest": { "enabled": true, "time": "08:00" },
    "due": { "enabled": true, "time": "07:30", "windowDays": 0, "includeOverdue": true },
    "templates": { "digest": "email/digest.template", "due": "email/due.template" }
  },
  "notice": "Created email-settings.json"
}
```

#### Update

`PATCH /email/settings`

Body (any subset of fields):

```json
{
  "enabled": true,
  "smtp": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "you@gmail.com",
    "password": "app-password",
    "from": "you@gmail.com",
    "to": "you@gmail.com",
    "useTLS": true
  },
  "digest": { "enabled": true, "time": "08:00" },
  "due": { "enabled": true, "time": "07:30", "windowDays": 0, "includeOverdue": true }
}
```

#### Send test email

`POST /email/test`

Response:

```json
{ "status": "sent" }
```

Response:

```json
{
  "version": 6,
  "darkMode": true,
  "defaultView": "edit",
  "sidebarWidth": 320,
  "defaultFolder": "Projects",
  "showTemplates": true,
  "notesSortBy": "updated",
  "notesSortOrder": "desc",
  "rootIcons": {
    "notes": "/icons/notes.png"
  }
}
```

## Curl examples

List tasks:

```bash
curl http://localhost:8080/api/v1/tasks
```

Create a note:

```bash
curl -X POST http://localhost:8080/api/v1/notes \
  -H 'Content-Type: application/json' \
  -d '{"path":"Daily/2026-01-06","content":"# Title"}'
```

Toggle a task:

```bash
curl -X PATCH http://localhost:8080/api/v1/tasks/toggle \
  -H 'Content-Type: application/json' \
  -d '{"path":"Daily/2026-01-06.md","lineNumber":12,"lineHash":"abc123...","completed":true}'
```
