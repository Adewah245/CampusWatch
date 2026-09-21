/**
 * First-run registration page behaviour.
 *
 * This page creates the deployment's first account, which the backend makes an
 * administrator of the institution it also creates. It is a bootstrap step, not
 * a sign-up form: the backend serves it only while no account exists, and
 * answers 409 afterwards. Both outcomes are normal, so the page has a designed
 * state for each rather than treating the closed case as a failure.
 */

import { api, ApiError } from '../api.js';
import { getToken } from '../session.js';
import { mustFind, withBusy, brandMark, attachPasswordToggle } from '../ui.js';

const form = mustFind('#register-form');
const errorBox = mustFind('#register-error');
const closedBox = mustFind('#register-closed');
const submitButton = mustFind('#register-submit');

// Show/hide controls for both password fields. Confirming a password you cannot
// read is the commonest way to fail this form, and on this page a failure costs
// the whole account: registration succeeds exactly once.
attachPasswordToggle(mustFind('#password'));
attachPasswordToggle(mustFind('#confirm-password'));

/**
 * The shortest password the backend accepts.
 *
 * Kept in step with `MinPasswordLength` in backend/internal/service/auth.go.
 * Checking here as well is a courtesy — it saves a round trip and puts the
 * message next to the field — but the backend enforces it either way, because a
 * client-side check is a convenience and never a control.
 */
const MIN_PASSWORD_LENGTH = 12;

/**
 * The longest password the backend accepts.
 *
 * Kept in step with `maxPasswordBytes` in backend/internal/service/auth.go,
 * which exists because bcrypt silently ignores everything past 72 bytes. A
 * passphrase longer than this is a normal thing to choose and the backend
 * refuses it, so the form has to say so up front rather than let the request
 * come back rejected.
 */
const MAX_PASSWORD_BYTES = 72;

/**
 * Length of a string in bytes, matching Go's `len()`.
 *
 * The backend measures this password in bytes, so measuring it in characters
 * here would disagree with it on any non-ASCII password — a check that passes
 * in the browser and fails on the server is worse than no check.
 *
 * @param {string} value
 * @returns {number}
 */
function byteLength(value) {
  return new TextEncoder().encode(value).length;
}

/**
 * Derives an institution's slug from its name.
 *
 * Mirrors `slugify` in backend/internal/service/institutions.go: lowercase,
 * every run of anything else becomes a single dash, and the result is trimmed
 * of dashes. The server derives the stored slug this way, so a name that
 * slugifies to nothing has no slug to store and is rejected.
 *
 * @param {string} value
 * @returns {string}
 */
function slugify(value) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * Swaps the placeholder mark in the markup for the configured brand mark.
 *
 * This page sits in pages/, so asset paths get a "../" prefix — `brandMark` is
 * the one place that knows the depth.
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
 * Reports that registration is closed and removes the form.
 *
 * The form is hidden rather than disabled: leaving filled-in fields on screen
 * invites the operator to keep trying a request that cannot succeed. The message
 * that replaces it says only that an account already exists, because reaching
 * the dashboard from here needs credentials this page must not ask for — an
 * operator who needs them already knows the sign-in page is their sibling.
 */
function showClosed() {
  clearError();
  form.hidden = true;
  closedBox.hidden = false;
  // tabindex="-1" in the markup makes this focusable, so a screen reader is
  // moved to the explanation rather than being left on a form that just
  // disappeared.
  closedBox.focus();
}

/**
 * Checks the form, returning a message for the first problem found.
 *
 * @returns {string} an error message, or "" when the form is acceptable
 */
function validate(values) {
  if (!values.institution) return 'Enter the institution this deployment monitors.';
  // The slug is derived from the name and stored, so a name made only of
  // punctuation or symbols has nothing for the server to derive.
  if (!slugify(values.institution)) {
    return 'Include at least one letter or digit in the institution name.';
  }
  if (!values.first_name) return 'Enter a first name.';
  if (!values.last_name) return 'Enter a last name.';
  if (!values.email) return 'Enter an email address.';
  // Deliberately shallow. The address is an identifier here, not something this
  // page can prove exists; a stricter pattern would reject valid addresses.
  if (!values.email.includes('@')) return 'Enter a valid email address.';
  if (!values.password) return 'Choose a password.';
  if (byteLength(values.password) < MIN_PASSWORD_LENGTH) {
    return `Choose a password of at least ${MIN_PASSWORD_LENGTH} characters.`;
  }
  if (byteLength(values.password) > MAX_PASSWORD_BYTES) {
    return `Choose a password of at most ${MAX_PASSWORD_BYTES} characters.`;
  }
  if (values.password !== values.confirm_password) return 'The two passwords do not match.';
  return '';
}

form.addEventListener('submit', async (event) => {
  event.preventDefault();
  clearError();

  // Named exactly as the backend's RegisterRequest expects them, so the form
  // fields and the request body cannot drift apart silently.
  const values = {
    institution: form.institution.value.trim(),
    first_name: form.first_name.value.trim(),
    last_name: form.last_name.value.trim(),
    email: form.email.value.trim(),
    password: form.password.value,
    confirm_password: form.confirm_password.value,
  };

  const problem = validate(values);
  if (problem) {
    showError(problem);
    return;
  }

  // Built field by field rather than by spreading the form values, so that
  // confirm_password cannot leak into the request body if the backend's
  // RegisterRequest ever grows a field of that name.
  const account = {
    institution: values.institution,
    first_name: values.first_name,
    last_name: values.last_name,
    email: values.email,
    password: values.password,
  };

  await withBusy(submitButton, async () => {
    try {
      await api.register(account);
      // No session is returned by design, so the operator signs in with the
      // credentials they just chose. `replace` keeps this spent page out of the
      // back history.
      window.location.replace('login.html?registered=1');
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.isNetworkError) {
          showError(error.message);
        } else if (error.status === 409) {
          // The expected answer on every deployment that is already set up.
          showClosed();
          return;
        } else if (error.status === 429) {
          showError('Too many attempts. Wait a minute and try again.');
        } else if (error.status === 404) {
          // Two causes produce this, and they need different fixes. Either the
          // request never reached the CampusWatch server — the page was opened
          // from a file or an editor's live-preview server, so the relative
          // /api/... URL resolved against that origin instead — or it reached a
          // build of the server from before this route existed. Routes live in
          // the binary while the page is read from disk, so an old server serves
          // the new page happily and then has nowhere to send the request.
          showError(
            'The CampusWatch server does not have this route. Either the page is not being ' +
              'served by it — open the http://localhost:8080 address it prints when it starts, ' +
              'not a file or another tool — or it is running an older build. Restart the server ' +
              'and try again.',
          );
        } else {
          // The backend explains its refusals in plain text (`http.Error`), and
          // that sentence is the whole value of this branch: "institution name
          // must contain a letter or digit" tells the operator what to change,
          // where "Bad Request" leaves them guessing at five fields. The status
          // is included because the difference between a 400 and a 500 is the
          // difference between fixing the form and fixing the server.
          const serverSaid = typeof error.body === 'string' ? error.body.trim() : '';
          showError(
            serverSaid
              ? `The server rejected those details (${error.status}): ${serverSaid}`
              : `The server rejected those details (${error.status}). Check each field and try again.`,
          );
        }
      } else {
        showError('Registration failed unexpectedly. Please try again.');
        console.error('Unexpected registration failure', error);
      }
      form.password.value = '';
      form.confirm_password.value = '';
      form.password.focus();
    }
  });
});

/**
 * If a session already exists the deployment is definitely set up, so there is
 * nothing for this page to do. Send the operator where they were going instead
 * of making them submit a form that would only come back 409.
 */
if (getToken()) {
  window.location.replace('dashboard.html');
} else {
  form.institution.focus();
}
