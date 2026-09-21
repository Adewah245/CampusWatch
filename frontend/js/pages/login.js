/**
 * Sign-in page behaviour.
 */

import { api, ApiError } from '../api.js';
import { storeSession, getToken, clearSession } from '../session.js';
import { mustFind, withBusy, brandMark, attachPasswordToggle } from '../ui.js';

const form = mustFind('#login-form');
const errorBox = mustFind('#login-error');
const noticeBox = mustFind('#login-notice');
const submitButton = mustFind('#login-submit');

// Lets the operator check what they typed before submitting. The backend
// deliberately does not distinguish an unknown address from a wrong password,
// so a mistyped password is otherwise indistinguishable from a wrong one.
attachPasswordToggle(form.password);

/**
 * Swaps the placeholder mark in the markup for the configured brand mark.
 *
 * The markup ships with the bundled vector mark so the page is never blank;
 * this upgrades it to the institution's logo, and `brandMark` falls back again
 * if that file cannot be loaded. This page sits in pages/, so asset paths get a
 * "../" prefix — `brandMark` is the one place that knows the depth.
 */
const brandSlot = document.querySelector('[data-brand="login"]');
if (brandSlot) {
  brandSlot.replaceWith(brandMark('../', { size: 40, className: null }));
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
 * Every dashboard page is a sibling of this one, so the answer is a bare
 * filename in this directory — `redirectToLogin` in session.js writes `?next=`
 * as the requested page's last path segment, which is exactly that, query
 * string included (system.html?id=...).
 *
 * The pattern is the open-redirect guard. A value that is not a plain sibling
 * filename is discarded rather than repaired: an absolute URL would turn this
 * page into a redirector, and `..` or a leading slash would walk out of the
 * dashboard entirely. Anything unexpected falls back to the dashboard, which is
 * a page every signed-in operator is entitled to reach.
 */
function resolveNextPage() {
  const requested = new URLSearchParams(window.location.search).get('next');
  if (!requested) return 'dashboard.html';
  if (!/^[a-z0-9][a-z0-9._-]*\.html(\?[^#]*)?$/i.test(requested)) return 'dashboard.html';
  return requested;
}

/** Explains why the operator was returned to this page, if they were. */
function explainReturn() {
  const params = new URLSearchParams(window.location.search);

  if (params.get('reason') === 'expired') {
    showError('Your session has ended. Please sign in again.');
  }

  // Registration deliberately stops short of signing the operator in, so this
  // is the hand-off between the two pages: it confirms the account exists and
  // tells them which credentials to use.
  if (params.get('registered') === '1') {
    noticeBox.textContent =
      'Administrator account created. Sign in with the credentials you just chose.';
    noticeBox.hidden = false;
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
