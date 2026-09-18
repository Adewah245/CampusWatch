/**
 * Systems list page.
 *
 * Filtering and searching happen in the browser against the full list returned
 * by `GET /api/v1/systems`. The backend's list endpoint takes no documented
 * filter parameters, and the fleet is bounded by the size of one institution,
 * so fetching once and filtering locally is both correct today and instant for
 * the operator.
 */

import { api, ApiError, describeError } from '../api.js';
import { mountShell } from '../shell.js';
import { canManage } from '../session.js';
import {
  el,
  render,
  table,
  statusBadge,
  formatRelative,
  formatDateTime,
  loadingState,
  errorState,
  emptyState,
  toast,
  withBusy,
} from '../ui.js';

const content = mountShell(
  'systems.html',
  'Systems',
  'Every computer registered to your institution.',
);

/** Every system as last fetched; the table renders a filtered view of this. */
let allSystems = [];

/** Current filter state, owned by this module and read by `applyFilters`. */
const filters = { search: '', status: '' };
/**
 * The statuses a system can be in, per README section 45. Listed explicitly
 * rather than derived from the data so the filter offers statuses that are
 * currently empty — an operator looking for maintenance windows needs to see
 * "none" rather than an option that has vanished.
 */
const MONITOR_STATUSES = [
  'PENDING',
  'ONLINE',
  'OFFLINE',
  'INACTIVE',
  'MAINTENANCE',
  'SUSPENDED',
  'RETIRED',
];

/** Approves a system, then refreshes the row in place. */
async function approve(system, button) {
  await withBusy(button, async () => {
    try {
      await api.approveSystem(system.id);
      toast(`${system.hostname || system.id} approved.`, 'success');
      await load(false);
    } catch (error) {
      toast(
        error instanceof ApiError ? `Could not approve: ${error.message}` : 'Could not approve.',
        'error',
      );
    }  });
}

/** Builds the filter controls. */
function filterBar() {
  const searchInput = el('input', {
    class: 'input',
    type: 'search',
    id: 'filter-search',
    placeholder: 'Hostname, device name or OS',
    value: filters.search,
    oninput: (event) => {
      // Stored raw so the input is not rewritten while the operator types;
      // the comparison lowercases it at match time instead.
      filters.search = event.target.value.trim();
      applyFilters();
    },
  });

  const statusSelect = el(
    'select',
    {
      class: 'select',
      id: 'filter-status',
      onchange: (event) => {
        filters.status = event.target.value;
        applyFilters();
      },
    },
    [
      el('option', { value: '', text: 'All statuses' }),
      ...MONITOR_STATUSES.map((status) =>
        el('option', { value: status, text: status, selected: filters.status === status }),
      ),
    ],
  );

  return el('div', { class: 'filters' }, [
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'filter-search', text: 'Search' }),
      searchInput,
    ]),
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'filter-status', text: 'Status' }),
      statusSelect,
    ]),
  ]);
}

/** Returns the systems matching the active filters. */
function filteredSystems() {
  const needle = filters.search.toLowerCase();
  return allSystems.filter((system) => {
    if (filters.status && String(system.monitor_status).toUpperCase() !== filters.status) {
      return false;
    }
    if (!needle) return true;
    const haystack = [system.hostname, system.device_name, system.operating_system, system.id]
      .join(' ')
      .toLowerCase();
    return haystack.includes(needle);
  });
}

/** Renders the table body for the current filters into the existing wrapper. */
function applyFilters() {
  const host = content.querySelector('[data-role="systems-table"]');
  if (!host) return;
  render(host, [systemsTable(filteredSystems())]);
}

/** Builds the systems table, or an empty state when nothing matches. */
function systemsTable(systems) {
  if (allSystems.length === 0) {
    return emptyState('No systems have been registered yet.');
  }
  if (systems.length === 0) {
    return emptyState('No systems match the current filters.');
  }

  const manageable = canManage();

  return el('div', { class: 'table-wrap' }, [
    table(
      [
        { heading: 'Hostname' },
        { heading: 'Status' },
        { heading: 'Device' },
        { heading: 'Operating system' },
        { heading: 'Last seen' },
        { heading: 'Registered' },
        { heading: 'Actions' },
      ],
      systems.map((system) => [
        el('a', {
          class: 'table__link',
          href: `system.html?id=${encodeURIComponent(system.id)}`,
          text: system.hostname || system.id,
        }),
        statusBadge(system.monitor_status),
        system.device_name || '—',
        [system.operating_system, system.os_version].filter(Boolean).join(' ') || '—',
        formatRelative(system.last_seen_at),
        formatDateTime(system.created_at),
        // Approval is the only action that makes sense on a list row; anything
        // else would need the detail page's context. Hidden for roles the
        // backend would reject with 403.
        manageable && String(system.monitor_status).toUpperCase() === 'PENDING'
          ? el('button', {
              class: 'btn btn--small',
              type: 'button',
              text: 'Approve',
              onclick: (event) => approve(system, event.currentTarget),
            })
          : el('span', { class: 'muted', text: '—' }),
      ]),
      'Registered systems',
    ),
  ]);
}

/**
 * Fetches and renders the page.
 *
 * @param {boolean} showSpinner
 */
async function load(showSpinner) {
  if (showSpinner) render(content, [loadingState('Loading systems…')]);

  try {
    allSystems = (await api.systems()) || [];
    render(content, [
      filterBar(),
      el('section', { class: 'card' }, [
        el('div', { class: 'card__head' }, [
          el('h2', { class: 'card__title', text: 'Registered systems' }),
          el('span', {
            class: 'card__hint',
            text: `${allSystems.length} total`,
          }),
        ]),
        el('div', { 'data-role': 'systems-table' }, [systemsTable(filteredSystems())]),
      ]),
    ]);
  } catch (error) {
    if (error?.isUnauthorized) return;
    render(content, [errorState(describeError(error), () => load(true))]);
  }
}

await load(true);
