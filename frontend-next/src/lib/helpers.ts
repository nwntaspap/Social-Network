/**
 * lib/helpers.ts
 *
 * JS equivalents of the Go BFF helper functions — mirrors helpers.js.
 */

// ─── Types ────────────────────────────────────────────────────────────────────

import type { Category, Topic } from './types';

// ─── Color helpers (mirrors helpers/color.go + helpers.js) ───────────────────

const HEX_FALLBACK = '#00C6FF';

/**
 * Normalises a raw hex color string from the backend.
 * Adds a leading '#' if missing, validates the format, uppercases.
 */
export function normalizeColor(color?: string): string {
  if (!color) return HEX_FALLBACK;

  let c = color.trim();
  if (!c.startsWith('#')) c = '#' + c;

  if (!isValidHexColor(c)) {
    console.warn(`Invalid color format: ${c}, using fallback`);
    return HEX_FALLBACK;
  }

  return c.toUpperCase();
}

function isValidHexColor(s: string): boolean {
  if (s.length !== 4 && s.length !== 7) return false;
  if (s[0] !== '#') return false;
  return /^[0-9A-Fa-f]+$/.test(s.slice(1));
}

/**
 * Helper to normalize a color field that might be a string
 */
function normalizeColorField(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined;
  return value.startsWith('#') ? value : `#${value}`;
}

/**
 * Mirrors prepareCategories() — normalises color on each category object.
 */
export function prepareCategories(raw: Record<string, unknown>[]): Category[] {
  return raw.map((cat) => {
    const image = cat.imagePath as string | undefined;

    const image_path = image ? '/' + image.replace(/^static\//, '') : undefined;

    return {
      ...cat,

      id: cat.ID ?? cat.id,
      ID: cat.ID ?? cat.id,

      color: normalizeColorField(cat.Color ?? cat.color),

      topics: (cat.topics as Topic[]) || (cat.Topics as Topic[]) || [],
      Topics: (cat.Topics as Topic[]) || (cat.topics as Topic[]) || [],

      image_path,
    };
  }) as Category[];
}

// ─── Date helpers ─────────────────────────────────────────────────────────────

export function formatRelativeDate(isoString?: string): string {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date.getTime())) return isoString;

  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
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

export function formatMessageDate(isoString?: string): string {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date.getTime())) return isoString;
  return date.toLocaleString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatMessageTime(isoString?: string): string {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date.getTime())) return isoString;
  return date.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
}

export function formatNotificationTime(isoString?: string): string {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date.getTime())) return '';
  const diffMin = Math.floor((Date.now() - date.getTime()) / 60000);
  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
}

// ─── String helpers ───────────────────────────────────────────────────────────

/**
 * Escapes a string for safe insertion into HTML.
 * In React we rarely need this (JSX handles it), but it's here for
 * parity with helpers.js and for any dangerouslySetInnerHTML cases.
 */
export function escapeHTML(str?: string): string {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

// ─── Function helpers ─────────────────────────────────────────────────────────

export function throttle<T extends (...args: unknown[]) => unknown>(
  fn: T,
  limitMs: number
): (...args: Parameters<T>) => void {
  let lastCall = 0;
  return function (this: unknown, ...args: Parameters<T>) {
    const now = Date.now();
    if (now - lastCall >= limitMs) {
      lastCall = now;
      fn.apply(this, args);
    }
  };
}

export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delayMs: number
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout>;
  return function (this: unknown, ...args: Parameters<T>) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delayMs);
  };
}
