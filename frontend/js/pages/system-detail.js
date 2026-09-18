/**
 * System detail page.
 *
 * Shows one machine: its registration details, its most recent health report,
 * its login history and its event log. The machine is chosen with the `?id=`
 * query parameter.
 */

import { api, ApiError, describeError } from '../api.js';
import { mountShell, queryParam } from '../shell.js';
import { canManage } from '../session.js';
import {
  el,
  render,
  table,
  statusBadge,
  metricMeter,
  formatDateTime,
  formatRelative,
  formatDuration,
  humanise,
  loadingState,
  errorState,
  emptyState,
  toast,
  withBusy,
} from '../ui.js';

const systemId = queryParam('id');

const content = mountShell('systems.html', 'System', 'Loading…');

/** Statuses an operator may set by hand, per README section 45. */
const EDITABLE_STATUSES = [
  'PENDING',
  'ONLINE',
  'OFFLINE',
  'INACTIVE',
  'MAINTENANCE',
  'SUSPENDED',
  'RETIRED',
];

/**
 * Renders the health panel.
 *
 * A missing report is normal for a system that has never had an approved agent
 * report in, so it reads as "no data yet" rather than as an error.
 *
 * @param {object|null} health the HealthReport, or null when none exists yet
 */
function healthPanel(health) {
  if (!health) {
    return el('section', { class: 'card' }, [
      el('h2', { class: 'card__title', text: 'Health' }),
      emptyState(
        'No health report has been received. This is expected until the system’s agent is approved and reporting.',
      ),
    ]);
  }

  const battery =
    health.battery_percent === undefined || health.battery_percent === null
      ? null
      : metricMeter('Battery', health.battery_percent, 25, 10);

  return el('section', { class: 'card' }, [
    el('div', { class: 'card__head' }, [
      el('h2', { class: 'card__title', text: 'Health' }),
      el('span', { class: 'card__hint', text: `reported ${formatRelative(health.recorded_at)}` }),
    ]),
    el('div', { class: 'grid grid--halves' }, [
      el('div', {}, [
        metricMeter('CPU', health.cpu_percent),
        metricMeter('Memory', health.memory_percent),
        metricMeter('Disk', health.disk_percent),
        battery,
      ]),
      el('dl', { class: 'detail-list' }, [
        el('dt', { text: 'Agent health' }),
        el('dd', {}, [statusBadge(health.agent_health)]),
        el('dt', { text: 'Reported OS' }),
        el('dd', { text: health.operating_system || '—' }),
        el('dt', { text: 'Network' }),
        // Deliberately not `statusBadge`: ONLINE/OFFLINE are system statuses,
        // and labelling a network state with them would read as a claim about
        // the machine rather than about its connection.
        el('dd', {}, [
          el('span', {
            class: `badge badge--${health.network_connected ? 'ok' : 'danger'}`,
            text: health.network_connected ? 'Connected' : 'Not connected',
          }),
        ]),
        el('dt', { text: 'Uptime' }),
        el('dd', { text: formatDuration(health.uptime_seconds) }),
        el('dt', { text: 'Recorded at' }),
        el('dd', { text: formatDateTime(health.recorded_at) }),
      ]),
    ]),
  ]);
}

/**
 * Renders the login history.
 *
 * @param {Array<object>} sessions
 */
function sessionsPanel(sessions) {
  return el('section', { class: 'card' }, [
    el('div', { class: 'card__head' }, [
      el('h2', { class: 'card__title', text: 'Login history' }),
      el('span', { class: 'card__hint', text: `${sessions.length} recorded` }),
    ]),
    sessions.length === 0
      ? emptyState('No logins have been recorded for this system.')
      : el('div', { class: 'table-wrap' }, [
          table(
            [
              { heading: 'Account' },
              { heading: 'State' },
              { heading: 'Signed in' },
              { heading: 'Signed out' },
              { heading: 'Duration', numeric: true },
            ],
            sessions.map((session) => [
              // Usernames come from the monitored machine, so they are written
              // as text and never as markup.
              el('span', { class: 'mono', text: session.username }),
              statusBadge(session.state),
              formatDateTime(session.login_at),
              session.logout_at ? formatDateTime(session.logout_at) : '—',
              session.duration_seconds ? formatDuration(session.duration_seconds) : '—',
            ]),
            'Login and logout history for this system',
          ),
        ]),
  ]);
}

/**
 * Renders the event log.
 *
 * @param {Array<object>} events
 */
function eventsPanel(events) {
  return el('section', { class: 'card' }, [
    el('div', { class: 'card__head' }, [
      el('h2', { class: 'card__title', text: 'Events' }),
      el('span', { class: 'card__hint', text: `${events.length} recent` }),
    ]),
    events.length === 0
      ? emptyState('No events have been reported for this system.')
      : el('div', { class: 'table-wrap' }, [
          table(
            [
              { heading: 'Event' },
              { heading: 'Occurred' },
              { heading: 'Detail' },
            ],
            events.map((event) => [
              statusBadge(event.event_type),
              formatDateTime(event.occurred_at),
              // The payload is opaque JSON from the agent, so it is shown
              // verbatim inside a <pre>. textContent keeps it inert.
              el('pre', {
                class: 'mono',
                text: formatPayload(event.payload),
              }),
            ]),
            'Events reported for this system',
          ),
        ]),
  ]);
}

/**
 * Renders an event payload compactly.
 *
 * An empty or absent payload is common for simple events and is shown as an
 * em dash rather than as `null`, which would look like a bug.
 *
 * @param {any} payload
 * @returns {string}
 */
function formatPayload(payload) {
  if (payload === null || payload === undefined) return '—';
  if (typeof payload === 'object' && Object.keys(payload).length === 0) return '—';
  try {
    return JSON.stringify(payload);
  } catch {
    return '(unreadable payload)';
  }
}

/** Builds the status editor shown to roles that may change a system. */
function statusEditor(system, onChanged) {
  const select = el(
    'select',
    { class: 'select', id: 'system-status' },
    EDITABLE_STATUSES.map((status) =>
      el('option', {
        value: status,
        text: humanise(status),
        selected: String(system.monitor_status).toUpperCase() === status,
      }),
    ),
  );

  const save = el('button', { class: 'btn btn--small', type: 'button', text: 'Save status' });
  save.addEventListener('click', () =>
    withBusy(save, async () => {
      try {
        await api.updateSystem(system.id, { monitor_status: select.value });
        toast('System status updated.', 'success');
        await onChanged();
      } catch (error) {
        toast(
          error instanceof ApiError ? `Could not update: ${error.message}` : 'Could not update.',
          'error',
        );
      }
    }),
  );

  return el('div', { class: 'field' }, [
    el('label', { class: 'field__label', for: 'system-status', text: 'Monitor status' }),
    el('div', { class: 'form-actions' }, [select, save]),
  ]);
}

/** Renders the registration details panel. */
function detailsPanel(system) {
  return el('section', { class: 'card' }, [
    el('div', { class: 'card__head' }, [
      el('h2', { class: 'card__title', text: 'Registration' }),
      statusBadge(system.monitor_status),
    ]),
    el('dl', { class: 'detail-list' }, [
      el('dt', { text: 'System ID' }),
      el('dd', { class: 'mono', text: system.id }),
      el('dt', { text: 'Hostname' }),
      el('dd', { text: system.hostname || '—' }),
      el('dt', { text: 'Device name' }),
      el('dd', { text: system.device_name || '—' }),
      el('dt', { text: 'Operating system' }),
      el('dd', { text: [system.operating_system, system.os_version].filter(Boolean).join(' ') || '—' }),
      el('dt', { text: 'System status' }),
      el('dd', {}, [statusBadge(system.system_status)]),
      el('dt', { text: 'Location' }),
      el('dd', { class: 'mono', text: system.location_id || '—' }),
      el('dt', { text: 'Last seen' }),
      el('dd', { text: `${formatRelative(system.last_seen_at)} (${formatDateTime(system.last_seen_at)})` }),
      el('dt', { text: 'Registered' }),
      el('dd', { text: formatDateTime(system.created_at) }),
    ]),
    canManage() ? statusEditor(system, () => load(false)) : null,
  ]);
}

/** Approves the system, which unblocks its agent. */
async function approve(system, button) {
  await withBusy(button, async () => {
    try {
      await api.approveSystem(system.id);
      toast('System approved. Its agent may now report.', 'success');
      await load(false);
    } catch (error) {
      toast(
        error instanceof ApiError ? `Could not approve: ${error.message}` : 'Could not approve.',
        'error',
      );
    }
  });
}

/**
 * Fetches and renders the page.
 *
 * @param {boolean} showSpinner
 */
async function load(showSpinner) {
  if (showSpinner) render(content, [loadingState('Loading system…')]);

  if (!systemId) {
    render(content, [errorState('No system was specified. Open a system from the systems list.')]);
    return;
  }

  // The system record is essential; the three supporting collections are not.
  // They are settled independently so that a missing health report or an empty
  // event log still leaves a usable page.
  let system;
  try {
    system = await api.system(systemId);
  } catch (error) {
    if (error?.isUnauthorized) return;
    // `describeError` already produces the right wording for a 404 and for a
    // role the backend refuses, so no special-casing is needed here.
    render(content, [errorState(describeError(error), () => load(true))]);
    return;
  }

  const [health, sessions, events] = await Promise.all([
    api.systemHealth(systemId).catch(() => null),
    api.systemSessions(systemId).catch(() => []),
    api.systemEvents(systemId).catch(() => []),
  ]);

  const heading = document.querySelector('[data-shell="heading"]');
  if (heading) {
    render(heading, [
      el('div', {}, [
        el('h1', { class: 'page__title', text: system.hostname || system.id }),
        el('p', {
          class: 'page__subtitle',
          text: [system.device_name, system.operating_system].filter(Boolean).join(' · ') ||
            'Registered system',
        }),
      ]),
      String(system.monitor_status).toUpperCase() === 'PENDING' && canManage()
        ? el('div', { class: 'page__actions' }, [
            el('button', {
              class: 'btn',
              type: 'button',
              text: 'Approve system',
              onclick: (event) => approve(system, event.currentTarget),
            }),
          ])
        : null,
    ]);
  }

  render(content, [
    el('div', { class: 'grid grid--split' }, [
      el('div', { class: 'stack' }, [
        healthPanel(health),
        sessionsPanel(sessions || []),
        eventsPanel(events || []),
      ]),
      detailsPanel(system),
    ]),
  ]);
}

await load(true);
