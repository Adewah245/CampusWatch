#!/usr/bin/env python3
"""Development server for the CampusWatch dashboard.

Serves the frontend directory as static files and forwards /api/* to the
CampusWatch backend, so the browser sees one origin for both. That matters
because the backend sends no CORS headers: a page served from :8000 calling an
API on :8080 is blocked by the browser, and no amount of frontend code can work
around it. Putting both behind one port removes the problem instead of papering
over it.

    python3 frontend/serve.py                  # dashboard on http://localhost:8000
    python3 frontend/serve.py --port 9000      # a different port
    python3 frontend/serve.py --api http://127.0.0.1:8080

This is a development convenience and is deliberately NOT what production uses.
In production the same job is done by nginx or an equivalent reverse proxy, which
also terminates TLS — see frontend/README.md, "Required: same-origin with the API".

Security notes:

  * It binds to 127.0.0.1 by default, so it is not reachable from the network.
    Passing --host 0.0.0.0 exposes an unauthenticated proxy on every interface;
    only do that on a trusted network, and never in production.
  * It performs no authentication of its own. It forwards whatever it is given
    to the backend, which remains the thing that checks credentials and roles.
  * It serves the whole frontend directory, so nothing secret may be placed
    there. Nothing is, and nothing should be.
"""

import argparse
import functools
import http.server
import os
import sys
import urllib.error
import urllib.parse
import urllib.request

# Headers that describe a single hop of an HTTP conversation and must not be
# copied onto a forwarded request or response.
HOP_BY_HOP = {
    "connection",
    "keep-alive",
    "proxy-authenticate",
    "proxy-authorization",
    "te",
    "trailer",
    "transfer-encoding",
    "upgrade",
}

# How long to wait on the backend before giving up. Generous, because a report
# over a month of data can legitimately take a moment.
BACKEND_TIMEOUT_SECONDS = 30


class DashboardRequestHandler(http.server.SimpleHTTPRequestHandler):
    """Serves static files, and proxies anything under /api to the backend."""

    # Overridden per instance in main().
    api_base = "http://127.0.0.1:8080"

    # -- Routing ----------------------------------------------------------

    def _is_api_request(self):
        """True for /api and anything beneath it."""
        return self.path == "/api" or self.path.startswith("/api/")

    def _dispatch(self, method):
        if self._is_api_request():
            self._proxy(method)
        elif method == "GET":
            super().do_GET()
        elif method == "HEAD":
            super().do_HEAD()
        else:
            # A non-API path with a non-GET verb is almost always a mistake in
            # a static file request, and saying so beats a silent 501.
            self.send_error(405, "Only GET and HEAD are supported for static files")

    def do_GET(self):
        self._dispatch("GET")

    def do_HEAD(self):
        self._dispatch("HEAD")

    def do_POST(self):
        self._dispatch("POST")

    def do_PUT(self):
        self._dispatch("PUT")

    def do_PATCH(self):
        self._dispatch("PATCH")

    def do_DELETE(self):
        self._dispatch("DELETE")

    def do_OPTIONS(self):
        self._dispatch("OPTIONS")

    # -- Proxying ---------------------------------------------------------

    def _proxy(self, method):
        """Forwards the request to the backend and relays the response."""
        length = int(self.headers.get("Content-Length") or 0)
        body = self.rfile.read(length) if length else None

        url = self.api_base + self.path
        headers = {
            name: value
            for name, value in self.headers.items()
            if name.lower() not in HOP_BY_HOP
        }
        # The backend routes on path, not host, but sending its own host keeps
        # logs and any host-based checks on the backend sensible.
        headers["Host"] = urllib.parse.urlsplit(self.api_base).netloc

        request = urllib.request.Request(url, data=body, headers=headers, method=method)

        try:
            with urllib.request.urlopen(request, timeout=BACKEND_TIMEOUT_SECONDS) as response:
                self._relay(response.status, response.headers, response.read())
        except urllib.error.HTTPError as error:
            # A 4xx or 5xx from the backend is a real answer, not a proxy
            # failure. Relaying it unchanged is what lets the dashboard show
            # the backend's own error message — an expired session (401) or a
            # role the backend refuses (403) both arrive this way.
            self._relay(error.code, error.headers, error.read())
        except urllib.error.URLError as error:
            self._backend_unreachable(error)
        except TimeoutError:
            self._backend_unreachable("timed out")

    def _relay(self, status, headers, payload):
        """Writes a backend response back to the browser."""
        self.send_response(status)
        for name, value in headers.items():
            lowered = name.lower()
            # The hop-by-hop headers describe a connection we are not reusing.
            if lowered in HOP_BY_HOP:
                continue
            # Content-Length is recomputed below from the body we actually
            # hold — correct for every method except HEAD, where urllib hands
            # back an empty body and the length would be reported as 0. HEAD
            # responses carry the length the GET *would* have returned, so the
            # backend's own header is the only right answer there.
            if lowered == "content-length" and self.command != "HEAD":
                continue
            self.send_header(name, value)
        if self.command != "HEAD":
            self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(payload)

    def _backend_unreachable(self, detail):
        """Reports that the backend could not be reached at all.

        502 rather than 404: the file exists, the API behind it does not
        answer. That distinction is the whole reason this message is worth
        having — it names the thing that is wrong.
        """
        self.send_error(
            502,
            "Cannot reach the CampusWatch backend",
            f"The dashboard is running, but no API answered at {self.api_base}. "
            f"Start it with: cd backend && go run ./cmd/server  ({detail})",
        )

    # -- Logging ----------------------------------------------------------

    def log_message(self, fmt, *args):
        """Prefixes each line so proxy traffic is distinguishable from files."""
        marker = "api " if self._is_api_request() else "file"
        sys.stderr.write(f"[{marker}] {self.address_string()} {fmt % args}\n")


def parse_args():
    parser = argparse.ArgumentParser(description="Serve the CampusWatch dashboard.")
    parser.add_argument(
        "--port",
        type=int,
        default=8000,
        help="port to serve on (default: 8000)",
    )
    parser.add_argument(
        "--host",
        default="127.0.0.1",
        help="interface to bind (default: 127.0.0.1, local access only)",
    )
    parser.add_argument(
        "--api",
        default="http://127.0.0.1:8080",
        help="backend base URL to forward /api to (default: http://127.0.0.1:8080)",
    )
    parser.add_argument(
        "--dir",
        default=None,
        help="directory to serve (default: the directory containing this script)",
    )
    return parser.parse_args()


def main():
    args = parse_args()

    directory = args.dir or os.path.dirname(os.path.abspath(__file__))
    api_base = args.api.rstrip("/")

    handler = functools.partial(DashboardRequestHandler, directory=directory)
    # The class attribute is read by _proxy, so it has to be set on the class
    # rather than passed through the partial.
    DashboardRequestHandler.api_base = api_base

    with http.server.ThreadingHTTPServer((args.host, args.port), handler) as server:
        print(f"CampusWatch dashboard : http://{args.host}:{args.port}/")
        print(f"API forwarded to      : {api_base}")
        print(f"Serving files from    : {directory}")
        print()
        print("If the dashboard loads but shows no data, check that the backend")
        print("is running and that PORT in .env matches the --api URL above.")
        print("Press Ctrl+C to stop.")

        try:
            server.serve_forever()
        except KeyboardInterrupt:
            print("\nStopped.")


if __name__ == "__main__":
    main()
