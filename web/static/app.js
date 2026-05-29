// Toast system — called by HTMX event
document.body.addEventListener("htmx:afterRequest", function (evt) {
  const hdr = evt.detail.xhr.getResponseHeader("X-Toast");
  if (hdr) showToast(JSON.parse(hdr));
});

function handleActionResponse(evt) {
  const hdr = evt.detail.xhr.getResponseHeader("X-Toast");
  if (hdr) showToast(JSON.parse(hdr));
}

function showToast({ message, type }) {
  const container = document.getElementById("toast-container");
  if (!container) return;

  const toast = document.createElement("div");
  const base =
    "pointer-events-auto text-xs px-3 py-2 rounded border bg-surface transition-all duration-200 opacity-0 translate-y-1";
  const tone = {
    success: "border-neon-green text-neon-green shadow-glow-green",
    error: "border-neon-red text-neon-red shadow-glow-red",
    warning: "border-neon-amber text-neon-amber",
    info: "border-neon-blue text-neon-blue shadow-glow-blue",
  };
  toast.className = `${base} ${tone[type] || tone.info}`;
  toast.textContent = message || "Done";
  container.appendChild(toast);

  requestAnimationFrame(() => {
    toast.classList.remove("opacity-0", "translate-y-1");
  });

  setTimeout(() => {
    toast.classList.add("opacity-0", "translate-y-1");
  }, 2500);

  setTimeout(() => {
    toast.remove();
  }, 3000);
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

// Chapter viewer chrome visibility (used in viewer.html)
// Called from viewer page only — keep it a named function, not auto-running
function toggleViewerChrome() {
  const chrome = document.getElementById("viewer-chrome");
  if (!chrome) return;

  chrome.classList.toggle("opacity-0");
  chrome.classList.toggle("pointer-events-none");
  chrome.classList.toggle("translate-y-2");
}
