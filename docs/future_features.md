## TODO
- [ ] Template manager UI: templates are powerful but hidden; add a simple “Templates” view or a quick editor that lists all default template files.
- [ ] CSV import/export polish for Sheets: round out with column type inference and per-sheet settings (freeze header, default column widths).
- [ ] Search improvements: show snippet matches with highlighted terms and a “search in folder” option.

## MAYBE SOMEDAY
- [ ] Flash cards Node - Ability to set texst for a "front" and a "back" and the ability to run through a deck
- [ ] AI (MCP)
- [ ] Meeting Transcription Node - Record mic and computer audio, storing the audio. Also, during recording, there is a text pane which the user can type into. This typing is stored with time stamps so if the recording is replayed, the notes show real time when notes were written (in relation to the audio)
- [ ] Canvas - Node that allows the creation of open canvases for drawing.


## COMPLETED
- [X] Universal prompt. Hitting a key compo brings up a text box the user can use to find a command.action they want to do.
- [X] Tag pills are too tight (spce around the word) and too loose (space between the pills)
- [x] Ability to change icons on any root nodes via context menu upload
- [x] Autosave causing date error dialog. If save happens while adding a ">2026-01" date, it should not show a parse error
- [x] Save on change (debounced) instead of timed autosave
- [x] have task data tiles be all on one line [title] [tags] [filepath] in all task views
- [X] Move toggle for sidebar as well as settings icons to bottom right of sidebar
- [X] All root nodes start collapsed on load
- [X] Rolling Journal Node - "Twitter" like feed with an "Add" button that allows adding a new entry which consists of just text, with a created and updated dat area
- [X] Global Scratch Pad - Access from anywhere. Simple MC edit pane
- [X] Drag-and-drop in the sidebar tree
- [X] Collapsible sidebar
- [X] For display of project names in tree, force Proper Case
- [X] On page refresh, return to previously selected node/task/note. If no longer there, return to its parent
- [X] On note save, refresh left sidebar, including tasks
- [X] Hide No Project and Completed nodes if empty
- [X] Template use (context menu items?)
- [X] Default startup folder (Daily?)
- [X] Notes root should always be named Notes, not the root folder name
- [X] If there is a "Daily" folder and a note for this day in the form YYYY-MM-DD does not exist in it, create a new note with today's date in the format YYYY-MM-DD and, if in the daily folder there is a default.template file, use its contents for the new note
- [X] task filters by tag/mention/project across parsed note tasks
    Example: * This is a task in the +Home project and is due >2025-12-27 and is priority ^3 #home #test
- [X] Mentions view parity with Tags: already aggregates mentions; add quick actions (open all, filter by project/date) to match Tasks/Tabs usability.
- [X] Better daily/journal cross-linking: link journal entries to a daily note (auto-create if missing) for more cohesion.
- [X] Task filtering + saved task views: users already rely on tags/mentions/projects; add saved filters (e.g., “Work + Due this week”) and pin them in the Tasks tree.