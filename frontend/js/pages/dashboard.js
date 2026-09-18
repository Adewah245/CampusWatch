/**
 * Dashboard page: fleet summary tiles, the systems list, and the issue queue.
 *
 * Data comes from three endpoints fetched together:
 *   GET /api/v1/dashboard/summary
 *   GET /api/v1/dashboard/systems
 *   GET /api/v1/issues
 */

import { api, describeError } from '../api.js';
import { DASHBOARD_REFRESH_MS } from '../config.js';
import { mountShell } from '../shell.js';
import {
  el,
  render,
  table,
  statusBadge,
  formatCount,
  formatRelative,
  loadingState,
  errorState,
  emptyState,
  toast,
} from '../ui.js';

const content = mountShell(
  'dashboard.html',
  'Dashboard',
  'Current state of the monitored fleet.',
);

/**
 * Builds one summary tile.
 *
 * @param {string} label
 * @param {number} value
 * @param {object} [options]
 * @param {string} [options.tone] ok | warn | danger | info
 * @param {string} [options.foot] small print under the number
 */
function statTile(label, value, options = {}) {
  return el('div', { class: `stat ${options.tone ? `stat--${options.tone}` : ''}` }, [
    el('span', { class: 'stat__label', text: label }),
    el('span', { class: 'stat__value', text: formatCount(value) }),
    options.foot ? el('span', { class: 'stat__foot', text: options.foot }) : null,
  ]);
}

/** Renders the summary tiles from `/dashboard/summary`. */
function summarySection(summary) {
  return el('section', { class: 'grid grid--stats', 'aria-label': 'Summary' }, [
    statTile('Total systems', summary.total_systems, { foot: 'registered in your institution' }),
    statTile('Online', summary.online_systems, { tone: 'ok', foot: 'reporting normally' }),
    statTile('Offline', summary.offline_systems, {
      tone: summary.offline_systems > 0 ? 'danger' : undefined,
      foot: 'no recent heartbeat',
    }),
    statTile('Inactive', summary.inactive_systems, {
      tone: summary.inactive_systems > 0 ? 'warn' : undefined,
      foot: 'online but unused',
    }),
    statTile('Maintenance', summary.maintenance_systems, {
      tone: 'info',
      foot: 'excluded from alerting',
    }),
    statTile('Active users', summary.active_users, { foot: 'sessions in use now' }),
    statTile('Open issues', summary.open_issues, {
      tone: summary.open_issues > 0 ? 'warn' : undefined,
      foot: 'unresolved',
    }),
    statTile('After-hours usage', summary.after_hours_usage, {
      foot: 'recorded outside open hours',
    }),
  ]);
}

/**
 * Renders the systems table.
 *
 * The full fleet is fetched and the most interesting rows are shown first and
 * limited to a readable number, since the systems page has the complete list
 * with filtering.
 */
function systemsSection(systems) {
  const ordered = [...systems].sort((a, b) => {
    // Problems first, then most recently seen. String comparison keeps this
    // independent of the backend's ordering.
    const rank = { OFFLINE: 0, PENDING: 1, INACTIVE: 2, MAINTENANCE: 3, ONLINE: 4 };
    const rankA = rank[a.monitor_status] ?? 5;
    const rankB = rank[b.monitor_status] ?? 5;
    if (rankA !== rankB) return rankA - rankB;
    return String(b.last_seen_at || '').localeCompare(String(a.last_seen_at || ''));
  });

  const shown = ordered.slice(0, 8);

  return el('section', { class: 'card' }, [
    el('div', { class: 'card__head' }, [
      el('h2', { class: 'card__title', text: 'Systems needing attention' }),
      el('a', { class: 'table__link', href: 'systems.html', text: 'View all systems' }),
    ]),
    shown.length === 0
      ? emptyState('No systems have been registered yet.')
      : el('div', { class: 'table-wrap' }, [
          table(
            [
              { heading: 'Hostname' },
              { heading: 'Status' },
              { heading: 'Operating system' },
              { heading: 'Last seen' },
            ],
            shown.map((system) => [
              el('a', {
                class: 'table__link',
                href: `system.html?id=${encodeURIComponent(system.id)}`,
                text: system.hostname || system.id,
              }),
              statusBadge(system.monitor_status),
              system.operating_system || '—',
              formatRelative(system.last_seen_at),
            ]),
            'Systems ordered by status, with the most recently seen first',
          ),
        ]),
  ]);
}

/** Renders the most recent open issues. */
function issuesSection(issues) {
  const open = issues
    .filter((issue) => !['RESOLVED', 'CLOSED'].includes(String(issue.status).toUpperCase()))
    .slice(0, 6);

  return el('section', { class: 'card' }, [
    el('div', { class: 'card__head' }, [
      el('h2', { class: 'card__title', text: 'Open issues' }),
      el('a', { class: 'table__link', href: 'issues.html', text: 'Manage issues' }),
    ]),
    open.length === 0
      ? emptyState('Nothing open. The fleet is clear.')
      : el('div', { class: 'table-wrap' }, [
          table(
            [
              { heading: 'Title' },
              { heading: 'Priority' },
              { heading: 'Status' },
              { heading: 'Raised' },
            ],
            open.map((issue) => [
              el('a', {
                class: 'table__link',
                href: `issues.html?id=${encodeURIComponent(issue.id)}`,
                text: issue.title || issue.id,
              }),
              statusBadge(issue.priority),
              statusBadge(issue.status),
              formatRelative(issue.created_at),
            ]),
            'Open issues, most recent first',
          ),
        ]),
  ]);
}

/**
 * Fetches everything the page needs and renders it.
 *
 * `firstLoad` distinguishes the initial spinner from a background refresh, so
 * the periodic refresh does not flash the page back to a loading state.
 *
 * @param {boolean} firstLoad
 */
async function load(firstLoad) {
  if (firstLoad) render(content, [loadingState('Loading dashboard…')]);

  try {
    // Fetched together rather than in sequence: the three are independent, and
    // this page is the one an operator opens first, so latency is noticeable.
    const [summary, systems, issues] = await Promise.all([
      api.dashboardSummary(),
      api.dashboardSystems(),
      api.issues(),
    ]);

    render(content, [
      summarySection(summary),
      systemsSection(systems || []),
      issuesSection(issues || []),
    ]);
  } catch (error) {
    if (error?.isUnauthorized) return; // api.js already redirected to sign-in.

    // A failed background refresh keeps the last good render on screen and says
    // so quietly; replacing a readable dashboard with an error would be worse.
    if (!firstLoad) {
      toast(`Refresh failed: ${describeError(error)}`, 'error');
      return;
    }
    render(content, [errorState(describeError(error), () => load(true))]);
  }
}

await load(true);

if (DASHBOARD_REFRESH_MS > 0) {
  // Paused while the tab is hidden so a dashboard left open overnight does not
  // poll the backend indefinitely.
  setInterval(() => {
    if (!document.hidden) load(false);
  }, DASHBOARD_REFRESH_MS);
}
