/**
 * Issues page: the maintenance queue.
 *
 * Filtering happens in the browser against the full list from `GET
 * /api/v1/issues`, matching the approach on the systems page.
 *
 * Limitation worth knowing: the backend exposes `POST /issues/{id}/updates` to
 * append a progress note but no endpoint to list them, and `GET /issues/{id}`
 * returns the issue alone. The update list shown here therefore contains only
 * the notes added in this browser session, and is labelled as such rather than
 * presented as the full history.
 */

import { api, ApiError, describeError } from '../api.js';
import { mountShell } from '../shell.js';
import { canManage, getSession } from '../session.js';
import {
  el,
  render,
  table,
  statusBadge,
  formatDateTime,
  formatRelative,
  humanise,
  loadingState,
  errorState,
  emptyState,
  toast,
  withBusy,
} from '../ui.js';

const content = mountShell(
  'issues.html',
  'Issues',
  'Faults and maintenance work raised against monitored systems.',
);

/** Every issue as last fetched. */
let allIssues = [];
/** Every system, used to label issues and to populate the create form. */
let allSystems = [];
/** Notes appended during this browser session, keyed by issue id. */
const sessionUpdates = new Map();
/** The issue currently expanded, or null. */
let selectedIssueId = null;

const ISSUE_STATUSES = ['OPEN', 'ACKNOWLEDGED', 'IN_PROGRESS', 'RESOLVED', 'CLOSED'];
const PRIORITIES = ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'];

const filters = { status: '', priority: '', search: '' };

/** Finds a system's hostname for display, falling back to the raw id. */
function hostnameFor(systemId) {
  const system = allSystems.find((candidate) => candidate.id === systemId);
  return system?.hostname || systemId || '—';
}

/** Builds the filter controls. */
function filterBar() {
  const search = el('input', {
    class: 'input',
    type: 'search',
    id: 'issue-search',
    placeholder: 'Title or description',
    value: filters.search,
    oninput: (event) => {
      filters.search = event.target.value.trim();
      applyFilters();
    },
  });

  const status = el(
    'select',
    {
      class: 'select',
      id: 'issue-status-filter',
      onchange: (event) => {
        filters.status = event.target.value;
        applyFilters();
      },
    },
    [
      el('option', { value: '', text: 'All statuses' }),
      ...ISSUE_STATUSES.map((value) =>
        el('option', { value, text: humanise(value), selected: filters.status === value }),
      ),
    ],
  );

  const priority = el(
    'select',
    {
      class: 'select',
      id: 'issue-priority-filter',
      onchange: (event) => {
        filters.priority = event.target.value;
        applyFilters();
      },
    },
    [
      el('option', { value: '', text: 'All priorities' }),
      ...PRIORITIES.map((value) =>
        el('option', { value, text: humanise(value), selected: filters.priority === value }),
      ),
    ],
  );

  return el('div', { class: 'filters' }, [
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'issue-search', text: 'Search' }),
      search,
    ]),
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'issue-status-filter', text: 'Status' }),
      status,
    ]),
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'issue-priority-filter', text: 'Priority' }),
      priority,
    ]),
  ]);
}

/** Returns the issues matching the active filters. */
function filteredIssues() {
  const needle = filters.search.toLowerCase();
  return allIssues.filter((issue) => {
    if (filters.status && String(issue.status).toUpperCase() !== filters.status) return false;
    if (filters.priority && String(issue.priority).toUpperCase() !== filters.priority) return false;
    if (!needle) return true;
    return `${issue.title} ${issue.description}`.toLowerCase().includes(needle);
  });
}

/** Re-renders the list without disturbing the filter inputs. */
function applyFilters() {
  const host = content.querySelector('[data-role="issues-list"]');
  if (host) render(host, [issuesCard()]);
}

/** The issues table plus the detail panel for whichever issue is open. */
function issuesCard() {
  const issues = filteredIssues();

  return el('div', {}, [
    issues.length === 0
      ? emptyState(
          allIssues.length === 0
            ? 'No issues have been raised.'
            : 'No issues match the current filters.',
        )
      : el('div', { class: 'table-wrap' }, [
          table(
            [
              { heading: 'Title' },
              { heading: 'System' },
              { heading: 'Priority' },
              { heading: 'Status' },
              { heading: 'Updated' },
            ],
            issues.map((issue) => {
              const isOpen = issue.id === selectedIssueId;
              return [
                el('button', {
                  class: 'table__link',
                  type: 'button',
                  text: issue.title || issue.id,
                  'aria-expanded': isOpen ? 'true' : 'false',
                  onclick: () => {
                    selectedIssueId = isOpen ? null : issue.id;
                    applyFilters();
                  },
                }),
                hostnameFor(issue.system_id),
                statusBadge(issue.priority),
                statusBadge(issue.status),
                formatRelative(issue.updated_at),
              ];
            }),
            'Issues matching the current filters',
          ),
        ]),
    selectedIssueId ? issueDetail(issues.find((issue) => issue.id === selectedIssueId)) : null,
  ]);
}

/** The expanded detail panel for one issue. */
function issueDetail(issue) {
  if (!issue) return null;

  const manageable = canManage();

  return el('section', { class: 'card', 'aria-label': `Issue ${issue.title}` }, [
    el('div', { class: 'card__head' }, [
      el('h3', { class: 'card__title', text: issue.title }),
      statusBadge(issue.status),
    ]),
    el('dl', { class: 'detail-list' }, [
      el('dt', { text: 'System' }),
      el('dd', {}, [
        el('a', {
          class: 'table__link',
          href: `system.html?id=${encodeURIComponent(issue.system_id)}`,
          text: hostnameFor(issue.system_id),
        }),
      ]),
      el('dt', { text: 'Priority' }),
      el('dd', {}, [statusBadge(issue.priority)]),
      el('dt', { text: 'Raised' }),
      el('dd', {
        text: `${formatDateTime(issue.created_at)}${issue.reported_by ? ` by ${issue.reported_by}` : ''}`,
      }),
      el('dt', { text: 'Last updated' }),
      el('dd', { text: formatDateTime(issue.updated_at) }),
    ]),
    issue.description
      ? el('div', { class: 'field' }, [
          el('span', { class: 'field__label', text: 'Description' }),
          el('p', { text: issue.description }),
        ])
      : null,
    manageable ? updateForm(issue) : null,
    updateHistory(issue),
  ]);
}

/** The status-change and note form for an issue. */
function updateForm(issue) {
  const statusSelect = el(
    'select',
    { class: 'select', id: `issue-status-${issue.id}` },
    ISSUE_STATUSES.map((value) =>
      el('option', {
        value,
        text: humanise(value),
        selected: String(issue.status).toUpperCase() === value,
      }),
    ),
  );

  const comment = el('textarea', {
    class: 'textarea',
    id: `issue-comment-${issue.id}`,
    placeholder: 'What was done, or what happens next…',
  });

  const submit = el('button', { class: 'btn', type: 'button', text: 'Post update' });
  submit.addEventListener('click', () =>
    withBusy(submit, async () => {
      const commentText = comment.value.trim();
      const status = statusSelect.value;
      const statusChanged = status !== String(issue.status).toUpperCase();

      if (!commentText && !statusChanged) {
        toast('Change the status or write a note before posting.', 'error');
        return;
      }

      try {
        // `updated_by` is taken from the local session rather than left to the
        // backend, which accepts it unvalidated. The backend's audit log remains
        // the authoritative record of who acted.
        const update = await api.addIssueUpdate(issue.id, {
          status: statusChanged ? status : undefined,
          comment: commentText,
          updated_by: getSession()?.user_id || undefined,
        });

        rememberUpdate(issue.id, update);
        toast('Issue updated.', 'success');

        // The status may have changed, so refetch before re-rendering rather
        // than guessing the server's new state.
        await refreshIssue(issue.id);
        comment.value = '';
        applyFilters();
      } catch (error) {
        toast(
          error instanceof ApiError ? `Could not update: ${error.message}` : 'Could not update.',
          'error',
        );
      }
    }),
  );

  return el('div', { class: 'field' }, [
    el('label', { class: 'field__label', for: `issue-status-${issue.id}`, text: 'Progress' }),
    el('div', { class: 'form-row' }, [
      el('div', {}, [statusSelect]),
      el('div', {}, [comment]),
    ]),
    el('div', { class: 'form-actions' }, [submit]),
  ]);
}

/** Records a posted update locally. */
function rememberUpdate(issueId, update) {
  if (!update) return;
  const list = sessionUpdates.get(issueId) || [];
  list.push(update);
  sessionUpdates.set(issueId, list);
}

/** Shows the notes posted in this session, clearly labelled as partial. */
function updateHistory(issue) {
  const updates = sessionUpdates.get(issue.id) || [];
  if (updates.length === 0) {
    return el('p', {
      class: 'card__hint',
      text: 'Earlier progress notes are stored by the backend but are not exposed by an API endpoint, so only notes added in this session appear here.',
    });
  }

  return el('div', {}, [
    el('span', { class: 'field__label', text: 'Notes added in this session' }),
    el('ul', {}, updates.map((update) =>
      el('li', {}, [
        el('span', {
          text: `${formatDateTime(update.created_at)}${update.status ? ` — ${humanise(update.status)}` : ''}`,
        }),
        update.comment ? el('p', { text: update.comment }) : null,
      ]),
    )),
  ]);
}

/** Refetches one issue and updates the local cache. */
async function refreshIssue(issueId) {
  const fresh = await api.issue(issueId);
  allIssues = allIssues.map((issue) => (issue.id === issueId ? fresh : issue));
}

/** The form for raising a new issue. */
function createForm() {
  const systemSelect = el(
    'select',
    { class: 'select', id: 'new-issue-system' },
    [
      el('option', { value: '', text: 'Select a system…' }),
      ...allSystems.map((system) =>
        el('option', { value: system.id, text: system.hostname || system.id }),
      ),
    ],
  );

  const title = el('input', { class: 'input', id: 'new-issue-title', type: 'text' });
  const description = el('textarea', { class: 'textarea', id: 'new-issue-description' });
  const priority = el(
    'select',
    { class: 'select', id: 'new-issue-priority' },
    PRIORITIES.map((value) =>
      el('option', { value, text: humanise(value), selected: value === 'MEDIUM' }),
    ),
  );

  const submit = el('button', { class: 'btn', type: 'submit', text: 'Raise issue' });

  const form = el('form', {}, [
    el('div', { class: 'form-row' }, [
      el('div', { class: 'field' }, [
        el('label', { class: 'field__label', for: 'new-issue-system', text: 'System' }),
        systemSelect,
      ]),
      el('div', { class: 'field' }, [
        el('label', { class: 'field__label', for: 'new-issue-priority', text: 'Priority' }),
        priority,
      ]),
    ]),
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'new-issue-title', text: 'Title' }),
      title,
    ]),
    el('div', { class: 'field' }, [
      el('label', { class: 'field__label', for: 'new-issue-description', text: 'Description' }),
      description,
    ]),
    el('div', { class: 'form-actions' }, [submit]),
  ]);

  form.addEventListener('submit', async (event) => {
    event.preventDefault();

    if (!systemSelect.value || !title.value.trim()) {
      toast('Choose a system and give the issue a title.', 'error');
      return;
    }

    await withBusy(submit, async () => {
      try {
        await api.createIssue({
          system_id: systemSelect.value,
          title: title.value.trim(),
          description: description.value.trim(),
          priority: priority.value,
          // Left to the backend to attribute if it prefers; sent for parity with
          // the issue-update flow.
          reported_by: getSession()?.user_id || undefined,
        });
        toast('Issue raised.', 'success');
        form.reset();
        await load(false);
      } catch (error) {
        toast(
          error instanceof ApiError ? `Could not raise issue: ${error.message}` : 'Could not raise issue.',
          'error',
        );
      }
    });
  });

  return el('details', { class: 'card' }, [
    el('summary', { class: 'card__title', text: 'Raise a new issue' }),
    form,
  ]);
}

/**
 * Fetches and renders the page.
 *
 * @param {boolean} showSpinner
 */
async function load(showSpinner) {
  if (showSpinner) render(content, [loadingState('Loading issues…')]);

  try {
    // Systems are fetched alongside issues because every issue row needs a
    // hostname and the create form needs the full list.
    const [issues, systems] = await Promise.all([api.issues(), api.systems()]);
    allIssues = issues || [];
    allSystems = systems || [];

    render(content, [
      filterBar(),
      el('div', { 'data-role': 'issues-list' }, [issuesCard()]),
      canManage() ? createForm() : null,
    ]);
  } catch (error) {
    if (error?.isUnauthorized) return;
    render(content, [errorState(describeError(error), () => load(true))]);
  }
}

// Opening the page with `?id=` — as the dashboard's issue links do — expands
// that issue immediately.
selectedIssueId = new URLSearchParams(window.location.search).get('id');

await load(true);
