/**
 * Hierarchy page.
 *
 * The four levels — institutions, campuses, clusters and locations — have the
 * same shape: a list, a create form, and a parent to choose. They are driven
 * from one table of level definitions rather than four near-identical blocks,
 * so a change to how a level renders is made once.
 */

import { api, ApiError, describeError } from '../api.js';
import { mountShell, queryParam } from '../shell.js';
import { canManage } from '../session.js';
import {
  el,
  render,
  table,
  statusBadge,
  formatDateTime,
  humanise,
  loadingState,
  errorState,
  emptyState,
  toast,
  withBusy,
} from '../ui.js';

const content = mountShell(
  'hierarchy.html',
  'Hierarchy',
  'How campuses, clusters and locations are organised, and when they are open.',
);

/**
 * One entry per hierarchy level.
 *
 * `parent` names the property that holds the id of the containing record, which
 * is what the create form offers as a picker and what the list shows instead of
 * a bare id.
 */
const LEVELS = [
  {
    key: 'institutions',
    label: 'Institutions',
    list: () => api.institutions(),
    create: (body) => api.createInstitution(body),
    parent: null,
    fields: [{ name: 'name', label: 'Name', required: true }],
  },
  {
    key: 'campuses',
    label: 'Campuses',
    list: () => api.campuses(),
    create: (body) => api.createCampus(body),
    parent: { property: 'institution_id', level: 'institutions', label: 'Institution' },
    fields: [{ name: 'name', label: 'Name', required: true }],
  },
  {
    key: 'clusters',
    label: 'Clusters',
    list: () => api.clusters(),
    create: (body) => api.createCluster(body),
    parent: { property: 'campus_id', level: 'campuses', label: 'Campus' },
    fields: [{ name: 'name', label: 'Name', required: true }],
  },
  {
    key: 'locations',
    label: 'Locations',
    list: () => api.locations(),
    create: (body) => api.createLocation(body),
    parent: { property: 'cluster_id', level: 'clusters', label: 'Cluster' },
    fields: [
      { name: 'name', label: 'Name', required: true },
      { name: 'location_type', label: 'Type', placeholder: 'lab, office, library…' },
    ],
  },
];

/** The level currently shown; seeded from the URL so a tab is linkable. */
let activeKey = LEVELS.some((level) => level.key === queryParam('level'))
  ? queryParam('level')
  : 'institutions';

/** Records for every level that has been loaded, keyed by level name. */
const cache = new Map();

/** Returns the definition of the active level. */
function activeLevel() {
  return LEVELS.find((level) => level.key === activeKey) || LEVELS[0];
}

/** Looks up a parent record's display name, falling back to its id. */
function parentNameFor(level, record) {
  if (!level.parent) return null;
  const siblings = cache.get(level.parent.level) || [];
  return siblings.find((candidate) => candidate.id === record[level.parent.property])?.name;
}

/** Builds the tab strip. */
function tabs() {
  return el('div', { class: 'tabs', role: 'tablist' }, LEVELS.map((level) =>
    el('button', {
      class: `tab ${level.key === activeKey ? 'is-active' : ''}`,
      type: 'button',
      role: 'tab',
      'aria-selected': level.key === activeKey ? 'true' : 'false',
      text: level.label,
      onclick: () => {
        activeKey = level.key;
        const url = new URL(window.location.href);
        url.searchParams.set('level', activeKey);
        window.history.replaceState(null, '', url);
        load(false);
      },
    }),
  ));
}

/** Renders the table of records for a level. */
function levelTable(level, records) {
  if (records.length === 0) {
    return emptyState(`No ${level.label.toLowerCase()} have been created yet.`);
  }

  const columns = [{ heading: 'Name' }, { heading: 'Slug' }];
  if (level.parent) columns.push({ heading: level.parent.label });
  if (level.key === 'locations') columns.push({ heading: 'Type' });
  columns.push({ heading: 'Status' }, { heading: 'Created' });

  return el('div', { class: 'table-wrap' }, [
    table(
      columns,
      records.map((record) => {
        const cells = [
          record.name || record.id,
          el('span', { class: 'mono', text: record.slug || '—' }),
        ];
        if (level.parent) cells.push(parentNameFor(level, record) || record[level.parent.property]);
        if (level.key === 'locations') cells.push(record.location_type || '—');
        cells.push(statusBadge(record.status), formatDateTime(record.created_at));
        return cells;
      }),
      `${level.label} in the hierarchy`,
    ),
  ]);
}

/**
 * Builds the create form for a level.
 *
 * The parent picker is populated from the cache. If the parent level has not
 * loaded — which happens when the operator opens a deep tab directly — the
 * form says so rather than offering an empty picker that cannot succeed.
 */
function createForm(level) {
  const inputs = {};
  const rows = [];

  if (level.parent) {
    const parents = cache.get(level.parent.level) || [];
    const select = el(
      'select',
      { class: 'select', id: `new-${level.key}-parent` },
      [
        el('option', {
          value: '',
          text: parents.length ? `Select a ${level.parent.label.toLowerCase()}…` : 'None available',
        }),
        ...parents.map((parent) =>
          el('option', { value: parent.id, text: parent.name || parent.id }),
        ),
      ],
    );
    inputs[level.parent.property] = select;
    rows.push(
      el('div', { class: 'field' }, [
        el('label', {
          class: 'field__label',
          for: `new-${level.key}-parent`,
          text: level.parent.label,
        }),
        select,
      ]),
    );
  }

  for (const field of level.fields) {
    const input = el('input', {
      class: 'input',
      id: `new-${level.key}-${field.name}`,
      type: 'text',
      placeholder: field.placeholder || '',
    });
    inputs[field.name] = input;
    rows.push(
      el('div', { class: 'field' }, [
        el('label', { class: 'field__label', for: `new-${level.key}-${field.name}`, text: field.label }),
        input,
      ]),
    );
  }

  const submit = el('button', { class: 'btn', type: 'submit', text: `Create ${level.label.replace(/s$/, '').toLowerCase()}` });

  const form = el('form', {}, [
    el('div', { class: 'form-row' }, rows),
    el('div', { class: 'form-actions' }, [submit]),
  ]);

  form.addEventListener('submit', async (event) => {
    event.preventDefault();

    const body = {};
    for (const [name, input] of Object.entries(inputs)) {
      const value = input.value.trim();
      if (value) body[name] = value;
    }

    // Validate against the level's own required fields plus the parent.
    const missing = [];
    if (level.parent && !body[level.parent.property]) missing.push(level.parent.label);
    for (const field of level.fields) {
      if (field.required && !body[field.name]) missing.push(field.label);
    }
    if (missing.length) {
      toast(`Missing: ${missing.join(', ')}.`, 'error');
      return;
    }

    await withBusy(submit, async () => {
      try {
        await level.create(body);
        toast(`${humanise(level.label.replace(/s$/, ''))} created.`, 'success');
        form.reset();
        await load(false);
      } catch (error) {
        toast(
          error instanceof ApiError ? `Could not create: ${error.message}` : 'Could not create.',
          'error',
        );
      }
    });
  });

  return el('details', { class: 'card' }, [
    el('summary', { class: 'card__title', text: `Add ${level.label.toLowerCase().replace(/s$/, '')}` }),
    form,
  ]);
}

/**
 * Loads the active level, plus its parent level so the picker and the parent
 * name column can be filled in.
 *
 * @param {boolean} showSpinner
 */
async function load(showSpinner) {
  const level = activeLevel();
  if (showSpinner) render(content, [loadingState('Loading hierarchy…')]);

  try {
    const needed = [level];
    if (level.parent) {
      const parentLevel = LEVELS.find((candidate) => candidate.key === level.parent.level);
      if (parentLevel) needed.push(parentLevel);
    }

    await Promise.all(
      needed.map(async (target) => {
        cache.set(target.key, (await target.list()) || []);
      }),
    );

    render(content, [
      tabs(),
      el('section', { class: 'card' }, [
        el('div', { class: 'card__head' }, [
          el('h2', { class: 'card__title', text: level.label }),
          el('span', { class: 'card__hint', text: `${(cache.get(level.key) || []).length} total` }),
        ]),
        levelTable(level, cache.get(level.key) || []),
      ]),
      canManage() ? createForm(level) : null,
    ]);
  } catch (error) {
    if (error?.isUnauthorized) return;
    render(content, [errorState(describeError(error), () => load(true))]);
  }
}

await load(true);
