# Third-party skills

Every skill under `skills/` comes from somewhere. This file records where, which version, and under what licence, so a reviewer can check the copy against its source. Nothing here is edited after copying.

## diagram-design

- **What it does:** helps an agent draw diagrams (architecture, flow, sequence and many more) as self-contained HTML or SVG and export them to SVG or PNG.
- **Source:** https://github.com/cathrynlavery/diagram-design (path `skills/diagram-design`), installed with `npx skills add cathrynlavery/diagram-design`.
- **Version:** 2.6, commit `0cf73c3`. The copy is byte-identical to that commit (`diff -r` shows nothing).
- **Licence:** MIT, copyright Cathryn Lavery.
- **Reviewed on 2026-11-02 before adoption:** the scripts import only the Python standard library (plus Playwright for `export_png.py`); there are no network calls, no subprocesses and no `eval` or `exec`; the markdown contains no install or shell commands. `urllib.parse` is used only to split URL strings.
- **Google Fonts:** 220 of the example HTML files in `assets/` link to `fonts.googleapis.com` and make a request when opened in a browser. They are examples and are not served by the Kuronuma site. Kuronuma's own profile (`profiles/kuronuma.md`) sets `Font source: system`, so diagrams made for this project make no font request.
- **Kuronuma profile:** `profiles/kuronuma.md` is the skill's own profile format with the Blush colours and Inter and Geist Mono. It is ours, not upstream's. The root file `.diagram-design` selects it.

### Licence text (MIT)

```
MIT License

Copyright (c) 2025 Cathryn Lavery

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
