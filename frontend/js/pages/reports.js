/**
 * Reports page.
 *
 * `GET /api/v1/reports/{kind}?period={period}` returns a single Report object
 * containing every metric for that period; which metrics are meaningful
 * depends on the kind. Rather than hiding the irrelevant ones, each kind lists
 * the fields it actually populates and the rest are shown as well under a
 * secondary heading, so the page never claims a number it was not given.
 */

import { api, ApiError, describeError } from '../api.js';
import { mountShell, queryParam } from '../shell.js';
import {
  el,
  render,
  formatCount,
  formatDateTime,
  formatDuration,
  humanise,
  loadingState,
  errorState,
  toast,
} from '../ui.js';

const content = mountShell(
  'reports.html',
  'Reports',
  'Usage, uptime and issue summaries for a chosen period.',
);

/** Report kinds the backend serves. */
const KINDS = [
  { value: 'usage', label: 'Usage' },
  { value: 'uptime', label: 'Uptime' },
  { value: 'sessions', label: 'Sessions' },
  { value: 'issues', label: 'Issues' },
  { value: 'after-hours', label: 'After-hours' },
];

/** Periods the backend serves. */
const PERIODS = [
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

/** The metrics each kind is expected to populate, with how to format them. */
const KIND_FIELDS = {
  usage: [
    ['usage_seconds', 'Usage time', formatDuration],
    ['user_sessions', 'Sessions', formatCount],
  ],
  uptime: [
    ['uptime_seconds', 'Uptime', formatDuration],
    ['downtime_seconds', 'Downtime', formatDuration],
  ],
  sessions: [
    ['user_sessions', 'Sessions', formatCount],
    ['usage_seconds', 'Usage time', formatDuration],
  ],
  issues: [['issues_created', 'Issues raised', formatCount]],
  'after-hours': [
    ['after_hours_events', 'After-hours events', formatCount],
    ['usage_seconds', 'Usage time', formatDuration],
  ],
};

/**
 * Every metric on a Report, with its label and formatter. Used for the
 * "also reported" section, which keeps the raw figures visible no matter which
 * kind was requested.
 */
const ALL_FIELDS = {
  user_sessions: ['Sessions', formatCount],
  usage_seconds: ['Usage time', formatDuration],
  issues_created: ['Issues raised', formatCount],
  after_hours_events: ['After-hours events', formatCount],
  uptime_seconds: ['Uptime', formatDuration],
  downtime_seconds: ['Downtime', formatDuration],
};

/** The chosen kind and period; seeded from the URL so a report is linkable. */
const state = {
  kind: KINDS.some((k) => k.value === queryParam('kind')) ? queryParam('kind') : 'usage',
  period: PERIODS.some((p) => p.value === queryParam('period')) ? queryParam('period') : 'weekly',
};

/** Keeps the URL in step with the selected report so it can be shared. */
function syncUrl() {
  const url = new URL(window.location.href);
  url.searchParams.set('kind', state.kind);
  url.searchParams.set('period', state.period);
  window.history.replaceState(null, '', url);
}

/** The kind and period pickers. */
function controls() {
  const kindSelect = el(
    'select',
    {
      class: 'select',
      id: 'report-kind',
      onchange: (event) => {
        state.kind = event.target.value;
        syncUrl();
        load();
      },
    },
    KINDS.map((kind) =>
      el('option', { value: kind.value, text: kind.label, selected: state.kind === kind.value }),
    ),
  );

  const periodSelect = el(
    'select',
    {
      class: 'select',
      id: 'report-period',
      onchange: (event) => {
        state.period = event.target.value;
        syncUrl();
        load();
      },
    },
    PERIODS.map((period) =>
      el('option', {
        value: period.value,
        text: period.label,
        selected: state.period === period.value,
      }),
    ),
  );

  const print = el('button', {
    class: 'btn btn--ghost',
    type: 'button',
    text: 'Print',
    // The browser's own print dialogue handles pagination and PDF export; a
    // bespoke exporter would add a dependency for no gain.
    onclick: () => window.print(),
  });

  return el('div', { class: 'filters' }, [
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'report-kind', text: 'Report' }),
      kindSelect,
    ]),
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'report-period', text: 'Period' }),
      periodSelect,
    ]),
    el('div', { class: 'field' }, [el('span', { class: 'field__label', text: ' ' }), print]),
  ]);
}

/** A tile showing one metric. */
function metricTile(label, value, formatter) {
  return el('div', { class: 'stat' }, [
    el('span', { class: 'stat__label', text: label }),
    el('span', { class: 'stat__value', text: formatter(value) }),
  ]);
}

/** Renders a Report. */
function reportView(report, requestedKind) {
  const primary = KIND_FIELDS[requestedKind] || [];
  const primaryKeys = new Set(primary.map(([key]) => key));
  const secondary = Object.entries(ALL_FIELDS).filter(([key]) => !primaryKeys.has(key));

  return el('div', {}, [
    el('section', { class: 'card' }, [
      el('div', { class: 'card__head' }, [
        el('h2', { class: 'card__title', text: `${humanise(report.kind || requestedKind)} report` }),
        el('span', {
          class: 'card__hint',
          text: `${formatDateTime(report.from)} – ${formatDateTime(report.to)}`,
        }),
      ]),
      el('div', { class: 'grid grid--stats' }, primary.map(([key, label, formatter]) =>
        metricTile(label, report[key], formatter),
      )),
    ]),
    el('section', { class: 'card' }, [
      el('h3', { class: 'card__title', text: 'Also reported for this period' }),
      el('p', {
        class: 'card__hint',
        text: 'Every figure the backend returns, including those this report kind does not lead with.',
      }),
      el('div', { class: 'grid grid--stats' }, secondary.map(([key, [label, formatter]]) =>
        metricTile(label, report[key], formatter),
      )),
    ]),
  ]);
}

/**
 * Fetches and renders the selected report.
 *
 * The controls are re-rendered with the body so the pickers always reflect
 * `state`, which the change handlers mutate.
 */
async function load() {
  render(content, [
    controls(),
    el('div', { 'data-role': 'report-body' }, [loadingState('Building report…')]),
  ]);

  const body = content.querySelector('[data-role="report-body"]');
  try {
    const report = await api.report(state.kind, state.period);
    if (body) render(body, [reportView(report, state.kind)]);
  } catch (error) {
    if (error?.isUnauthorized) return;
    if (body) render(body, [errorState(describeError(error), () => load())]);
    if (error instanceof ApiError && error.status === 400) {
      toast('That report kind is not supported by the backend.', 'error');
    }
  }
}

await load();
