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
  if (!menu) return;

  const isHidden = menu.classList.contains("hidden");
  menu.classList.toggle("hidden");
  menu.classList.toggle("opacity-0");
  menu.classList.toggle("translate-y-1");
  menu.classList.toggle("pointer-events-none");

  const toggle = document.getElementById("action-menu-toggle");
  if (toggle) toggle.setAttribute("aria-expanded", String(isHidden));
}

function toggleDesc() {
  const desc = document.getElementById("desc");
  const btn = document.getElementById("desc-btn");
  if (!desc || !btn) return;

  const isClamped = desc.classList.contains("line-clamp-3");
  desc.classList.toggle("line-clamp-3");
  btn.textContent = isClamped ? "Show less" : "Show more";
}

function filterJobs(status) {
  const list = document.getElementById("job-list");
  if (!list) return;

  const url = status === "all" ? "/api/jobs" : "/api/jobs?status=" + encodeURIComponent(status);
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

function toggleJobDetail(id) {
  const detail = document.getElementById("detail-" + id);
  const arrow = document.getElementById("arrow-" + id);
  if (!detail) return;

  detail.classList.toggle("hidden");
  if (arrow) {
    arrow.style.transform = detail.classList.contains("hidden") ? "" : "rotate(180deg)";
  }

  if (!detail.classList.contains("hidden") && window.htmx) {
    htmx.trigger(detail, "revealed");
  }
}

// Chapter viewer chrome visibility (used in viewer.html)
// Called from viewer page only — keep it a named function, not auto-running
function toggleViewerChrome() {
  const chrome = document.getElementById("viewer-chrome");
  if (!chrome) return;

  chrome.classList.toggle("opacity-0");
  chrome.classList.toggle("pointer-events-none");
  chrome.classList.toggle("translate-y-2");
}
