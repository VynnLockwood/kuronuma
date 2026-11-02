# Third-party skills

Every skill under `skills/` comes from somewhere. This file records where, which version, and under what licence, so a reviewer can check the copy against its source. Nothing here is edited after copying.

## diagram-design

- **What it does:** helps an agent draw diagrams (architecture, flow, sequence and many more) as self-contained HTML or SVG and export them to SVG or PNG.
- **Source:** https://github.com/cathrynlavery/diagram-design (path `skills/diagram-design`), installed with `npx skills add cathrynlavery/diagram-design`.
- **Version:** 2.6, commit `0cf73c3`. The copy is byte-identical to that commit (`diff -r` shows nothing).
- **Licence:** MIT, copyright Cathryn Lavery.
- **Reviewed on 2026-11-02 before adoption, and what that review found** (so a later reviewer can judge it again):
  - Scripts: they import only the Python standard library, except `export_png.py`, which needs Playwright. There is no `subprocess`, `eval` or `exec` in any script and nothing reads `~` or the environment. File writes go only to the output path the user gives. `urllib.parse` is used only to split URL strings. `drawio_extract.py` caps decompressed size.
  - `export_png.py` starts headless Chromium and loads the diagram HTML with the network enabled and JavaScript on. With `Font source: web`, or any remote reference in a diagram, that makes real requests. Kuronuma's profile uses `Font source: system`, so diagrams made for this project reference no remote fonts.
  - The markdown does contain commands an agent may run: `mkdir -p ~/.diagram-design/profiles` (writes outside the project, in the user's home; `references/profiles.md`), `python3 scripts/...` export and import commands (`references/export.md`, `references/import-*.md`), and `agent-browser navigate <url>` in onboarding (`references/onboarding.md`), which is behind a trust prompt that asks the user first. These were judged acceptable. They are why adding or updating a skill needs the owner's approval.
  - The skill's own flows (onboarding, `save`, markerless `load`) write into the skill's installed `references/style-guide.md`. In this repository that is the vendored copy, so `skills/README.md` forbids running them here.
- **Google Fonts:** 220 of the example HTML files in `assets/` link to `fonts.googleapis.com` and make a request when opened in a browser. They are examples and are not served by the Kuronuma site. Kuronuma's own profile (`profiles/kuronuma.md`) sets `Font source: system`, so diagrams made for this project make no font request.
- **Kuronuma profile:** `profiles/kuronuma.md` is the skill's own profile format with the Blush colours and Inter and Geist Mono. It is ours, not upstream's. The root file `.diagram-design` selects it.

### Third-party content inside the skill

The skill bundles icon artwork and logos from other projects (listed in upstream's `THIRD_PARTY_LICENSES.md`, which is outside the vendored folder). The parts that are inside `skills/diagram-design/`:

- **Tabler Icons**: MIT, https://github.com/tabler/tabler-icons. Stroked icons in `references/primitive-icons.md` and `assets/icons.html`. Licence: https://github.com/tabler/tabler-icons/blob/main/LICENSE.
- **Simple Icons**: CC0 1.0 Universal (public domain dedication), https://github.com/simple-icons/simple-icons. Filled brand silhouettes in the same two files. Licence: https://github.com/simple-icons/simple-icons/blob/develop/LICENSE.md.
- **log-z/logos**: MIT, https://github.com/log-z/logos. A few filled brand silhouettes (for example MySQL and Redis) in the same two files. Licence: https://github.com/log-z/logos/blob/main/LICENSE.
- **Trademarks**: brand logos remain the trademarks of their owners; they appear for documentation and illustration only and imply no endorsement or affiliation.

### Licence text (MIT, the skill itself)

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
