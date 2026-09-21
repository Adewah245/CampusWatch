/**
 * Session storage and access control for the dashboard.
 *
 * The backend issues an opaque session token from POST /api/v1/auth/login and
 * expects it back as `Authorization: Bearer <token>`. It is kept in
 * localStorage so a page reload does not sign the operator out.
 *
 * Note on threat model: a token in localStorage is readable by any script on
 * this origin, so it is only appropriate because the backend rejects
 * cross-origin requests (no CORS headers) and the dashboard is served from the
 * operator's own origin. The backend also exposes no cookie-based auth, so an
 * HttpOnly cookie is not an option without a backend change.
 */

import { TOKEN_STORAGE_KEY, SESSION_STORAGE_KEY } from './config.js';

/**
 * Reads a value from localStorage, tolerating browsers that deny access
 * (private mode, disabled storage).
 *
 * @param {string} key
 * @returns {string|null}
 */
function readStorage(key) {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

/**
 * Writes a value to localStorage, ignoring a denial.
 *
 * @param {string} key
 * @param {string} value
 */
function writeStorage(key, value) {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // Storage being unavailable costs us persistence, not correctness: the
    // operator stays signed in for the current page only.
  }
}

/**
 * Removes a value from localStorage, ignoring a denial.
 *
 * @param {string} key
 */
function removeStorage(key) {
  try {
    window.localStorage.removeItem(key);
  } catch {
    // See writeStorage.
  }
}

/**
 * Returns the stored session token, or null when signed out.
 *
 * @returns {string|null}
 */
export function getToken() {
  return readStorage(TOKEN_STORAGE_KEY);
}

/**
 * Returns the cached session details for display purposes.
 *
 * These are only ever a cache of what the backend told us at login. Nothing
 * security-relevant may be decided from them: the backend re-checks the role
 * on every request.
 *
 * @returns {{token: string, user_id: string, institution_id: string, role: string, expires_at: string}|null}
 */
export function getSession() {
  const raw = readStorage(SESSION_STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    // A corrupted entry is treated as "signed out" rather than crashing the page.
    removeStorage(SESSION_STORAGE_KEY);
    return null;
  }
}

/**
 * Persists a successful login.
 *
 * @param {{token: string, user_id: string, institution_id: string, role: string, expires_at: string}} loginResponse
 */
export function storeSession(loginResponse) {
  writeStorage(TOKEN_STORAGE_KEY, loginResponse.token);
  writeStorage(SESSION_STORAGE_KEY, JSON.stringify(loginResponse));
}

/** Clears all stored session state. */
export function clearSession() {
  removeStorage(TOKEN_STORAGE_KEY);
  removeStorage(SESSION_STORAGE_KEY);
}

/**
 * Reports whether the cached session has already expired.
 *
 * This is a courtesy check to avoid a pointless request; the backend remains
 * the authority on whether a token is valid.
 *
 * @returns {boolean}
 */
export function isExpired() {
  const session = getSession();
  if (!session?.expires_at) return false;
  const expiresAt = Date.parse(session.expires_at);
  if (Number.isNaN(expiresAt)) return false;
  return expiresAt <= Date.now();
}

/**
 * True when the signed-in operator may call the admin-guarded routes.
 *
 * Every route the dashboard uses is registered behind
 * `RequireRole("admin", "manager")` in `backend/internal/server/server.go`, so
 * in practice a user without one of those roles cannot load any dashboard page
 * at all — `describeError` in api.js explains that to them. This function still
 * gates the individual mutating controls, because the role is a cached value
 * from sign-in and the backend is the authority: it re-checks on every request.
 *
 * @returns {boolean}
 */
export function canManage() {
  const role = (getSession()?.role || '').toLowerCase();
  return role === 'admin' || role === 'manager';
}

/**
 * Redirects to the sign-in page, preserving where the operator was heading.
 *
 * @param {string} [reason]
 */
export function redirectToLogin(reason) {
  // A bare sibling filename, because the sign-in page sits in pages/ beside
  // every guarded page and resolves this relative to itself.
  const target = new URL('login.html', window.location.href);
  if (reason) target.searchParams.set('reason', reason);
  // Preserve the originally requested page so login can return there.
  const current = window.location.pathname.split('/').pop();
  if (current && current !== 'login.html') {
    target.searchParams.set('next', current + window.location.search);
  }
  window.location.replace(target.toString());
}

/**
 * Guards a page: sends the operator to sign-in unless a usable token exists.
 *
 * @returns {boolean} true when the page may render.
 */
export function requireSession() {
  // Distinguished before anything is cleared: a token that has expired is a
  // session that ended, while no token at all is a first visit. Only the first
  // is worth explaining — the site's root leads to the dashboard, so a stranger
  // opening the app arrives here having done nothing, and "Your session has
  // ended" would be a puzzling thing to tell them.
  const hadToken = Boolean(getToken());
  if (!hadToken || isExpired()) {
    clearSession();
    redirectToLogin(hadToken ? 'expired' : undefined);
    return false;
  }
  return true;
}

/**
 * Signs the operator out, ending the session on the backend as well as locally.
 *
 * @param {(path: string, options?: object) => Promise<any>} apiRequest
 * @returns {Promise<void>}
 */
export async function signOut(apiRequest) {
  try {
    // Best effort: a failed logout call must not strand the operator on a page
    // they have already asked to leave. The token is discarded regardless.
    await apiRequest('/auth/logout', { method: 'POST' });
  } catch {
    // Ignored on purpose; see above.
  } finally {
    clearSession();
    window.location.replace('login.html');
  }
}
