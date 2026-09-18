# Brand image

## `adewahlogo.png`

The institution's logo. It is displayed in the header of every page, on the
sign-in page, and as the browser tab icon.

Referenced from two places, which must agree:

| Location | Value |
| --- | --- |
| `js/config.js` → `BRAND_MARK` | `image/adewahlogo.png` |
| Favicon `<link>` in `index.html` and `pages/*.html` (7 files) | `image/adewahlogo.png` |

The favicon is declared in markup because a browser chooses it from the page
before any script runs, so it cannot read `config.js`. **Renaming the file means
editing `js/config.js` and seven favicon tags.**

If the file is missing or renamed, every page falls back to
`../assets/logo.svg` automatically — see `brandMark()` in `js/ui.js`. No page
will show a broken image.

### Size

The current file is **418 KB**, which is heavy for a mark that renders at 28px
in the header and 40px on the sign-in page — and it is downloaded on every page
load, uncached on first visit.

Resizing it to **192×192** and running it through a PNG optimiser
(`pngquant`, `oxipng`, or `squoosh.app`) would typically bring it under 20 KB
with no visible difference at these display sizes. That is a ~95% reduction on a
file every operator downloads.

A square image is ideal: the mark is drawn in a square box, so a wide logo will
be letterboxed rather than stretched. Transparency is supported and works in
both the light and dark themes.
