# CampusWatch Dashboard

The operator-facing web interface: sign in, watch the fleet, drill into a
machine, manage the issue queue, pull reports and maintain the campus
hierarchy.

It is plain HTML, CSS and JavaScript with **no build step and no
dependencies**. ES modules load directly in the browser, so there is nothing to
install, compile or audit beyond the files in this directory.

---

## Running it

The dashboard is a set of static files, so any web server can host it. It must
be served over HTTP rather than opened as a `file://` path, because ES modules
and `fetch` are both blocked on the `file://` scheme.

### Locally, in one command

```bash
# Terminal 1 — the backend.
cd backend && go run ./cmd/server

# Terminal 2 — the dashboard, with /api forwarded to the backend.
python3 frontend/serve.py
```

Then open <http://localhost:8000/>.

`serve.py` serves this directory and proxies `/api/*` to `127.0.0.1:8080`, so
the browser sees a single origin and CORS never applies. It is a development
tool only — it performs no TLS and no authentication of its own — but it is the
quickest way to get a working dashboard, and it needs nothing installed.

```bash
python3 frontend/serve.py --port 9000          # a different port
python3 frontend/serve.py --api http://127.0.0.1:9001   # a different backend
python3 frontend/serve.py --help
```

### Serving the files yourself

Any static file server works for the frontend half, but note that it cannot
serve the API as well:

```bash
python3 -m http.server 8000 --directory frontend
```

> **This will load the sign-in page and nothing else.** `http.server` cannot
> proxy, so `/api` has no backend behind it and every request fails with a
> network error. Use `serve.py` for local work, or one of the two options below
> for a real deployment.

### Opening the backend's port by mistake

`http://localhost:8080/` answers `404 page not found`. That is the *backend*,
and it is correct: it serves no static files, only `/health` and `/api/v1/*`.
The dashboard is this directory, served separately. Nothing is broken.

---

## Required: same-origin with the API

The backend sends **no CORS headers**. A browser will therefore refuse to let a
page on `localhost:8000` read responses from an API on `localhost:8080` — which
is exactly what the command above produces. This is a deliberate consequence of
the backend's current middleware stack, not a bug in the dashboard, and the
dashboard cannot work around it from the client side.

There are two supported ways to run it in production. (`serve.py` above is the
local equivalent of Option A, without TLS.)

### Option A — reverse proxy (no backend changes)

Serve the dashboard and the API from one origin, with the proxy forwarding
`/api` to the backend. Leave `API_BASE` empty (its default) so requests are
same-origin.

An `nginx` location block that does this:

```nginx
server {
    listen 443 ssl;
    server_name campuswatch.example.edu;

    root /srv/campuswatch/frontend;
    index index.html;

    # The API lives behind the same origin, so the browser sees no cross-origin
    # request at all and CORS never applies.
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Option B — add CORS to the backend

If the dashboard must be served from a different origin, the backend needs a
CORS middleware that:

- sets `Access-Control-Allow-Origin` to the dashboard's exact origin — never
  `*`, since the dashboard sends `Authorization` headers
- handles `OPTIONS` preflight, allowing the `Authorization` and
  `Content-Type` request headers
- sets `Access-Control-Allow-Credentials` only if cookie auth is ever added
  (it is not used today; the dashboard authenticates with a bearer token)

This is a backend change and is **not** part of this frontend. Until one of
these two options is in place, the sign-in page will load but every request
will fail with a network error.

---

## Configuration

`js/config.js` is the only file with deployment-specific values.

| Setting | Default | Notes |
| ------- | ------- | ----- |
| `API_BASE` | `''` (same origin) | Override with `window.CAMPUSWATCH_API_BASE` or `localStorage["campuswatch.apiBase"]` |
| `DASHBOARD_REFRESH_MS` | `30000` | How often the dashboard refreshes itself; `0` disables it |

To point a build at an absolute API URL without editing the file, uncomment the
inline script in `index.html`:

```html
<script>window.CAMPUSWATCH_API_BASE = 'https://campuswatch.example.edu';</script>
```

---

## Branding

The logo is `frontend/image/adewahlogo.png`, displayed in the header of every
page, on the sign-in page, and as the tab icon. Paths are set in `js/config.js`
and resolved by `brandMark()` in `js/ui.js`.

If the file cannot be loaded, every page falls back to `assets/logo.svg`
automatically, so a renamed or missing logo never produces a broken image. The
fallback filenames appear in the favicon tags of all seven HTML files, which
cannot read `config.js` — see [`image/README.md`](image/README.md) for the list.

> **The current logo is 418 KB** for a mark rendered at 28–40px. Resizing to
> 192×192 and optimising would cut it to roughly 20 KB. Every operator downloads
> it on every page load, so this is worth doing.

---

## Pages

| File | Purpose |
| ---- | ------- |
| `index.html` | Sign in |
| `pages/dashboard.html` | Fleet summary, systems needing attention, open issues |
| `pages/systems.html` | Every registered system, with search and status filter |
| `pages/system.html?id=` | One system: health, login history, events, status editor |
| `pages/issues.html` | The maintenance queue, with create and progress notes |
| `pages/reports.html` | Usage, uptime, session, issue and after-hours reports |
| `pages/hierarchy.html` | Institutions, campuses, clusters, locations |

## Modules

| File | Responsibility |
| ---- | -------------- |
| `js/config.js` | API base URL and tunables |
| `js/session.js` | Token storage, expiry check, sign-out, page guard |
| `js/api.js` | The single HTTP client, and one named call per backend route |
| `js/ui.js` | Element building, formatting, badges, meters, tables, toasts |
| `js/shell.js` | Shared header, navigation and page scaffold |
| `js/pages/*.js` | One module per page |
| `serve.py` | Local development server; not used in production |

---

## Responsive design

Everything is fluid by default, so the layout reflows continuously rather than
snapping between fixed widths. The media queries in `css/main.css` change only
what genuinely cannot work at a given size.

| Breakpoint | Applies to | Change |
| --- | --- | --- |
| `min-width: 1600px` | Large desktop, wall display | Page widens from 1280px to 1520px |
| `max-width: 900px` | Tablet portrait | Two-column layouts collapse to one |
| `max-width: 720px` | Phone, small tablet | Tables become cards; nav scrolls sideways |
| `max-width: 480px` | Small phone | Summary tiles go two-up |
| `pointer: coarse` | Any touch device | Controls grow to a 44px tap target |
| `orientation: landscape` + `max-height: 500px` | Landscape phone | Nav returns to one row beside the brand |
| `forced-colors: active` | Windows High Contrast | Badges and meters outlined |
| `print` | Paper, PDF | Chrome removed; table columns restored |

A landscape phone is the case worth calling out: it is wide enough for the
desktop header but only ~375px tall, so the constraint flips from horizontal to
vertical and the navigation goes back to a single row.

### Printing

The reports page has a Print button, so there is a print stylesheet. It drops
the header, navigation, filter bar and buttons; forces black on white so a
dark-theme operator does not print a page of solid ink; and **reverses the
table-to-card transformation**, because paper is wide enough for real columns
and cannot scroll. `.card__hint` is deliberately kept — on a report it carries
the date range the figures cover.

### Tables on small screens

A seven-column table is unreadable on a 360px phone — scrolling sideways loses
track of which column a value belongs to. Below 720px the header row is hidden
and each row becomes a card, with the column name printed beside every value.

That relies on `table()` in `js/ui.js` stamping a `data-label` attribute on each
cell; the CSS prints it with `content: attr(data-label)`. **A table built by
hand instead of via that helper will lose its labels on a phone.** The header
row is clipped rather than removed, so screen readers still get a real table.

### Browser support

"Any device" has a floor worth stating. The dashboard uses ES modules,
optional chaining and `replaceChildren`, so it needs a **browser from roughly
2020 onwards** — Chrome/Edge 86+, Firefox 78+, Safari 14+. Everything older
renders a blank page, since the modules never execute.

That covers essentially every device still receiving updates, including older
phones and tablets. It does **not** cover Internet Explorer, or an Android
device old enough to be stuck on a pre-2020 WebView. If such a device matters,
the fix is a build step that transpiles and bundles — which would end the
no-build-step property this frontend is built on, so it is a deliberate
trade-off rather than an oversight.

### Touch and display cutouts

- Controls reach the 44px minimum tap target on any coarse pointer, including
  wide tablets where the width-based breakpoints would not fire.
- All seven pages set `viewport-fit=cover`, which is what makes the
  `env(safe-area-inset-*)` padding resolve on notched phones. Without it those
  insets are always zero and content slides under the notch in landscape.
- `color-scheme: light dark` makes native controls, scrollbars and the mobile
  browser chrome follow the light or dark theme, rather than staying light on a
  dark page.

> **Not verified in a browser.** These rules are correct by inspection and the
> stylesheet's braces balance, but nothing has rendered them — I have no way to
> run a browser here. Worth checking on a real phone before relying on it,
> particularly the table-to-card transformation, which is the most intricate
> part, and the print layout, which is easy to get subtly wrong.

---

## How it authenticates

`POST /api/v1/auth/login` returns `{token, user_id, institution_id, role,
expires_at}`. The token is stored in `localStorage` and sent as
`Authorization: Bearer <token>` on every subsequent request.

Three consequences worth being explicit about:

- **The token is readable by any script on the origin.** That is acceptable
  only because the backend accepts no cross-origin requests. Adding permissive
  CORS would make this storage choice unsafe.
- **The cached role is a display convenience, never a control.** `canManage()`
  hides buttons the backend would reject with `403`, but the backend re-checks
  the role on every request and remains the authority.
- **A `401` signs the operator out immediately.** Any rejected request clears
  the stored session and returns to the sign-in page with an explanation,
  rather than leaving a page that silently fails.

---

## Backend limitations this frontend works around

These are gaps in the current backend API, not dashboard choices.

1. **Issue progress notes cannot be listed.** `POST /api/v1/issues/{id}/updates`
   appends a note, but no endpoint returns them, and `GET /api/v1/issues/{id}`
   returns the issue alone. The issue detail panel therefore shows only the
   notes added during the current browser session, labelled as such. Showing the
   full history needs a `GET /api/v1/issues/{id}/updates` endpoint.

2. **The backend serves no static files.** There is no route for the dashboard
   itself, so it must be hosted separately — hence the proxy in Option A.

3. **System registration needs a `location_id`.** Creating a system requires a
   location to exist first, so the hierarchy page has to be populated before
   the systems page is useful.

---

## Who can use the dashboard

**Only the `admin` and `manager` roles.** Every route this dashboard calls —
systems, issues, dashboard, reports, schedules, and the whole hierarchy — is
registered behind `RequireRole("admin", "manager")` in
`backend/internal/server/server.go`. There is no read-only view.

A user with any other role can authenticate successfully and will then be
refused by every page. The dashboard detects that 403 and says so explicitly
(see `describeError` in `js/api.js`) rather than showing an unexplained error,
but it cannot grant access that the backend withholds.

If a read-only role is wanted later, it needs a change in the backend's route
registration — the dashboard already hides mutating controls via `canManage()`.

---

## Security posture

- **No server data is ever written as HTML.** Every helper in `js/ui.js` uses
  `textContent` or `createElement`. This matters here more than in most
  dashboards: hostnames, usernames and issue text originate on *monitored*
  machines, so a compromised campus computer could otherwise inject script into
  an operator's session. Event payloads are rendered inside a `<pre>` via
  `textContent` and stay inert.
- **Sign-in failures do not distinguish causes.** An unknown email and a wrong
  password produce the same message, matching the backend and avoiding account
  enumeration.
- **The redirect target after sign-in is validated.** `?next=` accepts a page
  filename only; anything with a scheme or `//` is discarded, closing an open
  redirect.
- **Pages send `noindex, nofollow` and `Cache-Control: no-store`.** The
  dashboard shows operational state and should not be indexed or cached by an
  intermediary.
- **No inline scripts or third-party resources.** The only subresource is
  `assets/logo.svg`, which is local, and the CSS has no `@import`. A strict
  `Content-Security-Policy` can therefore be applied at the proxy without
  `unsafe-inline` — worth doing, since the backend's `SecurityHeaders`
  middleware only covers API responses and not these static pages.
