package cerberus

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
)

// The consent page is where a person decides whether an assistant may reach their data. It has to answer "should I
// approve this?" in plain words, in the person's language, and offer a way to approve less than was asked for.

// consentText is the wording of the page in one language. Fields ending in Fmt take the client name or origin.
type consentText struct {
	PageTitle, TitleFmt, SignedInAs, AfterFmt         string
	CheckTitle, CheckYes, CheckNo, CanDo              string
	AccessTitle, FullTitle, FullSub, ReadTitle        string
	ReadSub, Approve, Deny, Fine, SwitchLabel, Switch string
}

var consentCopy = map[string]consentText{
	"en": {
		PageTitle:   "Connect to Nemo",
		TitleFmt:    "Connect %s to your Nemo?",
		SignedInAs:  "Signed in as",
		AfterFmt:    "After you decide, you’ll go back to %s.",
		CheckTitle:  "Check before you approve",
		CheckYes:    "Approve if you just added Nemo as a connector in Gemini, Claude or ChatGPT.",
		CheckNo:     "Deny if you didn’t start this, or someone sent you the link.",
		CanDo:       "It will be able to",
		AccessTitle: "Access level",
		FullTitle:   "Read and change",
		FullSub:     "Everything listed above.",
		ReadTitle:   "Read only",
		ReadSub:     "It can look, but can’t add, change or delete. A safe way to try it.",
		Approve:     "Approve",
		Deny:        "Deny",
		Fine:        "The app’s name is chosen by the app and isn’t verified by Nemo.",
		SwitchLabel: "ไทย",
		Switch:      "th",
	},
	"th": {
		PageTitle:   "เชื่อมต่อกับ Nemo",
		TitleFmt:    "เชื่อมต่อ %s กับ Nemo ของคุณไหม?",
		SignedInAs:  "เข้าสู่ระบบเป็น",
		AfterFmt:    "หลังตัดสินใจ คุณจะถูกพากลับไปที่ %s",
		CheckTitle:  "ตรวจสอบก่อนอนุมัติ",
		CheckYes:    "อนุมัติถ้าคุณเพิ่งเพิ่ม Nemo เป็นตัวเชื่อมต่อใน Gemini, Claude หรือ ChatGPT",
		CheckNo:     "ปฏิเสธถ้าคุณไม่ได้เป็นคนเริ่ม หรือมีคนส่งลิงก์นี้มาให้",
		CanDo:       "สิ่งที่แอปนี้จะทำได้",
		AccessTitle: "ระดับการเข้าถึง",
		FullTitle:   "อ่านและแก้ไขได้",
		FullSub:     "ทุกอย่างที่แสดงไว้ข้างบน",
		ReadTitle:   "อ่านอย่างเดียว",
		ReadSub:     "ดูได้ แต่เพิ่ม แก้ไข หรือลบไม่ได้ เหมาะกับการลองใช้ก่อน",
		Approve:     "อนุมัติ",
		Deny:        "ปฏิเสธ",
		Fine:        "ชื่อแอปเป็นชื่อที่แอปตั้งเอง Nemo ไม่ได้ตรวจสอบ",
		SwitchLabel: "EN",
		Switch:      "en",
	},
}

// areaNames labels a group of scopes by the part of the scope name before the colon.
var areaNames = map[string]map[string]string{
	"plutus":     {"en": "Spending", "th": "ค่าใช้จ่าย"},
	"categories": {"en": "Categories", "th": "หมวดหมู่"},
	"atlas":      {"en": "Tasks, to-dos and notes", "th": "งาน สิ่งที่ต้องทำ และโน้ต"},
}

// scopeLines says what each known scope allows, in everyday words. An unknown scope falls back to the description
// it was registered with. Saving spending or a memory can also create a category, which the lines say.
var scopeLines = map[string]map[string]string{
	"plutus:read":      {"en": "See your spending records and totals", "th": "ดูรายการและยอดรวมค่าใช้จ่ายของคุณ"},
	"plutus:write":     {"en": "Add and delete spending records, and create categories as it saves", "th": "เพิ่มและลบรายการค่าใช้จ่าย และสร้างหมวดหมู่ใหม่ตอนบันทึก"},
	"categories:read":  {"en": "See your category list", "th": "ดูรายการหมวดหมู่ของคุณ"},
	"categories:write": {"en": "Create, rename and merge categories", "th": "สร้าง เปลี่ยนชื่อ และรวมหมวดหมู่"},
	"atlas:read":       {"en": "See your tasks, to-dos and notes", "th": "ดูงาน สิ่งที่ต้องทำ และโน้ตของคุณ"},
	"atlas:write":      {"en": "Add, change and delete tasks, to-dos and notes, and create categories as it saves", "th": "เพิ่ม แก้ไข และลบงาน สิ่งที่ต้องทำ และโน้ต และสร้างหมวดหมู่ใหม่ตอนบันทึก"},
}

type consentGroup struct {
	Area  string
	Lines []string
}

// pickLang chooses the page language: an explicit ?lang=th or ?lang=en wins, then the first of Thai or English in
// the browser's Accept-Language, then English.
func pickLang(r *http.Request, explicit string) string {
	if _, ok := consentCopy[explicit]; ok {
		return explicit
	}
	for _, part := range strings.Split(r.Header.Get("Accept-Language"), ",") {
		tag := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		switch {
		case tag == "th" || strings.HasPrefix(tag, "th-"):
			return "th"
		case tag == "en" || strings.HasPrefix(tag, "en-"):
			return "en"
		}
	}
	return "en"
}

// readOnlyScopes keeps only the scopes that read.
func readOnlyScopes(scopes []string) []string {
	var out []string
	for _, s := range scopes {
		if strings.HasSuffix(s, ":read") {
			out = append(out, s)
		}
	}
	return out
}

// consentGroups lists the requested scopes grouped by area, reads before changes.
func (o *OAuth) consentGroups(names []string, lang string) []consentGroup {
	ordered := slices.Clone(names)
	slices.SortStableFunc(ordered, func(a, b string) int {
		ra, rb := strings.HasSuffix(a, ":read"), strings.HasSuffix(b, ":read")
		switch {
		case ra == rb:
			return 0
		case ra:
			return -1
		}
		return 1
	})
	var groups []consentGroup
	index := map[string]int{}
	for _, name := range ordered {
		area, _, _ := strings.Cut(name, ":")
		i, ok := index[area]
		if !ok {
			label := areaNames[area][lang]
			if label == "" {
				label = area
			}
			groups = append(groups, consentGroup{Area: label})
			i = len(groups) - 1
			index[area] = i
		}
		line := scopeLines[name][lang]
		if line == "" {
			for _, s := range o.opt.Scopes {
				if s.Name == name {
					line = s.Description
				}
			}
		}
		if line == "" {
			line = name
		}
		groups[i].Lines = append(groups[i].Lines, line)
	}
	return groups
}

type consentView struct {
	Lang        string
	T           consentText
	Title       string
	After       string
	Email       string
	EmailOff    template.HTML // Cloudflare's markers: keep the address as plain text
	EmailOn     template.HTML
	Groups      []consentGroup
	CanReadOnly bool
	SwitchURL   string
	Redirect    string
	ClientID    string
	RedirectURI string
	Challenge   string
	State       string
	Scope       string
	Resource    string
	CSRF        string
}

func (o *OAuth) renderConsent(w http.ResponseWriter, r *http.Request, user User, req authzReq) {
	lang := pickLang(r, r.URL.Query().Get("lang"))
	t := consentCopy[lang]
	q := r.URL.Query()
	q.Set("lang", t.Switch)

	redirect := redirectLabel(req.RedirectURI)
	securePageHeaders(w)
	err := consentTmpl.Execute(w, consentView{
		Lang:        lang,
		T:           t,
		Title:       fmt.Sprintf(t.TitleFmt, req.Client.Name),
		After:       fmt.Sprintf(t.AfterFmt, redirect),
		Email:       user.Email,
		EmailOff:    "<!--email_off-->",
		EmailOn:     "<!--/email_off-->",
		Groups:      o.consentGroups(req.Scopes, lang),
		CanReadOnly: len(readOnlyScopes(req.Scopes)) > 0,
		SwitchURL:   "/api/oauth/authorize?" + q.Encode(),
		Redirect:    redirect,
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURI,
		Challenge:   req.Challenge,
		State:       req.State,
		Scope:       strings.Join(req.Scopes, " "),
		Resource:    req.Resource,
		CSRF:        o.csrf(user.ID, req),
	})
	if err != nil {
		o.log.Error("render consent failed", "err", err)
	}
}

// pageStyle is shared by the consent and error pages. The page's security policy allows no scripts or fonts, so it
// is plain inline CSS with the system fonts.
const pageStyle = `*{box-sizing:border-box}
body{margin:0;background:#fff;color:#111;font:16px/1.55 system-ui,-apple-system,"Segoe UI",Roboto,"Noto Sans Thai",Sarabun,sans-serif}
main{max-width:32rem;margin:0 auto;padding:1.25rem 1.25rem 3rem}
header{display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:.5rem 0 1rem}
.brand{display:flex;align-items:center;gap:.6rem;text-decoration:none;color:#111}
.brand img{width:44px;height:44px;display:block}
.brand span{font-weight:600;letter-spacing:.2em;font-size:1rem}
.lang{font-size:.85rem;color:#444;border:1px solid #bbb;border-radius:.25rem;padding:.2rem .6rem;text-decoration:none}
.card{border:1px solid #d4d4d4;border-radius:.5rem;padding:1.25rem;background:#fafafa}
h1{font-size:1.35rem;line-height:1.3;margin:0 0 .35rem;font-weight:600}
p{margin:.4rem 0}
.muted{color:#555;font-size:.9rem}
.check{background:#fff7e6;border:1px solid #ecd49a;border-radius:.4rem;padding:.7rem .9rem;margin:1rem 0;font-size:.92rem}
.check strong{display:block;margin-bottom:.25rem}
h2{font-size:.85rem;font-weight:600;color:#555;margin:1.1rem 0 .4rem}
.area{border:1px solid #d4d4d4;border-radius:.4rem;background:#fff;padding:.55rem .8rem;margin:.4rem 0}
.area b{display:block;font-size:.95rem}
.area ul{margin:.2rem 0 0;padding-left:1.1rem;color:#444;font-size:.9rem}
label.opt{display:flex;gap:.6rem;align-items:flex-start;margin:.4rem 0;cursor:pointer}
label.opt input{margin-top:.4rem}
label.opt small{display:block;color:#555}
.row{display:flex;gap:.75rem;margin-top:1.25rem}
button{font:inherit;border-radius:.4rem;padding:.65rem 1.4rem;border:1px solid #111;cursor:pointer;min-height:2.75rem}
button:focus-visible,.lang:focus-visible,a:focus-visible{outline:3px solid #2f6fed;outline-offset:2px}
.approve{background:#111;color:#fff}.deny{background:#fff;color:#111}
.fine{margin-top:1rem;color:#666;font-size:.8rem}
code{background:#eee;padding:.05rem .3rem;border-radius:.2rem;word-break:break-all}
@media(max-width:30rem){.row button{flex:1}}`

var consentTmpl = template.Must(template.New("consent").Parse(`<!doctype html>
<html lang="{{.Lang}}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.T.PageTitle}}</title>
<style>` + pageStyle + `</style></head><body><main>
<header><a class="brand" href="/"><img src="/brand/nemo-logo.png" width="44" height="44" alt=""><span>NEMO</span></a><a class="lang" href="{{.SwitchURL}}">{{.T.SwitchLabel}}</a></header>
<div class="card">
<h1>{{.Title}}</h1>
<p class="muted">{{.T.SignedInAs}} {{.EmailOff}}{{.Email}}{{.EmailOn}}</p>
<p class="muted">{{.After}}</p>
<div class="check"><strong>{{.T.CheckTitle}}</strong>
<div>&#10003; {{.T.CheckYes}}</div>
<div>&#10005; {{.T.CheckNo}}</div></div>
<h2>{{.T.CanDo}}</h2>
{{range .Groups}}<div class="area"><b>{{.Area}}</b><ul>{{range .Lines}}<li>{{.}}</li>{{end}}</ul></div>
{{end}}<form method="post" action="/api/oauth/authorize">
<input type="hidden" name="response_type" value="code">
<input type="hidden" name="client_id" value="{{.ClientID}}">
<input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
<input type="hidden" name="code_challenge" value="{{.Challenge}}">
<input type="hidden" name="code_challenge_method" value="S256">
<input type="hidden" name="state" value="{{.State}}">
<input type="hidden" name="scope" value="{{.Scope}}">
<input type="hidden" name="resource" value="{{.Resource}}">
<input type="hidden" name="csrf" value="{{.CSRF}}">
<input type="hidden" name="lang" value="{{.Lang}}">
{{if .CanReadOnly}}<h2>{{.T.AccessTitle}}</h2>
<label class="opt"><input type="radio" name="access" value="full" checked><span>{{.T.FullTitle}}<small>{{.T.FullSub}}</small></span></label>
<label class="opt"><input type="radio" name="access" value="read"><span>{{.T.ReadTitle}}<small>{{.T.ReadSub}}</small></span></label>
{{else}}<input type="hidden" name="access" value="full">
{{end}}<div class="row">
<button class="approve" type="submit" name="decision" value="approve">{{.T.Approve}}</button>
<button class="deny" type="submit" name="decision" value="deny">{{.T.Deny}}</button>
</div></form>
<p class="fine">{{.T.Fine}}</p>
</div></main></body></html>`))

var errorTmpl = template.Must(template.New("error").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Nemo</title>
<style>` + pageStyle + `</style></head><body><main>
<header><a class="brand" href="/"><img src="/brand/nemo-logo.png" width="44" height="44" alt=""><span>NEMO</span></a></header>
<div class="card"><p>{{.}}</p></div></main></body></html>`))
