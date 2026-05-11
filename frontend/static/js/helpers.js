/**
 * helpers.js
 *
 * JS equivalents of the Go BFF helper functions.
 * These lived in cmd/client/helpers/ — now they run in the browser.
 */

// ─── Color helpers (mirrors helpers/color.go) ─────────────────────────────────

const HEX_FALLBACK = '#00C6FF';

/**
 * Normalises a raw hex color string from the backend.
 * Adds a leading '#' if missing, validates the format,
 * and uppercases the result — just like the Go NormalizeColor function.
 *
 * @param {string} color
 * @returns {string}
 */
export function normalizeColor(color) {
  if (!color) return HEX_FALLBACK;

  let c = color.trim();

  if (!c.startsWith('#')) {
    c = '#' + c;
  }

  if (!isValidHexColor(c)) {
    console.warn(`Invalid color format: ${c}, using fallback`);
    return HEX_FALLBACK;
  }

  return c.toUpperCase();
}

/**
 * Accepts lengths 4 (#RGB) or 7 (#RRGGBB).
 */
function isValidHexColor(s) {
  if (s.length !== 4 && s.length !== 7) return false;
  if (s[0] !== '#') return false;
  return /^[0-9A-Fa-f]+$/.test(s.slice(1));
}

/**
 * Mutates and returns an array of category objects with normalised colors.
 * Mirrors helpers/color.go PrepareCategories.
 *
 * @param {Array} categories
 * @returns {Array}
 */
export function prepareCategories(categories) {
  if (!Array.isArray(categories)) return [];
  return categories.map((cat) => ({
    ...cat,
    Color: normalizeColor(cat.color || cat.Color),
    ImagePath: cat.ImagePath || cat.imagePath,
  }));
}

// ─── Date helpers ─────────────────────────────────────────────────────────────

/**
 * Formats an ISO date string into a human-readable relative time,
 * falling back to a locale date string.
 *
 * @param {string} isoString
 * @returns {string}
 */
export function formatRelativeDate(isoString) {
  if (!isoString) return '';

  const date = new Date(isoString);
  if (isNaN(date)) return isoString;

  const now = new Date();
  const diffMs = now - date;
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHr = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHr / 24);

  if (diffSec < 60) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHr < 24) return `${diffHr}h ago`;
  if (diffDay < 7) return `${diffDay}d ago`;

  return date.toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

/**
 * Formats an ISO date string as a full timestamp for chat messages.
 *
 * @param {string} isoString
 * @returns {string}
 */
export function formatMessageDate(isoString) {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date)) return isoString;
  return date.toLocaleString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * Formats an ISO date string to time only (HH:MM) for chat messages.
 *
 * @param {string} isoString
 * @returns {string}
 */
export function formatMessageTime(isoString) {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date)) return isoString;
  return date.toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

// ─── DOM helpers ──────────────────────────────────────────────────────────────

/**
 * Safely sets the inner HTML of an element and returns it.
 * Use this instead of direct innerHTML assignments for clarity.
 *
 * @param {HTMLElement} el
 * @param {string} html
 * @returns {HTMLElement}
 */
export function setHTML(el, html) {
  el.innerHTML = html;
  return el;
}

/**
 * Creates an element with optional class names and inner HTML.
 *
 * @param {string} tag
 * @param {string|string[]} [classes]
 * @param {string} [html]
 * @returns {HTMLElement}
 */
export function createElement(tag, classes, html) {
  const el = document.createElement(tag);
  if (classes) {
    const list = Array.isArray(classes) ? classes : [classes];
    el.classList.add(...list.filter(Boolean));
  }
  if (html !== undefined) el.innerHTML = html;
  return el;
}

/**
 * Escapes a string for safe insertion into HTML.
 *
 * @param {string} str
 * @returns {string}
 */
export function escapeHTML(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

// ─── Scroll helpers ───────────────────────────────────────────────────────────

/**
 * Returns a throttled version of a function.
 * Used for scroll events (e.g. pagination on chat, topic lists).
 *
 * @param {Function} fn
 * @param {number} limitMs
 * @returns {Function}
 */
export function throttle(fn, limitMs) {
  let lastCall = 0;
  return function (...args) {
    const now = Date.now();
    if (now - lastCall >= limitMs) {
      lastCall = now;
      return fn.apply(this, args);
    }
  };
}

/**
 * Returns a debounced version of a function.
 * Used for search inputs.
 *
 * @param {Function} fn
 * @param {number} delayMs
 * @returns {Function}
 */
export function debounce(fn, delayMs) {
  let timer;
  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delayMs);
  };
}
