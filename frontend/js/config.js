/**
 * CampusWatch frontend configuration.
 *
 * The frontend is plain HTML, CSS and JavaScript with no build step, so this
 * file is the only place deployment-specific values live.
 */

/**
 * Base URL prepended to every API path.
 *
 * It defaults to an empty string, which means "same origin". That is the
 * correct default because the backend serves no CORS headers, so a browser
 * will only accept its responses when the dashboard is reached through the
 * same origin as the API. Serve the dashboard behind a reverse proxy that
 * forwards /api to the backend, or set an explicit base below.
 *
 * Override without editing this file, in order of precedence:
 *   1. window.CAMPUSWATCH_API_BASE, set by an inline script in the HTML
 *   2. localStorage["campuswatch.apiBase"]
 *
 * An absolute value must include the scheme and host, for example
 * "https://campuswatch.example.edu".
 */
function resolveApiBase() {
  if (typeof window !== 'undefined') {
    if (typeof window.CAMPUSWATCH_API_BASE === 'string') {
      return window.CAMPUSWATCH_API_BASE;
    }
    try {
      const stored = window.localStorage.getItem('campuswatch.apiBase');
      if (stored) return stored;
    } catch {
      // Private browsing modes can deny localStorage access. Falling back to
      // same-origin is correct and should not break the page.
    }
  }
  return '';
}

/** API base URL, without a trailing slash. */
export const API_BASE = resolveApiBase().replace(/\/+$/, '');

/** Version of the API this frontend speaks, from the backend routes. */
export const API_PREFIX = '/api/v1';

/** Where the session token is kept between page loads. */
export const TOKEN_STORAGE_KEY = 'campuswatch.token';

/** Where the last known session details are cached for display. */
export const SESSION_STORAGE_KEY = 'campuswatch.session';

/**
 * How often the dashboard refreshes itself, in milliseconds. Set to 0 to
 * disable automatic refreshing.
 */
export const DASHBOARD_REFRESH_MS = 30000;

/**
 * The institution's brand mark.
 *
 * Both paths are relative to the frontend root, so callers only supply a prefix
 * for their own depth (`'../'` from every page in `pages/`). Keeping them whole
 * here means the two files can live in different directories — as they do —
 * without every caller having to know that.
 *
 * `BRAND_MARK` is the institution's logo. It is displayed wherever the
 * CampusWatch mark appears. `BRAND_MARK_FALLBACK` is used if it cannot be
 * loaded, so a renamed or missing file degrades to the bundled vector mark
 * instead of a broken-image icon.
 *
 * Supply the logo at least 2x its largest display size (40px, so 80px or more)
 * to stay sharp on high-density screens.
 */
export const BRAND_MARK = 'image/adewahlogo.png';

/** Bundled vector mark, used when the brand mark above cannot be loaded. */
export const BRAND_MARK_FALLBACK = 'assets/logo.svg';
