const apiBase = "/api/v1";

const app = document.getElementById("app");
const treeContainer = document.getElementById("tree-container");
const treeToggleBtn = document.getElementById("tree-toggle-btn");
const editor = document.getElementById("editor");
const preview = document.getElementById("preview");
const notePath = document.getElementById("note-path");
const saveBtn = document.getElementById("save-btn");
const moveCompletedBtn = document.getElementById("move-completed-btn");
const datePill = document.getElementById("date-pill");
const calendarBtn = document.getElementById("calendar-btn");
let datePopover = document.getElementById("date-popover");
let datePicker = document.getElementById("date-picker");
const viewButtons = Array.from(document.querySelectorAll(".view-btn"));
const viewSelector = document.querySelector(".view-selector");
const contextMenu = document.getElementById("context-menu");
const sidebar = document.getElementById("sidebar");
const sidebarToggle = document.getElementById("sidebar-toggle");
const sidebarResizer = document.getElementById("sidebar-resizer");
const paneResizer = document.getElementById("pane-resizer");
const editorPane = document.getElementById("editor-pane");
const previewPane = document.getElementById("preview-pane");
const mainContent = document.getElementById("main-content");
const mainHeader = document.querySelector(".main-header");
const searchInput = document.getElementById("search-input");
const searchBtn = document.getElementById("search-btn");
const searchResults = document.getElementById("search-results");
const tagBar = document.getElementById("tag-bar");
const tagAddBtn = document.getElementById("tag-add-btn");
const tagPills = document.getElementById("tag-pills");
const assetPreview = document.getElementById("asset-preview");
const pdfPreview = document.getElementById("pdf-preview");
const csvPreview = document.getElementById("csv-preview");
const summaryPanel = document.getElementById("summary-panel");
const taskList = document.getElementById("task-list");
const journalPanel = document.getElementById("journal-panel");
const journalCompose = document.getElementById("journal-compose");
const journalInput = document.getElementById("journal-input");
const journalSave = document.getElementById("journal-save");
const journalArchiveAll = document.getElementById("journal-archive-all");
const journalFeed = document.getElementById("journal-feed");
const settingsBtn = document.getElementById("settings-btn");
const settingsPanel = document.getElementById("settings-panel");
const brandBtn = document.getElementById("brand-btn");
const settingsDarkMode = document.getElementById("settings-dark-mode");
const settingsDefaultView = document.getElementById("settings-default-view");
const settingsDefaultFolder = document.getElementById("settings-default-folder");
const settingsShowTemplates = document.getElementById("settings-show-templates");
const settingsNotesSortBy = document.getElementById("settings-notes-sort-by");
const settingsNotesSortOrder = document.getElementById("settings-notes-sort-order");
const scratchBtn = document.getElementById("scratch-btn");
const inboxDialog = document.getElementById("inbox-dialog");
const inboxDialogText = document.getElementById("inbox-dialog-text");
const inboxDialogClose = document.getElementById("inbox-dialog-close");
const inboxDialogCancel = document.getElementById("inbox-dialog-cancel");
const inboxDialogSave = document.getElementById("inbox-dialog-save");
const inboxDialogBackdrop = inboxDialog ? inboxDialog.querySelector(".modal-backdrop") : null;
const scratchDialog = document.getElementById("scratch-dialog");
const scratchDialogText = document.getElementById("scratch-dialog-text");
const scratchDialogClose = document.getElementById("scratch-dialog-close");
const scratchDialogCancel = document.getElementById("scratch-dialog-cancel");
const scratchDialogMove = document.getElementById("scratch-dialog-move");
const scratchDialogSave = document.getElementById("scratch-dialog-save");
const scratchDialogBackdrop = scratchDialog ? scratchDialog.querySelector(".modal-backdrop") : null;
const whatsNewModal = document.getElementById("whats-new-modal");
const whatsNewList = document.getElementById("whats-new-list");
const whatsNewClose = document.getElementById("whats-new-close");
const whatsNewConfirm = document.getElementById("whats-new-confirm");
const whatsNewBackdrop = whatsNewModal ? whatsNewModal.querySelector(".modal-backdrop") : null;
const offlineBanner = document.getElementById("offline-banner");
const offlineRetry = document.getElementById("offline-retry");
const statusBanners = document.getElementById("status-banners");
const updateBanner = document.getElementById("update-banner");
const updateToast = document.getElementById("update-toast");
const updateReload = document.getElementById("update-reload");
const appVersionLabel = document.getElementById("app-version");
const appToast = document.getElementById("app-toast");
const commandPalette = document.getElementById("command-palette");
const commandInput = document.getElementById("command-input");
const commandResults = document.getElementById("command-results");
const commandClose = document.getElementById("command-close");
const commandBackdrop = commandPalette ? commandPalette.querySelector(".modal-backdrop") : null;

let currentNotePath = "";
let currentActivePath = "";
let currentTree = null;
let currentTags = [];
let currentTasks = [];
let journalEntries = [];
let journalSummary = { archives: [], totalCount: 0 };
let journalViewMode = "main";
let currentMode = "note";
let lastNoteView = "preview";
let currentSettings = { darkMode: false };
let settingsLoaded = false;
let noteSaveTimer = null;
let noteSaveInFlight = false;
let noteSaveQueued = false;
let noteSaveSnapshot = null;
let isDirty = false;
let syncingScroll = false;
let activeScrollSource = null;
let clearScrollSourceTimer = null;
let isSidebarOpen = false;
let touchStartY = null;
let dragState = null;
let dragExpandTimer = null;
let dragOverRow = null;
const dragExpandDelayMs = 500;
const inboxNotePath = "Inbox.md";
const scratchNotePath = "scratch.md";
const journalRootPath = "__journal__";
const journalFolderName = "journal";
const dailyFolderName = "Daily";
let lastActiveElement = null;
let scratchNoteExists = false;
let scratchAutosaveTimer = null;
let scratchDirty = false;
let scratchSaveInFlight = false;
const scratchAutosaveDelayMs = 1200;
let previewSyncTimer = null;
let previewSyncing = false;
let taskMetaOverflowTimer = null;
let previewDirty = false;
let serviceWorkerRegistration = null;
let offlineState = false;
const appVersion = "pwa-6";
const whatsNewItems = [
  "App shortcuts for Inbox, Daily, and Tasks.",
  "Offline fallback with a Retry action.",
  "Update toast for new versions.",
];
let startupShortcutHandled = false;
let commandItems = [];
let commandMatches = [];
let commandSelectedIndex = 0;
let commandSearchRequestId = 0;
let toastTimer = null;
if (brandBtn) {
  brandBtn.addEventListener("click", () => {
    window.location.reload();
  });
}

const showTasksRoot = true;

const tagPalette = [
  "#fde68a",
  "#fecdd3",
  "#bfdbfe",
  "#bbf7d0",
  "#e9d5ff",
  "#fbcfe8",
  "#bae6fd",
  "#fed7aa",
  "#c7d2fe",
  "#a7f3d0",
  "#f5d0fe",
  "#d9f99d",
  "#fecaca",
  "#cbd5f5",
  "#e0f2fe",
  "#fae8ff",
];

function setView(view, force = false) {
  if (!force && viewSelector.classList.contains("hidden")) {
    return;
  }
  app.dataset.view = view;
  viewButtons.forEach((btn) => {
    btn.classList.toggle("active", btn.dataset.view === view);
  });
  if (currentMode === "note") {
    lastNoteView = view;
  }
}

function isMobileView() {
  return window.matchMedia("(max-width: 720px)").matches;
}

function isSidebarCollapsed() {
  return app.classList.contains("sidebar-collapsed");
}

function updateSidebarToggle() {
  if (!sidebarToggle) {
    return;
  }
  const collapsed = isSidebarCollapsed();
  sidebarToggle.title = collapsed ? "Expand sidebar" : "Collapse sidebar";
  sidebarToggle.setAttribute("aria-label", sidebarToggle.title);
}

function toggleSidebarCollapse() {
  if (isMobileView()) {
    return;
  }
  app.classList.toggle("sidebar-collapsed");
  updateSidebarToggle();
}

function getPreferredView() {
  if (isMobileView()) {
    return "preview";
  }
  return getDefaultView(currentSettings.defaultView);
}

function openSidebar() {
  if (!isMobileView()) {
    return;
  }
  isSidebarOpen = true;
  app.classList.add("sidebar-open");
}

function closeSidebar() {
  if (!isMobileView()) {
    return;
  }
  isSidebarOpen = false;
  app.classList.remove("sidebar-open");
}

function updateMobileLayout() {
  const mobile = isMobileView();
  app.classList.toggle("mobile", mobile);
  if (mobile) {
    if (!isSidebarOpen) {
      app.classList.remove("sidebar-open");
    }
    app.classList.remove("sidebar-collapsed");
  } else {
    isSidebarOpen = false;
    app.classList.remove("sidebar-open");
  }
  updateSidebarToggle();
}

function escapeHtml(value) {
  if (value === null || value === undefined) {
    return "";
  }
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/\"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function displayNodeName(node) {
  if (node.type === "file" && node.name && node.name.toLowerCase().endsWith(".md")) {
    return node.name.slice(0, -3);
  }
  return node.name || "(root)";
}

function formatCountLabel(label, count) {
  return `${label} (${count})`;
}

let markdownRenderer = null;

function setUpdateAvailable(isAvailable) {
  if (!updateBanner) {
    return;
  }
  updateBanner.classList.toggle("hidden", !isAvailable);
  if (updateToast) {
    updateToast.classList.toggle("hidden", !isAvailable);
  }
  updateStatusBanners();
}

function normalizeHeaders(headers) {
  if (!headers) {
    return {};
  }
  if (headers instanceof Headers) {
    return Object.fromEntries(headers.entries());
  }
  if (Array.isArray(headers)) {
    return Object.fromEntries(headers);
  }
  return { ...headers };
}

async function queueRequestForSync(path, options) {
  const method = String(options.method || "GET").toUpperCase();
  if (method === "GET") {
    return false;
  }
  if (!navigator.serviceWorker || !navigator.serviceWorker.controller) {
    return false;
  }
  const headers = normalizeHeaders(options.headers);
  if (options.body && !headers["Content-Type"] && !headers["content-type"]) {
    headers["Content-Type"] = "application/json";
  }
  const payload = {
    url: `${apiBase}${path}`,
    method,
    headers,
    body: typeof options.body === "string" ? options.body : "",
    queuedAt: Date.now(),
  };
  navigator.serviceWorker.controller.postMessage({ type: "QUEUE_REQUEST", payload });
  if (serviceWorkerRegistration && serviceWorkerRegistration.sync) {
    try {
      await serviceWorkerRegistration.sync.register("scoli-sync");
    } catch (err) {
      console.warn("Unable to register background sync", err);
    }
  }
  return true;
}

function updateStatusBanners() {
  if (!statusBanners) {
    return;
  }
  const hasOffline = offlineBanner && !offlineBanner.classList.contains("hidden");
  const hasUpdate = updateBanner && !updateBanner.classList.contains("hidden");
  statusBanners.classList.toggle("hidden", !(hasOffline || hasUpdate));
}

function setOfflineState(isOffline) {
  offlineState = isOffline;
  if (!offlineBanner) {
    return;
  }
  offlineBanner.classList.toggle("hidden", !isOffline);
  updateStatusBanners();
}

async function checkServerHealth() {
  try {
    const response = await fetch(`${apiBase}/health`, { cache: "no-store" });
    if (!response.ok) {
      throw new Error("Health check failed");
    }
    setOfflineState(false);
  } catch (err) {
    setOfflineState(true);
  }
}

function showUpdateBanner(registration) {
  if (!updateBanner || !updateReload) {
    return;
  }
  if (registration) {
    serviceWorkerRegistration = registration;
  }
  setUpdateAvailable(true);
}

async function registerServiceWorker() {
  if (!("serviceWorker" in navigator)) {
    return;
  }
  try {
    const registration = await navigator.serviceWorker.register("/sw.js");
    serviceWorkerRegistration = registration;
    setUpdateAvailable(!!registration.waiting);
    registration.addEventListener("updatefound", () => {
      const newWorker = registration.installing;
      if (!newWorker) {
        return;
      }
      newWorker.addEventListener("statechange", () => {
        if (newWorker.state === "installed" && navigator.serviceWorker.controller) {
          showUpdateBanner(registration);
        }
      });
    });
    navigator.serviceWorker.addEventListener("controllerchange", () => {
      setUpdateAvailable(false);
      window.location.reload();
    });
  } catch (err) {
    console.warn("Service worker registration failed", err);
  }
}

function buildMarkdownRenderer() {
  if (!window.markdownit) {
    return null;
  }

  const md = window.markdownit({
    html: true,
    linkify: true,
    breaks: true,
    langPrefix: "language-",
  });

  md.core.ruler.after("inline", "task-checkboxes", (state) => {
    const Token = state.Token;
    state.tokens.forEach((token, index) => {
      if (token.type !== "inline" || !token.children || token.children.length === 0) {
        return;
      }
      const prev = state.tokens[index - 1];
      const prevPrev = state.tokens[index - 2];
      if (!prev || !prevPrev || prev.type !== "paragraph_open" || prevPrev.type !== "list_item_open") {
        return;
      }
      const first = token.children[0];
      if (!first || first.type !== "text") {
        return;
      }
      const match = first.content.match(/^\[([ xX])\]\s+/);
      if (!match) {
        return;
      }
      const checked = match[1].toLowerCase() === "x";
      first.content = first.content.slice(match[0].length);
      if (first.content.length === 0) {
        token.children.shift();
      }
      const checkbox = new Token("html_inline", "", 0);
      checkbox.content = `<input type="checkbox" class="task-checkbox" contenteditable="false"${checked ? " checked" : ""} /> `;
      token.children.unshift(checkbox);
    });
  });

  const defaultLinkOpen = md.renderer.rules.link_open || ((tokens, idx, options, env, self) =>
    self.renderToken(tokens, idx, options)
  );
  md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
    const token = tokens[idx];
    const href = token.attrGet("href") || "";
    if (href.startsWith("http")) {
      token.attrSet("target", "_blank");
      token.attrSet("rel", "noopener");
    }
    return defaultLinkOpen(tokens, idx, options, env, self);
  };

  md.renderer.rules.image = (tokens, idx, options, env, self) => {
    const token = tokens[idx];
    const src = token.attrGet("src") || "";
    token.attrSet("src", resolveAssetPath(src));
    if (token.content) {
      token.attrSet("alt", token.content);
    }
    return self.renderToken(tokens, idx, options);
  };

  return md;
}

function getMarkdownRenderer() {
  if (!markdownRenderer) {
    markdownRenderer = buildMarkdownRenderer();
  }
  return markdownRenderer;
}

function showWhatsNewModal() {
  if (!whatsNewModal || !whatsNewList) {
    return;
  }
  whatsNewList.innerHTML = "";
  whatsNewItems.forEach((item) => {
    const li = document.createElement("li");
    li.textContent = item;
    whatsNewList.appendChild(li);
  });
  whatsNewModal.classList.remove("hidden");
}

function hideWhatsNewModal() {
  if (!whatsNewModal) {
    return;
  }
  whatsNewModal.classList.add("hidden");
  localStorage.setItem("scoli-version", appVersion);
}

function maybeShowWhatsNew() {
  const stored = localStorage.getItem("scoli-version");
  if (stored === appVersion) {
    return;
  }
  if (whatsNewModal && !whatsNewModal.classList.contains("hidden")) {
    return;
  }
  showWhatsNewModal();
}

function handleStartupShortcut() {
  if (startupShortcutHandled) {
    return;
  }
  startupShortcutHandled = true;
  const url = new URL(window.location.href);
  const shortcut = url.searchParams.get("shortcut");
  if (!shortcut) {
    return;
  }
  if (shortcut === "inbox") {
    openNote(inboxNotePath);
  } else if (shortcut === "daily") {
    openDailyNote();
  } else if (shortcut === "tasks") {
    currentActivePath = "__tasks__";
    restoreTaskSelection(currentActivePath, currentTasks || []);
  } else if (shortcut === "today") {
    currentActivePath = "task-group:Today";
    restoreTaskSelection(currentActivePath, currentTasks || []);
  }
  url.searchParams.delete("shortcut");
  window.history.replaceState({}, "", url.pathname + url.search);
}

function renderMarkdown(text) {
  const md = getMarkdownRenderer();
  if (!md) {
    return `<pre>${escapeHtml(text)}</pre>`;
  }
  return md.render(text);
}

function isPreviewEditable() {
  return preview && preview.getAttribute("contenteditable") === "true";
}

function setPreviewEditable(enabled) {
  if (!preview) {
    return;
  }
  preview.setAttribute("contenteditable", enabled ? "true" : "false");
  preview.classList.toggle("preview-editable", enabled);
}

function schedulePreviewSync() {
  if (previewSyncing) {
    return;
  }
  if (previewSyncTimer) {
    window.clearTimeout(previewSyncTimer);
  }
  previewSyncTimer = window.setTimeout(() => {
    previewSyncTimer = null;
    syncEditorFromPreview();
  }, 120);
}

function updatePreviewFromMarkdown(text) {
  if (!preview) {
    return;
  }
  previewSyncing = true;
  preview.innerHTML = renderMarkdown(text);
  previewSyncing = false;
  applyHighlighting();
}

function syncEditorFromPreview() {
  if (!preview || previewSyncing || currentMode !== "note") {
    return;
  }
  previewSyncing = true;
  const nextContent = serializePreviewToMarkdown();
  if (editor.value !== nextContent) {
    editor.value = nextContent;
    renderTagBar(extractTags(nextContent));
    isDirty = true;
    saveBtn.disabled = !currentNotePath;
    scheduleNoteSave();
  }
  previewDirty = false;
  previewSyncing = false;
}

function serializePreviewToMarkdown() {
  if (!preview) {
    return "";
  }
  const blocks = [];
  preview.childNodes.forEach((node) => {
    const block = serializeBlock(node, { indent: "" });
    if (block) {
      blocks.push(block);
    }
  });
  return blocks.join("\n\n").replace(/\n{3,}/g, "\n\n").replace(/\s+$/, "");
}

function serializeBlock(node, ctx) {
  if (!node) {
    return "";
  }
  if (node.nodeType === Node.TEXT_NODE) {
    return String(node.textContent || "").trim();
  }
  if (node.nodeType !== Node.ELEMENT_NODE) {
    return "";
  }
  const tag = node.tagName.toLowerCase();
  if (tag === "p" || tag === "div") {
    return serializeInline(node).trim();
  }
  if (tag === "h1" || tag === "h2" || tag === "h3" || tag === "h4" || tag === "h5" || tag === "h6") {
    const level = Number(tag.slice(1)) || 1;
    return `${"#".repeat(level)} ${serializeInline(node).trim()}`;
  }
  if (tag === "ul" || tag === "ol") {
    return serializeList(node, ctx.indent);
  }
  if (tag === "blockquote") {
    const inner = serializeChildrenAsBlocks(node, ctx);
    if (!inner) {
      return "";
    }
    return inner
      .split("\n")
      .map((line) => (line ? `> ${line}` : ">"))
      .join("\n");
  }
  if (tag === "pre") {
    const code = node.textContent || "";
    const codeNode = node.querySelector("code");
    const language = codeNode && codeNode.className ? codeNode.className.replace("language-", "") : "";
    const fence = "```";
    return `${fence}${language}\n${code.replace(/\s+$/, "")}\n${fence}`;
  }
  if (tag === "table") {
    return serializeTable(node);
  }
  if (tag === "hr") {
    return "---";
  }
  return serializeInline(node).trim();
}

function serializeChildrenAsBlocks(node, ctx) {
  const blocks = [];
  node.childNodes.forEach((child) => {
    const block = serializeBlock(child, ctx);
    if (block) {
      blocks.push(block);
    }
  });
  return blocks.join("\n\n");
}

function serializeInline(node) {
  if (!node) {
    return "";
  }
  if (node.nodeType === Node.TEXT_NODE) {
    return String(node.textContent || "");
  }
  if (node.nodeType !== Node.ELEMENT_NODE) {
    return "";
  }
  const tag = node.tagName.toLowerCase();
  if (tag === "br") {
    return "\n";
  }
  if (tag === "strong" || tag === "b") {
    return `**${serializeInlineChildren(node)}**`;
  }
  if (tag === "em" || tag === "i") {
    return `*${serializeInlineChildren(node)}*`;
  }
  if (tag === "code") {
    return `\`${serializeInlineChildren(node)}\``;
  }
  if (tag === "a") {
    const href = node.getAttribute("href") || "";
    return `[${serializeInlineChildren(node)}](${href})`;
  }
  if (tag === "img") {
    const alt = node.getAttribute("alt") || "";
    const src = node.getAttribute("src") || "";
    return `![${alt}](${src})`;
  }
  if (tag === "input" && node.type === "checkbox") {
    return node.checked ? "[x] " : "[ ] ";
  }
  return serializeInlineChildren(node);
}

function serializeInlineChildren(node) {
  let out = "";
  node.childNodes.forEach((child) => {
    out += serializeInline(child);
  });
  return out;
}

function serializeList(node, indent) {
  const isOrdered = node.tagName.toLowerCase() === "ol";
  const items = Array.from(node.children).filter((child) => child.tagName && child.tagName.toLowerCase() === "li");
  const lines = items.map((item, index) => serializeListItem(item, indent, isOrdered, index));
  return lines.join("\n");
}

function serializeListItem(item, indent, isOrdered, index) {
  const marker = isOrdered ? `${index + 1}. ` : "- ";
  const checkbox = item.querySelector(":scope > input[type=\"checkbox\"]");
  const checkboxPrefix = checkbox ? `[${checkbox.checked ? "x" : " "}] ` : "";
  const textParts = [];
  const nestedLists = [];
  item.childNodes.forEach((child) => {
    if (child.nodeType === Node.ELEMENT_NODE) {
      const tag = child.tagName.toLowerCase();
      if (tag === "ul" || tag === "ol") {
        nestedLists.push(child);
        return;
      }
      if (tag === "input" && child.type === "checkbox") {
        return;
      }
    }
    textParts.push(serializeInline(child));
  });
  let text = textParts.join("").replace(/\s+$/, "");
  if (!text) {
    text = "";
  }
  const linePrefix = `${indent}${marker}${checkboxPrefix}`;
  const lines = text.split("\n");
  let output = `${linePrefix}${lines[0] || ""}`.trimEnd();
  if (lines.length > 1) {
    const continuation = lines
      .slice(1)
      .map((line) => `${indent}  ${line}`);
    output += `\n${continuation.join("\n")}`;
  }
  nestedLists.forEach((list) => {
    const nested = serializeList(list, `${indent}  `);
    if (nested) {
      output += `\n${nested}`;
    }
  });
  return output;
}

function serializeTable(table) {
  const rows = Array.from(table.querySelectorAll("tr"));
  if (rows.length === 0) {
    return "";
  }
  const cells = rows.map((row) =>
    Array.from(row.children).map((cell) => String(cell.textContent || "").trim())
  );
  const header = cells[0];
  const separator = header.map(() => "---");
  const lines = [`| ${header.join(" | ")} |`, `| ${separator.join(" | ")} |`];
  cells.slice(1).forEach((row) => {
    lines.push(`| ${row.join(" | ")} |`);
  });
  return lines.join("\n");
}

function getSelectionRange() {
  const selection = window.getSelection();
  if (!selection || selection.rangeCount === 0) {
    return null;
  }
  return selection.getRangeAt(0);
}

function closestElement(node, selector) {
  let current = node && node.nodeType === Node.ELEMENT_NODE ? node : node?.parentElement;
  while (current && current !== preview) {
    if (current.matches && current.matches(selector)) {
      return current;
    }
    current = current.parentElement;
  }
  return null;
}

function isCaretAtStart(container, range) {
  if (!container || !range) {
    return false;
  }
  const probe = range.cloneRange();
  probe.selectNodeContents(container);
  probe.setEnd(range.startContainer, range.startOffset);
  return probe.toString().length === 0;
}

function isCaretAtEnd(container, range) {
  if (!container || !range) {
    return false;
  }
  const probe = range.cloneRange();
  probe.selectNodeContents(container);
  probe.setStart(range.endContainer, range.endOffset);
  return probe.toString().length === 0;
}

function placeCaretAtStart(node) {
  const range = document.createRange();
  const selection = window.getSelection();
  range.selectNodeContents(node);
  range.collapse(true);
  selection.removeAllRanges();
  selection.addRange(range);
}

function placeCaretAtEnd(node) {
  const range = document.createRange();
  const selection = window.getSelection();
  range.selectNodeContents(node);
  range.collapse(false);
  selection.removeAllRanges();
  selection.addRange(range);
}

function insertCheckboxListItemAfter(item) {
  const list = item.parentElement;
  if (!list) {
    return;
  }
  const newItem = document.createElement("li");
  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.className = "task-checkbox";
  newItem.appendChild(checkbox);
  newItem.appendChild(document.createTextNode(" "));
  list.insertBefore(newItem, item.nextSibling);
  placeCaretAtEnd(newItem);
}

function removeListItem(item) {
  const list = item.parentElement;
  const prev = item.previousElementSibling;
  const next = item.nextElementSibling;
  item.remove();
  if (list && list.children.length === 0) {
    const paragraph = document.createElement("p");
    paragraph.appendChild(document.createElement("br"));
    list.replaceWith(paragraph);
    placeCaretAtStart(paragraph);
    return;
  }
  if (prev) {
    placeCaretAtEnd(prev);
    return;
  }
  if (next) {
    placeCaretAtStart(next);
  }
}

function handlePreviewKeydown(event) {
  if (currentMode !== "note" || !isPreviewEditable()) {
    return;
  }
  const range = getSelectionRange();
  if (!range || !range.collapsed) {
    return;
  }
  if (event.key === "Enter") {
    const listItem = closestElement(range.startContainer, "li");
    if (listItem) {
      const checkbox = listItem.querySelector(":scope > input[type=\"checkbox\"]");
      if (checkbox && isCaretAtEnd(listItem, range)) {
        event.preventDefault();
        insertCheckboxListItemAfter(listItem);
        schedulePreviewSync();
      }
      return;
    }
    const blockquote = closestElement(range.startContainer, "blockquote");
    if (blockquote && isCaretAtEnd(blockquote, range)) {
      event.preventDefault();
      const paragraph = document.createElement("p");
      paragraph.appendChild(document.createElement("br"));
      blockquote.appendChild(paragraph);
      placeCaretAtStart(paragraph);
      schedulePreviewSync();
    }
  }
  if (event.key === "Backspace") {
    const listItem = closestElement(range.startContainer, "li");
    if (listItem) {
      const checkbox = listItem.querySelector(":scope > input[type=\"checkbox\"]");
      const text = listItem.textContent.replace(/\u200b/g, "").trim();
      if (checkbox && text === "" && isCaretAtStart(listItem, range)) {
        event.preventDefault();
        removeListItem(listItem);
        schedulePreviewSync();
      }
    }
  }
}

function extractTags(text) {
  if (!text) {
    return [];
  }
  const pattern = /(^|\s)#([A-Za-z]+)\b/g;
  const seen = new Set();
  const tags = [];
  let match;
  while ((match = pattern.exec(text)) !== null) {
    const tag = match[2];
    const key = tag.toLowerCase();
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    tags.push(tag);
  }
  return tags;
}

function isBlankLine(line) {
  return !line || line.trim() === "";
}

function isTagLine(line) {
  return /^\s*(#[A-Za-z]+(\s+#[A-Za-z]+)*)\s*$/.test(line || "");
}

function splitTrailingTagBlock(lines) {
  let end = lines.length;
  const trailingBlanks = [];
  while (end > 0 && isBlankLine(lines[end - 1])) {
    trailingBlanks.unshift(lines[end - 1]);
    end -= 1;
  }
  if (end > 0 && isTagLine(lines[end - 1])) {
    const tagLines = [...trailingBlanks];
    while (end > 0 && (isTagLine(lines[end - 1]) || isBlankLine(lines[end - 1]))) {
      tagLines.unshift(lines[end - 1]);
      end -= 1;
    }
    return { bodyLines: lines.slice(0, end), tagLines };
  }
  return { bodyLines: lines.slice(0, end).concat(trailingBlanks), tagLines: [] };
}

function isFenceLine(line) {
  return /^\s*```/.test(line || "");
}

function isTaskLine(line) {
  return /^\s*(~\s*)?-\s*\[[ xX]\]/.test(line || "");
}

function isCompletedTaskLine(line) {
  return /^\s*(~\s*)?-\s*\[[xX]\]/.test(line || "");
}

function markTaskCompleted(line) {
  if (!isTaskLine(line)) {
    return line;
  }
  return line.replace(/^(\s*)(~\s*)?-\s*\[[ xX]\]/, (match, indent, tilde) => {
    const prefix = tilde || "";
    return `${indent}${prefix}- [x]`;
  });
}

function trimBlankEdges(lines) {
  let start = 0;
  let end = lines.length;
  while (start < end && isBlankLine(lines[start])) {
    start += 1;
  }
  while (end > start && isBlankLine(lines[end - 1])) {
    end -= 1;
  }
  return lines.slice(start, end);
}

function appendBlock(lines, block) {
  const trimmed = trimBlankEdges(block);
  if (trimmed.length === 0) {
    return;
  }
  if (lines.length > 0 && !isBlankLine(lines[lines.length - 1])) {
    lines.push("");
  }
  trimmed.forEach((line) => lines.push(line));
}

function headingLevel(line) {
  const match = /^\s*(#{1,6})\s+/.exec(line || "");
  if (!match) {
    return 0;
  }
  return match[1].length;
}

function extractDoneSection(lines) {
  let startIndex = -1;
  let endIndex = lines.length;
  let inCodeBlock = false;
  for (let i = 0; i < lines.length; i += 1) {
    const line = lines[i];
    if (isFenceLine(line)) {
      inCodeBlock = !inCodeBlock;
      continue;
    }
    if (inCodeBlock) {
      continue;
    }
    if (/^\s*##\s+Done\s*$/i.test(line)) {
      startIndex = i;
      break;
    }
  }
  if (startIndex === -1) {
    return { lines, doneContent: [], hadDoneHeading: false };
  }
  inCodeBlock = false;
  for (let i = startIndex + 1; i < lines.length; i += 1) {
    const line = lines[i];
    if (isFenceLine(line)) {
      inCodeBlock = !inCodeBlock;
      continue;
    }
    if (inCodeBlock) {
      continue;
    }
    const level = headingLevel(line);
    if (level > 0 && level <= 2) {
      endIndex = i;
      break;
    }
  }
  const doneContent = lines.slice(startIndex + 1, endIndex);
  const remaining = lines.slice(0, startIndex).concat(lines.slice(endIndex));
  return { lines: remaining, doneContent, hadDoneHeading: true };
}

function moveCompletedTasksToDoneSection(text) {
  if (!text) {
    return { text, movedBlocks: 0 };
  }
  const normalized = text.replace(/\r\n/g, "\n");
  const hasTrailingNewline = normalized.endsWith("\n");
  const lines = normalized.split("\n");
  const { bodyLines, tagLines } = splitTrailingTagBlock(lines);
  const extracted = extractDoneSection(bodyLines);
  const remainingLines = [];
  const movedBlocks = [];
  let inCodeBlock = false;

  for (let i = 0; i < extracted.lines.length; i += 1) {
    const line = extracted.lines[i];
    if (isFenceLine(line)) {
      inCodeBlock = !inCodeBlock;
      remainingLines.push(line);
      continue;
    }
    if (!inCodeBlock && isCompletedTaskLine(line)) {
      const parentIndent = (line.match(/^\s*/) || [""])[0].length;
      const block = [];
      let blockInCode = false;
      block.push(markTaskCompleted(line));
      let j = i + 1;
      for (; j < extracted.lines.length; j += 1) {
        const nextLine = extracted.lines[j];
        const nextIndent = (nextLine.match(/^\s*/) || [""])[0].length;
        if (nextIndent <= parentIndent) {
          break;
        }
        if (isFenceLine(nextLine)) {
          blockInCode = !blockInCode;
          block.push(nextLine);
          continue;
        }
        if (blockInCode) {
          block.push(nextLine);
          continue;
        }
        if (isTaskLine(nextLine)) {
          block.push(markTaskCompleted(nextLine));
        } else {
          block.push(nextLine);
        }
      }
      movedBlocks.push(block);
      i = j - 1;
      continue;
    }
    remainingLines.push(line);
  }

  const doneLines = trimBlankEdges(extracted.doneContent);
  movedBlocks.forEach((block) => appendBlock(doneLines, block));

  let outputLines = remainingLines;
  const shouldWriteDone = extracted.hadDoneHeading || doneLines.length > 0 || movedBlocks.length > 0;
  if (shouldWriteDone) {
    if (outputLines.length > 0 && !isBlankLine(outputLines[outputLines.length - 1])) {
      outputLines.push("");
    }
    outputLines.push("## Done");
    if (doneLines.length > 0) {
      outputLines.push("");
      outputLines = outputLines.concat(doneLines);
    }
  }

  if (tagLines.length > 0) {
    if (outputLines.length > 0 && !isBlankLine(outputLines[outputLines.length - 1])) {
      outputLines.push("");
    }
    outputLines = outputLines.concat(tagLines);
  }

  let result = outputLines.join("\n");
  if (hasTrailingNewline && !result.endsWith("\n")) {
    result += "\n";
  }
  return { text: result, movedBlocks: movedBlocks.length };
}

function renderTagBar(tags) {
  tagPills.innerHTML = "";
  if (!currentNotePath) {
    tagBar.classList.add("hidden");
    return;
  }
  (tags || []).forEach((tag) => {
    const pill = document.createElement("button");
    pill.type = "button";
    pill.className = "tag-pill";
    pill.textContent = `#${tag}`;
    pill.style.backgroundColor = getTagColor(tag);
    pill.addEventListener("click", () => openTagGroup(tag));
    tagPills.appendChild(pill);
  });
  tagBar.classList.remove("hidden");
}

function showTaskList(title, tasks, summaryText) {
  if (currentMode === "note") {
    lastNoteView = app.dataset.view;
  }
  currentMode = "tasks";
  setPreviewEditable(false);
  currentNotePath = "";
  notePath.textContent = title;
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  tagBar.classList.add("hidden");
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  summaryPanel.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  if (journalPanel) {
    journalPanel.classList.add("hidden");
  }
  editor.classList.add("hidden");
  preview.classList.add("hidden");
  assetPreview.classList.add("hidden");
  assetPreview.innerHTML = "";
  pdfPreview.classList.add("hidden");
  pdfPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  taskList.classList.remove("hidden");
  editorPane.classList.add("hidden");
  paneResizer.classList.add("hidden");
  previewPane.classList.remove("hidden");
  setView("preview", true);
  renderTaskList(title, tasks, summaryText);
}

function showNoteEditor() {
  currentMode = "note";
  summaryPanel.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  taskList.classList.add("hidden");
  if (journalPanel) {
    journalPanel.classList.add("hidden");
  }
  if (journalPanel) {
    journalPanel.classList.add("hidden");
  }
  editor.classList.remove("hidden");
  preview.classList.remove("hidden");
  setPreviewEditable(true);
  editorPane.classList.remove("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.remove("hidden");
  viewSelector.classList.remove("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = false;
  });
  setView(lastNoteView || getPreferredView());
}

function showSummary(title, items, action) {
  currentMode = "summary";
  setPreviewEditable(false);
  currentNotePath = "";
  notePath.textContent = title;
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  tagBar.classList.add("hidden");
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  taskList.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  if (journalPanel) {
    journalPanel.classList.add("hidden");
  }
  editor.classList.add("hidden");
  preview.classList.add("hidden");
  assetPreview.classList.add("hidden");
  assetPreview.innerHTML = "";
  pdfPreview.classList.add("hidden");
  pdfPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  summaryPanel.innerHTML = "";
  summaryPanel.classList.remove("hidden");
  editorPane.classList.remove("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.remove("hidden");
  setView("preview", true);

  const header = document.createElement("div");
  header.className = "summary-header";

  const heading = document.createElement("h2");
  heading.className = "summary-title";
  heading.textContent = title;
  header.appendChild(heading);

  if (action && action.label && action.handler) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "primary summary-action";
    button.textContent = action.label;
    button.addEventListener("click", action.handler);
    header.appendChild(button);
  }

  summaryPanel.appendChild(header);

  const grid = document.createElement("div");
  grid.className = "summary-grid";
  items.forEach((item) => {
    const card = document.createElement("div");
    card.className = "summary-item";
    const label = document.createElement("div");
    label.className = "summary-label";
    label.textContent = item.label;
    const value = document.createElement("div");
    value.className = "summary-value";
    value.textContent = String(item.value);
    card.appendChild(label);
    card.appendChild(value);
    grid.appendChild(card);
  });
  summaryPanel.appendChild(grid);
}

function formatJournalTimestamp(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return date.toLocaleString();
}

function showJournal() {
  if (currentMode === "note") {
    lastNoteView = app.dataset.view;
  }
  currentMode = "journal";
  journalViewMode = "main";
  setPreviewEditable(false);
  currentNotePath = "";
  currentActivePath = journalRootPath;
  notePath.textContent = "Journal";
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  tagBar.classList.add("hidden");
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  summaryPanel.classList.add("hidden");
  taskList.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  editor.classList.add("hidden");
  preview.classList.add("hidden");
  assetPreview.classList.add("hidden");
  assetPreview.innerHTML = "";
  pdfPreview.classList.add("hidden");
  pdfPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  if (journalPanel) {
    journalPanel.classList.remove("hidden");
  }
  if (journalCompose) {
    journalCompose.classList.remove("hidden");
  }
  editorPane.classList.add("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.add("hidden");
  setView("preview", true);
  setActiveNode(currentActivePath);
  loadJournalEntries().catch((err) => alert(err.message));
}

function showJournalArchive(date) {
  if (currentMode === "note") {
    lastNoteView = app.dataset.view;
  }
  currentMode = "journal-archive";
  journalViewMode = "archive";
  setPreviewEditable(false);
  currentNotePath = "";
  currentActivePath = `${journalRootPath}:${date}`;
  notePath.textContent = `Journal Archive: ${date}`;
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  tagBar.classList.add("hidden");
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  summaryPanel.classList.add("hidden");
  taskList.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  editor.classList.add("hidden");
  preview.classList.add("hidden");
  assetPreview.classList.add("hidden");
  assetPreview.innerHTML = "";
  pdfPreview.classList.add("hidden");
  pdfPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  if (journalPanel) {
    journalPanel.classList.remove("hidden");
  }
  if (journalCompose) {
    journalCompose.classList.add("hidden");
  }
  editorPane.classList.add("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.add("hidden");
  setView("preview", true);
  setActiveNode(currentActivePath);
  loadJournalEntries(date).catch((err) => alert(err.message));
}

async function loadJournalEntries(archiveDate = "") {
  const response = archiveDate
    ? await apiFetch(`/journal/archive?date=${encodeURIComponent(archiveDate)}`)
    : await apiFetch("/journal");
  journalEntries = response.entries || [];
  renderJournal();
}

function renderJournal() {
  if (!journalFeed) {
    return;
  }
  journalFeed.innerHTML = "";
  const readOnly = journalViewMode === "archive";
  const entries = [...(journalEntries || [])].sort((a, b) => {
    const aDate = new Date(a.createdAt).getTime();
    const bDate = new Date(b.createdAt).getTime();
    if (Number.isNaN(aDate) || Number.isNaN(bDate)) {
      return 0;
    }
    return bDate - aDate;
  });

  if (entries.length === 0) {
    const empty = document.createElement("div");
    empty.className = "empty";
    empty.textContent = "No journal entries yet.";
    journalFeed.appendChild(empty);
    return;
  }

  entries.forEach((entry) => {
    const card = document.createElement("div");
    card.className = "journal-entry";
    card.dataset.id = entry.id;

    const meta = document.createElement("div");
    meta.className = "journal-entry-meta";
    const created = formatJournalTimestamp(entry.createdAt);
    const updated = formatJournalTimestamp(entry.updatedAt);
    const createdSpan = document.createElement("span");
    createdSpan.textContent = created ? `Created ${created}` : "Created";
    meta.appendChild(createdSpan);
    if (updated && updated !== created) {
      const updatedSpan = document.createElement("span");
      updatedSpan.textContent = `Updated ${updated}`;
      meta.appendChild(updatedSpan);
    }

    const content = document.createElement("div");
    content.className = "journal-entry-content";
    content.textContent = entry.content || "";

    card.appendChild(meta);
    card.appendChild(content);
    if (!readOnly) {
      const actions = document.createElement("div");
      actions.className = "journal-entry-actions";

      const editBtn = document.createElement("button");
      editBtn.type = "button";
      editBtn.className = "ghost";
      editBtn.textContent = "Edit";
      editBtn.addEventListener("click", () => {
        enterJournalEdit(card, entry);
      });

      const archiveBtn = document.createElement("button");
      archiveBtn.type = "button";
      archiveBtn.className = "ghost";
      archiveBtn.textContent = "Archive";
      archiveBtn.addEventListener("click", async () => {
        if (!confirm("Archive this journal entry?")) {
          return;
        }
        try {
          await archiveJournalEntry(entry.id);
        } catch (err) {
          alert(err.message);
        }
      });

      const deleteBtn = document.createElement("button");
      deleteBtn.type = "button";
      deleteBtn.className = "ghost";
      deleteBtn.textContent = "Delete";
      deleteBtn.addEventListener("click", async () => {
        if (!confirm("Delete this journal entry?")) {
          return;
        }
        try {
          await deleteJournalEntry(entry.id);
        } catch (err) {
          alert(err.message);
        }
      });

      actions.appendChild(editBtn);
      actions.appendChild(archiveBtn);
      actions.appendChild(deleteBtn);
      card.appendChild(actions);
    }
    journalFeed.appendChild(card);
  });
}

function enterJournalEdit(card, entry) {
  if (!card || !entry) {
    return;
  }
  const existingEditor = card.querySelector(".journal-entry-edit");
  if (existingEditor) {
    existingEditor.focus();
    return;
  }

  const content = card.querySelector(".journal-entry-content");
  const actions = card.querySelector(".journal-entry-actions");
  if (!content || !actions) {
    return;
  }

  const editor = document.createElement("textarea");
  editor.className = "journal-entry-edit";
  editor.value = entry.content || "";
  editor.addEventListener("keydown", async (event) => {
    if (event.key !== "Enter" || event.altKey) {
      return;
    }
    event.preventDefault();
    try {
      await updateJournalEntry(entry.id, editor.value);
    } catch (err) {
      alert(err.message);
    }
  });

  content.replaceWith(editor);
  actions.innerHTML = "";

  const saveBtn = document.createElement("button");
  saveBtn.type = "button";
  saveBtn.className = "primary";
  saveBtn.textContent = "Save";
  saveBtn.addEventListener("click", async () => {
    try {
      await updateJournalEntry(entry.id, editor.value);
    } catch (err) {
      alert(err.message);
    }
  });

  const cancelBtn = document.createElement("button");
  cancelBtn.type = "button";
  cancelBtn.className = "ghost";
  cancelBtn.textContent = "Cancel";
  cancelBtn.addEventListener("click", () => {
    renderJournal();
  });

  actions.appendChild(cancelBtn);
  actions.appendChild(saveBtn);
  setTimeout(() => editor.focus(), 0);
}

async function createJournalEntry(content) {
  const payload = { content: String(content || "") };
  const entry = await apiFetch("/journal", {
    method: "POST",
    body: JSON.stringify(payload),
  });
  journalEntries = [entry].concat(journalEntries || []);
  renderJournal();
  try {
    await refreshJournalSummary();
  } catch (err) {
    console.warn("Unable to refresh journal count", err);
  }
}

async function updateJournalEntry(id, content) {
  await apiFetch("/journal", {
    method: "PATCH",
    body: JSON.stringify({ id, content }),
  });
  await loadJournalEntries();
}

async function deleteJournalEntry(id) {
  await apiFetch(`/journal?id=${encodeURIComponent(id)}`, { method: "DELETE" });
  await loadJournalEntries();
  try {
    await refreshJournalSummary();
  } catch (err) {
    console.warn("Unable to refresh journal count", err);
  }
}

async function archiveJournalEntry(id) {
  await apiFetch("/journal/archive", {
    method: "POST",
    body: JSON.stringify({ id }),
  });
  await loadJournalEntries();
}

function getDefaultView(value) {
  if (value === "edit" || value === "preview" || value === "split") {
    return value;
  }
  return "preview";
}

function scheduleNoteSave() {
  if (currentMode !== "note" || !currentNotePath) {
    return;
  }
  noteSaveSnapshot = {
    path: currentNotePath,
    content: editor.value,
  };
  if (noteSaveTimer) {
    window.clearTimeout(noteSaveTimer);
  }
  noteSaveTimer = window.setTimeout(() => {
    noteSaveTimer = null;
    flushNoteSave();
  }, 800);
}

function applySidebarWidth(width) {
  if (!width || Number.isNaN(width)) {
    return;
  }
  const clamped = Math.min(600, Math.max(220, Math.round(width)));
  sidebar.style.width = `${clamped}px`;
  document.documentElement.style.setProperty("--sidebar-width", `${clamped}px`);
  currentSettings.sidebarWidth = clamped;
}

function parentPathForPath(path) {
  if (!path) {
    return "";
  }
  const parts = path.split("/").filter(Boolean);
  parts.pop();
  return parts.join("/");
}

function clearDragState() {
  dragState = null;
  if (dragOverRow) {
    dragOverRow.classList.remove("drag-over");
    dragOverRow = null;
  }
  if (dragExpandTimer) {
    clearTimeout(dragExpandTimer);
    dragExpandTimer = null;
  }
}

function getDropTargetPath(row) {
  if (!row || !row.dataset) {
    return null;
  }
  if (row.dataset.type === "folder") {
    return row.dataset.path || "";
  }
  if (row.dataset.type === "file") {
    return parentPathForPath(row.dataset.path);
  }
  return null;
}

function isInvalidFolderMove(folderPath, targetFolderPath) {
  if (!folderPath) {
    return true;
  }
  if (targetFolderPath === folderPath) {
    return true;
  }
  return targetFolderPath.startsWith(`${folderPath}/`);
}

function isValidDropTarget(targetFolderPath) {
  if (!dragState || targetFolderPath === null) {
    return false;
  }
  if (dragState.type === "folder" && isInvalidFolderMove(dragState.path, targetFolderPath)) {
    return false;
  }
  return true;
}

function scheduleFolderExpand(row) {
  if (!row || row.dataset.type !== "folder") {
    return;
  }
  const wrapper = row.closest(".tree-node.folder");
  if (!wrapper || !wrapper.classList.contains("collapsed")) {
    return;
  }
  if (dragExpandTimer) {
    clearTimeout(dragExpandTimer);
  }
  dragExpandTimer = setTimeout(() => {
    wrapper.classList.remove("collapsed");
    dragExpandTimer = null;
  }, dragExpandDelayMs);
}

async function moveNoteToFolder(path, targetFolderPath) {
  const baseName = path.split("/").pop() || "";
  const newPath = targetFolderPath ? `${targetFolderPath}/${baseName}` : baseName;
  if (!baseName || newPath === path) {
    return;
  }
  try {
    const data = await apiFetch("/notes/rename", {
      method: "PATCH",
      body: JSON.stringify({ path, newPath }),
    });
    await loadTree();
    const updatedPath = data.newPath || newPath;
    if (currentNotePath === path) {
      await openNote(updatedPath);
    }
  } catch (err) {
    alert(err.message);
  }
}

async function moveFolderToFolder(path, targetFolderPath) {
  const baseName = path.split("/").pop() || "";
  const newPath = targetFolderPath ? `${targetFolderPath}/${baseName}` : baseName;
  if (!baseName || newPath === path) {
    return;
  }
  if (isInvalidFolderMove(path, targetFolderPath)) {
    alert("Folders cannot be moved into themselves.");
    return;
  }
  if (currentActivePath && (currentActivePath === path || currentActivePath.startsWith(`${path}/`))) {
    currentActivePath = `${newPath}${currentActivePath.slice(path.length)}`;
  }
  try {
    await apiFetch("/folders", {
      method: "PATCH",
      body: JSON.stringify({ path, newPath }),
    });
    await loadTree();
    if (currentNotePath && currentNotePath.startsWith(`${path}/`)) {
      const suffix = currentNotePath.slice(path.length);
      await openNote(`${newPath}${suffix}`);
    } else {
      focusTreePath(newPath);
    }
  } catch (err) {
    alert(err.message);
  }
}

function handleDragStart(event, node) {
  if (!event.dataTransfer) {
    return;
  }
  dragState = { path: node.path || "", type: node.type || "" };
  event.dataTransfer.effectAllowed = "move";
  try {
    event.dataTransfer.setData("text/plain", JSON.stringify(dragState));
  } catch {
    // Ignore data transfer errors; we still track drag state locally.
  }
  const row = event.currentTarget;
  if (row) {
    row.classList.add("dragging");
  }
}

function handleDragEnd(event) {
  const row = event.currentTarget;
  if (row) {
    row.classList.remove("dragging");
  }
  clearDragState();
}

function handleDragOver(event) {
  if (!dragState) {
    return;
  }
  const row = event.currentTarget;
  const targetFolderPath = getDropTargetPath(row);
  if (!isValidDropTarget(targetFolderPath)) {
    if (dragOverRow) {
      dragOverRow.classList.remove("drag-over");
      dragOverRow = null;
    }
    return;
  }
  event.preventDefault();
  event.dataTransfer.dropEffect = "move";
  if (dragOverRow !== row) {
    if (dragOverRow) {
      dragOverRow.classList.remove("drag-over");
    }
    dragOverRow = row;
    dragOverRow.classList.add("drag-over");
  }
  scheduleFolderExpand(row);
}

function handleDragLeave(event) {
  const row = event.currentTarget;
  if (dragOverRow === row) {
    dragOverRow.classList.remove("drag-over");
    dragOverRow = null;
  }
}

async function handleDrop(event) {
  if (!dragState) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  const row = event.currentTarget;
  const targetFolderPath = getDropTargetPath(row);
  if (!isValidDropTarget(targetFolderPath)) {
    clearDragState();
    return;
  }
  const { path, type } = dragState;
  clearDragState();
  if (type === "file") {
    await moveNoteToFolder(path, targetFolderPath);
  } else if (type === "folder") {
    await moveFolderToFolder(path, targetFolderPath);
  }
}

function findFolderNode(tree, path) {
  if (!tree || !path) {
    return null;
  }
  const queue = [tree];
  while (queue.length > 0) {
    const node = queue.shift();
    if (!node) {
      continue;
    }
    if (node.type === "folder" && node.path === path) {
      return node;
    }
    if (node.children && node.children.length > 0) {
      queue.push(...node.children);
    }
  }
  return null;
}

function getRootIconPath(key) {
  if (!currentSettings.rootIcons) {
    return "";
  }
  return currentSettings.rootIcons[key] || "";
}

function applyRootIconToRow(row, key) {
  if (!row) {
    return;
  }
  const icon = row.querySelector(".folder-icon");
  if (!icon) {
    return;
  }
  icon.classList.add("root-icon");
  const path = getRootIconPath(key);
  if (path) {
    icon.classList.add("custom-root-icon");
    icon.style.backgroundImage = `url("${path}")`;
    return;
  }
  icon.classList.remove("custom-root-icon");
  icon.style.removeProperty("background-image");
}

function rootIconMenuItems(rootKey, row) {
  const items = [
    {
      label: "Change Icon",
      action: () => promptRootIconChange(rootKey, row),
    },
  ];
  if (getRootIconPath(rootKey)) {
    items.push({
      label: "Reset Icon",
      action: () => resetRootIcon(rootKey, row),
    });
  }
  return items;
}

function applySettings(settings) {
  currentSettings = {
    darkMode: !!settings.darkMode,
    defaultView: getDefaultView(settings.defaultView),
    sidebarWidth: Number(settings.sidebarWidth) || 300,
    defaultFolder: settings.defaultFolder || "",
    showTemplates: settings.showTemplates !== false,
    notesSortBy: settings.notesSortBy || "name",
    notesSortOrder: settings.notesSortOrder || "asc",
    externalCommandsPath: settings.externalCommandsPath || "",
    rootIcons: settings.rootIcons || {},
  };
  document.body.classList.toggle("theme-dark", currentSettings.darkMode);
  if (settingsDarkMode) {
    settingsDarkMode.checked = currentSettings.darkMode;
  }
  if (settingsDefaultView) {
    settingsDefaultView.value = currentSettings.defaultView;
  }
  if (settingsDefaultFolder) {
    settingsDefaultFolder.value = currentSettings.defaultFolder;
  }
  if (settingsShowTemplates) {
    settingsShowTemplates.checked = currentSettings.showTemplates;
  }
  if (settingsNotesSortBy) {
    settingsNotesSortBy.value = currentSettings.notesSortBy;
  }
  if (settingsNotesSortOrder) {
    settingsNotesSortOrder.value = currentSettings.notesSortOrder;
  }
  applySidebarWidth(currentSettings.sidebarWidth);
}

function showSettings() {
  currentMode = "settings";
  setPreviewEditable(false);
  currentNotePath = "";
  notePath.textContent = "Settings";
  tagBar.classList.add("hidden");
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  summaryPanel.classList.add("hidden");
  taskList.classList.add("hidden");
  editor.classList.add("hidden");
  preview.classList.add("hidden");
  if (journalPanel) {
    journalPanel.classList.add("hidden");
  }
  assetPreview.classList.add("hidden");
  assetPreview.innerHTML = "";
  pdfPreview.classList.add("hidden");
  pdfPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  settingsPanel.classList.remove("hidden");
  editorPane.classList.remove("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.remove("hidden");
  setView("edit", true);
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  isDirty = false;
  if (settingsDarkMode) {
    settingsDarkMode.checked = currentSettings.darkMode;
  }
  if (settingsDefaultView) {
    settingsDefaultView.value = currentSettings.defaultView || "preview";
  }
  if (settingsDefaultFolder) {
    settingsDefaultFolder.value = currentSettings.defaultFolder || "";
  }
  if (settingsShowTemplates) {
    settingsShowTemplates.checked = currentSettings.showTemplates;
  }
  if (settingsNotesSortBy) {
    settingsNotesSortBy.value = currentSettings.notesSortBy || "name";
  }
  if (settingsNotesSortOrder) {
    settingsNotesSortOrder.value = currentSettings.notesSortOrder || "asc";
  }
}

async function saveSettings() {
  if (
    !settingsDarkMode ||
    !settingsDefaultView ||
    !settingsDefaultFolder ||
    !settingsShowTemplates ||
    !settingsNotesSortBy ||
    !settingsNotesSortOrder
  ) {
    return;
  }
  try {
    saveBtn.disabled = true;
    saveBtn.textContent = "Saving...";
    const payload = {
      darkMode: settingsDarkMode.checked,
      defaultView: settingsDefaultView.value,
      sidebarWidth: currentSettings.sidebarWidth || 300,
      defaultFolder: settingsDefaultFolder.value.trim(),
      showTemplates: settingsShowTemplates.checked,
      notesSortBy: settingsNotesSortBy.value,
      notesSortOrder: settingsNotesSortOrder.value,
    };
    const updated = await apiFetch("/settings", {
      method: "PATCH",
      body: JSON.stringify(payload),
    });
    applySettings(updated);
    await refreshTreePreserveMode();
    isDirty = false;
    saveBtn.textContent = "Save";
    saveBtn.disabled = false;
  } catch (err) {
    saveBtn.textContent = "Save";
    saveBtn.disabled = false;
    alert(err.message);
  }
}

async function saveSidebarWidth(width) {
  const clamped = Math.min(600, Math.max(220, Math.round(width)));
  try {
    await apiFetch("/settings", {
      method: "PATCH",
      body: JSON.stringify({ sidebarWidth: clamped }),
    });
    currentSettings.sidebarWidth = clamped;
  } catch (err) {
    console.warn("Unable to save sidebar width", err);
  }
}

function getTagColor(tag) {
  const value = String(tag || "");
  let hash = 0;
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash * 31 + value.charCodeAt(i)) % tagPalette.length;
  }
  return tagPalette[Math.abs(hash) % tagPalette.length];
}

function openTagGroup(tag) {
  if (!tag) {
    return;
  }
  const tagRoot = treeContainer.querySelector(".tree-node.tag-root");
  if (tagRoot) {
    tagRoot.classList.remove("collapsed");
  }
  const tagRow = treeContainer.querySelector(
    `.node-row[data-type="tag"][data-tag="${CSS.escape(tag)}"]`
  );
  if (!tagRow) {
    return;
  }
  const tagWrapper = tagRow.closest(".tree-node.tag-group");
  if (tagWrapper) {
    tagWrapper.classList.remove("collapsed");
  }
  tagRow.scrollIntoView({ block: "center" });
}

function applyHighlighting() {
  if (!window.hljs) {
    return;
  }
  preview.querySelectorAll("pre code").forEach((block) => {
    hljs.highlightElement(block);
  });
}

async function apiFetch(path, options = {}) {
  let response;
  try {
    response = await fetch(`${apiBase}${path}`, {
      headers: {
        "Content-Type": "application/json",
      },
      ...options,
    });
    setOfflineState(false);
  } catch (err) {
    setOfflineState(true);
    const queued = await queueRequestForSync(path, options);
    if (queued) {
      throw new Error("Queued for sync. Will retry when online.");
    }
    throw new Error("Unable to reach server");
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: "Request failed" }));
    throw new Error(error.error || "Request failed");
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

function buildTreeNode(node, depth = 0) {
  const wrapper = document.createElement("div");
  wrapper.className = `tree-node ${node.type}`;
  if (node.type === "folder" && depth === 0) {
    wrapper.classList.add("note-root");
  }
  const isDailyNode = isDailyPath(node.path);
  const isDailyRoot = node.path === dailyFolderName;

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = `${12 + depth * 12}px`;
  row.dataset.path = node.path;
  row.dataset.type = node.type;

  if (node.type === "folder") {
    const icon = document.createElement("span");
    icon.className = "folder-icon";
    row.appendChild(icon);
  } else if (node.type === "asset") {
    const icon = document.createElement("span");
    icon.className = "asset-icon";
    row.appendChild(icon);
  } else if (node.type === "pdf") {
    const icon = document.createElement("span");
    icon.className = "pdf-icon";
    row.appendChild(icon);
  } else if (node.type === "csv") {
    const icon = document.createElement("span");
    icon.className = "csv-icon";
    row.appendChild(icon);
  } else {
    const icon = document.createElement("span");
    icon.className = "note-icon";
    row.appendChild(icon);
  }

  const name = document.createElement("span");
  name.className = "node-name";
  const baseName = displayNodeName(node);
  const shouldShowCount = node.type === "folder";
  if (shouldShowCount) {
    const count = (node.children || []).length;
    name.textContent = formatCountLabel(baseName, count);
  } else {
    name.textContent = baseName;
  }
  row.appendChild(name);

  wrapper.appendChild(row);

  if (node.type === "folder" && depth === 0 && !isDailyRoot) {
    applyRootIconToRow(row, "notes");
  }
  if (node.type === "folder" && isDailyRoot) {
    applyRootIconToRow(row, "daily");
  }

  if (
    (node.type === "file" || node.type === "folder") &&
    !(node.type === "folder" && depth === 0)
  ) {
    row.setAttribute("draggable", "true");
    row.addEventListener("dragstart", (event) => handleDragStart(event, node));
    row.addEventListener("dragend", handleDragEnd);
  }
  if (node.type === "file" || node.type === "folder") {
    row.addEventListener("dragover", handleDragOver);
    row.addEventListener("dragleave", handleDragLeave);
    row.addEventListener("drop", handleDrop);
  }

  if (node.type === "folder") {
    if (depth > 0) {
      wrapper.classList.add("collapsed");
    }
    const children = document.createElement("div");
    children.className = "node-children";

    (node.children || []).forEach((child) => {
      children.appendChild(buildTreeNode(child, depth + 1));
    });

    wrapper.appendChild(children);

    row.addEventListener("click", () => {
      hideContextMenu();
      wrapper.classList.toggle("collapsed");
      const counts = countTreeItems(node);
      currentActivePath = node.path || "";
      setActiveNode(currentActivePath);
      const title = isDailyRoot ? "Daily" : depth === 0 ? "Notes" : `Folder: ${node.name}`;
      const action = isDailyRoot
        ? null
        : { label: "New", handler: () => createNote(node.path || "") };
      showSummary(title, [
        { label: "Folders", value: counts.folders },
        { label: "Notes", value: counts.notes },
        { label: "Assets", value: counts.assets },
        { label: "PDFs", value: counts.pdfs },
        { label: "CSVs", value: counts.csvs },
      ], action);
    });

    row.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      const isCollapsed = wrapper.classList.contains("collapsed");
      if (isDailyRoot) {
        showContextMenu(event.clientX, event.clientY, [
          ...rootIconMenuItems("daily", row),
          {
            label: "New Folder",
            action: () => createFolder(node.path),
          },
          {
            label: "New Note",
            action: () => createNote(node.path),
          },
          {
            label: "Edit Template",
            action: () => editTemplate(node.path),
          },
          {
            label: isCollapsed ? "Expand" : "Collapse",
            action: () => wrapper.classList.toggle("collapsed"),
          },
        ]);
        return;
      }
      if (depth === 0) {
        showContextMenu(event.clientX, event.clientY, [
          ...rootIconMenuItems("notes", row),
          {
            label: "New Folder",
            action: () => createFolder(node.path),
          },
          {
            label: "New Note",
            action: () => createNote(node.path),
          },
          {
            label: "Sort Notes...",
            action: () => showNotesSortMenu(event.clientX, event.clientY),
          },
          {
            label: "Edit Template",
            action: () => editTemplate(node.path),
          },
          {
            label: "Rename",
            action: () => renameFolder(node.path),
          },
          {
            label: "Delete",
            action: () => deleteFolder(node.path),
          },
          {
            label: isCollapsed ? "Expand" : "Collapse",
            action: () => wrapper.classList.toggle("collapsed"),
          },
        ]);
        return;
      }
      showContextMenu(event.clientX, event.clientY, [
        {
          label: "New Folder",
          action: () => createFolder(node.path),
        },
        {
          label: "New Note",
          action: () => createNote(node.path),
        },
        {
          label: "Sort Notes...",
          action: () => showNotesSortMenu(event.clientX, event.clientY),
        },
        {
          label: "Edit Template",
          action: () => editTemplate(node.path),
        },
        {
          label: "Rename",
          action: () => renameFolder(node.path),
        },
        {
          label: "Delete",
          action: () => deleteFolder(node.path),
        },
        {
          label: isCollapsed ? "Expand" : "Collapse",
          action: () => wrapper.classList.toggle("collapsed"),
        },
      ]);
    });
  } else {
    row.addEventListener("click", (event) => {
      event.stopPropagation();
      hideContextMenu();
      if (node.type === "asset") {
        openAsset(node.path);
      } else if (node.type === "pdf") {
        openPdf(node.path);
      } else if (node.type === "csv") {
        openCsv(node.path);
      } else {
        openNote(node.path);
      }
    });

    if (node.type === "file") {
      row.addEventListener("contextmenu", (event) => {
        event.preventDefault();
        const parentPath = node.path.split("/").slice(0, -1).join("/");
        if (isDailyNode) {
          showContextMenu(event.clientX, event.clientY, [
            {
              label: "Open Date",
              action: () => showDatePopover(),
            },
            {
              label: "New Note",
              action: () => createNote(parentPath),
            },
            {
              label: "Rename",
              action: () => renameNote(node.path),
            },
            {
              label: "Delete",
              action: () => deleteNote(node.path),
            },
          ]);
          return;
        }
        showContextMenu(event.clientX, event.clientY, [
          {
            label: "New Note",
            action: () => createNote(parentPath),
          },
          {
            label: "Rename",
            action: () => renameNote(node.path),
          },
          {
            label: "Delete",
            action: () => deleteNote(node.path),
          },
        ]);
      });
    }
  }

  return wrapper;
}

function getSortableDueDate(value) {
  if (!value) {
    return null;
  }
  const date = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(date.getTime())) {
    return null;
  }
  return date;
}

function compareTasksBySchedule(a, b) {
  const dateA = getSortableDueDate(a.dueDateISO);
  const dateB = getSortableDueDate(b.dueDateISO);
  if (dateA && !dateB) {
    return -1;
  }
  if (!dateA && dateB) {
    return 1;
  }
  if (dateA && dateB && dateA.getTime() !== dateB.getTime()) {
    return dateA - dateB;
  }
  const priorityA = Number(a.priority) || 0;
  const priorityB = Number(b.priority) || 0;
  if (priorityA && !priorityB) {
    return -1;
  }
  if (!priorityA && priorityB) {
    return 1;
  }
  if (priorityA && priorityB && priorityA !== priorityB) {
    return priorityA - priorityB;
  }
  return String(a.text || "").localeCompare(String(b.text || ""), undefined, { sensitivity: "base" });
}

function sortTasksForView(tasks, viewKey) {
  const sorted = [...tasks];
  sorted.sort((a, b) => {
    if (viewKey === "__tasks__") {
      const projectA = (a.project || "").toLowerCase();
      const projectB = (b.project || "").toLowerCase();
      const isEmptyA = projectA === "";
      const isEmptyB = projectB === "";
      if (isEmptyA !== isEmptyB) {
        return isEmptyA ? 1 : -1;
      }
      if (projectA !== projectB) {
        return projectA.localeCompare(projectB, undefined, { sensitivity: "base" });
      }
    }
    return compareTasksBySchedule(a, b);
  });
  return sorted;
}

function formatDateKey(value) {
  const year = value.getFullYear();
  const month = `${value.getMonth() + 1}`.padStart(2, "0");
  const day = `${value.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function isTaskDueTodayOrPast(task, todayKey) {
  if (!task || !task.dueDateISO) {
    return true;
  }
  return task.dueDateISO <= todayKey;
}

function hasTaskTag(task, tag) {
  if (!task || !tag) {
    return false;
  }
  const target = String(tag).toLowerCase();
  return (task.tags || []).some((value) => String(value || "").toLowerCase() === target);
}

function getTodayTasks(tasks) {
  const todayKey = formatDateKey(new Date());
  return (tasks || []).filter(
    (task) => !task.completed && !hasTaskTag(task, "someday") && isTaskDueTodayOrPast(task, todayKey)
  );
}

function getSomedayTasks(tasks) {
  return (tasks || []).filter((task) => !task.completed && hasTaskTag(task, "someday"));
}

function showToast(message) {
  if (!appToast) {
    alert(message);
    return;
  }
  appToast.textContent = message;
  appToast.classList.remove("hidden");
  if (toastTimer) {
    clearTimeout(toastTimer);
  }
  toastTimer = setTimeout(() => {
    appToast.classList.add("hidden");
    toastTimer = null;
  }, 3000);
}

const commandSearchPrefix = ">";

function getBuiltinCommands() {
  return [
    { label: "Open Inbox", keywords: ["inbox"], run: () => openNote(inboxNotePath) },
    { label: "Open Daily Note", keywords: ["daily", "today"], run: () => openDailyNote() },
    {
      label: "Open Today Tasks",
      keywords: ["today", "tasks"],
      run: () => {
        currentActivePath = "task-group:Today";
        restoreTaskSelection(currentActivePath, currentTasks || []);
      },
    },
    {
      label: "Open Tasks",
      keywords: ["tasks"],
      run: () => {
        currentActivePath = "__tasks__";
        restoreTaskSelection(currentActivePath, currentTasks || []);
      },
    },
    { label: "Open Journal", keywords: ["journal"], run: () => showJournal() },
    { label: "Open Scratch Pad", keywords: ["scratch"], run: () => openScratchDialog() },
    { label: "Open Settings", keywords: ["settings"], run: () => showSettings() },
    {
      label: "New Note",
      keywords: ["create", "note"],
      run: () => createNote(),
    },
    {
      label: "New Folder",
      keywords: ["create", "folder"],
      run: () => createFolder(),
    },
    {
      label: "Toggle Sidebar",
      keywords: ["sidebar", "collapse"],
      run: () => toggleSidebarCollapse(),
    },
    {
      label: "Collapse All",
      keywords: ["collapse", "tree"],
      run: () => collapseAllTreeNodes(),
    },
    {
      label: "Expand All Roots",
      keywords: ["expand", "tree"],
      run: () => expandAllRootNodes(),
    },
  ];
}

const externalCommandActions = {
  "open-note": (args) => {
    if (!args || !args.path) {
      throw new Error("Missing path for open-note");
    }
    openNote(args.path);
  },
  "open-inbox": () => openNote(inboxNotePath),
  "open-daily": () => openDailyNote(),
  "open-today": () => {
    currentActivePath = "task-group:Today";
    restoreTaskSelection(currentActivePath, currentTasks || []);
  },
  "open-tasks": () => {
    currentActivePath = "__tasks__";
    restoreTaskSelection(currentActivePath, currentTasks || []);
  },
  "open-journal": () => showJournal(),
  "open-scratch": () => openScratchDialog(),
  "open-settings": () => showSettings(),
  "new-note": (args) => createNote(args && args.path ? args.path : ""),
  "new-folder": (args) => createFolder(args && args.path ? args.path : ""),
  "toggle-sidebar": () => toggleSidebarCollapse(),
};

async function loadExternalCommands() {
  const path = (currentSettings.externalCommandsPath || "").trim();
  if (!path) {
    return [];
  }
  try {
    const response = await fetch(`${apiBase}/files?path=${encodeURIComponent(path)}`);
    if (!response.ok) {
      if (response.status === 404) {
        return [];
      }
      throw new Error("Unable to load external commands");
    }
    const text = await response.text();
    const parsed = JSON.parse(text);
    if (!Array.isArray(parsed)) {
      showToast("External commands file must be a JSON array.");
      return [];
    }
    const commands = [];
    parsed.forEach((entry) => {
      if (!entry || typeof entry !== "object") {
        showToast("Ignoring invalid external command entry.");
        return;
      }
      const label = String(entry.label || "").trim();
      const actionName = String(entry.action || "").trim();
      const keywords = Array.isArray(entry.keywords) ? entry.keywords.map((k) => String(k)) : [];
      if (!label || !actionName) {
        showToast("Ignoring external command missing label/action.");
        return;
      }
      const action = externalCommandActions[actionName];
      if (!action) {
        showToast(`Ignoring unknown external command action: ${actionName}`);
        return;
      }
      const args = entry.args && typeof entry.args === "object" ? entry.args : null;
      commands.push({
        label,
        keywords,
        description: entry.description ? String(entry.description) : actionName,
        run: () => action(args),
      });
    });
    return commands;
  } catch (err) {
    showToast(err.message);
    return [];
  }
}

function openCommandPalette() {
  if (!commandPalette || !commandInput || !commandResults) {
    return;
  }
  hideContextMenu();
  lastActiveElement = document.activeElement;
  commandPalette.classList.remove("hidden");
  commandInput.value = "";
  commandSelectedIndex = 0;
  commandMatches = [];
  commandItems = getBuiltinCommands();
  loadExternalCommands()
    .then((external) => {
      commandItems = commandItems.concat(external || []);
      updateCommandResults();
    })
    .catch(() => {
      updateCommandResults();
    });
  updateCommandResults();
  setTimeout(() => commandInput.focus(), 0);
}

function closeCommandPalette() {
  if (!commandPalette) {
    return;
  }
  commandPalette.classList.add("hidden");
  if (lastActiveElement && typeof lastActiveElement.focus === "function") {
    lastActiveElement.focus();
  }
  lastActiveElement = null;
}

function matchCommand(command, query) {
  const haystack = `${command.label} ${(command.keywords || []).join(" ")}`.toLowerCase();
  return haystack.includes(query);
}

function updateCommandResults() {
  if (!commandInput || !commandResults) {
    return;
  }
  const raw = commandInput.value || "";
  const trimmed = raw.trim();
  if (trimmed.startsWith(commandSearchPrefix)) {
    const term = trimmed.slice(commandSearchPrefix.length).trim();
    runCommandSearch(term);
    return;
  }
  const query = trimmed.toLowerCase();
  commandMatches = query
    ? commandItems.filter((cmd) => matchCommand(cmd, query))
    : commandItems.slice(0, 8);
  commandSelectedIndex = 0;
  renderCommandResults(commandMatches);
}

async function runCommandSearch(term) {
  const requestId = ++commandSearchRequestId;
  if (!term) {
    commandMatches = [];
    renderCommandResults(commandMatches, "Type to search notes...");
    return;
  }
  try {
    const results = await apiFetch(`/search?query=${encodeURIComponent(term)}`);
    if (requestId !== commandSearchRequestId) {
      return;
    }
    commandMatches = (results || []).map((item) => ({
      label: item.path,
      description: "Open note",
      run: () => openNote(item.path),
    }));
    commandSelectedIndex = 0;
    renderCommandResults(commandMatches, "No results.");
  } catch (err) {
    if (requestId === commandSearchRequestId) {
      showToast(err.message);
    }
  }
}

function renderCommandResults(items, emptyLabel = "No matches.") {
  if (!commandResults) {
    return;
  }
  commandResults.innerHTML = "";
  if (!items || items.length === 0) {
    const empty = document.createElement("div");
    empty.className = "command-item";
    empty.textContent = emptyLabel;
    commandResults.appendChild(empty);
    return;
  }
  items.forEach((item, index) => {
    const row = document.createElement("button");
    row.type = "button";
    row.className = "command-item";
    if (index === commandSelectedIndex) {
      row.classList.add("active");
    }
    const title = document.createElement("strong");
    title.textContent = item.label;
    row.appendChild(title);
    if (item.description) {
      const desc = document.createElement("span");
      desc.textContent = item.description;
      row.appendChild(desc);
    }
    row.addEventListener("click", () => {
      runCommandItem(index);
    });
    commandResults.appendChild(row);
  });
}

function runCommandItem(index) {
  const item = commandMatches[index];
  if (!item || typeof item.run !== "function") {
    return;
  }
  closeCommandPalette();
  item.run();
}

function countTreeItems(node) {
  const counts = {
    folders: 0,
    notes: 0,
    assets: 0,
    pdfs: 0,
    csvs: 0,
  };
  if (!node || !node.children) {
    return counts;
  }
  const stack = [...node.children];
  while (stack.length > 0) {
    const current = stack.pop();
    if (!current) {
      continue;
    }
    switch (current.type) {
      case "folder":
        counts.folders += 1;
        if (current.children && current.children.length > 0) {
          stack.push(...current.children);
        }
        break;
      case "file":
        if (!(current.name || "").toLowerCase().endsWith(".template")) {
          counts.notes += 1;
        }
        break;
      case "asset":
        counts.assets += 1;
        break;
      case "pdf":
        counts.pdfs += 1;
        break;
      case "csv":
        counts.csvs += 1;
        break;
      default:
        break;
    }
  }
  return counts;
}

function buildTaskGroup(name, tasks, depth) {
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node folder task-group";

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = `${12 + depth * 12}px`;
  row.dataset.path = `task-group:${name}`;
  row.dataset.type = "task-group";

  const caret = document.createElement("span");
  caret.className = "folder-icon";
  row.appendChild(caret);

  const label = document.createElement("span");
  label.className = "node-name";
  label.textContent = formatCountLabel(formatProjectLabel(name), (tasks || []).length);
  row.appendChild(label);

  wrapper.appendChild(row);

  row.addEventListener("click", () => {
    hideContextMenu();
    const total = (tasks || []).length;
    currentActivePath = `task-group:${name}`;
    setActiveNode(currentActivePath);
    let title = `Project: ${name}`;
    if (name === "No Project") {
      title = "No Project";
    } else if (name === "Completed") {
      title = "Completed Tasks";
    } else if (name === "Today") {
      title = "Today";
    } else if (name === "Someday") {
      title = "Someday";
    } else if (name === "All") {
      title = "All Tasks";
    }
    showTaskList(title, sortTasksForView(tasks || [], name), taskGroupSummary(tasks || [], name));
  });

  row.addEventListener("contextmenu", (event) => {
    event.preventDefault();
    const items = [];
    if (name === "Completed") {
      items.unshift({
        label: "Archive Completed",
        action: () => archiveCompletedTasks(),
      });
    }
    showContextMenu(event.clientX, event.clientY, items);
  });

  return wrapper;
}

function formatProjectLabel(name) {
  if (!name) {
    return "";
  }
  if (name === "No Project" || name === "Completed" || name === "Today" || name === "Someday" || name === "All") {
    return name;
  }
  return String(name)
    .split(/[\s-]+/)
    .filter(Boolean)
    .map((word) => word[0].toUpperCase() + word.slice(1).toLowerCase())
    .join(" ");
}

function buildTasksRoot(tasks) {
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node folder task-root";
  wrapper.classList.add("collapsed");

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = "12px";
  row.dataset.path = "__tasks__";
  row.dataset.type = "task-root";

  const caret = document.createElement("span");
  caret.className = "folder-icon";
  row.appendChild(caret);

  const activeTasks = (tasks || []).filter((task) => !task.completed);
  const completedTasks = (tasks || []).filter((task) => task.completed);
  const todayTasks = getTodayTasks(tasks || []);
  const somedayTasks = getSomedayTasks(tasks || []);

  const name = document.createElement("span");
  name.className = "node-name";
  name.textContent = formatCountLabel("Tasks", activeTasks.length);
  row.appendChild(name);

  wrapper.appendChild(row);

  applyRootIconToRow(row, "tasks");

  const children = document.createElement("div");
  children.className = "node-children";

  const projectMap = new Map();
  const noProject = [];

  activeTasks.forEach((task) => {
    const projectName = (task.project || "").trim();
    if (!projectName) {
      noProject.push(task);
      return;
    }
    if (!projectMap.has(projectName)) {
      projectMap.set(projectName, []);
    }
    projectMap.get(projectName).push(task);
  });

  const projectNames = Array.from(projectMap.keys()).sort((a, b) =>
    a.localeCompare(b, undefined, { sensitivity: "base" })
  );

  children.appendChild(buildTaskGroup("Today", todayTasks, 1));
  children.appendChild(buildTaskGroup("Someday", somedayTasks, 1));

  projectNames.forEach((project) => {
    children.appendChild(buildTaskGroup(project, projectMap.get(project) || [], 1));
  });

  if (noProject.length > 0) {
    children.appendChild(buildTaskGroup("No Project", noProject, 1));
  }
  if (completedTasks.length > 0) {
    children.appendChild(buildTaskGroup("Completed", completedTasks, 1));
  }
  children.appendChild(buildTaskGroup("All", activeTasks, 1));

  wrapper.appendChild(children);

  row.addEventListener("click", () => {
    hideContextMenu();
    wrapper.classList.toggle("collapsed");
    const projectSet = new Set();
    activeTasks.forEach((task) => {
      const projectName = (task.project || "").trim();
      if (projectName) {
        projectSet.add(projectName);
      }
    });
    currentActivePath = "__tasks__";
    setActiveNode(currentActivePath);
    const summaryText = `${activeTasks.length} active task${activeTasks.length === 1 ? "" : "s"} across ${projectSet.size} project${projectSet.size === 1 ? "" : "s"}`;
    showTaskList("Tasks", sortTasksForView(activeTasks, "__tasks__"), summaryText);
  });

  row.addEventListener("contextmenu", (event) => {
    event.preventDefault();
    const isCollapsed = wrapper.classList.contains("collapsed");
    showContextMenu(event.clientX, event.clientY, [
      ...rootIconMenuItems("tasks", row),
      {
        label: isCollapsed ? "Expand" : "Collapse",
        action: () => wrapper.classList.toggle("collapsed"),
      },
    ]);
  });

  return wrapper;
}

function splitTasksByProject(tasks) {
  const activeTasks = (tasks || []).filter((task) => !task.completed);
  const completedTasks = (tasks || []).filter((task) => task.completed);
  const projectMap = new Map();
  const noProject = [];
  const todayTasks = getTodayTasks(tasks || []);
  const somedayTasks = getSomedayTasks(tasks || []);

  activeTasks.forEach((task) => {
    const projectName = (task.project || "").trim();
    if (!projectName) {
      noProject.push(task);
      return;
    }
    if (!projectMap.has(projectName)) {
      projectMap.set(projectName, []);
    }
    projectMap.get(projectName).push(task);
  });

  return { activeTasks, completedTasks, projectMap, noProject, todayTasks, somedayTasks };
}

function taskGroupSummary(tasks, label) {
  const count = (tasks || []).length;
  const suffix = count === 1 ? "" : "s";
  if (label === "Completed") {
    return `${count} completed task${suffix}`;
  }
  if (label === "Today") {
    return `${count} task${suffix} due today or earlier`;
  }
  if (label === "Someday") {
    return `${count} someday task${suffix}`;
  }
  if (label === "All") {
    return `${count} open task${suffix}`;
  }
  return `${count} task${suffix}`;
}

function restoreTaskSelection(previousActivePath, tasks) {
  if (!showTasksRoot || !previousActivePath) {
    return false;
  }
  if (previousActivePath === "__tasks__") {
    const { activeTasks, projectMap } = splitTasksByProject(tasks || []);
    currentActivePath = "__tasks__";
    setActiveNode(currentActivePath);
    const summaryText = `${activeTasks.length} active task${activeTasks.length === 1 ? "" : "s"} across ${projectMap.size} project${projectMap.size === 1 ? "" : "s"}`;
    showTaskList("Tasks", sortTasksForView(activeTasks, "__tasks__"), summaryText);
    return true;
  }
  if (previousActivePath.startsWith("task-group:")) {
    const name = previousActivePath.replace("task-group:", "");
    const { activeTasks, completedTasks, projectMap, noProject, todayTasks, somedayTasks } = splitTasksByProject(
      tasks || []
    );
    let groupTasks = [];
    let title = "";
    if (name === "Completed") {
      groupTasks = completedTasks;
      title = "Completed Tasks";
    } else if (name === "Today") {
      groupTasks = todayTasks;
      title = "Today";
    } else if (name === "Someday") {
      groupTasks = somedayTasks;
      title = "Someday";
    } else if (name === "All") {
      groupTasks = activeTasks;
      title = "All Tasks";
    } else if (name === "No Project") {
      groupTasks = noProject;
      title = "No Project";
    } else if (projectMap.has(name)) {
      groupTasks = projectMap.get(name);
      title = `Project: ${name}`;
    } else {
      title = `Project: ${name}`;
    }
    currentActivePath = `task-group:${name}`;
    setActiveNode(currentActivePath);
    showTaskList(title, sortTasksForView(groupTasks, name), taskGroupSummary(groupTasks, name));
    return true;
  }
  return false;
}

function renderTaskList(title, tasks, summaryText) {
  taskList.innerHTML = "";

  const header = document.createElement("div");
  header.className = "task-list-header";

  const heading = document.createElement("h2");
  heading.className = "task-list-title";
  heading.textContent = title;
  header.appendChild(heading);

  if (summaryText) {
    const summary = document.createElement("div");
    summary.className = "task-list-summary";
    summary.textContent = summaryText;
    header.appendChild(summary);
  }

  taskList.appendChild(header);

  const list = document.createElement("div");
  list.className = "task-items";

  if (!tasks || tasks.length === 0) {
    const empty = document.createElement("div");
    empty.className = "search-empty";
    empty.textContent = "No tasks to show.";
    list.appendChild(empty);
  } else {
    tasks.forEach((task) => {
      list.appendChild(buildTaskListItem(task));
    });
  }

  taskList.appendChild(list);
  taskList.scrollTop = 0;
  requestAnimationFrame(() => updateTaskMetaOverflow());
}

function buildTaskListItem(task) {
  const item = document.createElement("div");
  item.className = "task-item";
  if (task.completed) {
    item.classList.add("completed");
  }

  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.className = "task-checkbox";
  checkbox.checked = !!task.completed;
  checkbox.addEventListener("click", (event) => event.stopPropagation());
  checkbox.addEventListener("change", async () => {
    checkbox.disabled = true;
    await toggleTaskCompletion(task, checkbox.checked);
    checkbox.disabled = false;
  });
  item.appendChild(checkbox);

  const main = document.createElement("div");
  main.className = "task-main";
  main.addEventListener("click", () => {
    openNoteAtLine(task.path, task.lineNumber);
  });

  const text = document.createElement("div");
  text.className = "task-text";
  text.textContent = task.text || "(untitled)";
  main.appendChild(text);

  const meta = document.createElement("div");
  meta.className = "task-meta";

  if (task.project) {
    meta.appendChild(buildTaskChip(`+${task.project}`));
  }
  if (task.dueDate) {
    const label = task.dueDateISO ? `>${task.dueDateISO}` : `>${task.dueDate}`;
    const invalid = task.dueDate && !task.dueDateISO;
    meta.appendChild(buildTaskChip(label, invalid));
  }
  if (task.priority) {
    meta.appendChild(buildTaskChip(`^${task.priority}`));
  }
  (task.tags || []).forEach((tag) => meta.appendChild(buildTaskChip(`#${tag}`)));
  (task.mentions || []).forEach((mention) => meta.appendChild(buildTaskChip(`@${mention}`)));

  if (meta.children.length > 0) {
    main.appendChild(meta);
  }

  const location = document.createElement("div");
  location.className = "task-location";
  location.textContent = `${task.path} : ${task.lineNumber}`;
  main.appendChild(location);

  item.appendChild(main);
  return item;
}

function buildTaskChip(text, invalid = false) {
  const chip = document.createElement("span");
  chip.className = "task-chip";
  if (invalid) {
    chip.classList.add("invalid");
    chip.title = "Unrecognized due date format";
  }
  chip.textContent = text;
  return chip;
}

function updateTaskMetaOverflow() {
  const items = document.querySelectorAll(".task-item");
  items.forEach((item) => {
    const meta = item.querySelector(".task-meta");
    if (!meta) {
      return;
    }
    const chips = Array.from(meta.querySelectorAll(".task-chip"));
    if (chips.length === 0) {
      return;
    }
    let overflow = meta.querySelector(".task-meta-overflow");
    if (!overflow) {
      overflow = document.createElement("span");
      overflow.className = "task-meta-overflow";
      meta.appendChild(overflow);
    }
    chips.forEach((chip) => {
      chip.style.display = "";
    });
    overflow.textContent = "";
    overflow.style.display = "none";

    const fits = () => meta.scrollWidth <= meta.clientWidth;
    if (!fits()) {
      overflow.style.display = "inline-flex";
      overflow.textContent = "+1";
      let hidden = 0;
      for (let i = chips.length - 1; i >= 0 && !fits(); i -= 1) {
        chips[i].style.display = "none";
        hidden += 1;
        overflow.textContent = `+${hidden}`;
      }
      if (hidden === 0) {
        overflow.style.display = "none";
      }
    }
  });
}

async function toggleTaskCompletion(task, completed) {
  try {
    await apiFetch("/tasks/toggle", {
      method: "PATCH",
      body: JSON.stringify({
        path: task.path,
        lineNumber: task.lineNumber,
        lineHash: task.lineHash,
        completed,
      }),
    });
    await loadTree();
  } catch (err) {
    alert(err.message);
  }
}

async function archiveCompletedTasks() {
  try {
    await apiFetch("/tasks/archive", { method: "PATCH" });
    await loadTree();
  } catch (err) {
    alert(err.message);
  }
}

function buildTagRoot(tags) {
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node folder tag-root collapsed";

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = "12px";
  row.dataset.path = "__tags__";
  row.dataset.type = "tag-root";

  const caret = document.createElement("span");
  caret.className = "folder-icon";
  row.appendChild(caret);

  const name = document.createElement("span");
  name.className = "node-name";
  name.textContent = formatCountLabel("Tags", (tags || []).length);
  row.appendChild(name);

  wrapper.appendChild(row);

  applyRootIconToRow(row, "tags");

  const children = document.createElement("div");
  children.className = "node-children";
  tags.forEach((group) => {
    children.appendChild(buildTagGroup(group, 1));
  });
  wrapper.appendChild(children);

  row.addEventListener("click", () => {
    hideContextMenu();
    wrapper.classList.toggle("collapsed");
    const totalTags = (tags || []).length;
    const noteSet = new Set();
    let totalEntries = 0;
    (tags || []).forEach((group) => {
      (group.notes || []).forEach((note) => {
        if (note && note.path) {
          noteSet.add(note.path);
        }
        totalEntries += 1;
      });
    });
    currentActivePath = "__tags__";
    setActiveNode(currentActivePath);
    showSummary("Tags", [
      { label: "Tags", value: totalTags },
      { label: "Tagged Notes", value: noteSet.size },
      { label: "Tag Entries", value: totalEntries },
    ]);
  });

  row.addEventListener("contextmenu", (event) => {
    event.preventDefault();
    const isCollapsed = wrapper.classList.contains("collapsed");
    showContextMenu(event.clientX, event.clientY, [
      ...rootIconMenuItems("tags", row),
      {
        label: isCollapsed ? "Expand" : "Collapse",
        action: () => wrapper.classList.toggle("collapsed"),
      },
    ]);
  });

  return wrapper;
}

function buildTagGroup(group, depth) {
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node folder tag-group";
  wrapper.classList.add("collapsed");

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = "10px";
  row.dataset.path = `tag:${group.tag}`;
  row.dataset.type = "tag";
  row.dataset.tag = group.tag;

  const caret = document.createElement("span");
  caret.className = "folder-icon";
  row.appendChild(caret);

  const name = document.createElement("span");
  name.className = "node-name tag-label";
  name.textContent = formatCountLabel(`#${group.tag}`, (group.notes || []).length);
  name.style.backgroundColor = getTagColor(group.tag);
  row.appendChild(name);

  wrapper.appendChild(row);

  const children = document.createElement("div");
  children.className = "node-children";
  (group.notes || []).forEach((note) => {
    children.appendChild(buildTagNote(note, depth + 1));
  });
  wrapper.appendChild(children);

  row.addEventListener("click", () => {
    hideContextMenu();
    wrapper.classList.toggle("collapsed");
    const totalTags = (tags || []).length;
    const noteSet = new Set();
    let totalEntries = 0;
    (tags || []).forEach((group) => {
      (group.notes || []).forEach((note) => {
        if (note && note.path) {
          noteSet.add(note.path);
        }
        totalEntries += 1;
      });
    });
    currentActivePath = "__tags__";
    setActiveNode(currentActivePath);
    showSummary("Tags", [
      { label: "Tags", value: totalTags },
      { label: "Tagged Notes", value: noteSet.size },
      { label: "Tag Entries", value: totalEntries },
    ]);
  });

  row.addEventListener("contextmenu", (event) => {
    event.preventDefault();
    const isCollapsed = wrapper.classList.contains("collapsed");
    showContextMenu(event.clientX, event.clientY, [
      {
        label: isCollapsed ? "Expand" : "Collapse",
        action: () => wrapper.classList.toggle("collapsed"),
      },
    ]);
  });

  return wrapper;
}

function buildTagNote(note, depth) {
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node file";

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = `${12 + depth * 12}px`;
  row.dataset.path = note.path;
  row.dataset.type = "file";

  const spacer = document.createElement("span");
  spacer.className = "note-icon";
  row.appendChild(spacer);

  const name = document.createElement("span");
  name.className = "node-name";
  name.textContent = displayNodeName({ type: "file", name: note.name });
  row.appendChild(name);

  wrapper.appendChild(row);

  row.addEventListener("click", (event) => {
    event.stopPropagation();
    hideContextMenu();
    openNote(note.path);
  });

  return wrapper;
}

function buildInboxNode() {
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node inbox-root";

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = "12px";
  row.dataset.path = inboxNotePath;
  row.dataset.type = "inbox";

  const icon = document.createElement("span");
  icon.className = "folder-icon";
  icon.addEventListener("click", (event) => {
    event.stopPropagation();
    wrapper.classList.toggle("collapsed");
  });
  row.appendChild(icon);

  const name = document.createElement("span");
  name.className = "node-name";
  name.textContent = "Inbox";
  row.appendChild(name);

  wrapper.appendChild(row);

  applyRootIconToRow(row, "inbox");

  row.addEventListener("click", (event) => {
    event.stopPropagation();
    hideContextMenu();
    openNote(inboxNotePath);
  });

  row.addEventListener("contextmenu", (event) => {
    event.preventDefault();
    showContextMenu(event.clientX, event.clientY, [
      ...rootIconMenuItems("inbox", row),
    ]);
  });

  return wrapper;
}

function buildJournalNode(summary = {}) {
  const archives = Array.isArray(summary.archives) ? summary.archives : [];
  const totalCount = Number.isFinite(summary.totalCount) ? summary.totalCount : archives.length;
  const wrapper = document.createElement("div");
  wrapper.className = "tree-node folder journal-root";
  wrapper.classList.add("collapsed");

  const row = document.createElement("div");
  row.className = "node-row";
  row.style.paddingLeft = "12px";
  row.dataset.path = journalRootPath;
  row.dataset.type = "journal-root";

  const icon = document.createElement("span");
  icon.className = "folder-icon";
  icon.addEventListener("click", (event) => {
    event.stopPropagation();
    wrapper.classList.toggle("collapsed");
    currentActivePath = journalRootPath;
    setActiveNode(currentActivePath);
    showJournal();
  });
  row.appendChild(icon);

  const name = document.createElement("span");
  name.className = "node-name";
  name.textContent = formatCountLabel("Journal", totalCount);
  row.appendChild(name);

  wrapper.appendChild(row);

  applyRootIconToRow(row, "journal");

  const children = document.createElement("div");
  children.className = "node-children";
  archives.forEach((archive) => {
    const child = document.createElement("div");
    child.className = "tree-node";
    const childRow = document.createElement("div");
    childRow.className = "node-row";
    childRow.style.paddingLeft = "28px";
    childRow.dataset.path = `${journalRootPath}:${archive.date}`;
    childRow.dataset.type = "journal-archive";

    const childIcon = document.createElement("span");
    childIcon.className = "file-icon";
    childRow.appendChild(childIcon);

    const childName = document.createElement("span");
    childName.className = "node-name";
    childName.textContent = `Archive ${archive.date}`;
    childRow.appendChild(childName);

    child.appendChild(childRow);
    children.appendChild(child);

    childRow.addEventListener("click", (event) => {
      event.stopPropagation();
      hideContextMenu();
      currentActivePath = `${journalRootPath}:${archive.date}`;
      setActiveNode(currentActivePath);
      showJournalArchive(archive.date);
    });
  });

  wrapper.appendChild(children);

  row.addEventListener("click", (event) => {
    event.stopPropagation();
    hideContextMenu();
    wrapper.classList.toggle("collapsed");
    currentActivePath = journalRootPath;
    setActiveNode(currentActivePath);
    showJournal();
  });

  row.addEventListener("contextmenu", (event) => {
    event.preventDefault();
    const isCollapsed = wrapper.classList.contains("collapsed");
    showContextMenu(event.clientX, event.clientY, [
      ...rootIconMenuItems("journal", row),
      {
        label: isCollapsed ? "Expand" : "Collapse",
        action: () => wrapper.classList.toggle("collapsed"),
      },
    ]);
  });

  return wrapper;
}

function updateJournalRootCount() {
  if (!treeContainer) {
    return;
  }
  const row = treeContainer.querySelector('.node-row[data-type="journal-root"]');
  if (!row) {
    return;
  }
  const name = row.querySelector(".node-name");
  if (!name) {
    return;
  }
  const count = Number.isFinite(journalSummary.totalCount)
    ? journalSummary.totalCount
    : (journalSummary.archives || []).length;
  name.textContent = formatCountLabel("Journal", count);
}

async function refreshJournalSummary() {
  const response = await apiFetch("/journal/archives");
  journalSummary = {
    archives: response.archives || [],
    totalCount: Number.isFinite(response.totalCount)
      ? response.totalCount
      : (response.archives || []).length,
  };
  updateJournalRootCount();
}

function stripInboxNode(tree) {
  if (!tree || !Array.isArray(tree.children)) {
    return tree;
  }
  const filtered = tree.children.filter(
    (child) => !(child.type === "file" && child.path === inboxNotePath)
  );
  if (filtered.length === tree.children.length) {
    return tree;
  }
  return { ...tree, children: filtered };
}

function stripJournalFolder(tree) {
  if (!tree || !Array.isArray(tree.children)) {
    return tree;
  }
  const filtered = tree.children.filter((child) => {
    if (child.type !== "folder" || !child.path) {
      return true;
    }
    return child.path.toLowerCase() !== journalFolderName;
  });
  if (filtered.length === tree.children.length) {
    return tree;
  }
  return { ...tree, children: filtered };
}

function stripScratchNode(tree) {
  if (!tree || !Array.isArray(tree.children)) {
    return tree;
  }
  const filtered = tree.children.filter(
    (child) => !(child.type === "file" && child.path === scratchNotePath)
  );
  if (filtered.length === tree.children.length) {
    return tree;
  }
  return { ...tree, children: filtered };
}

function extractDailyNode(tree) {
  if (!tree || !Array.isArray(tree.children)) {
    return { dailyNode: null, tree };
  }
  const index = tree.children.findIndex(
    (child) => child.type === "folder" && child.path === dailyFolderName
  );
  if (index === -1) {
    return {
      dailyNode: { name: "Daily", path: dailyFolderName, type: "folder", children: [] },
      tree,
    };
  }
  const dailyNode = tree.children[index];
  const children = tree.children.filter((_, childIndex) => childIndex !== index);
  return { dailyNode, tree: { ...tree, children } };
}

function compareDailyNodes(a, b) {
  const aIsFolder = a.type === "folder";
  const bIsFolder = b.type === "folder";
  if (aIsFolder !== bIsFolder) {
    return aIsFolder ? 1 : -1;
  }
  const nameA = String(a.name || "").toLowerCase();
  const nameB = String(b.name || "").toLowerCase();
  if (nameA === nameB) {
    return 0;
  }
  return nameA < nameB ? 1 : -1;
}

function sortDailyNodeTree(node) {
  if (!node || !Array.isArray(node.children)) {
    return;
  }
  node.children.sort(compareDailyNodes);
  node.children.forEach((child) => {
    if (child.type === "folder") {
      sortDailyNodeTree(child);
    }
  });
}

function buildDailyNode(node) {
  const dailyNode = node || { name: "Daily", path: dailyFolderName, type: "folder", children: [] };
  sortDailyNodeTree(dailyNode);
  const wrapper = buildTreeNode(dailyNode, 0);
  wrapper.classList.add("daily-root");
  wrapper.classList.remove("note-root");
  wrapper.classList.add("collapsed");
  return wrapper;
}

function renderTree(tree, tags, tasks) {
  treeContainer.innerHTML = "";
  if (tree) {
    treeContainer.appendChild(buildInboxNode());
    const { dailyNode, tree: notesTree } = extractDailyNode(
      stripJournalFolder(stripScratchNode(stripInboxNode(tree)))
    );
    treeContainer.appendChild(buildDailyNode(dailyNode));
    const rootNode = buildTreeNode(notesTree, 0);
    rootNode.classList.add("collapsed");
    treeContainer.appendChild(rootNode);
  }
  if (showTasksRoot && tasks) {
    const tasksRoot = buildTasksRoot(tasks);
    treeContainer.appendChild(tasksRoot);
  }
  treeContainer.appendChild(buildJournalNode(journalSummary));
  if (tags) {
    const tagRoot = buildTagRoot(tags);
    treeContainer.appendChild(tagRoot);
  }
  setActiveNode(currentActivePath);
}

function getExpandedTreePaths() {
  const paths = new Set();
  treeContainer.querySelectorAll(".tree-node.folder").forEach((node) => {
    if (node.classList.contains("collapsed")) {
      return;
    }
    const row = node.querySelector(":scope > .node-row");
    if (row && row.dataset.path) {
      paths.add(row.dataset.path);
    }
  });
  return paths;
}

function toggleAllTreeNodes() {
  if (!treeContainer) {
    return;
  }
  const folders = Array.from(treeContainer.querySelectorAll(".tree-node.folder"));
  const hasExpanded = folders.some((node) => !node.classList.contains("collapsed"));
  if (hasExpanded) {
    collapseAllTreeNodes();
    return;
  }
  expandAllRootNodes();
}

function collapseAllTreeNodes() {
  if (!treeContainer) {
    return;
  }
  treeContainer.querySelectorAll(".tree-node.folder").forEach((node) => {
    node.classList.add("collapsed");
  });
}

function expandAllRootNodes() {
  if (!treeContainer) {
    return;
  }
  const rootSelectors = [
    ".tree-node.note-root",
    ".tree-node.daily-root",
    ".tree-node.task-root",
    ".tree-node.tag-root",
    ".tree-node.journal-root",
  ];
  treeContainer.querySelectorAll(rootSelectors.join(",")).forEach((node) => {
    node.classList.remove("collapsed");
  });
}

function restoreExpandedTreePaths(paths) {
  if (!paths || paths.size === 0) {
    return;
  }
  paths.forEach((path) => {
    const row = treeContainer.querySelector(`.node-row[data-path="${CSS.escape(path)}"]`);
    if (!row) {
      return;
    }
    const wrapper = row.closest(".tree-node.folder");
    if (wrapper) {
      wrapper.classList.remove("collapsed");
    }
  });
}

function refreshTasksTree() {
  renderTree(currentTree, currentTags, currentTasks);
  if (currentMode === "tasks") {
    restoreTaskSelection(currentActivePath, currentTasks);
  }
}

function refreshTasksAndTagsPreserveView(options = {}) {
  const modeSnapshot = currentMode;
  const activePathSnapshot = currentActivePath;
  const expandedPaths = getExpandedTreePaths();
  const { suppressNotice = false } = options;
  return Promise.all([apiFetch("/tasks"), apiFetch("/tags")])
    .then(([tasksResponse, tags]) => {
      if (tasksResponse.notice && !suppressNotice) {
        alert(tasksResponse.notice);
      }
      currentTags = tags || [];
      currentTasks = tasksResponse.tasks || [];
      renderTree(currentTree, currentTags, currentTasks);
      restoreExpandedTreePaths(expandedPaths);
      if (modeSnapshot === "tasks") {
        restoreTaskSelection(activePathSnapshot, currentTasks);
      } else if (activePathSnapshot) {
        setActiveNode(activePathSnapshot);
      }
    })
    .catch((err) => {
      console.warn("Unable to refresh tasks and tags", err);
    });
}

async function refreshTreePreserveMode() {
  const modeSnapshot = currentMode;
  const notePathSnapshot = currentNotePath;
  const activePathSnapshot = currentActivePath;
  const expandedPaths = getExpandedTreePaths();
  await loadTree();
  if (modeSnapshot === "note" && notePathSnapshot) {
    await openNote(notePathSnapshot);
  } else if (modeSnapshot === "settings") {
    showSettings();
  } else if (modeSnapshot === "journal") {
    showJournal();
  } else if (modeSnapshot === "journal-archive" && activePathSnapshot.startsWith(`${journalRootPath}:`)) {
    const date = activePathSnapshot.slice(`${journalRootPath}:`.length);
    if (date) {
      showJournalArchive(date);
    }
  } else if (activePathSnapshot) {
    setActiveNode(activePathSnapshot);
  }
  restoreExpandedTreePaths(expandedPaths);
  if (activePathSnapshot) {
    setActiveNode(activePathSnapshot);
  }
}

async function refreshTasksForNote(path, options = {}) {
  if (!path) {
    return;
  }
  try {
    const response = await apiFetch(`/tasks/for-note?path=${encodeURIComponent(path)}`);
    if (response.notice && !options.suppressNotice) {
      alert(response.notice);
    }
    currentTasks = (currentTasks || []).filter((task) => task.path !== path);
    currentTasks = currentTasks.concat(response.tasks || []);
    refreshTasksTree();
  } catch (err) {
    console.warn("Unable to refresh tasks for note", err);
  }
}

function setActiveNode(path) {
  const rows = treeContainer.querySelectorAll(".node-row");
  rows.forEach((row) => {
    const isSelectable =
      row.dataset.type === "file" ||
      row.dataset.type === "inbox" ||
      row.dataset.type === "journal-root" ||
      row.dataset.type === "journal-archive" ||
      row.dataset.type === "asset" ||
      row.dataset.type === "pdf" ||
      row.dataset.type === "csv" ||
      row.dataset.type === "task-root" ||
      row.dataset.type === "tag-root" ||
      row.dataset.type === "task-group" ||
      row.dataset.type === "folder";
    row.classList.toggle("active", isSelectable && row.dataset.path === path);
  });
}

function expandToPath(path) {
  if (!path) {
    return;
  }
  const parts = path.split("/").filter(Boolean);
  let current = "";
  parts.slice(0, -1).forEach((part) => {
    current = current ? `${current}/${part}` : part;
    const row = treeContainer.querySelector(`.node-row[data-path="${CSS.escape(current)}"]`);
    if (row) {
      const wrapper = row.closest(".tree-node.folder");
      if (wrapper) {
        wrapper.classList.remove("collapsed");
      }
    }
  });
  const targetRow = treeContainer.querySelector(`.node-row[data-path="${CSS.escape(path)}"]`);
  if (targetRow && targetRow.dataset.type === "folder") {
    const wrapper = targetRow.closest(".tree-node.folder");
    if (wrapper) {
      wrapper.classList.remove("collapsed");
    }
  }
}

function focusTreePath(path) {
  if (!path) {
    return;
  }
  expandToPath(path);
  setActiveNode(path);
  const row = treeContainer.querySelector(`.node-row[data-path="${CSS.escape(path)}"]`);
  if (row) {
    row.scrollIntoView({ block: "center" });
  }
}

async function loadTree(path = "") {
  try {
    const previousActivePath = currentActivePath;
    const expandedPaths = getExpandedTreePaths();
    const query = path ? `?path=${encodeURIComponent(path)}` : "";
    const [tree, tags, tasksResponse, settingsResponse, archiveResponse] = await Promise.all([
      apiFetch(`/tree${query}`),
      apiFetch("/tags"),
      apiFetch("/tasks"),
      apiFetch("/settings"),
      apiFetch("/journal/archives"),
    ]);
    setOfflineState(false);
    if (tasksResponse.notice) {
      alert(tasksResponse.notice);
    }
    if (settingsResponse.notice && !settingsLoaded) {
      alert(settingsResponse.notice);
    }
    settingsLoaded = true;
    applySettings(settingsResponse.settings || {});
    currentTree = tree;
    currentTags = tags;
    currentTasks = tasksResponse.tasks || [];
    journalSummary = {
      archives: archiveResponse.archives || [],
      totalCount: Number.isFinite(archiveResponse.totalCount)
        ? archiveResponse.totalCount
        : (archiveResponse.archives || []).length,
    };
    renderTree(currentTree, currentTags, tasksResponse.tasks || []);
    if (restoreTaskSelection(previousActivePath, currentTasks)) {
      restoreExpandedTreePaths(expandedPaths);
      return;
    }
    if (previousActivePath === journalRootPath) {
      currentActivePath = journalRootPath;
      setActiveNode(currentActivePath);
      restoreExpandedTreePaths(expandedPaths);
      showJournal();
      return;
    }
    if (previousActivePath && previousActivePath.startsWith(`${journalRootPath}:`)) {
      const date = previousActivePath.slice(`${journalRootPath}:`.length);
      if (date) {
        currentActivePath = previousActivePath;
        setActiveNode(currentActivePath);
        restoreExpandedTreePaths(expandedPaths);
        showJournalArchive(date);
        return;
      }
    }
    if (currentTree && currentTree.type === "folder" && currentTree.children) {
      const defaultFolder = (currentSettings.defaultFolder || "").trim();
      const isTagPath = previousActivePath === "__tags__";
      let targetNode = null;
      if (!isTagPath) {
        if (previousActivePath === "") {
          targetNode = currentTree;
        } else if (previousActivePath) {
          targetNode = findFolderNode(currentTree, previousActivePath);
        }
      }
      if (!targetNode) {
        targetNode = defaultFolder ? findFolderNode(currentTree, defaultFolder) : currentTree;
      }
      const counts = countTreeItems(targetNode || currentTree);
      currentActivePath = targetNode ? targetNode.path : "";
      setActiveNode(currentActivePath);
      if (targetNode && targetNode.path) {
        expandToPath(targetNode.path);
      }
      const title = targetNode && targetNode.path ? `Folder: ${targetNode.name}` : "Notes";
      const actionPath = targetNode && targetNode.path ? targetNode.path : "";
      showSummary(
        title,
        [
          { label: "Folders", value: counts.folders },
          { label: "Notes", value: counts.notes },
          { label: "Assets", value: counts.assets },
          { label: "PDFs", value: counts.pdfs },
          { label: "CSVs", value: counts.csvs },
        ],
        { label: "New", handler: () => createNote(actionPath) }
      );
      restoreExpandedTreePaths(expandedPaths);
    }
    handleStartupShortcut();
    maybeShowWhatsNew();
  } catch (err) {
    alert(err.message);
  }
}

async function openNote(path) {
  try {
    showNoteEditor();
    const data = await apiFetch(`/notes?path=${encodeURIComponent(path)}`);
    currentNotePath = data.path;
    currentActivePath = data.path;
    notePath.textContent = data.path;
    editor.value = data.content;
    updatePreviewFromMarkdown(data.content);
    preview.classList.remove("hidden");
    assetPreview.classList.add("hidden");
    assetPreview.innerHTML = "";
    pdfPreview.classList.add("hidden");
    pdfPreview.innerHTML = "";
    csvPreview.classList.add("hidden");
    csvPreview.innerHTML = "";
    viewSelector.classList.remove("hidden");
    viewButtons.forEach((btn) => {
      btn.disabled = false;
    });
    setView(getPreferredView(), true);
    applyHighlighting();
    renderTagBar(extractTags(data.content));
    isDirty = false;
    saveBtn.disabled = false;
    if (moveCompletedBtn) {
      moveCompletedBtn.disabled = false;
    }
    focusTreePath(currentNotePath);
  } catch (err) {
    alert(err.message);
  }
}

function isInboxDialogOpen() {
  return inboxDialog && !inboxDialog.classList.contains("hidden");
}

function isScratchDialogOpen() {
  return scratchDialog && !scratchDialog.classList.contains("hidden");
}

function openInboxDialog() {
  if (!inboxDialog || !inboxDialogText) {
    return;
  }
  hideContextMenu();
  lastActiveElement = document.activeElement;
  inboxDialogText.value = "";
  inboxDialog.classList.remove("hidden");
  setTimeout(() => inboxDialogText.focus(), 0);
}

function closeInboxDialog() {
  if (!inboxDialog) {
    return;
  }
  inboxDialog.classList.add("hidden");
  if (lastActiveElement && typeof lastActiveElement.focus === "function") {
    lastActiveElement.focus();
  }
  lastActiveElement = null;
}

function clearScratchAutosave() {
  if (scratchAutosaveTimer) {
    clearTimeout(scratchAutosaveTimer);
    scratchAutosaveTimer = null;
  }
}

function scheduleScratchAutosave() {
  if (!scratchDialogText) {
    return;
  }
  scratchDirty = true;
  clearScratchAutosave();
  scratchAutosaveTimer = setTimeout(() => {
    saveScratchNote({ refresh: false }).catch((err) => alert(err.message));
  }, scratchAutosaveDelayMs);
}

async function loadScratchNote() {
  if (!scratchDialogText || !scratchDialogSave) {
    return;
  }
  scratchDialogSave.disabled = true;
  scratchNoteExists = false;
  scratchDirty = false;
  clearScratchAutosave();
  scratchDialogText.value = "";
  try {
    const data = await apiFetch(`/notes?path=${encodeURIComponent(scratchNotePath)}`);
    scratchNoteExists = true;
    scratchDialogText.value = data.content || "";
  } catch (err) {
    if (err.message !== "note not found") {
      throw err;
    }
  } finally {
    scratchDialogSave.disabled = false;
  }
}

function openScratchDialog() {
  if (!scratchDialog || !scratchDialogText) {
    return;
  }
  hideContextMenu();
  lastActiveElement = document.activeElement;
  scratchDialogText.value = "";
  scratchDirty = false;
  clearScratchAutosave();
  scratchDialog.classList.remove("hidden");
  setTimeout(() => scratchDialogText.focus(), 0);
  loadScratchNote().catch((err) => alert(err.message));
}

async function closeScratchDialog() {
  if (!scratchDialog) {
    return;
  }
  if (scratchDirty) {
    try {
      await saveScratchNote({ refresh: false });
    } catch (err) {
      alert(err.message);
      return;
    }
  }
  clearScratchAutosave();
  scratchDialog.classList.add("hidden");
  if (lastActiveElement && typeof lastActiveElement.focus === "function") {
    lastActiveElement.focus();
  }
  lastActiveElement = null;
}

async function saveScratchNote({ refresh = true } = {}) {
  if (!scratchDialogText) {
    return;
  }
  if (scratchSaveInFlight) {
    return;
  }
  const content = String(scratchDialogText.value || "").replace(/\r\n/g, "\n");
  const payload = { path: scratchNotePath, content };
  const method = scratchNoteExists ? "PATCH" : "POST";
  scratchSaveInFlight = true;
  try {
    await apiFetch("/notes", {
      method,
      body: JSON.stringify(payload),
    });
    scratchNoteExists = true;
    scratchDirty = false;
  } catch (err) {
    if (method === "PATCH" && err.message === "note not found") {
      await apiFetch("/notes", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      scratchNoteExists = true;
      scratchDirty = false;
    } else {
      throw err;
    }
  } finally {
    scratchSaveInFlight = false;
  }
  if (refresh) {
    await refreshTreePreserveMode();
  }
}

async function moveScratchToInbox() {
  if (!scratchDialogText) {
    return;
  }
  const content = String(scratchDialogText.value || "").replace(/\r\n/g, "\n");
  if (!content.trim()) {
    scratchDialogText.value = "";
    scratchDirty = false;
    return;
  }
  clearScratchAutosave();
  await appendToInbox(content);
  scratchDialogText.value = "";
  scratchDirty = true;
  await saveScratchNote({ refresh: true });
  setTimeout(() => scratchDialogText.focus(), 0);
}

async function appendToInbox(text) {
  const content = String(text || "").replace(/\r\n/g, "\n");
  if (!content.trim()) {
    return;
  }
  let existing = "";
  let found = false;
  try {
    const data = await apiFetch(`/notes?path=${encodeURIComponent(inboxNotePath)}`);
    existing = data.content || "";
    found = true;
  } catch (err) {
    if (err.message !== "note not found") {
      throw err;
    }
  }
  const separator = existing && !existing.endsWith("\n") ? "\n" : "";
  const nextContent = found ? `${existing}${separator}${content}` : content;
  const method = found ? "PATCH" : "POST";
  await apiFetch("/notes", {
    method,
    body: JSON.stringify({ path: inboxNotePath, content: nextContent }),
  });
  if (currentNotePath === inboxNotePath) {
    editor.value = nextContent;
    updatePreviewFromMarkdown(nextContent);
    renderTagBar(extractTags(nextContent));
    isDirty = false;
    saveBtn.disabled = false;
  }
  await refreshTasksAndTagsPreserveView();
}

async function openNoteAtLine(path, lineNumber) {
  await openNote(path);
  if (currentNotePath !== path) {
    return;
  }
  scrollEditorToLine(lineNumber);
}

function scrollEditorToLine(lineNumber) {
  if (!editor || !lineNumber) {
    return;
  }
  const lines = editor.value.split("\n");
  const clamped = Math.max(1, Math.min(lineNumber, lines.length));
  let index = 0;
  for (let i = 0; i < clamped - 1; i += 1) {
    index += lines[i].length + 1;
  }
  editor.setSelectionRange(index, index);
  const lineHeight = Number.parseFloat(getComputedStyle(editor).lineHeight) || 20;
  editor.scrollTop = Math.max(0, (clamped - 1) * lineHeight);
}

function openAsset(path) {
  if (!path) {
    return;
  }
  currentMode = "asset";
  summaryPanel.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  taskList.classList.add("hidden");
  if (journalPanel) {
    journalPanel.classList.add("hidden");
  }
  editor.classList.remove("hidden");
  editorPane.classList.remove("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.remove("hidden");
  currentNotePath = "";
  currentActivePath = path;
  notePath.textContent = path;
  editor.value = "";
  preview.innerHTML = "";
  preview.classList.add("hidden");
  assetPreview.classList.remove("hidden");
  assetPreview.innerHTML = "";
  pdfPreview.classList.add("hidden");
  pdfPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  const img = document.createElement("img");
  img.src = `${apiBase}/files?path=${encodeURIComponent(path)}`;
  img.alt = path.split("/").pop() || "Image";
  assetPreview.appendChild(img);
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  app.dataset.view = "preview";
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  isDirty = false;
  renderTagBar([]);
  expandToPath(path);
  setActiveNode(currentActivePath);
}

function openPdf(path) {
  if (!path) {
    return;
  }
  currentMode = "asset";
  summaryPanel.classList.add("hidden");
  settingsPanel.classList.add("hidden");
  taskList.classList.add("hidden");
  editor.classList.remove("hidden");
  editorPane.classList.remove("hidden");
  previewPane.classList.remove("hidden");
  paneResizer.classList.remove("hidden");
  currentNotePath = "";
  currentActivePath = path;
  notePath.textContent = path;
  editor.value = "";
  preview.innerHTML = "";
  preview.classList.add("hidden");
  assetPreview.classList.add("hidden");
  assetPreview.innerHTML = "";
  csvPreview.classList.add("hidden");
  csvPreview.innerHTML = "";
  pdfPreview.classList.remove("hidden");
  pdfPreview.innerHTML = "";
  const src = `${apiBase}/files?path=${encodeURIComponent(path)}`;
  const embed = document.createElement("embed");
  embed.type = "application/pdf";
  embed.src = src;
  pdfPreview.appendChild(embed);
  const fallback = document.createElement("div");
  fallback.className = "pdf-fallback";
  const link = document.createElement("a");
  link.href = src;
  link.textContent = "Open PDF";
  link.target = "_blank";
  link.rel = "noopener";
  fallback.appendChild(document.createTextNode("PDF preview unavailable. "));
  fallback.appendChild(link);
  pdfPreview.appendChild(fallback);
  viewSelector.classList.add("hidden");
  viewButtons.forEach((btn) => {
    btn.disabled = true;
  });
  app.dataset.view = "preview";
  saveBtn.disabled = true;
  if (moveCompletedBtn) {
    moveCompletedBtn.disabled = true;
  }
  isDirty = false;
  renderTagBar([]);
  expandToPath(path);
  setActiveNode(currentActivePath);
}

function parseCsv(text) {
  const rows = [];
  let row = [];
  let value = "";
  let inQuotes = false;

  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    if (char === "\"") {
      if (inQuotes && text[i + 1] === "\"") {
        value += "\"";
        i += 1;
      } else {
        inQuotes = !inQuotes;
      }
      continue;
    }
    if (char === "," && !inQuotes) {
      row.push(value);
      value = "";
      continue;
    }
    if ((char === "\n" || char === "\r") && !inQuotes) {
      if (char === "\r" && text[i + 1] === "\n") {
        i += 1;
      }
      row.push(value);
      rows.push(row);
      row = [];
      value = "";
      continue;
    }
    value += char;
  }
  row.push(value);
  rows.push(row);
  return rows;
}

function renderCsvTable(rows) {
  csvPreview.innerHTML = "";
  if (!rows || rows.length === 0) {
    return;
  }
  const table = document.createElement("table");
  table.className = "csv-table";
  const thead = document.createElement("thead");
  const headerRow = document.createElement("tr");
  rows[0].forEach((cell) => {
    const th = document.createElement("th");
    th.textContent = cell;
    headerRow.appendChild(th);
  });
  thead.appendChild(headerRow);
  table.appendChild(thead);

  const tbody = document.createElement("tbody");
  rows.slice(1).forEach((row) => {
    const tr = document.createElement("tr");
    row.forEach((cell) => {
      const td = document.createElement("td");
      td.textContent = cell;
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  });
  table.appendChild(tbody);
  csvPreview.appendChild(table);
}

async function openCsv(path) {
  if (!path) {
    return;
  }
  try {
    const response = await fetch(`${apiBase}/files?path=${encodeURIComponent(path)}`);
    if (!response.ok) {
      throw new Error("Unable to load CSV file");
    }
    const text = await response.text();
    renderCsvTable(parseCsv(text));
    currentMode = "asset";
    summaryPanel.classList.add("hidden");
    settingsPanel.classList.add("hidden");
    taskList.classList.add("hidden");
    if (journalPanel) {
      journalPanel.classList.add("hidden");
    }
    editor.classList.remove("hidden");
    editorPane.classList.remove("hidden");
    previewPane.classList.remove("hidden");
    paneResizer.classList.remove("hidden");
    currentNotePath = "";
    currentActivePath = path;
    notePath.textContent = path;
    editor.value = "";
    preview.innerHTML = "";
    preview.classList.add("hidden");
    assetPreview.classList.add("hidden");
    assetPreview.innerHTML = "";
    pdfPreview.classList.add("hidden");
    pdfPreview.innerHTML = "";
    csvPreview.classList.remove("hidden");
    viewSelector.classList.add("hidden");
    viewButtons.forEach((btn) => {
      btn.disabled = true;
    });
    app.dataset.view = "preview";
    saveBtn.disabled = true;
    if (moveCompletedBtn) {
      moveCompletedBtn.disabled = true;
    }
    isDirty = false;
    renderTagBar([]);
    expandToPath(path);
    setActiveNode(currentActivePath);
  } catch (err) {
    alert(err.message);
  }
}

async function saveNoteContent(path, content, options = {}) {
  if (!path) {
    return;
  }
  const { showStatus = true, suppressNotice = false } = options;
  try {
    if (showStatus && path === currentNotePath && isPreviewEditable() && previewDirty) {
      syncEditorFromPreview();
      content = editor.value;
    }
    if (showStatus && path === currentNotePath) {
      saveBtn.disabled = true;
      saveBtn.textContent = "Saving...";
    }
    await apiFetch("/notes", {
      method: "PATCH",
      body: JSON.stringify({
        path,
        content,
      }),
    });
    if (path === currentNotePath) {
      await refreshTasksAndTagsPreserveView({ suppressNotice });
    } else {
      await refreshTasksForNote(path, { suppressNotice });
    }
    if (path === currentNotePath && editor.value === content) {
      isDirty = false;
    }
    if (showStatus && path === currentNotePath) {
      saveBtn.textContent = "Save";
      saveBtn.disabled = false;
    }
  } catch (err) {
    if (showStatus && path === currentNotePath) {
      saveBtn.textContent = "Save";
      saveBtn.disabled = false;
    }
    alert(err.message);
  }
}

async function saveNote() {
  if (!currentNotePath) {
    return;
  }
  if (noteSaveTimer) {
    window.clearTimeout(noteSaveTimer);
    noteSaveTimer = null;
  }
  noteSaveSnapshot = null;
  noteSaveQueued = false;
  await saveNoteContent(currentNotePath, editor.value, { showStatus: true, suppressNotice: false });
}

function flushNoteSave() {
  if (!noteSaveSnapshot) {
    return;
  }
  if (noteSaveInFlight) {
    noteSaveQueued = true;
    return;
  }
  const snapshot = noteSaveSnapshot;
  noteSaveSnapshot = null;
  noteSaveInFlight = true;
  saveNoteContent(snapshot.path, snapshot.content, { showStatus: false, suppressNotice: true })
    .catch(() => {})
    .finally(() => {
      noteSaveInFlight = false;
      if (noteSaveQueued) {
        noteSaveQueued = false;
        scheduleNoteSave();
      }
    });
}

function saveCurrent() {
  if (currentMode === "settings") {
    saveSettings();
  } else {
    saveNote();
  }
}

function promptForName(label) {
  const name = window.prompt(label);
  if (!name) {
    return "";
  }
  if (name.includes("/") || name.includes("\\")) {
    alert("Names cannot include slashes.");
    return "";
  }
  return name.trim();
}

function promptForNameWithDefault(label, defaultValue) {
  const name = window.prompt(label, defaultValue);
  if (!name) {
    return "";
  }
  if (name.includes("/") || name.includes("\\")) {
    alert("Names cannot include slashes.");
    return "";
  }
  return name.trim();
}

function ensureMarkdownName(name) {
  if (name.toLowerCase().endsWith(".md")) {
    return name;
  }
  return `${name}.md`;
}

function ensureTemplateName(name) {
  if (name.toLowerCase().endsWith(".template")) {
    return name;
  }
  return `${name}.template`;
}

function formatDailyDate(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatDatePill(date) {
  return new Intl.DateTimeFormat(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
  }).format(date);
}

function joinPath(parent, child) {
  if (!parent) {
    return child;
  }
  return `${parent.replace(/\/+$/, "")}/${child}`;
}

function isDailyPath(path) {
  if (!path) {
    return false;
  }
  return path === dailyFolderName || path.startsWith(`${dailyFolderName}/`);
}

function findDailyNodeInTree(tree) {
  if (!tree || !Array.isArray(tree.children)) {
    return null;
  }
  return tree.children.find(
    (child) => child.type === "folder" && child.path === dailyFolderName
  ) || null;
}

function findDailyNotePathForDate(dateString) {
  if (!dateString || !currentTree) {
    return "";
  }
  const dailyNode = findDailyNodeInTree(currentTree);
  if (!dailyNode || !Array.isArray(dailyNode.children)) {
    return "";
  }
  const targetName = ensureMarkdownName(dateString);
  const direct = dailyNode.children.find(
    (child) => child.type === "file" && child.name === targetName
  );
  if (direct) {
    return direct.path;
  }
  const queue = dailyNode.children.filter((child) => child.type === "folder");
  while (queue.length > 0) {
    const folder = queue.shift();
    const children = folder.children || [];
    const match = children.find(
      (child) => child.type === "file" && child.name === targetName
    );
    if (match) {
      return match.path;
    }
    children.forEach((child) => {
      if (child.type === "folder") {
        queue.push(child);
      }
    });
  }
  return "";
}

function updateDatePill() {
  if (!datePill) {
    return;
  }
  const today = new Date();
  const label = formatDatePill(today);
  datePill.textContent = label;
  datePill.title = `Open ${label} note`;
  datePill.dataset.date = formatDailyDate(today);
  if (refreshDateElements()) {
    datePicker.value = datePill.dataset.date;
  }
}

function refreshDateElements() {
  if (!datePopover) {
    datePopover = document.getElementById("date-popover");
  }
  if (!datePicker) {
    datePicker = document.getElementById("date-picker");
  }
  return !!datePopover && !!datePicker;
}

function isDatePopoverOpen() {
  return refreshDateElements() && !datePopover.classList.contains("hidden");
}

function showDatePopover() {
  if (!refreshDateElements()) {
    return;
  }
  const fallbackDate = datePill ? datePill.dataset.date : formatDailyDate(new Date());
  datePicker.value = fallbackDate;
  datePopover.classList.remove("hidden");
  datePicker.focus();
  document.body.classList.add("date-open");
}

function hideDatePopover() {
  if (!refreshDateElements()) {
    return;
  }
  datePopover.classList.add("hidden");
  datePicker.blur();
  document.body.classList.remove("date-open");
}

function toggleDatePopover() {
  if (!datePopover || !datePicker) {
    return;
  }
  if (isDatePopoverOpen()) {
    hideDatePopover();
    return;
  }
  showDatePopover();
}

async function openDailyNoteForDate(dateString) {
  const notePath = joinPath(dailyFolderName, dateString);
  await loadTree();
  const existingPath = findDailyNotePathForDate(dateString);
  if (existingPath) {
    await openNote(existingPath);
    return;
  }
  try {
    const data = await apiFetch("/notes", {
      method: "POST",
      body: JSON.stringify({
        path: notePath,
        content: "",
      }),
    });
    if (data.notice) {
      alert(data.notice);
    }
    await loadTree();
    await openNote(data.path);
  } catch (err) {
    const message = String(err.message || "").toLowerCase();
    if (message.includes("already exists")) {
      await loadTree();
      const resolvedPath = findDailyNotePathForDate(dateString);
      if (resolvedPath) {
        await openNote(resolvedPath);
        return;
      }
      await openNote(ensureMarkdownName(notePath));
      return;
    }
    alert(err.message);
  }
}

async function openDailyNote() {
  await openDailyNoteForDate(formatDailyDate(new Date()));
}

function moveCompletedTasksForCurrentNote() {
  if (currentMode !== "note" || !currentNotePath) {
    return;
  }
  if (isPreviewEditable() && previewDirty) {
    syncEditorFromPreview();
  }
  const result = moveCompletedTasksToDoneSection(editor.value);
  if (result.text === editor.value) {
    return;
  }
  editor.value = result.text;
  updatePreviewFromMarkdown(result.text);
  renderTagBar(extractTags(result.text));
  isDirty = true;
  saveBtn.disabled = false;
}

async function createNote(parentPath = "") {
  const name = promptForName("New note name");
  if (!name) {
    return;
  }
  const path = parentPath ? `${parentPath}/${name}` : name;
  try {
    const data = await apiFetch("/notes", {
      method: "POST",
      body: JSON.stringify({
        path,
        content: "",
      }),
    });
    if (data.notice) {
      alert(data.notice);
    }
    await loadTree();
    await openNote(data.path);
  } catch (err) {
    alert(err.message);
  }
}

async function editTemplate(parentPath = "") {
  const templateName = "default.template";
  const path = parentPath ? `${parentPath}/${templateName}` : templateName;
  try {
    const data = await apiFetch("/notes", {
      method: "POST",
      body: JSON.stringify({
        path,
        content: "",
      }),
    });
    if (data.notice) {
      alert(data.notice);
    }
    await loadTree();
    await openNote(data.path);
  } catch (err) {
    if (String(err.message || "").toLowerCase().includes("already exists")) {
      await openNote(path);
      return;
    }
    alert(err.message);
  }
}

async function renameNote(path) {
  if (!path) {
    return;
  }
  const currentName = path.split("/").pop() || "";
  const isTemplate = currentName.toLowerCase().endsWith(".template");
  const displayName = isTemplate
    ? currentName
    : displayNodeName({ type: "file", name: currentName });
  const name = promptForNameWithDefault(`Rename note (${displayName})`, displayName);
  if (!name) {
    return;
  }
  const base = path.split("/").slice(0, -1).join("/");
  const newName = isTemplate ? ensureTemplateName(name) : ensureMarkdownName(name);
  const newPath = base ? `${base}/${newName}` : newName;
  try {
    const data = await apiFetch("/notes/rename", {
      method: "PATCH",
      body: JSON.stringify({ path, newPath }),
    });
    currentActivePath = parentPathForPath(path);
    await loadTree();
    if (currentNotePath === path) {
      await openNote(data.newPath || newPath);
    }
  } catch (err) {
    alert(err.message);
  }
}

async function createFolder(parentPath = "") {
  const name = promptForName("New folder name");
  if (!name) {
    return;
  }
  const path = parentPath ? `${parentPath}/${name}` : name;
  try {
    await apiFetch("/folders", {
      method: "POST",
      body: JSON.stringify({ path }),
    });
    await loadTree();
  } catch (err) {
    alert(err.message);
  }
}

async function renameFolder(path) {
  if (!path) {
    alert("Root folder cannot be renamed.");
    return;
  }
  if (path.toLowerCase() === dailyFolderName.toLowerCase()) {
    alert("Daily cannot be renamed.");
    return;
  }
  const currentName = path.split("/").pop();
  const name = promptForName(`Rename folder (${currentName})`);
  if (!name) {
    return;
  }
  const base = path.split("/").slice(0, -1).join("/");
  const newPath = base ? `${base}/${name}` : name;
  try {
    await apiFetch("/folders", {
      method: "PATCH",
      body: JSON.stringify({ path, newPath }),
    });
    await loadTree();
  } catch (err) {
    alert(err.message);
  }
}

async function deleteFolder(path) {
  if (!path) {
    alert("Root folder cannot be deleted.");
    return;
  }
  const confirmDelete = window.confirm("Delete this folder and all of its contents?");
  if (!confirmDelete) {
    return;
  }
  try {
    await apiFetch(`/folders?path=${encodeURIComponent(path)}`, {
      method: "DELETE",
    });
    await loadTree();
  } catch (err) {
    alert(err.message);
  }
}

async function deleteNote(path) {
  if (!path) {
    return;
  }
  const confirmDelete = window.confirm("Delete this note?");
  if (!confirmDelete) {
    return;
  }
  try {
    await apiFetch(`/notes?path=${encodeURIComponent(path)}`, {
      method: "DELETE",
    });
    currentActivePath = parentPathForPath(path);
    if (currentNotePath === path) {
      currentNotePath = "";
      notePath.textContent = "";
      editor.value = "";
      preview.innerHTML = "";
      preview.classList.remove("hidden");
      assetPreview.classList.add("hidden");
      assetPreview.innerHTML = "";
      pdfPreview.classList.add("hidden");
      pdfPreview.innerHTML = "";
      csvPreview.classList.add("hidden");
      csvPreview.innerHTML = "";
      viewSelector.classList.remove("hidden");
      viewButtons.forEach((btn) => {
        btn.disabled = false;
      });
      saveBtn.disabled = true;
      if (moveCompletedBtn) {
        moveCompletedBtn.disabled = true;
      }
      isDirty = false;
      renderTagBar([]);
    }
    await loadTree();
  } catch (err) {
    alert(err.message);
  }
}

function showContextMenu(x, y, items) {
  contextMenu.innerHTML = "";
  items.forEach((item) => {
    const button = document.createElement("button");
    button.type = "button";
    button.textContent = item.label;
    button.addEventListener("click", (event) => {
      event.preventDefault();
      event.stopPropagation();
      hideContextMenu();
      item.action();
    });
    contextMenu.appendChild(button);
  });
  contextMenu.style.left = `${x}px`;
  contextMenu.style.top = `${y}px`;
  contextMenu.classList.remove("hidden");
}

function hideContextMenu() {
  contextMenu.classList.add("hidden");
}

function promptRootIconChange(rootKey, row) {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "image/png,image/svg+xml,image/x-icon,image/vnd.microsoft.icon,.png,.svg,.ico";
  input.style.display = "none";
  input.addEventListener("change", async () => {
    const file = input.files && input.files[0];
    if (!file) {
      input.remove();
      return;
    }
    if (file.size > 1024 * 1024) {
      alert("Icon must be 1MB or smaller.");
      input.remove();
      return;
    }
    try {
      const formData = new FormData();
      formData.append("icon", file, file.name);
      const response = await fetch(`${apiBase}/icons/root?root=${encodeURIComponent(rootKey)}`, {
        method: "POST",
        body: formData,
      });
      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: "Request failed" }));
        throw new Error(error.error || "Request failed");
      }
      const data = await response.json();
      currentSettings.rootIcons = {
        ...(currentSettings.rootIcons || {}),
        [rootKey]: data.path,
      };
      applyRootIconToRow(row, rootKey);
    } catch (err) {
      alert(err.message);
    } finally {
      input.remove();
    }
  });
  document.body.appendChild(input);
  input.click();
}

async function resetRootIcon(rootKey, row) {
  try {
    const response = await apiFetch(`/icons/root?root=${encodeURIComponent(rootKey)}`, {
      method: "DELETE",
    });
    if (response && response.root === rootKey) {
      if (currentSettings.rootIcons) {
        delete currentSettings.rootIcons[rootKey];
      }
      applyRootIconToRow(row, rootKey);
    }
  } catch (err) {
    alert(err.message);
  }
}

function showNotesSortMenu(x, y) {
  const activeSortBy = currentSettings.notesSortBy || "name";
  const activeSortOrder = currentSettings.notesSortOrder || "asc";
  const labelFor = (sortBy, sortOrder, label) => {
    if (activeSortBy === sortBy && activeSortOrder === sortOrder) {
      return `* ${label}`;
    }
    return label;
  };
  showContextMenu(x, y, [
    { label: labelFor("name", "asc", "Name A-Z"), action: () => updateNotesSort("name", "asc") },
    { label: labelFor("name", "desc", "Name Z-A"), action: () => updateNotesSort("name", "desc") },
    { label: labelFor("created", "asc", "Created (Oldest)"), action: () => updateNotesSort("created", "asc") },
    { label: labelFor("created", "desc", "Created (Newest)"), action: () => updateNotesSort("created", "desc") },
    { label: labelFor("updated", "asc", "Updated (Oldest)"), action: () => updateNotesSort("updated", "asc") },
    { label: labelFor("updated", "desc", "Updated (Newest)"), action: () => updateNotesSort("updated", "desc") },
  ]);
}

async function updateNotesSort(sortBy, sortOrder) {
  try {
    const updated = await apiFetch("/settings", {
      method: "PATCH",
      body: JSON.stringify({ notesSortBy: sortBy, notesSortOrder: sortOrder }),
    });
    applySettings(updated);
    await refreshTreePreserveMode();
  } catch (err) {
    alert(err.message);
  }
}

function renderSearchResults(matches) {
  const safeMatches = Array.isArray(matches) ? matches : [];
  searchResults.innerHTML = "";
  if (safeMatches.length === 0) {
    const empty = document.createElement("div");
    empty.className = "search-empty";
    empty.textContent = "No matches";
    searchResults.appendChild(empty);
    return;
  }
  safeMatches.forEach((match) => {
    const button = document.createElement("button");
    button.type = "button";
    const rawName = match.name || match.path.split("/").pop();
    button.textContent = displayNodeName({ type: "file", name: rawName });
    button.title = match.path;
    button.addEventListener("click", () => {
      hideSearchResults();
      openNote(match.path);
    });
    searchResults.appendChild(button);
  });
}

function showSearchResults() {
  searchResults.classList.remove("hidden");
}

function hideSearchResults() {
  searchResults.classList.add("hidden");
}

async function runSearch() {
  const query = searchInput.value.trim();
  if (!query) {
    hideSearchResults();
    return;
  }
  try {
    const matches = await apiFetch(`/search?query=${encodeURIComponent(query)}`);
    renderSearchResults(matches);
    showSearchResults();
  } catch (err) {
    alert(err.message);
  }
}

function resolveAssetPath(href) {
  if (!href) {
    return "";
  }
  if (/^(https?:|data:|\/)/i.test(href)) {
    return href;
  }
  const base = currentNotePath.split("/").slice(0, -1).join("/");
  const combined = base ? `${base}/${href}` : href;
  return `${apiBase}/files?path=${encodeURIComponent(combined)}`;
}

function setupSplitters() {
  const isStacked = () => window.matchMedia("(max-width: 720px)").matches;
  let dragOverlay = null;

  function attachDragOverlay(cursor) {
    dragOverlay = document.createElement("div");
    dragOverlay.style.position = "fixed";
    dragOverlay.style.top = "0";
    dragOverlay.style.left = "0";
    dragOverlay.style.width = "100%";
    dragOverlay.style.height = "100%";
    dragOverlay.style.zIndex = "999";
    dragOverlay.style.cursor = cursor;
    dragOverlay.style.background = "transparent";
    document.body.appendChild(dragOverlay);
    document.body.style.userSelect = "none";
  }

  function detachDragOverlay() {
    if (dragOverlay) {
      dragOverlay.remove();
      dragOverlay = null;
    }
    document.body.style.userSelect = "";
  }

  sidebarResizer.addEventListener("mousedown", (event) => {
    event.preventDefault();
    attachDragOverlay("col-resize");
    const startX = event.clientX;
    const startY = event.clientY;
    const startWidth = sidebar.getBoundingClientRect().width;
    const startHeight = sidebar.getBoundingClientRect().height;

    function onMove(moveEvent) {
      if (isStacked()) {
        const delta = moveEvent.clientY - startY;
        const newHeight = Math.max(160, startHeight + delta);
        sidebar.style.height = `${newHeight}px`;
      } else {
        const delta = moveEvent.clientX - startX;
        const newWidth = Math.max(220, startWidth + delta);
        sidebar.style.width = `${newWidth}px`;
        document.documentElement.style.setProperty("--sidebar-width", `${newWidth}px`);
      }
    }

    function onUp() {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      detachDragOverlay();
      if (!isStacked()) {
        const width = sidebar.getBoundingClientRect().width;
        saveSidebarWidth(width);
      }
    }

    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  });

  paneResizer.addEventListener("mousedown", (event) => {
    event.preventDefault();
    attachDragOverlay(isStacked() ? "row-resize" : "col-resize");
    const startX = event.clientX;
    const startY = event.clientY;
    const startWidth = editorPane.getBoundingClientRect().width;
    const startHeight = editorPane.getBoundingClientRect().height;

    function onMove(moveEvent) {
      if (isStacked()) {
        const delta = moveEvent.clientY - startY;
        const newHeight = Math.max(120, startHeight + delta);
        editorPane.style.height = `${newHeight}px`;
      } else {
        const delta = moveEvent.clientX - startX;
        const containerWidth = mainContent.getBoundingClientRect().width;
        const newWidth = Math.min(containerWidth - 200, Math.max(200, startWidth + delta));
        editorPane.style.flex = "0 0 auto";
        editorPane.style.flexBasis = `${newWidth}px`;
        editorPane.style.width = "";
        previewPane.style.flex = "1 1 0";
        previewPane.style.width = "";
      }
    }

    function onUp() {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      detachDragOverlay();
    }

    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  });
}

if (treeToggleBtn) {
  treeToggleBtn.addEventListener("click", () => toggleAllTreeNodes());
}
if (sidebarToggle) {
  sidebarToggle.addEventListener("click", () => toggleSidebarCollapse());
  updateSidebarToggle();
}
settingsBtn.addEventListener("click", () => {
  hideContextMenu();
  showSettings();
});

if (scratchBtn) {
  scratchBtn.addEventListener("click", () => {
    if (!isScratchDialogOpen()) {
      openScratchDialog();
    }
  });
}

editor.addEventListener("input", () => {
  if (currentMode !== "note") {
    return;
  }
  isDirty = true;
  updatePreviewFromMarkdown(editor.value);
  renderTagBar(extractTags(editor.value));
  saveBtn.disabled = !currentNotePath;
  scheduleNoteSave();
});

preview.addEventListener("input", () => {
  if (currentMode !== "note" || previewSyncing) {
    return;
  }
  previewDirty = true;
  schedulePreviewSync();
});

preview.addEventListener("change", () => {
  if (currentMode !== "note" || previewSyncing) {
    return;
  }
  previewDirty = true;
  schedulePreviewSync();
});

preview.addEventListener("keydown", handlePreviewKeydown);
preview.addEventListener("click", (event) => {
  const link = event.target.closest("a");
  if (!link || !preview.contains(link)) {
    return;
  }
  const href = link.getAttribute("href");
  if (!href) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  if (link.target === "_blank" || href.startsWith("http")) {
    window.open(link.href, "_blank", "noopener");
  } else {
    window.location.assign(link.href);
  }
});

if (settingsDarkMode) {
  settingsDarkMode.addEventListener("change", () => {
    if (currentMode !== "settings") {
      return;
    }
    applySettings({ ...currentSettings, darkMode: settingsDarkMode.checked });
    isDirty = true;
    saveBtn.disabled = false;
  });
}

if (settingsDefaultView) {
  settingsDefaultView.addEventListener("change", () => {
    if (currentMode !== "settings") {
      return;
    }
    currentSettings.defaultView = getDefaultView(settingsDefaultView.value);
    isDirty = true;
    saveBtn.disabled = false;
  });
}

if (settingsDefaultFolder) {
  settingsDefaultFolder.addEventListener("input", () => {
    if (currentMode !== "settings") {
      return;
    }
    currentSettings.defaultFolder = settingsDefaultFolder.value.trim();
    isDirty = true;
    saveBtn.disabled = false;
  });
}

if (settingsShowTemplates) {
  settingsShowTemplates.addEventListener("change", () => {
    if (currentMode !== "settings") {
      return;
    }
    currentSettings.showTemplates = settingsShowTemplates.checked;
    isDirty = true;
    saveBtn.disabled = false;
  });
}

if (settingsNotesSortBy) {
  settingsNotesSortBy.addEventListener("change", () => {
    if (currentMode !== "settings") {
      return;
    }
    currentSettings.notesSortBy = settingsNotesSortBy.value;
    isDirty = true;
    saveBtn.disabled = false;
  });
}

if (settingsNotesSortOrder) {
  settingsNotesSortOrder.addEventListener("change", () => {
    if (currentMode !== "settings") {
      return;
    }
    currentSettings.notesSortOrder = settingsNotesSortOrder.value;
    isDirty = true;
    saveBtn.disabled = false;
  });
}

function normalizeTagInput(value) {
  const trimmed = String(value || "").trim();
  if (!trimmed) {
    return "";
  }
  const cleaned = trimmed.startsWith("#") ? trimmed.slice(1) : trimmed;
  const normalized = cleaned.toLowerCase();
  if (!/^[a-z0-9]+$/.test(normalized)) {
    return "";
  }
  return normalized;
}

function appendTagToNote(tag) {
  if (!currentNotePath) {
    alert("Select a note before adding tags.");
    return;
  }
  const current = editor.value || "";
  const normalizedTag = normalizeTagInput(tag);
  if (!normalizedTag) {
    alert("Tags must contain only letters or numbers.");
    return;
  }
  const lines = current.split("\n");
  const lastIndex = Math.max(0, lines.length - 1);
  const prefix = lines[lastIndex].trim().length === 0 ? "" : " ";
  lines[lastIndex] = `${lines[lastIndex]}${prefix}#${normalizedTag}`;
  editor.value = lines.join("\n");
  isDirty = true;
  updatePreviewFromMarkdown(editor.value);
  applyHighlighting();
  renderTagBar(extractTags(editor.value));
  saveBtn.disabled = !currentNotePath;
}

function getScrollRatio(element) {
  const maxScroll = element.scrollHeight - element.clientHeight;
  if (maxScroll <= 0) {
    return 0;
  }
  return element.scrollTop / maxScroll;
}

function syncScroll(source, target) {
  if (syncingScroll) {
    return;
  }
  if (activeScrollSource && source !== activeScrollSource) {
    return;
  }
  syncingScroll = true;
  const maxSourceScroll = source.scrollHeight - source.clientHeight;
  const maxTargetScroll = target.scrollHeight - target.clientHeight;
  const ratio = maxSourceScroll <= 0 ? 0 : source.scrollTop / maxSourceScroll;
  const topScaled =
    source.scrollHeight <= 0 ? 0 : source.scrollTop * (target.scrollHeight / source.scrollHeight);
  const bottomScaled = Math.max(0, maxTargetScroll) * ratio;
  const blended = topScaled * (1 - ratio) + bottomScaled * ratio;
  target.scrollTop = Math.max(0, Math.min(maxTargetScroll, blended));
  requestAnimationFrame(() => {
    syncingScroll = false;
  });
}

function markActiveScrollSource(source) {
  activeScrollSource = source;
  if (clearScrollSourceTimer) {
    clearTimeout(clearScrollSourceTimer);
  }
  clearScrollSourceTimer = setTimeout(() => {
    activeScrollSource = null;
  }, 140);
}

editor.addEventListener("wheel", () => markActiveScrollSource(editor), { passive: true });
preview.addEventListener("wheel", () => markActiveScrollSource(preview), { passive: true });
editor.addEventListener("touchstart", () => markActiveScrollSource(editor), { passive: true });
preview.addEventListener("touchstart", () => markActiveScrollSource(preview), { passive: true });

editor.addEventListener("scroll", () => {
  if (!activeScrollSource) {
    markActiveScrollSource(editor);
  }
  syncScroll(editor, preview);
});
preview.addEventListener("scroll", () => {
  if (!activeScrollSource) {
    markActiveScrollSource(preview);
  }
  syncScroll(preview, editor);
});

saveBtn.addEventListener("click", () => saveCurrent());
if (moveCompletedBtn) {
  moveCompletedBtn.disabled = !currentNotePath;
  moveCompletedBtn.addEventListener("click", () => moveCompletedTasksForCurrentNote());
}

viewButtons.forEach((btn) => {
  btn.addEventListener("click", () => setView(btn.dataset.view));
});

if (datePill) {
  datePill.addEventListener("click", () => openDailyNote());
  updateDatePill();
  setInterval(updateDatePill, 60000);
}

if (calendarBtn) {
  calendarBtn.addEventListener("click", (event) => {
    event.preventDefault();
    event.stopPropagation();
    toggleDatePopover();
  });
}

function attachSwipeHandlers(element, onDown, onUp) {
  if (!element) {
    return;
  }
  element.addEventListener("touchstart", (event) => {
    if (!isMobileView()) {
      return;
    }
    if (event.touches.length !== 1) {
      return;
    }
    touchStartY = event.touches[0].clientY;
  }, { passive: true });

  element.addEventListener("touchend", (event) => {
    if (!isMobileView() || touchStartY === null) {
      touchStartY = null;
      return;
    }
    const touch = event.changedTouches[0];
    if (!touch) {
      touchStartY = null;
      return;
    }
    const delta = touch.clientY - touchStartY;
    if (delta > 40 && onDown) {
      onDown();
    } else if (delta < -40 && onUp) {
      onUp();
    }
    touchStartY = null;
  });
}

attachSwipeHandlers(mainHeader, openSidebar, closeSidebar);
attachSwipeHandlers(sidebar, null, closeSidebar);

document.addEventListener("change", (event) => {
  if (!refreshDateElements()) {
    return;
  }
  if (event.target !== datePicker) {
    return;
  }
  const value = event.target.value;
  if (!value) {
    return;
  }
  openDailyNoteForDate(value);
  hideDatePopover();
});

tagAddBtn.addEventListener("click", () => {
  const input = window.prompt("Add tag");
  if (input === null) {
    return;
  }
  appendTagToNote(input);
});

window.addEventListener("keydown", (event) => {
  if (!(event.ctrlKey && event.altKey)) {
    return;
  }
  const key = event.key.toLowerCase();
  if (key === "k") {
    event.preventDefault();
    if (!commandPalette || commandPalette.classList.contains("hidden") === false) {
      return;
    }
    openCommandPalette();
    return;
  }
  if (key === "i") {
    event.preventDefault();
    if (!isInboxDialogOpen()) {
      openInboxDialog();
    }
    return;
  }
  if (key === "p") {
    event.preventDefault();
    if (!isScratchDialogOpen()) {
      openScratchDialog();
    }
    return;
  }
  if (key === "j") {
    event.preventDefault();
    currentActivePath = journalRootPath;
    setActiveNode(currentActivePath);
    showJournal();
    return;
  }
  if (key === "s") {
    event.preventDefault();
    saveCurrent();
    return;
  }
  if (key === "e") {
    event.preventDefault();
    setView("edit");
    return;
  }
  if (key === "v") {
    event.preventDefault();
    setView("preview");
    return;
  }
  if (key === "b") {
    event.preventDefault();
    setView("split");
    return;
  }
  if (key === "d") {
    event.preventDefault();
    openDailyNote();
    return;
  }
  if (key === "c") {
    event.preventDefault();
    toggleDatePopover();
  }
}, { capture: true });

treeContainer.addEventListener("contextmenu", (event) => {
  const row = event.target.closest(".node-row");
  if (row) {
    return;
  }
  event.preventDefault();
  showContextMenu(event.clientX, event.clientY, [
    { label: "New Folder", action: () => createFolder() },
    { label: "New Note", action: () => createNote() },
    { label: "Sort Notes...", action: () => showNotesSortMenu(event.clientX, event.clientY) },
  ]);
});

treeContainer.addEventListener("mousedown", () => hideContextMenu(), true);
mainHeader.addEventListener("mousedown", () => hideContextMenu(), true);
tagBar.addEventListener("mousedown", () => hideContextMenu(), true);
previewPane.addEventListener("mousedown", () => hideContextMenu(), true);
mainContent.addEventListener("mousedown", () => hideContextMenu(), true);
document.addEventListener("click", () => hideContextMenu());

document.addEventListener("mousedown", (event) => {
  if (!isDatePopoverOpen()) {
    return;
  }
  if (refreshDateElements() && datePopover.contains(event.target)) {
    return;
  }
  if (calendarBtn && calendarBtn.contains(event.target)) {
    return;
  }
  hideDatePopover();
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && isDatePopoverOpen()) {
    hideDatePopover();
  }
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && isInboxDialogOpen()) {
    closeInboxDialog();
  }
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && commandPalette && !commandPalette.classList.contains("hidden")) {
    closeCommandPalette();
  }
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && isScratchDialogOpen()) {
    closeScratchDialog();
  }
});

if (inboxDialogBackdrop) {
  inboxDialogBackdrop.addEventListener("click", () => closeInboxDialog());
}

if (inboxDialogClose) {
  inboxDialogClose.addEventListener("click", () => closeInboxDialog());
}

if (inboxDialogCancel) {
  inboxDialogCancel.addEventListener("click", () => closeInboxDialog());
}

if (inboxDialogSave) {
  inboxDialogSave.addEventListener("click", async () => {
    try {
      const value = inboxDialogText ? inboxDialogText.value : "";
      await appendToInbox(value);
      closeInboxDialog();
    } catch (err) {
      alert(err.message);
    }
  });
}

if (inboxDialogText) {
  inboxDialogText.addEventListener("keydown", async (event) => {
    if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      try {
        await appendToInbox(inboxDialogText.value);
        closeInboxDialog();
      } catch (err) {
        alert(err.message);
      }
    }
  });
}

if (scratchDialogBackdrop) {
  scratchDialogBackdrop.addEventListener("click", () => {
    closeScratchDialog();
  });
}

if (scratchDialogClose) {
  scratchDialogClose.addEventListener("click", () => {
    closeScratchDialog();
  });
}

if (scratchDialogCancel) {
  scratchDialogCancel.addEventListener("click", () => {
    closeScratchDialog();
  });
}

if (scratchDialogMove) {
  scratchDialogMove.addEventListener("click", async () => {
    try {
      await moveScratchToInbox();
    } catch (err) {
      alert(err.message);
    }
  });
}

if (scratchDialogSave) {
  scratchDialogSave.addEventListener("click", async () => {
    try {
      await saveScratchNote();
      closeScratchDialog();
    } catch (err) {
      alert(err.message);
    }
  });
}

if (scratchDialogText) {
  scratchDialogText.addEventListener("keydown", async (event) => {
    if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      try {
        await saveScratchNote();
        closeScratchDialog();
      } catch (err) {
        alert(err.message);
      }
    }
  });
  scratchDialogText.addEventListener("input", () => {
    scheduleScratchAutosave();
  });
}

if (commandBackdrop) {
  commandBackdrop.addEventListener("click", () => closeCommandPalette());
}

if (commandClose) {
  commandClose.addEventListener("click", () => closeCommandPalette());
}

if (commandInput) {
  commandInput.addEventListener("input", () => updateCommandResults());
  commandInput.addEventListener("keydown", (event) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (commandMatches.length > 0) {
        commandSelectedIndex = (commandSelectedIndex + 1) % commandMatches.length;
        renderCommandResults(commandMatches);
      }
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      if (commandMatches.length > 0) {
        commandSelectedIndex =
          (commandSelectedIndex - 1 + commandMatches.length) % commandMatches.length;
        renderCommandResults(commandMatches);
      }
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      if (commandMatches.length > 0) {
        runCommandItem(commandSelectedIndex);
      }
      return;
    }
  });
}

if (journalSave && journalInput) {
  journalSave.addEventListener("click", async () => {
    const value = journalInput.value || "";
    if (!value.trim()) {
      return;
    }
    try {
      await createJournalEntry(value);
      journalInput.value = "";
      journalInput.focus();
    } catch (err) {
      alert(err.message);
    }
  });
  journalInput.addEventListener("keydown", async (event) => {
    if (event.key !== "Enter" || event.altKey) {
      return;
    }
    event.preventDefault();
    const value = journalInput.value || "";
    if (!value.trim()) {
      return;
    }
    try {
      await createJournalEntry(value);
      journalInput.value = "";
      journalInput.focus();
    } catch (err) {
      alert(err.message);
    }
  });
}

if (journalArchiveAll) {
  journalArchiveAll.addEventListener("click", async () => {
    if (!confirm("Archive all journal entries?")) {
      return;
    }
    try {
      await apiFetch("/journal/archive-all", { method: "POST" });
      await loadJournalEntries();
      await refreshTreePreserveMode();
    } catch (err) {
      alert(err.message);
    }
  });
}

searchBtn.addEventListener("click", (event) => {
  event.preventDefault();
  runSearch();
});

searchInput.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    runSearch();
  }
});

document.addEventListener("click", (event) => {
  if (!searchResults.contains(event.target) && event.target !== searchBtn) {
    hideSearchResults();
  }
});

if (offlineRetry) {
  offlineRetry.addEventListener("click", (event) => {
    event.preventDefault();
    checkServerHealth().then(() => {
      if (!offlineState) {
        loadTree();
        return;
      }
      loadTree();
    });
  });
}

if (updateReload) {
  updateReload.addEventListener("click", (event) => {
    event.preventDefault();
    const waiting = serviceWorkerRegistration && serviceWorkerRegistration.waiting;
    if (waiting) {
      waiting.postMessage({ type: "SKIP_WAITING" });
      return;
    }
    setUpdateAvailable(false);
    window.location.reload();
  });
}

if (appVersionLabel) {
  appVersionLabel.textContent = `Version: ${appVersion}`;
}

if (whatsNewClose) {
  whatsNewClose.addEventListener("click", hideWhatsNewModal);
}

if (whatsNewConfirm) {
  whatsNewConfirm.addEventListener("click", hideWhatsNewModal);
}

if (whatsNewBackdrop) {
  whatsNewBackdrop.addEventListener("click", hideWhatsNewModal);
}

updateMobileLayout();
window.addEventListener("resize", () => {
  updateMobileLayout();
  if (taskMetaOverflowTimer) {
    window.clearTimeout(taskMetaOverflowTimer);
  }
  taskMetaOverflowTimer = window.setTimeout(() => {
    taskMetaOverflowTimer = null;
    updateTaskMetaOverflow();
  }, 120);
});
window.addEventListener("online", checkServerHealth);
window.addEventListener("offline", () => setOfflineState(true));
setView("preview");
setupSplitters();
registerServiceWorker();
checkServerHealth();
loadTree();
