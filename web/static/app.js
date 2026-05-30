// Toast system — called by HTMX after every request.
document.body.addEventListener("htmx:afterRequest", function (evt) {
  const hdr = evt.detail.xhr.getResponseHeader("X-Toast");
  if (!hdr) return;

  try {
    showToast(JSON.parse(hdr));
  } catch (_err) {
    showToast({ message: hdr, type: "info" });
  }
});

function handleActionResponse(_evt) {
  // Toasts are handled globally via the htmx:afterRequest listener above.
}

function showToast({ message, type }) {
  const container = document.getElementById("toast-container");
  if (!container) return;

  const el = document.createElement("div");
  el.className = `pointer-events-auto border rounded px-4 py-2 text-sm font-mono animate-fade-up max-w-xs ${
    type === "success"
      ? "border-neon-green text-neon-green bg-neon-green/5"
      : type === "error"
        ? "border-neon-red text-neon-red bg-neon-red/5"
        : "border-neon-blue text-neon-blue bg-neon-blue/5"
  }`;
  el.textContent = message || "Done";
  container.appendChild(el);
  setTimeout(() => el.remove(), 3000);
}

// Action menu toggle
function toggleActionMenu() {
  const menu = document.getElementById("action-menu");
  const toggle = document.getElementById("action-menu-toggle");
  if (!menu || !toggle) return;

  const opening = !menu.classList.contains("open");
  menu.classList.toggle("open", opening);
  toggle.setAttribute("aria-expanded", String(opening));
}

document.addEventListener("click", (e) => {
  const menu = document.getElementById("action-menu");
  const toggle = document.getElementById("action-menu-toggle");
  if (!menu || !toggle) return;
  if (!menu.contains(e.target) && !toggle.contains(e.target)) {
    menu.classList.remove("open");
    toggle.setAttribute("aria-expanded", "false");
  }
});

document.addEventListener("htmx:afterRequest", (e) => {
  const menu = document.getElementById("action-menu");
  if (!menu) return;
  if (menu.contains(e.target)) {
    menu.classList.remove("open");
    const toggle = document.getElementById("action-menu-toggle");
    if (toggle) toggle.setAttribute("aria-expanded", "false");
  }
});

let descOpen = false;
function toggleDesc() {
  descOpen = !descOpen;
  const p = document.getElementById("desc");
  const btn = document.getElementById("desc-btn");
  if (!p || !btn) return;
  p.classList.toggle("clamped", !descOpen);
  btn.textContent = descOpen ? "Show less ↑" : "Show more ↓";
}

// Hide the "Show more" button if the text is short enough to not overflow
document.addEventListener("DOMContentLoaded", () => {
  const p = document.getElementById("desc");
  const btn = document.getElementById("desc-btn");
  if (!p || !btn) return;
  // If clamped height === scroll height, nothing is hidden — no button needed
  if (p.scrollHeight <= p.clientHeight) {
    btn.style.display = "none";
  }
});

function filterJobs(status) {
  const list = document.getElementById("job-list");
  if (!list) return;

  const url =
    status === "all"
      ? "/api/jobs"
      : "/api/jobs?status=" + encodeURIComponent(status);
  list.setAttribute("hx-get", url);

  if (window.htmx) {
    htmx.process(list);
    htmx.trigger(list, "load");
  }

  const filters = document.querySelectorAll("#job-filters button[data-status]");
  filters.forEach((button) => {
    const active = button.dataset.status === status;
    button.classList.toggle("text-neon-blue", active);
    button.classList.toggle("border-b-2", active);
    button.classList.toggle("border-neon-blue", active);
    button.classList.toggle("text-dim", !active);
    button.classList.toggle("hover:text-white", !active);
  });
}

let currentExpanded = 0;

function openJob(id) {
  if (currentExpanded === id) {
    currentExpanded = 0;
  } else {
    currentExpanded = id;
  }

  const jobList = document.getElementById("job-list");
  const url = currentExpanded
    ? `/api/jobs?expanded=${currentExpanded}`
    : `/api/jobs`;

  htmx.ajax("GET", url, { target: "#job-list", swap: "innerHTML" });
}

document.addEventListener("htmx:configRequest", (e) => {
  if (e.target.id === "job-list" && currentExpanded) {
    e.detail.parameters["expanded"] = currentExpanded;
  }
});

// Chapter viewer chrome visibility (used in viewer.html)
// Called from viewer page only — keep it a named function, not auto-running
function toggleViewerChrome() {
  const chrome = document.getElementById("viewer-chrome");
  if (!chrome) return;

  chrome.classList.toggle("opacity-0");
  chrome.classList.toggle("pointer-events-none");
  chrome.classList.toggle("translate-y-2");
}
