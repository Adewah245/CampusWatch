# CampusWatch Dashboard

The operator-facing web interface: sign in, watch the fleet, drill into a
machine, manage the issue queue, pull reports and maintain the campus
hierarchy.

It is plain HTML, CSS and JavaScript with **no build step and no
dependencies**. ES modules load directly in the browser, so there is nothing to
install, compile or audit beyond the files in this directory.

---

## Running it

The dashboard is a set of static files, but it is not served separately. The Go
backend serves this directory **and** the API, so there is one process, one port
and one origin. It must be served over HTTP rather than opened as a `file://`
path, because ES modules and `fetch` are both blocked on the `file://` scheme.

### Locally

Run from the **repository root**:

```bash
go run ./backend/cmd/server
```

Then open <http://localhost:8080/>.

`/` is a redirect to `/pages/dashboard.html`, so opening the site lands on the
dashboard. Nothing needs installing beyond Go.

An unauthenticated visitor is turned away by that page and lands on
`/pages/login.html`. Signing in needs an account to exist, and the first one is
created from the browser: the sign-in page links to `/pages/register.html`, which
creates the administrator account and the institution it belongs to. It works
only while no account exists — after that it answers `409` and says registration
is closed — so it is a bootstrap step rather than a sign-up form.

The registration page deliberately does not link back to sign-in: it sets a
deployment up, it does not enter one. On success it hands off to sign-in so the
new operator proves the credentials they just chose.

Further accounts come from inside the dashboard or from `createuser`:

```bash
CAMPUSWATCH_PASSWORD='choose-a-strong-password' \
go run ./backend/cmd/createuser --email you@example.edu --institution 'Your Institution'
```

A `401` from `/api/v1/auth/login` means no account exists yet, not that the
password was mistyped.

### Not serving the files separately

A plain static server can host this directory, but it cannot serve the API as well
— and three things the Go server does for these files would be missing:

```bash
python3 -m http.server 8000 --directory frontend   # don't
```

- `/api` has no backend behind it, so every request fails with a network error.
- `/` answers with a file listing, because there is no `index.html` here — the
  root is a redirect the Go server performs.
- Every subdirectory lists its contents, for the same reason.

Use the Go server. Set `FRONTEND_DIR` to point it at a different dashboard
directory, for example when the binary and the dashboard are deployed separately.

---

## Required: same-origin with the API

The backend sends **no CORS headers**. A browser will therefore refuse to let a
dashboard on one origin read responses from an API on another.

The Go server satisfies this by default: it serves this directory and the API from
the same origin, so the local command above works as shipped and `API_BASE` stays
empty. The requirement below only bites if you serve the dashboard from somewhere
else — a different port, or a CDN.

There are two supported ways to do that. (Option A is also how you would put TLS in
front of the Go server, which speaks plain HTTP.)

### Option A — reverse proxy (recommended)

Forward everything to the Go server, which already serves the dashboard and the API
from one origin. Leave `API_BASE` empty (its default) so requests are same-origin.

An `nginx` location block that does this, and terminates TLS the Go server does not:

```nginx
server {
    listen 443 ssl;
    server_name campuswatch.example.edu;

    # Everything goes to the Go server, which serves both the dashboard and the
    # API from this one origin, so the browser sees no cross-origin request at
    # all and CORS never applies.
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Forwarding everything rather than serving the files from a `root` and proxying only
`/api/` is deliberate. The Go server's static handler also answers `/` with the
redirect to the dashboard and refuses directory listings; a proxy that served the
files itself would have to reproduce both, and would be a second copy of the
dashboard to keep in step.

### Option B — a separate origin for the dashboard

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
inline script in the page's HTML. Every page carries it, and a page served from
somewhere other than the Go server needs it in all of them (or can set the
`localStorage` key once instead):

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
fallback filenames appear in the favicon tags of all eight HTML files, which
cannot read `config.js` — see [`image/README.md`](image/README.md) for the list.

> **The current logo is 418 KB** for a mark rendered at 28–40px. Resizing to
> 192×192 and optimising would cut it to roughly 20 KB. Every operator downloads
> it on every page load, so this is worth doing.

---

## Pages

Every page lives in `pages/`, which is why each one references assets as
`../css/main.css` and `../js/pages/….js`. They all link to each other by bare
sibling filename.

| File | Purpose |
| ---- | ------- |
| `pages/dashboard.html` | Fleet summary, systems needing attention, open issues. What `/` redirects to |
| `pages/login.html` | Sign in. Also the way to the first-run setup |
| `pages/register.html` | First-run registration: creates the administrator for a new deployment |
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
- All eight pages set `viewport-fit=cover`, which is what makes the
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

2. **An agent cannot enrol itself, and nothing in the dashboard can enrol it.**
   `POST /api/v1/agents` issues an agent credential and `POST
   /api/v1/agents/{id}/approve` issues another on approval; both have to happen
   before a machine reports anything. The dashboard has no screen for either, so
   today they are done with `curl` and the credential is carried to the machine by
   hand. Until that has happened a system is silent, which is why every question
   about a machine being online or in use depends on it.

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
