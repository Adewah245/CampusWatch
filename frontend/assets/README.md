# Fallback assets

Everything the dashboard needs to stay intact when a brand asset is missing.
This directory holds no branding of its own.

## `logo.svg`

A vector mark drawn for this project. It is **not** the institution's logo.

It is used in two situations:

1. **As a fallback.** If `../image/adewahlogo.png` cannot be loaded — renamed,
   deleted, or not yet added — `brandMark()` in `js/ui.js` swaps to this file so
   no page shows a broken-image icon.
2. **As the initial markup on the sign-in page.** `index.html` ships with this
   mark already in place, so the page is never blank while scripts load. The
   sign-in script then upgrades it to the real logo.

It uses `currentColor` for its strokes, so it follows the light and dark themes
without a second file.

For the institution's logo, see [`../image/README.md`](../image/README.md).
