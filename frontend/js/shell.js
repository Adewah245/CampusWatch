/**
 * The shared authenticated page shell.
 *
 * Every page behind the sign-in calls `mountShell()` before it renders its own
 * content. That gives one definition of the header, navigation, operator
 * identity and sign-out control, so adding a page means editing one list.
 */

import { el, mustFind, render, markActiveNav, brandMark } from './ui.js';
import { getSession, signOut, requireSession } from './session.js';
import { request } from './api.js';

/** Pages shown in the main navigation, in display order. */
const NAV_ITEMS = [
  { href: 'dashboard.html', label: 'Dashboard' },
  { href: 'systems.html', label: 'Systems' },
  { href: 'issues.html', label: 'Issues' },
  { href: 'reports.html', label: 'Reports' },
  { href: 'hierarchy.html', label: 'Hierarchy' },
];

/**
 * Renders the header and navigation.
 *
 * @param {string} activePage filename of the page calling this
 */
function mountHeader(activePage) {
  const header = mustFind('[data-shell="header"]');
  const session = getSession();

  render(header, [
    el('a', { class: 'brand', href: 'dashboard.html' }, [
      // Pages live one level below the frontend root.
      brandMark('../', { size: 28 }),
      el('span', { class: 'brand__name', text: 'CampusWatch' }),
    ]),
    el('nav', { class: 'nav', 'aria-label': 'Main' }, NAV_ITEMS.map((item) =>
      el('a', { class: 'nav__link', href: item.href, text: item.label }),
    )),
    el('div', { class: 'header__account' }, [
      el('span', { class: 'header__role', text: session?.role ? session.role : 'signed in' }),
      el('button', {
        class: 'btn btn--ghost btn--small',
        type: 'button',
        text: 'Sign out',
        onclick: () => signOut(request),
      }),
    ]),
  ]);

  markActiveNav(activePage);
}

/**
 * Renders the page title block.
 *
 * @param {string} title
 * @param {string} [subtitle]
 * @param {Node[]} [actions]
 */
export function mountPageHeading(title, subtitle, actions = []) {
  const host = mustFind('[data-shell="heading"]');
  render(host, [
    el('div', {}, [
      el('h1', { class: 'page__title', text: title }),
      subtitle ? el('p', { class: 'page__subtitle', text: subtitle }) : null,
    ]),
    actions.length ? el('div', { class: 'page__actions' }, actions) : null,
  ]);
}

/**
 * Prepares the page: enforces the session, renders the header, and hands back
 * the container the page should render its content into.
 *
 * @param {string} activePage filename of the calling page
 * @param {string} [title] page heading
 * @param {string} [subtitle]
 * @returns {HTMLElement|null} the content container, or null when redirecting
 */
export function mountShell(activePage, title, subtitle) {
  if (!requireSession()) return null;

  mountHeader(activePage);
  if (title) mountPageHeading(title, subtitle);

  return mustFind('[data-shell="content"]');
}

/**
 * Reads a query-string parameter from the current URL.
 *
 * @param {string} name
 * @returns {string|null}
 */
export function queryParam(name) {
  return new URLSearchParams(window.location.search).get(name);
}
