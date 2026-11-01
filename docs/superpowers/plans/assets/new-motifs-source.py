# Preview sheet for the new motifs. Same style as components/motifs: 48x48, flat fill, round 1.5 outline, colours from tokens.
TOK = """
--motif-gold:#ffd66b;--motif-gold-line:#c99a1d;--motif-leaf:#9ad8ae;--motif-leaf-line:#4f9a6a;
--motif-rose:#fbd3e0;--motif-rose-line:#d98aa6;--motif-lilac:#e5defa;--motif-lilac-line:#8d7fd6;
--motif-honey:#f2cfa0;--motif-honey-line:#b9824f;--motif-sky:#cfe8f8;--motif-sky-line:#6da9d1;
--card:#ffffff;--foreground:#2a2024;"""
def v(n): return f"var(--motif-{n})"
def st(fill, line, w=1.5, extra=""):
    f = fill if fill in ("none",) else (fill if fill.startswith("var(") else v(fill))
    l = line if line.startswith("var(") else v(line + "-line")
    return f'fill="{f}" stroke="{l}" stroke-width="{w}" stroke-linejoin="round" stroke-linecap="round" {extra}'
CARD, FG = "var(--card)", "var(--foreground)"
def nostroke(fill): return f'fill="{fill}" stroke="none"'

M = {}
M["teddy"] = f'''<g transform="translate(0 -1.2)">
<circle cx="13.5" cy="9.5" r="5.5" {st("honey","honey")}/><circle cx="34.5" cy="9.5" r="5.5" {st("honey","honey")}/>
<circle cx="13.5" cy="10" r="2.6" {nostroke(v("rose"))}/><circle cx="34.5" cy="10" r="2.6" {nostroke(v("rose"))}/>
<ellipse cx="24" cy="37" rx="12" ry="9.5" {st("honey","honey")}/>
<ellipse cx="14.5" cy="44.5" rx="5.5" ry="3" {st("honey","honey")}/><ellipse cx="33.5" cy="44.5" rx="5.5" ry="3" {st("honey","honey")}/>
<ellipse cx="11" cy="34" rx="4" ry="6.5" transform="rotate(22 11 34)" {st("honey","honey")}/><ellipse cx="37" cy="34" rx="4" ry="6.5" transform="rotate(-22 37 34)" {st("honey","honey")}/>
<ellipse cx="24" cy="39" rx="6.5" ry="5.8" {st(CARD,"honey",1.1)}/>
<circle cx="24" cy="19" r="12.5" {st("honey","honey")}/>
<ellipse cx="24" cy="23.2" rx="6" ry="4.6" {st(CARD,"honey",1.1)}/>
<ellipse cx="24" cy="21.2" rx="2.2" ry="1.5" {nostroke(v("honey-line"))}/>
<path d="M24 22.7 V24.6 M21.4 25.6 Q24 27.4 26.6 25.6" fill="none" stroke="{v("honey-line")}" stroke-width="1.2" stroke-linecap="round"/>
<circle cx="19" cy="16.3" r="1.5" {nostroke(v("honey-line"))}/><circle cx="29" cy="16.3" r="1.5" {nostroke(v("honey-line"))}/>
<circle cx="15" cy="21.5" r="2.1" {nostroke(v("rose"))} opacity=".9"/><circle cx="33" cy="21.5" r="2.1" {nostroke(v("rose"))} opacity=".9"/>
<path d="M24 31.6 L18.8 28.6 V34.6Z M24 31.6 L29.2 28.6 V34.6Z" {st("rose","rose",1.2)}/><circle cx="24" cy="31.6" r="1.7" {st("rose","rose",1.2)}/></g>'''
M["plane"] = f'''
<path d="M3 39 C8 33 12 41 17 36 S25 41 29 38" fill="none" stroke="{v("sky-line")}" stroke-width="1.5" stroke-linecap="round" stroke-dasharray="1.5 3.2"/>
<path d="M5 22 L43 6 L24 29Z" {st(CARD,"sky")}/><path d="M43 6 L24 29 L31 43Z" {st("sky","sky")}/>
<path d="M24 29 L26.6 36" fill="none" stroke="{v("sky-line")}" stroke-width="1.2" stroke-linecap="round"/>'''
M["heart"] = f'''
<path d="M24 41 C8 29 6 18 13 13 C18 9.5 22 12 24 16 C26 12 30 9.5 35 13 C42 18 40 29 24 41Z" {st("rose","rose")}/>
<path d="M14 18 Q15.5 14.5 19 14.3" fill="none" stroke="{CARD}" stroke-width="2" stroke-linecap="round"/>'''
M["moon"] = f'''
<path d="M31 6 A18 18 0 1 0 42 33 A14 14 0 0 1 31 6Z" {st("gold","gold")}/>
<path d="M38 10 L39.3 13.2 L42.5 14.5 L39.3 15.8 L38 19 L36.7 15.8 L33.5 14.5 L36.7 13.2Z" {st(CARD,"gold",1.1)}/>
<circle cx="20" cy="26" r="1.3" {nostroke(v("gold-line"))}/><circle cx="14" cy="32" r="1.8" {nostroke(v("gold-line"))} opacity=".55"/>'''
M["bunny"] = f'''
<ellipse cx="16.5" cy="12" rx="4.2" ry="10" transform="rotate(-8 16.5 12)" {st(CARD,"rose")}/><ellipse cx="31.5" cy="12" rx="4.2" ry="10" transform="rotate(8 31.5 12)" {st(CARD,"rose")}/>
<ellipse cx="16.5" cy="13" rx="2" ry="6.5" transform="rotate(-8 16.5 13)" {nostroke(v("rose"))}/><ellipse cx="31.5" cy="13" rx="2" ry="6.5" transform="rotate(8 31.5 13)" {nostroke(v("rose"))}/>
<ellipse cx="24" cy="31" rx="13.5" ry="11.5" {st(CARD,"rose")}/>
<circle cx="19" cy="29.5" r="1.6" {nostroke(FG)}/><circle cx="29" cy="29.5" r="1.6" {nostroke(FG)}/>
<ellipse cx="24" cy="33" rx="1.9" ry="1.3" {st("rose","rose",1)}/>
<path d="M24 34.2 V35.8 M21.8 36.6 Q24 38.2 26.2 36.6" fill="none" stroke="{v("rose-line")}" stroke-width="1.2" stroke-linecap="round"/>
<circle cx="15" cy="34" r="2.6" {nostroke(v("rose"))}/><circle cx="33" cy="34" r="2.6" {nostroke(v("rose"))}/>'''
M["cat"] = f'''
<path d="M9 22 L10 7 L20 13 Q24 12 28 13 L38 7 L39 22 Q41 37 24 41 Q7 37 9 22Z" {st("lilac","lilac")}/>
<path d="M12.3 12 L17.4 15.4 L12.6 19Z M35.7 12 L30.6 15.4 L35.4 19Z" {nostroke(v("rose"))}/>
<ellipse cx="18" cy="25" rx="1.7" ry="2.3" {nostroke(FG)}/><ellipse cx="30" cy="25" rx="1.7" ry="2.3" {nostroke(FG)}/>
<circle cx="18.6" cy="24" r=".6" {nostroke(CARD)}/><circle cx="30.6" cy="24" r=".6" {nostroke(CARD)}/>
<path d="M22.3 29.4 H25.7 L24 31.2Z" {st("rose","rose",1)}/>
<path d="M24 31.2 Q22 34 19.8 32.8 M24 31.2 Q26 34 28.2 32.8" fill="none" stroke="{v("lilac-line")}" stroke-width="1.2" stroke-linecap="round"/>
<path d="M6.5 28 L14.5 29.5 M6.5 33.5 L14.5 32.2 M41.5 28 L33.5 29.5 M41.5 33.5 L33.5 32.2" fill="none" stroke="{v("lilac-line")}" stroke-width="1.1" stroke-linecap="round"/>
<circle cx="14.2" cy="31" r="2.3" {nostroke(v("rose"))} opacity=".85"/><circle cx="33.8" cy="31" r="2.3" {nostroke(v("rose"))} opacity=".85"/>'''
M["cupcake"] = f'''
<path d="M12 28 H36 L33 43 Q32.8 44.2 31.6 44.2 H16.4 Q15.2 44.2 15 43Z" {st("honey","honey")}/>
<path d="M19 28.6 L20.4 43.6 M24 28.6 V43.6 M29 28.6 L27.6 43.6" fill="none" stroke="{v("honey-line")}" stroke-width="1" opacity=".7" stroke-linecap="round"/>
<path d="M10 28 Q8 21 14 19.5 Q14 12.5 22 12.5 Q25 8 29.5 11.5 Q36 11.5 35 18.5 Q40.5 20.5 38 28Z" {st("rose","rose")}/>
<path d="M15 24.2 Q24 27.4 33 24.2" fill="none" stroke="{CARD}" stroke-width="1.7" stroke-linecap="round" opacity=".9"/>
<circle cx="26" cy="8.6" r="3.2" {st("rose","rose",1.2)}/><path d="M26 5.6 Q27 2.6 30 2.8" fill="none" stroke="{v("leaf-line")}" stroke-width="1.3" stroke-linecap="round"/>
<path d="M17 17 l2 -1 M28 16 l2 1.2 M22 20 l2 .2 M32 21.5 l1.6 -1.4" fill="none" stroke="{v("gold-line")}" stroke-width="1.5" stroke-linecap="round"/>'''
M["strawberry"] = f'''
<path d="M24 43 C10 36 7.5 24 11 18 C14 13 20 14 24 16.5 C28 14 34 13 37 18 C40.5 24 38 36 24 43Z" {st("rose","rose")}/>
<path d="M24 16.5 V9.5" fill="none" stroke="{v("leaf-line")}" stroke-width="2.4" stroke-linecap="round"/>
<path d="M24 16.5 L17.5 11.5 L21 15.8 L15.5 15.3 L22 19.2Z M24 16.5 L30.5 11.5 L27 15.8 L32.5 15.3 L26 19.2Z" {st("leaf","leaf",1.2)}/>
<g fill="{v("gold")}" stroke="{v("gold-line")}" stroke-width=".8"><ellipse cx="17" cy="25" rx=".9" ry="1.4"/><ellipse cx="24" cy="24" rx=".9" ry="1.4"/><ellipse cx="31" cy="25" rx=".9" ry="1.4"/><ellipse cx="20.5" cy="31" rx=".9" ry="1.4"/><ellipse cx="27.5" cy="31" rx=".9" ry="1.4"/><ellipse cx="24" cy="37" rx=".9" ry="1.4"/><ellipse cx="15" cy="30" rx=".9" ry="1.4"/><ellipse cx="33" cy="30" rx=".9" ry="1.4"/></g>'''
M["balloon"] = f'''
<path d="M24 4 C34.5 4 38.5 12 36.5 20 C34.5 27.5 28.5 31.5 24 32.5 C19.5 31.5 13.5 27.5 11.5 20 C9.5 12 13.5 4 24 4Z" {st("rose","rose")}/>
<path d="M16.5 12 Q17.6 8.4 21 7.4" fill="none" stroke="{CARD}" stroke-width="2.2" stroke-linecap="round"/>
<path d="M24 32.5 L21.4 36 H26.6Z" {st("rose","rose",1.2)}/>
<path d="M24 36 C19.5 39 28.5 41.5 24 45" fill="none" stroke="{v("rose-line")}" stroke-width="1.2" stroke-linecap="round"/>'''
M["rainbow"] = f'''<g transform="translate(24 24) scale(.9) translate(-24 -24)">
<path d="M5 37 A19 19 0 0 1 43 37" fill="none" stroke="{v("rose-line")}" stroke-width="3.4" stroke-linecap="round"/>
<path d="M9.6 37 A14.4 14.4 0 0 1 38.4 37" fill="none" stroke="{v("gold")}" stroke-width="3.4" stroke-linecap="round"/>
<path d="M14.2 37 A9.8 9.8 0 0 1 33.8 37" fill="none" stroke="{v("leaf")}" stroke-width="3.4" stroke-linecap="round"/>
<path d="M18.8 37 A5.2 5.2 0 0 1 29.2 37" fill="none" stroke="{v("sky")}" stroke-width="3.4" stroke-linecap="round"/>
<path d="M2.5 43 a3.6 3.6 0 0 1 1.2 -7 a5 5 0 0 1 9.4 -.8 a4 4 0 0 1 3 7.8Z" {st(CARD,"sky",1.3)}/>
<path d="M31.5 43 a3.6 3.6 0 0 1 1.2 -7 a5 5 0 0 1 9.4 -.8 a4 4 0 0 1 3 7.8Z" {st(CARD,"sky",1.3)}/></g>'''
M["daisy"] = "".join(f'<ellipse cx="24" cy="11.5" rx="4.7" ry="8" transform="rotate({a} 24 24)" {st(CARD,"rose",1.3)}/>' for a in range(0,360,45)) + f'''
<circle cx="24" cy="24" r="6.2" {st("gold","gold")}/><circle cx="22" cy="22" r="1.4" {nostroke(CARD)} opacity=".8"/>'''
M["bird"] = f'''
<path d="M11 31 L3 30 L5 36.5Z" {st("sky","sky",1.3)}/>
<path d="M10 30 C8 18 18 12 24 14 C26 9 34 8 38 13 C40 16 39 19 37 21 C39 31 32 40 22 40 C16 40 11 36 10 30Z" {st("sky","sky")}/>
<path d="M14 28 Q20 21.5 27.5 27.5 Q22 36 14 28Z" {st(CARD,"sky",1.3)}/>
<circle cx="33.6" cy="15" r="1.5" {nostroke(FG)}/><path d="M38 14 L44 16.6 L38 19.2Z" {st("gold","gold",1.2)}/>
<circle cx="35" cy="19.4" r="2" {nostroke(v("rose"))} opacity=".9"/>
<path d="M19 40 V44.5 M25 40 V44.5" fill="none" stroke="{v("gold-line")}" stroke-width="1.5" stroke-linecap="round"/>'''
M["teacup"] = f'''
<path d="M19 6 Q16 9 19 12 Q22 15 19 18 M27 5 Q24 8 27 11 Q30 14 27 17" fill="none" stroke="{v("sky-line")}" stroke-width="1.4" stroke-linecap="round" opacity=".7"/>
<ellipse cx="24" cy="40" rx="16.5" ry="3.4" {st(CARD,"lilac")}/>
<path d="M36 25 C44 24 44 34.5 35 33" fill="none" stroke="{v("lilac-line")}" stroke-width="2.4" stroke-linecap="round"/>
<path d="M10 21 H37 V28 C37 35.5 31.5 38.5 23.5 38.5 C15.5 38.5 10 35.5 10 28Z" {st(CARD,"lilac")}/>
<path d="M10.8 25 H36.2" fill="none" stroke="{v("lilac")}" stroke-width="3.4"/>
<ellipse cx="23.5" cy="21" rx="13.5" ry="2.4" {st("honey","honey",1.2)}/>
<path d="M23.5 35 C18 31.5 19 29 21 29 C22.4 29 23.2 30 23.5 30.8 C23.8 30 24.6 29 26 29 C28 29 29 31.5 23.5 35Z" {st("rose","rose",1)}/>'''
M["mushroom"] = f'''
<path d="M17.5 28 Q16.5 40 19.5 43.5 H28.5 Q31.5 40 30.5 28Z" {st(CARD,"rose")}/>
<path d="M4.5 28.5 C4.5 14.5 13.5 6 24 6 C34.5 6 43.5 14.5 43.5 28.5 Q24 34 4.5 28.5Z" {st("rose","rose")}/>
<circle cx="15" cy="19" r="3.1" {nostroke(CARD)}/><circle cx="26.5" cy="13.5" r="2.4" {nostroke(CARD)}/><circle cx="34" cy="22" r="3.3" {nostroke(CARD)}/><circle cx="22" cy="25" r="1.9" {nostroke(CARD)}/>
<circle cx="21.2" cy="35" r="1.2" {nostroke(FG)}/><circle cx="26.8" cy="35" r="1.2" {nostroke(FG)}/>
<path d="M22.6 38 Q24 39.4 25.4 38" fill="none" stroke="{v("rose-line")}" stroke-width="1.1" stroke-linecap="round"/>
<circle cx="19.3" cy="37" r="1.4" {nostroke(v("rose"))}/><circle cx="28.7" cy="37" r="1.4" {nostroke(v("rose"))}/>'''
M["bow"] = f'''
<path d="M23 27 L14.5 42.5 L21 40.2 L25.5 28Z M25 27 L33.5 42.5 L27 40.2 L22.5 28Z" {st("rose","rose")}/>
<path d="M24 24 C16 11.5 3.5 13.5 5.5 24 C3.5 34.5 16 36.5 24 24Z" {st("rose","rose")}/>
<path d="M24 24 C32 11.5 44.5 13.5 42.5 24 C44.5 34.5 32 36.5 24 24Z" {st("rose","rose")}/>
<path d="M9 22.5 Q14 20.5 19 23 M39 22.5 Q34 20.5 29 23" fill="none" stroke="{v("rose-line")}" stroke-width=".9" stroke-linecap="round" opacity=".8"/>
<ellipse cx="24" cy="24.2" rx="3.7" ry="4.5" {st("rose","rose")}/>'''
M["pencil"] = f'''
<g transform="rotate(38 24 24)">
<rect x="18.5" y="2" width="11" height="7.5" rx="2.4" {st("rose","rose")}/>
<rect x="18.5" y="9.5" width="11" height="4" {st(CARD,"gold",1.2)}/>
<rect x="18.5" y="13.5" width="11" height="19" {st("gold","gold")}/>
<path d="M24 13.8 V32.2" fill="none" stroke="{v("gold-line")}" stroke-width=".9" opacity=".6"/>
<path d="M18.5 32.5 L24 45 L29.5 32.5Z" {st("honey","honey")}/>
<path d="M21.7 39.2 L24 45 L26.3 39.2Z" {nostroke(FG)}/>
</g>'''

ORDER = list(M.keys())
def sheet():
    cells = ""
    for k in ORDER:
        svg = lambda s: f'<svg width="{s}" height="{s}" viewBox="0 0 48 48">{M[k]}</svg>'
        cells += f'<div class="c"><div class="big">{svg(132)}</div><div class="sm">{svg(40)}{svg(28)}{svg(20)}</div><div class="n">{k}</div></div>'
    return f'<html><body style="margin:0;background:#fffafb;font-family:system-ui;color:#2a2024"><style>:root{{{TOK}}} .g{{display:grid;grid-template-columns:repeat(4,1fr);gap:0}} .c{{padding:14px;border:1px solid #f6e3ea;display:flex;flex-direction:column;align-items:center;gap:8px}} .sm{{display:flex;gap:14px;align-items:end}} .n{{font-size:12px;color:#7a5c68}}</style><div class="g" style="width:760px">{cells}</div></body></html>'
if __name__ == "__main__":
    import sys
    open("/home/claude/scratch/motifs/sheet.html","w").write(sheet())
    from playwright.sync_api import sync_playwright
    with sync_playwright() as p:
        b=p.chromium.launch(); pg=b.new_page(viewport={"width":760,"height":900},device_scale_factor=2); pg.goto("file:///home/claude/scratch/motifs/sheet.html"); pg.screenshot(path="/home/claude/scratch/motifs/sheet.png",full_page=True); b.close()
    print("ok", len(ORDER))
