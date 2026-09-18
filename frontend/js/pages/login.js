/**
 * Sign-in page behaviour.
 */

import { api, ApiError } from '../api.js';
import { storeSession, getToken, clearSession } from '../session.js';
import { mustFind, withBusy, brandMark } from '../ui.js';

const form = mustFind('#login-form');
const errorBox = mustFind('#login-error');
const submitButton = mustFind('#login-submit');

/**
 * Swaps the placeholder mark in the markup for the configured brand mark.
 *
 * The markup ships with the bundled vector mark so the page is never blank;
 * this upgrades it to the institution's logo, and `brandMark` falls back again
 * if that file cannot be loaded. `index.html` sits at the frontend root, so no
 * path prefix is needed.
 */
const brandSlot = document.querySelector('[data-brand="login"]');
if (brandSlot) {
  brandSlot.replaceWith(brandMark('', { size: 40, className: null }));
}

/** Shows an error message above the form. */
function showError(message) {
  errorBox.textContent = message;
  errorBox.hidden = false;
}

/** Hides the error message. */
function clearError() {
  errorBox.textContent = '';
  errorBox.hidden = true;
}

/**
 * Where to send the operator after a successful sign-in.
 *
 * Only a bare page filename plus query string is honoured. An absolute URL in
 * `?next=` would turn this page into an open redirect, so anything containing a
 * scheme or a path separator is discarded in favour of the dashboard.
 */
function resolveNextPage() {
  const requested = new URLSearchParams(window.location.search).get('next');
  if (!requested) return 'pages/dashboard.html';
  if (/^[a-z][a-z0-9+.-]*:/i.test(requested) || requested.includes('//')) {
    return 'pages/dashboard.html';
  }
  return `pages/${requested.replace(/^\/+/, '')}`;
}

/** Explains why the operator was returned to this page, if they were. */
function explainReturn() {
  const reason = new URLSearchParams(window.location.search).get('reason');
  if (reason === 'expired') {
    showError('Your session has ended. Please sign in again.');
  }
}

form.addEventListener('submit', async (event) => {
  event.preventDefault();
  clearError();

  const email = form.email.value.trim();
  const password = form.password.value;

  if (!email || !password) {
    showError('Enter both your email address and password.');
    return;
  }

  await withBusy(submitButton, async () => {
    try {
      const session = await api.login(email, password);
      storeSession(session);
      // `replace` rather than `assign`: the sign-in page should not remain in
      // the back history where a stray Back press would return to it.
      window.location.replace(resolveNextPage());
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.isNetworkError) {
          showError(error.message);
        } else if (error.status === 401) {
          // Deliberately does not distinguish "unknown email" from "wrong
          // password"; the backend does not, and guessing here would leak
          // which addresses have accounts.
          showError('Those credentials were not accepted.');
        } else if (error.status === 429) {
          showError('Too many sign-in attempts. Wait a minute and try again.');
        } else {
          showError(error.message);
        }
      } else {
        showError('Sign-in failed unexpectedly. Please try again.');
        // Kept for the browser console so an unexpected failure is diagnosable
        // without asking the operator to reproduce it.
        console.error('Unexpected sign-in failure', error);
      }
      form.password.value = '';
      form.password.focus();
    }
  });
});

/**
 * If a token is already held, skip the form.
 *
 * A token present but rejected is handled by the API client, which clears the
 * session and redirects back here with a reason.
 */
if (getToken()) {
  window.location.replace(resolveNextPage());
} else {
  explainReturn();
  // Nothing to clear, but this also normalises state after a previously
  // rejected token left a partial session behind.
  clearSession();
  form.email.focus();
}
