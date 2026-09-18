/**
 * The CampusWatch API client.
 *
 * Every request to the backend goes through `request()` so that authentication,
 * error handling and the base URL are applied in exactly one place. Page code
 * calls the named helpers at the bottom of this file rather than building URLs
 * by hand, which keeps the route table in one reviewable list.
 *
 * The route shapes here mirror `backend/internal/server/server.go`. If a route
 * moves there, it must move here in step.
 */

import { API_BASE, API_PREFIX } from './config.js';
import { getToken, getSession, clearSession, redirectToLogin } from './session.js';

/**
 * An error returned by the CampusWatch API, or raised while reaching it.
 *
 * `status` is the HTTP status code, or 0 when the request never completed
 * (offline, DNS failure, connection refused). Callers can branch on `status`
 * to distinguish "the operator is not allowed to do this" (403) from "the
 * server is down" (0).
 */
export class ApiError extends Error {
  /**
   * @param {string} message
   * @param {number} status
   * @param {any} [body]
   */
  constructor(message, status, body) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }

  /** True when the failure was a network problem rather than a rejection. */
  get isNetworkError() {
    return this.status === 0;
  }

  /** True when the operator's session is not valid. */
  get isUnauthorized() {
    return this.status === 401;
  }

  /** True when the session is valid but lacks the required role. */
  get isForbidden() {
    return this.status === 403;
  }
}

/**
 * Pulls a human-readable message out of an error response body.
 *
 * The backend returns `{"error": "..."}` for handled failures, but a proxy or
 * a panic can return HTML or an empty body, so this falls back to the status
 * text rather than assuming the shape.
 *
 * @param {Response} response
 * @param {string} rawBody
 * @returns {string}
 */
function describeFailure(response, rawBody) {
  if (rawBody) {
    try {
      const parsed = JSON.parse(rawBody);
      const message = parsed?.error || parsed?.message;
      if (typeof message === 'string' && message.trim()) return message;
    } catch {
      // Not JSON. Fall through to the status text.
    }
  }
  return response.statusText || `Request failed with status ${response.status}`;
}

/**
 * Reads and parses a response body, tolerating the empty bodies that 204
 * responses and some handlers return.
 *
 * @param {Response} response
 * @returns {Promise<any>}
 */
async function readBody(response) {
  const text = await response.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    // A non-JSON success body would mean the frontend is pointed at something
    // that is not the CampusWatch API — most often a static file server that
    // answered the request with index.html. Say so rather than failing opaquely.
    throw new ApiError(
      'The server returned a response the dashboard could not read. Check that the API base URL points at the CampusWatch backend.',
      response.status,
      text.slice(0, 200),
    );
  }
}

/**
 * Builds a query string, dropping empty values.
 *
 * @param {Record<string, string|number|boolean|undefined|null>} [params]
 * @returns {string} a string beginning with "?", or an empty string
 */
function buildQuery(params) {
  if (!params) return '';
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue;
    search.set(key, String(value));
  }
  const encoded = search.toString();
  return encoded ? `?${encoded}` : '';
}

/**
 * Performs an authenticated request against the CampusWatch API.
 *
 * @param {string} path API path beginning with "/", for example "/systems".
 * @param {object} [options]
 * @param {string} [options.method] HTTP method, default GET.
 * @param {any} [options.body] Value to serialise as JSON.
 * @param {Record<string, any>} [options.query] Query parameters.
 * @param {boolean} [options.auth] Send the bearer token, default true.
 * @param {AbortSignal} [options.signal] For cancelling in-flight requests.
 * @returns {Promise<any>} the parsed response body, or null for an empty body.
 * @throws {ApiError}
 */
export async function request(path, options = {}) {
  const { method = 'GET', body, query, auth = true, signal } = options;

  const headers = { Accept: 'application/json' };
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (auth) {
    const token = getToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  let response;
  try {
    response = await fetch(`${API_BASE}${API_PREFIX}${path}${buildQuery(query)}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal,
    });
  } catch (error) {
    // fetch only rejects for network-level failures; HTTP error codes resolve.
    if (error?.name === 'AbortError') throw error;
    throw new ApiError(
      'Could not reach the CampusWatch server. Check your connection and that the API base URL is correct.',
      0,
    );
  }

  if (response.status === 401) {
    // The token is expired, revoked, or was never valid. Drop it and send the
    // operator back to sign-in rather than showing an error they cannot act on.
    const raw = await response.text().catch(() => '');
    clearSession();
    redirectToLogin('expired');
    throw new ApiError(describeFailure(response, raw), 401);
  }

  if (!response.ok) {
    const raw = await response.text().catch(() => '');
    throw new ApiError(describeFailure(response, raw), response.status, raw);
  }

  return readBody(response);
}

/**
 * Turns a thrown value into a message worth showing an operator.
 *
 * The important case is 403. Every route the dashboard uses is registered
 * behind `RequireRole("admin", "manager")` in `backend/internal/server/server.go`,
 * so a signed-in user with any other role — a technician or a viewer — can
 * authenticate successfully and still be refused by every page. Without this,
 * they would see a bare "forbidden" on each page and have no idea why or what
 * to ask for.
 *
 * @param {unknown} error
 * @returns {string}
 */
export function describeError(error) {
  if (!(error instanceof ApiError)) {
    return 'Something went wrong. Please try again.';
  }

  if (error.status === 403) {
    const role = getSession()?.role;
    const who = role ? `Your account role (${role})` : 'Your account';
    return (
      `${who} does not have access to the monitoring dashboard. ` +
      'Every dashboard route requires the admin or manager role — ask an administrator ' +
      'to grant it, or use an account that already has it.'
    );
  }

  if (error.status === 404) {
    return 'That record does not exist, or is not visible to your account.';
  }

  return error.message;
}

/**
 * Named API calls, one per backend route the dashboard uses.
 *
 * Keeping them together makes it obvious which parts of the backend the
 * frontend depends on.
 */
export const api = {
  // --- Authentication -----------------------------------------------------

  /** Signs in and returns `{token, user_id, institution_id, role, expires_at}`. */
  login: (email, password) =>
    request('/auth/login', { method: 'POST', body: { email, password }, auth: false }),

  /** Returns the current session, useful for confirming a token is still live. */
  me: () => request('/auth/me'),

  /** Ends the session on the backend. */
  logout: () => request('/auth/logout', { method: 'POST' }),

  // --- Dashboard ----------------------------------------------------------

  /** Aggregate counts for the summary tiles. */
  dashboardSummary: () => request('/dashboard/summary'),

  /** Every system the signed-in operator's institution may see. */
  dashboardSystems: () => request('/dashboard/systems'),

  // --- Systems ------------------------------------------------------------

  /** Lists systems, optionally filtered by status. */
  systems: (query) => request('/systems', { query }),

  /** Fetches a single system. */
  system: (id) => request(`/systems/${encodeURIComponent(id)}`),

  /** Applies a partial update to a system. */
  updateSystem: (id, patch) =>
    request(`/systems/${encodeURIComponent(id)}`, { method: 'PATCH', body: patch }),

  /** Approves a pending system so its agent may report. */
  approveSystem: (id) => request(`/systems/${encodeURIComponent(id)}/approve`, { method: 'POST' }),

  /** The most recent health report for a system. */
  systemHealth: (id) => request(`/systems/${encodeURIComponent(id)}/health`),

  /** Login/logout history for a system. */
  systemSessions: (id, query) => request(`/systems/${encodeURIComponent(id)}/sessions`, { query }),

  /** Events reported by or about a system. */
  systemEvents: (id, query) => request(`/systems/${encodeURIComponent(id)}/events`, { query }),

  // --- Issues -------------------------------------------------------------

  /** Lists issues, optionally filtered by `status` or `system_id`. */
  issues: (query) => request('/issues', { query }),

  /** Opens a new issue. */
  createIssue: (issue) => request('/issues', { method: 'POST', body: issue }),

  /** Fetches a single issue with its update history. */
  issue: (id) => request(`/issues/${encodeURIComponent(id)}`),

  /** Applies a partial update, typically a status change. */
  updateIssue: (id, patch) =>
    request(`/issues/${encodeURIComponent(id)}`, { method: 'PATCH', body: patch }),

  /** Appends a progress note to an issue. */
  addIssueUpdate: (id, update) =>
    request(`/issues/${encodeURIComponent(id)}/updates`, { method: 'POST', body: update }),

  // --- Hierarchy ----------------------------------------------------------

  institutions: () => request('/institutions'),
  createInstitution: (body) => request('/institutions', { method: 'POST', body }),

  campuses: (query) => request('/campuses', { query }),
  createCampus: (body) => request('/campuses', { method: 'POST', body }),

  clusters: (query) => request('/clusters', { query }),
  createCluster: (body) => request('/clusters', { method: 'POST', body }),

  locations: (query) => request('/locations', { query }),
  createLocation: (body) => request('/locations', { method: 'POST', body }),

  // --- Reports ------------------------------------------------------------

  /**
   * Fetches a report.
   *
   * @param {string} kind one of usage, uptime, sessions, issues, after-hours
   * @param {string} period one of daily, weekly, monthly
   */
  report: (kind, period) => request(`/reports/${encodeURIComponent(kind)}`, { query: { period } }),

  // --- Schedules ----------------------------------------------------------

  schedules: (query) => request('/schedules', { query }),
  createSchedule: (body) => request('/schedules', { method: 'POST', body }),

  /** Asks whether a moment falls inside an open-hours window. */
  checkSchedule: (institutionId, timestamp) =>
    request('/schedules/check', {
      query: { institution_id: institutionId, timestamp: timestamp?.toISOString() },
    }),
};
