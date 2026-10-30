import Link from "next/link";
import { AuthButton } from "@/components/AuthButton";
import { AuthErrorRedirect } from "@/components/AuthErrorRedirect";
import { Icon } from "@/components/icons";
import { Logo } from "@/components/Logo";
import { BRAND, REPO_URL } from "@/lib/brand";
import { copy, type Lang } from "./copy";

const home = (lang: Lang) => (lang === "th" ? "/th" : "/");

function LangSwitch({ lang, label }: { lang: Lang; label: string }) {
  return (
    <div className="lang" role="group" aria-label={label}>
      <Link href="/" hrefLang="en" lang="en" aria-current={lang === "en" ? "true" : undefined}>
        EN
      </Link>
      <Link href="/th" hrefLang="th" lang="th" aria-current={lang === "th" ? "true" : undefined}>
        ไทย
      </Link>
    </div>
  );
}

function SectionHead({ eyebrow, title, sub }: { eyebrow: string; title: string; sub?: string }) {
  return (
    <div className="section-head">
      <p className="eyebrow">{eyebrow}</p>
      <h2>{title}</h2>
      {sub && <p className="lead">{sub}</p>}
    </div>
  );
}

export function Landing({ lang }: { lang: Lang }) {
  const c = copy[lang];

  return (
    <div lang={c.htmlLang} className="landing">
      <AuthErrorRedirect />
      <a className="skip" href="#main">
        Skip to content
      </a>

      <header className="site-header">
        <div className="container header-row">
          <Logo href={home(lang)} />
          <nav className="nav" aria-label="Primary">
            {c.nav.map((n) => (
              <a key={n.id} href={`#${n.id}`}>
                {n.label}
              </a>
            ))}
          </nav>
          <div className="header-actions">
            <LangSwitch lang={lang} label={c.switchLabel} />
            <AuthButton signIn={c.signIn} openApp={c.openApp} className="btn btn-primary btn-sm" />
          </div>
        </div>
      </header>

      <main id="main">
        {/* Hero */}
        <section className="hero">
          <div className="container hero-grid">
            <div className="hero-copy">
              <p className="eyebrow pill">
                <Icon name="sparkle" size={16} /> {c.hero.eyebrow}
              </p>
              <h1>
                {c.hero.title[0]} <span className="accent">{c.hero.title[1]}</span>
              </h1>
              <p className="lead">{c.hero.sub}</p>
              <div className="hero-cta">
                <AuthButton signIn={c.hero.primary} openApp={c.openApp} className="btn btn-primary btn-lg" />
                <a className="btn btn-ghost btn-lg" href="#how">
                  {c.hero.secondary} <Icon name="arrow" size={18} />
                </a>
              </div>
              <ul className="chips" aria-label="Highlights">
                {c.hero.chips.map((chip) => (
                  <li key={chip}>
                    <Icon name="check" size={14} /> {chip}
                  </li>
                ))}
              </ul>
              <p className="works-with">{c.hero.worksWith}</p>
            </div>

            <div className="hero-visual" aria-hidden="false">
              <img
                className="hero-logo"
                src="/brand/nemo-logo.png"
                width={440}
                height={440}
                alt={`${BRAND.name} logo`}
              />

              <div className="float-toast" role="img" aria-label={c.mock.example}>
                <span className="icon-tile">
                  <Icon name="check" size={20} />
                </span>
                <div>
                  <strong>{c.mock.saved}</strong>
                  <span>฿450.00 · Grab · 28 Oct 2026</span>
                </div>
              </div>

              <div className="mock mock-summary" role="img" aria-label={c.mock.example}>
                <p className="mock-title">{c.mock.month}</p>
                <p className="mock-total">{c.mock.total}</p>
                <ul className="bars">
                  {c.mock.categories.map((cat) => (
                    <li key={cat.name}>
                      <span className="bar-label">
                        {cat.name} <em>{cat.value}</em>
                      </span>
                      <span className="bar-track">
                        <span className="bar-fill" style={{ width: `${cat.pct}%` }} />
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </section>

        {/* Demo */}
        <section id="demo" className="section">
          <div className="container demo-grid">
            <div className="demo-copy">
              <p className="eyebrow">{c.demo.eyebrow}</p>
              <h2>{c.demo.title}</h2>
              <p className="lead">{c.demo.sub}</p>
              <ul className="ticks">
                {c.demo.points.map((t) => (
                  <li key={t}>
                    <span className="tick ok">
                      <Icon name="check" size={14} />
                    </span>
                    {t}
                  </li>
                ))}
              </ul>
            </div>
            <div className="mock mock-chat" role="img" aria-label={c.mock.example}>
              <div className="mock-bar">
                <span className="dot" />
                <strong>{c.mock.assistant}</strong>
                <span className="tag">{c.mock.connected}</span>
              </div>
              <div className="bubble bubble-user">
                <span className="slip" aria-hidden="true">
                  <i />
                  <i />
                  <i />
                  <b />
                </span>
                <span>{c.mock.youSaid}</span>
              </div>
              <div className="bubble bubble-bot">
                <span>{c.mock.assistantSaid}</span>
                <dl className="facts">
                  <div>
                    <dt>{c.mock.amount}</dt>
                    <dd>฿450.00</dd>
                  </div>
                  <div>
                    <dt>{c.mock.payee}</dt>
                    <dd>Grab</dd>
                  </div>
                  <div>
                    <dt>{c.mock.date}</dt>
                    <dd>28 Oct 2026</dd>
                  </div>
                </dl>
                <div className="mock-actions">
                  <span className="mini mini-primary">{c.mock.save}</span>
                  <span className="mini">{c.mock.edit}</span>
                </div>
              </div>
              <div className="status ok">
                <Icon name="check" size={14} /> {c.mock.saved}
              </div>
              <span className="example-label">{c.mock.example}</span>
            </div>
          </div>
        </section>

        {/* How it works */}
        <section id="how" className="section surface">
          <div className="container">
            <SectionHead eyebrow={c.how.eyebrow} title={c.how.title} sub={c.how.sub} />
            <ol className="steps-grid">
              {c.how.steps.map((s, i) => (
                <li key={s.title} className="card step">
                  <span className="step-num" aria-hidden="true">
                    {i + 1}
                  </span>
                  <h3>{s.title}</h3>
                  <p>{s.body}</p>
                </li>
              ))}
            </ol>
          </div>
        </section>

        {/* Features */}
        <section id="features" className="section">
          <div className="container">
            <SectionHead eyebrow={c.features.eyebrow} title={c.features.title} sub={c.features.sub} />
            <ul className="feature-grid">
              {c.features.items.map((f) => (
                <li key={f.title} className="card feature">
                  <span className="icon-tile">
                    <Icon name={f.icon} />
                  </span>
                  <h3>{f.title}</h3>
                  <p>{f.body}</p>
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* Modules */}
        <section id="modules" className="section surface">
          <div className="container">
            <SectionHead eyebrow={c.modules.eyebrow} title={c.modules.title} sub={c.modules.sub} />
            <ul className="module-grid">
              {c.modules.items.map((m) => (
                <li key={m.god} className={`card module-card${m.live ? " live" : ""}`}>
                  <span className={`badge${m.live ? " badge-ok" : ""}`}>
                    {m.live ? c.modules.available : c.modules.soon}
                  </span>
                  <h3>{m.god}</h3>
                  <p className="module-area">{m.area}</p>
                  <p>{m.body}</p>
                </li>
              ))}
            </ul>
          </div>
        </section>

        {/* Privacy */}
        <section id="privacy" className="section">
          <div className="container">
            <SectionHead eyebrow={c.privacy.eyebrow} title={c.privacy.title} sub={c.privacy.sub} />
            <div className="privacy-grid">
              <div className="card privacy-card">
                <h3>{c.privacy.storesTitle}</h3>
                <ul className="ticks">
                  {c.privacy.stores.map((t) => (
                    <li key={t}>
                      <span className="tick ok">
                        <Icon name="check" size={14} />
                      </span>
                      {t}
                    </li>
                  ))}
                </ul>
              </div>
              <div className="card privacy-card">
                <h3>{c.privacy.notTitle}</h3>
                <ul className="ticks">
                  {c.privacy.nots.map((t) => (
                    <li key={t}>
                      <span className="tick no">
                        <Icon name="cross" size={14} />
                      </span>
                      {t}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
            <p className="note">
              <Icon name="lock" size={18} /> <span>{c.privacy.note}</span>
            </p>
          </div>
        </section>

        {/* FAQ */}
        <section id="faq" className="section surface">
          <div className="container narrow">
            <SectionHead eyebrow={c.faq.eyebrow} title={c.faq.title} />
            <div className="faq">
              {c.faq.items.map((f) => (
                <details key={f.q}>
                  <summary>
                    <span>{f.q}</span>
                    <Icon name="chevron" size={20} />
                  </summary>
                  <p>{f.a}</p>
                </details>
              ))}
            </div>
          </div>
        </section>

        {/* Final call to action */}
        <section className="section cta-section">
          <div className="container">
            <div className="cta-band">
              <img src="/brand/nemo-mark.svg" width={56} height={56} alt="" />
              <h2>{c.cta.title}</h2>
              <p className="lead">{c.cta.sub}</p>
              <div className="hero-cta center">
                <AuthButton signIn={c.signIn} openApp={c.openApp} className="btn btn-primary btn-lg" />
                <Link className="btn btn-ghost btn-lg" href="/connect">
                  {c.cta.guide} <Icon name="arrow" size={18} />
                </Link>
              </div>
            </div>
          </div>
        </section>
      </main>

      <footer className="site-footer">
        <div className="container footer-grid">
          <div>
            <Logo href={home(lang)} />
            <p className="muted footer-blurb">{c.footer.blurb}</p>
          </div>
          <nav aria-label={c.footer.product}>
            <h4>{c.footer.product}</h4>
            <ul>
              {c.nav.slice(0, 4).map((n) => (
                <li key={n.id}>
                  <a href={`#${n.id}`}>{n.label}</a>
                </li>
              ))}
            </ul>
          </nav>
          <nav aria-label={c.footer.resources}>
            <h4>{c.footer.resources}</h4>
            <ul>
              <li>
                <Link href="/connect">{c.footer.guide}</Link>
              </li>
              {REPO_URL && (
                <li>
                  <a href={REPO_URL} rel="noopener noreferrer">
                    {c.footer.source}
                  </a>
                </li>
              )}
              <li>{c.footer.license}</li>
            </ul>
          </nav>
          <div>
            <h4>{c.switchLabel}</h4>
            <LangSwitch lang={lang} label={c.switchLabel} />
          </div>
        </div>
        <div className="container footer-bottom">
          <span>
            © {new Date().getFullYear()} {BRAND.name} ({BRAND.thai})
          </span>
          <span>{c.footer.built}</span>
        </div>
      </footer>
    </div>
  );
}
