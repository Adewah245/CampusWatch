/**
 * Shared user-interface helpers: element construction, the brand mark,
 * formatting, status badges, loading and error states, and toasts.
 *
 * Security note that governs this whole file: values rendered here originate
 * on *monitored computers* — hostnames, usernames, location names and issue
 * text are all attacker-influenced if any campus machine is compromised. A
 * machine named `<img src=x onerror=...>` would run script in the operator's
 * dashboard. Every helper below therefore writes through `textContent` or
 * `createElement`, and nothing in this frontend assigns server data to
 * `innerHTML`.
 */

import { BRAND_MARK, BRAND_MARK_FALLBACK } from './config.js';

// ---------------------------------------------------------------------------
// Element construction
// ---------------------------------------------------------------------------

/**
 * Creates an element.
 *
 * @param {string} tag tag name, for example "div"
 * @param {object} [attrs] attributes; `class`, `text`, `dataset` and event
 *   handlers (`onclick`) are handled specially, everything else is set as an
 *   attribute.
 * @param {Array<Node|string|null|undefined>} [children]
 * @returns {HTMLElement}
 */
export function el(tag, attrs = {}, children = []) {
  const node = document.createElement(tag);

  for (const [key, value] of Object.entries(attrs)) {
    if (value === undefined || value === null || value === false) continue;
    if (key === 'class') node.className = value;
    else if (key === 'text') node.textContent = value;
    else if (key === 'dataset') Object.assign(node.dataset, value);
    else if (key.startsWith('on') && typeof value === 'function') {
      node.addEventListener(key.slice(2), value);
    } else if (value === true) node.setAttribute(key, '');
    else node.setAttribute(key, String(value));
  }

  for (const child of children.flat()) {
    if (child === null || child === undefined || child === false) continue;
    node.append(child instanceof Node ? child : document.createTextNode(String(child)));
  }
  return node;
}

/**
 * Replaces an element's contents.
 *
 * @param {HTMLElement} container
 * @param {Array<Node|string|null|undefined>} children
 */
export function render(container, children) {
  container.replaceChildren(...children.flat().filter(Boolean));
}

/** Looks up a required element, failing loudly if the page markup is wrong. */
export function mustFind(selector) {
  const node = document.querySelector(selector);
  if (!node) throw new Error(`Dashboard markup is missing ${selector}`);
  return node;
}

/**
 * Builds the brand mark.
 *
 * Displays the institution's logo from `config.js`, falling back to the bundled
 * vector mark if that file cannot be loaded. The fallback exists so a missing or
 * misnamed logo degrades to the built-in mark instead of showing a broken-image
 * icon on every page.
 *
 * The swap is attached here rather than as an inline `onerror` attribute so
 * that a strict Content-Security-Policy without `unsafe-inline` still works.
 *
 * @param {string} prefix path from the calling page to the frontend root:
 *   `'../'` from a page in pages/, which is every page that calls this.
 * @param {object} [options]
 * @param {number} [options.size] rendered size in pixels
 * @param {string|null} [options.className]
 * @returns {HTMLImageElement}
 */
export function brandMark(prefix = '', options = {}) {
  const { size = 28, className = 'brand__mark' } = options;

  const image = el('img', {
    class: className,
    src: `${prefix}${BRAND_MARK}`,
    // Decorative: the adjacent wordmark already names the product, so an
    // alternative text here would only repeat it to a screen reader.
    alt: '',
    width: String(size),
    height: String(size),
  });

  image.addEventListener(
    'error',
    () => {
      // Guards against a loop if the fallback is itself unavailable, which can
      // otherwise fire `error` indefinitely.
      if (image.dataset.fallbackApplied) return;
      image.dataset.fallbackApplied = 'true';
      image.src = `${prefix}${BRAND_MARK_FALLBACK}`;
    },
    { once: true },
  );

  return image;
}

// ---------------------------------------------------------------------------
// Formatting
// ---------------------------------------------------------------------------

/** Formats an ISO timestamp in the operator's local time, or an em dash. */
export function formatDateTime(value) {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/** Formats a timestamp as "3 minutes ago". */
export function formatRelative(value) {
  if (!value) return 'never';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'never';

  const seconds = Math.round((Date.now() - date.getTime()) / 1000);
  if (seconds < 0) return 'just now';
  if (seconds < 60) return `${seconds}s ago`;

  const units = [
    ['minute', 60],
    ['hour', 60],
    ['day', 24],
    ['month', 30],
    ['year', 12],
  ];
  let amount = seconds / 60;
  let label = 'minute';
  for (let i = 0; i < units.length; i += 1) {
    if (amount < units[i][1]) {
      label = units[i][0];
      break;
    }
    amount /= units[i][1];
    label = units[i + 1]?.[0] ?? 'year';
  }
  const rounded = Math.floor(amount);
  return `${rounded} ${label}${rounded === 1 ? '' : 's'} ago`;
}

/** Formats a number of seconds as "2d 4h 15m". */
export function formatDuration(totalSeconds) {
  const seconds = Number(totalSeconds);
  if (!Number.isFinite(seconds) || seconds < 0) return '—';
  if (seconds === 0) return '0m';

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const parts = [];
  if (days) parts.push(`${days}d`);
  if (hours) parts.push(`${hours}h`);
  if (minutes || parts.length === 0) parts.push(`${minutes}m`);
  return parts.join(' ');
}

/** Formats a 0-100 value as "42.5%", or an em dash when unmeasured. */
export function formatPercent(value) {
  const number = Number(value);
  if (!Number.isFinite(number)) return '—';
  return `${number.toFixed(1)}%`;
}

/** Formats an integer with thousands separators. */
export function formatCount(value) {
  const number = Number(value);
  if (!Number.isFinite(number)) return '—';
  return number.toLocaleString();
}

/** Converts a SCREAMING_SNAKE_CASE enum value to "Screaming snake case". */
export function humanise(value) {
  if (!value) return '—';
  const spaced = String(value).replace(/_/g, ' ').toLowerCase();
  return spaced.charAt(0).toUpperCase() + spaced.slice(1);
}

// ---------------------------------------------------------------------------
// Status badges
// ---------------------------------------------------------------------------

/**
 * Maps a status value to a badge tone. The status vocabularies come from
 * README sections 8, 9 and 45; anything unrecognised falls back to "neutral"
 * rather than being hidden, so a new backend status is visible immediately.
 */
const STATUS_TONES = {
  // System monitor status
  ONLINE: 'ok',
  OFFLINE: 'danger',
  INACTIVE: 'warn',
  MAINTENANCE: 'info',
  PENDING: 'warn',
  SUSPENDED: 'muted',
  RETIRED: 'muted',
  // Issue status
  OPEN: 'danger',
  ACKNOWLEDGED: 'warn',
  IN_PROGRESS: 'info',
  RESOLVED: 'ok',
  CLOSED: 'muted',
  // Session status
  LOGGED_IN_ACTIVE: 'ok',
  LOGGED_IN_IDLE: 'warn',
  LOGGED_OUT: 'muted',
  // Priority
  LOW: 'muted',
  MEDIUM: 'info',
  HIGH: 'warn',
  CRITICAL: 'danger',
  // Agent health
  HEALTHY: 'ok',
  DEGRADED: 'warn',
};

/**
 * Builds a coloured status pill.
 *
 * @param {string} status
 * @returns {HTMLElement}
 */
export function statusBadge(status) {
  const tone = STATUS_TONES[String(status).toUpperCase()] || 'neutral';
  return el('span', { class: `badge badge--${tone}`, text: humanise(status) });
}

/**
 * Builds a horizontal meter for a percentage metric.
 *
 * The numeric value is always printed alongside the bar: colour alone is not
 * an accessible signal, and operators reading a wall display need the number.
 *
 * @param {string} label
 * @param {number|null|undefined} value percentage from 0 to 100
 * @param {number} [warnAt] threshold at which the meter turns amber
 * @param {number} [dangerAt] threshold at which the meter turns red
 * @returns {HTMLElement}
 */
export function metricMeter(label, value, warnAt = 75, dangerAt = 90) {
  const number = Number(value);
  const measured = Number.isFinite(number);
  const clamped = measured ? Math.min(100, Math.max(0, number)) : 0;

  let tone = 'ok';
  if (measured && number >= dangerAt) tone = 'danger';
  else if (measured && number >= warnAt) tone = 'warn';

  return el('div', { class: 'meter' }, [
    el('div', { class: 'meter__head' }, [
      el('span', { class: 'meter__label', text: label }),
      el('span', { class: 'meter__value', text: measured ? formatPercent(number) : 'not reported' }),
    ]),
    el('div', { class: 'meter__track' }, [
      el('div', {
        class: `meter__fill meter__fill--${tone}`,
        style: `width: ${clamped}%`,
        role: 'img',
        'aria-label': `${label}: ${measured ? formatPercent(number) : 'not reported'}`,
      }),
    ]),
  ]);
}

// ---------------------------------------------------------------------------
// Page states
// ---------------------------------------------------------------------------

/** A centred "loading" placeholder. */
export function loadingState(message = 'Loading…') {
  return el('div', { class: 'state state--loading' }, [
    el('span', { class: 'spinner', 'aria-hidden': 'true' }),
    el('span', { text: message }),
  ]);
}

/** A centred "nothing here" placeholder. */
export function emptyState(message) {
  return el('div', { class: 'state state--empty', text: message });
}

/**
 * A centred error placeholder with an optional retry action.
 *
 * @param {string} message
 * @param {() => void} [onRetry]
 */
export function errorState(message, onRetry) {
  return el('div', { class: 'state state--error' }, [
    el('p', { text: message }),
    onRetry ? el('button', { class: 'btn btn--ghost', text: 'Try again', onclick: onRetry }) : null,
  ]);
}

// ---------------------------------------------------------------------------
// Toasts
// ---------------------------------------------------------------------------

/**
 * Shows a transient notification.
 *
 * @param {string} message
 * @param {'info'|'success'|'error'} [tone]
 */
export function toast(message, tone = 'info') {
  let host = document.querySelector('.toasts');
  if (!host) {
    host = el('div', { class: 'toasts', role: 'status', 'aria-live': 'polite' });
    document.body.append(host);
  }

  const node = el('div', { class: `toast toast--${tone}`, text: message });
  host.append(node);

  // Announced for screen readers by aria-live above, so the removal is purely
  // visual and needs no further announcement.
  setTimeout(() => node.remove(), 6000);
}

/**
 * Runs an async action with a button disabled and showing progress.
 *
 * Guards against double-submission, which matters for the mutating actions
 * here: a double-clicked "Approve" should not send two requests.
 *
 * @param {HTMLButtonElement} button
 * @param {() => Promise<any>} action
 */
export async function withBusy(button, action) {
  const original = button.textContent;
  button.disabled = true;
  button.textContent = 'Working…';
  try {
    return await action();
  } finally {
    button.disabled = false;
    button.textContent = original;
  }
}

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

/**
 * Builds a table.
 *
 * Each cell carries its column heading in a `data-label` attribute. The desktop
 * layout ignores it (the heading row is visible), but the narrow-screen layout
 * in main.css hides the header row and prints the label beside each value,
 * turning a table into a readable stack of cards. Without it, a seven-column
 * table on a phone is unreadable — scrolling sideways loses which column a
 * value belongs to.
 *
 * @param {Array<{heading: string, numeric?: boolean}>} columns
 * @param {Array<Array<Node|string|null>>} rows
 * @param {string} [caption] accessible description of the table
 * @returns {HTMLElement}
 */
export function table(columns, rows, caption) {
  return el('table', { class: 'table' }, [
    caption ? el('caption', { class: 'sr-only', text: caption }) : null,
    el('thead', {}, [
      el('tr', {}, columns.map((column) =>
        el('th', {
          scope: 'col',
          class: column.numeric ? 'is-numeric' : '',
          text: column.heading,
        }),
      )),
    ]),
    el('tbody', {}, rows.map((row) =>
      el('tr', {}, row.map((cell, index) => {
        const column = columns[index];
        return el(
          'td',
          {
            class: column?.numeric ? 'is-numeric' : '',
            // Empty for a column without a heading, which keeps the stacked
            // layout from printing a stray label.
            'data-label': column?.heading || '',
          },
          [cell],
        );
      })),
    )),
  ]);
}

/**
 * Highlights the nav entry matching the current page.
 *
 * @param {string} activePage filename of the current page, for example "systems.html"
 */
export function markActiveNav(activePage) {
  for (const link of document.querySelectorAll('.nav__link')) {
    const target = link.getAttribute('href') || '';
    const isActive = target.endsWith(activePage);
    link.classList.toggle('is-active', isActive);
    if (isActive) link.setAttribute('aria-current', 'page');
  }
}

// ---------------------------------------------------------------------------
// Password visibility
// ---------------------------------------------------------------------------

/** Namespace for the inline icon markup below. */
const SVG_NS = 'http://www.w3.org/2000/svg';

/**
 * Builds the eye / crossed-out eye the password toggle shows.
 *
 * Inline SVG rather than an icon font or an image file: the frontend has no
 * build step and no icon set, and a two-path icon is smaller than the request
 * that would fetch it. `aria-hidden` because the button carries the words — an
 * icon alone is not an accessible name.
 *
 * @param {boolean} revealed whether the password is currently visible
 * @returns {SVGElement}
 */
function eyeIcon(revealed) {
  const svg = document.createElementNS(SVG_NS, 'svg');
  svg.setAttribute('viewBox', '0 0 24 24');
  svg.setAttribute('fill', 'none');
  svg.setAttribute('stroke', 'currentColor');
  svg.setAttribute('stroke-width', '2');
  svg.setAttribute('stroke-linecap', 'round');
  svg.setAttribute('stroke-linejoin', 'round');
  svg.setAttribute('aria-hidden', 'true');

  const add = (tag, attrs) => {
    const node = document.createElementNS(SVG_NS, tag);
    for (const [name, value] of Object.entries(attrs)) node.setAttribute(name, value);
    svg.append(node);
  };

  if (revealed) {
    // Eye with a slash: the password is on screen, so the button hides it.
    add('path', {
      d: 'M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24',
    });
    add('line', { x1: '1', y1: '1', x2: '23', y2: '23' });
  } else {
    add('path', { d: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z' });
    add('circle', { cx: '12', cy: '12', r: '3' });
  }

  return svg;
}

/**
 * Adds a show/hide button to a password input.
 *
 * The point is typing accuracy: a password field is the one place where the
 * operator cannot see their own mistake, and on a registration form they have
 * to reproduce it exactly a second time. Revealing it is a deliberate,
 * reversible act with a visible state, so the toggle reports `aria-pressed`
 * and changes its label rather than only its icon.
 *
 * Reverting to `password` on every page load is intentional: nothing here
 * remembers the choice, so a revealed password is never left on screen for the
 * next person at the machine.
 *
 * Expects the input to sit in a `.field__control` wrapper (see css/main.css);
 * returns null if it does not, because a caller that forgot the wrapper should
 * not get a button floating in the wrong place.
 *
 * @param {HTMLInputElement} input
 * @returns {HTMLButtonElement|null}
 */
export function attachPasswordToggle(input) {
  const wrapper = input.parentElement;
  if (!wrapper || !wrapper.classList.contains('field__control')) return null;

  input.classList.add('input--with-toggle');

  // The accessible name lives in text, not in the icon: "Show password" /
  // "Hide password" is what a screen reader announces, and it flips with the
  // state. `.sr-only` keeps it out of the visual layout, which the icon fills.
  const label = document.createElement('span');
  label.className = 'sr-only';
  label.textContent = 'Show password';

  const button = el(
    'button',
    {
      type: 'button',
      class: 'field__toggle',
      'aria-pressed': 'false',
      onclick: () => {
        const revealed = input.type === 'text';
        input.type = revealed ? 'password' : 'text';
        button.setAttribute('aria-pressed', String(!revealed));
        label.textContent = revealed ? 'Show password' : 'Hide password';
        button.replaceChildren(eyeIcon(!revealed), label);
        // Takes the focus back off the button. Clicking it would otherwise
        // leave focus there, so the operator would type into nothing until they
        // noticed and clicked back into the field. The cost is that a keyboard
        // operator needs one Shift+Tab to toggle a second time, which is the
        // cheaper of the two.
        input.focus();
      },
    },
    [eyeIcon(false), label],
  );

  wrapper.append(button);
  return button;
}
