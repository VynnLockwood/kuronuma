import type { IconName } from "@/components/icons";

export type Lang = "en" | "th";

type Item = { icon: IconName; title: string; body: string };

export type Copy = {
  htmlLang: string;
  meta: { title: string; description: string };
  nav: { id: string; label: string }[];
  signIn: string;
  openApp: string;
  hero: {
    eyebrow: string;
    title: [string, string];
    sub: string;
    primary: string;
    secondary: string;
    chips: string[];
    worksWith: string;
  };
  mock: {
    example: string;
    assistant: string;
    connected: string;
    youSaid: string;
    assistantSaid: string;
    amount: string;
    payee: string;
    date: string;
    save: string;
    edit: string;
    saved: string;
    month: string;
    categories: { name: string; value: string; pct: number }[];
    total: string;
    slipAlt: string;
  };
  demo: { eyebrow: string; title: string; sub: string; points: string[] };
  how: { eyebrow: string; title: string; sub: string; steps: { title: string; body: string }[] };
  features: { eyebrow: string; title: string; sub: string; items: Item[] };
  modules: {
    eyebrow: string;
    title: string;
    sub: string;
    available: string;
    soon: string;
    items: { god: string; area: string; body: string; live: boolean }[];
  };
  privacy: {
    eyebrow: string;
    title: string;
    sub: string;
    storesTitle: string;
    stores: string[];
    notTitle: string;
    nots: string[];
    note: string;
  };
  faq: { eyebrow: string; title: string; items: { q: string; a: string }[] };
  cta: { title: string; sub: string; guide: string };
  footer: { blurb: string; product: string; resources: string; guide: string; source: string; license: string; built: string };
  switchLabel: string;
};

const en: Copy = {
  htmlLang: "en",
  meta: {
    title: "Nemo (เนโม่): memory for your AI assistant",
    description:
      "Nemo, short for Nemosyne, is an open-source memory hub for AI assistants. Snap a bank slip, save it through MCP, and ask how much you spent.",
  },
  nav: [
    { id: "features", label: "Features" },
    { id: "how", label: "How it works" },
    { id: "modules", label: "Modules" },
    { id: "privacy", label: "Privacy" },
    { id: "faq", label: "FAQ" },
  ],
  signIn: "Sign in with Google",
  openApp: "Open dashboard",
  hero: {
    eyebrow: "Open-source memory for your AI assistant",
    title: ["Your assistant’s memory,", "owned by you."],
    sub: "Nemo (short for Nemosyne) keeps your spending and more in one place, so any assistant that supports MCP can answer from your real data. Snap a bank slip, say “save this”, then ask how much you spent this month.",
    primary: "Get started",
    secondary: "See how it works",
    chips: ["MIT licensed", "Google sign-in", "Self-host with Docker", "Slip photos never stored"],
    worksWith: "Works with Claude and other assistants that support MCP connectors.",
  },
  mock: {
    example: "Example",
    assistant: "Assistant",
    connected: "Nemo connected",
    youSaid: "Save this slip, please.",
    assistantSaid: "I read this slip. Is it right?",
    amount: "Amount",
    payee: "Paid to",
    date: "Date",
    save: "Save",
    edit: "Edit",
    saved: "Saved to Nemo",
    month: "This month",
    categories: [
      { name: "Food", value: "฿3,240.50", pct: 100 },
      { name: "Bills", value: "฿2,600.00", pct: 80 },
      { name: "Transport", value: "฿1,120.00", pct: 35 },
    ],
    total: "฿6,960.50",
    slipAlt: "A bank slip",
  },
  demo: {
    eyebrow: "See it in action",
    title: "From a photo to an answer",
    sub: "An illustration of a chat with an assistant connected to Nemo.",
    points: [
      "Attach a bank slip",
      "Check the amount, payee and date",
      "Say yes and it is saved",
      "Ask for totals whenever you like",
    ],
  },
  how: {
    eyebrow: "How it works",
    title: "Three steps. No spreadsheets.",
    sub: "Nemo plugs into the assistant you already use.",
    steps: [
      {
        title: "Connect",
        body: "Paste your Nemo URL into your assistant’s connector settings. Sign in with Google and approve.",
      },
      {
        title: "Snap and confirm",
        body: "Attach a bank slip. Your assistant reads it, shows the amount, payee and date, and saves it when you say yes.",
      },
      {
        title: "Just ask",
        body: "Ask for totals by category, month or merchant in plain language, whenever you like.",
      },
    ],
  },
  features: {
    eyebrow: "Features",
    title: "Everything your assistant needs to remember",
    sub: "Small, careful tools that do one job well.",
    items: [
      {
        icon: "receipt",
        title: "From slip to record",
        body: "Attach a bank slip. Your assistant reads it, shows what it found, and saves only after you confirm.",
      },
      {
        icon: "chat",
        title: "Ask in plain language",
        body: "Totals by category, month or merchant on demand. Amounts in different currencies are never mixed.",
      },
      {
        icon: "layers",
        title: "No double entries",
        body: "The same slip can’t be saved twice, and payments that look identical are flagged for you to confirm.",
      },
      {
        icon: "shield",
        title: "You stay in control",
        body: "Every assistant shows what it wants to access and connects only when you approve.",
      },
      {
        icon: "camera",
        title: "Slip photos stay with you",
        body: "Your assistant reads the image. Nemo only receives the amount, payee and date you confirmed.",
      },
      {
        icon: "code",
        title: "Open and self-hostable",
        body: "MIT licensed. Run it yourself with Docker Compose, or use a Nemo that someone hosts for you.",
      },
    ],
  },
  modules: {
    eyebrow: "Modules",
    title: "A pantheon of modules",
    sub: "Each area of your life is a module named after a Greek god. Start with spending; more are on the way.",
    available: "Available",
    soon: "Coming soon",
    items: [
      { god: "Plutus", area: "Spending", body: "Slips, totals and reports.", live: true },
      { god: "Atlas", area: "Tasks", body: "Tasks your assistant keeps for you.", live: false },
      { god: "Horae", area: "Routines", body: "Daily and weekly routines.", live: false },
      { god: "Hephaestus", area: "Work data", body: "Notes and data from your work.", live: false },
    ],
  },
  privacy: {
    eyebrow: "Privacy",
    title: "Honest about your data",
    sub: "Here is exactly what Nemo keeps, and what it never sees.",
    storesTitle: "What Nemo stores",
    stores: [
      "Your Google name, email and photo, to sign you in",
      "The spending records you confirm",
      "Which assistants you have approved",
    ],
    notTitle: "What Nemo does not store",
    nots: [
      "Your slip photos",
      "Your Google password",
      "Your chats with your assistant. Only tool calls reach Nemo",
    ],
    note: "On a hosted Nemo, whoever runs it can technically access stored records, so choose a host you trust. For full control, self-host: your data stays on your own machine.",
  },
  faq: {
    eyebrow: "FAQ",
    title: "Questions, answered",
    items: [
      {
        q: "What is Nemo?",
        a: "Nemo is the friendly name of Nemosyne, an open-source service that gives your AI assistant a memory it can read and write through MCP, starting with your spending.",
      },
      {
        q: "Why “Nemosyne”?",
        a: "It comes from Mnemosyne, the Greek goddess of memory. We shorten it to Nemo (เนโม่) so it is easy to say.",
      },
      {
        q: "Which assistants work with it?",
        a: "Claude and any assistant or app that supports remote MCP servers with OAuth sign-in.",
      },
      {
        q: "How are slips read?",
        a: "Your assistant reads the photo with its own vision, shows you the amount, payee and date, and saves only after you confirm.",
      },
      {
        q: "Where is my data stored?",
        a: "In a Postgres database run by whoever hosts your Nemo. If you self-host, it stays on your own machine.",
      },
      {
        q: "Can I host it myself?",
        a: "Yes. Nemo is MIT licensed and ships with a Docker Compose setup.",
      },
    ],
  },
  cta: {
    title: "Give your assistant a memory.",
    sub: "Sign in, paste one URL into your assistant, and snap your first slip.",
    guide: "Read the connection guide",
  },
  footer: {
    blurb: "Memory for your AI assistant. Open source, built on Nemosyne.",
    product: "Product",
    resources: "Resources",
    guide: "Connection guide",
    source: "Source code",
    license: "MIT License",
    built: "Nemo is the friendly name of Nemosyne.",
  },
  switchLabel: "Language",
};

const th: Copy = {
  htmlLang: "th",
  meta: {
    title: "เนโม่ (Nemo): ความจำสำหรับผู้ช่วย AI ของคุณ",
    description:
      "เนโม่ (Nemo ชื่อย่อของ Nemosyne) คือศูนย์รวมความจำแบบโอเพนซอร์สสำหรับผู้ช่วย AI ถ่ายสลิปธนาคาร บันทึกผ่าน MCP แล้วถามได้ว่าใช้จ่ายไปเท่าไร",
  },
  nav: [
    { id: "features", label: "ฟีเจอร์" },
    { id: "how", label: "วิธีใช้งาน" },
    { id: "modules", label: "โมดูล" },
    { id: "privacy", label: "ความเป็นส่วนตัว" },
    { id: "faq", label: "คำถามที่พบบ่อย" },
  ],
  signIn: "เข้าสู่ระบบด้วย Google",
  openApp: "เปิดแดชบอร์ด",
  hero: {
    eyebrow: "ความจำแบบโอเพนซอร์สสำหรับผู้ช่วย AI ของคุณ",
    title: ["ความจำของผู้ช่วย AI", "คุณเป็นเจ้าของเอง"],
    sub: "เนโม่ (Nemo ชื่อย่อของ Nemosyne) เก็บค่าใช้จ่ายและข้อมูลอื่น ๆ ของคุณไว้ในที่เดียว ให้ผู้ช่วยที่รองรับ MCP ตอบจากข้อมูลจริงของคุณ ถ่ายสลิปธนาคาร สั่งว่า “บันทึกให้หน่อย” แล้วถามได้เลยว่าเดือนนี้ใช้จ่ายไปเท่าไร",
    primary: "เริ่มใช้งาน",
    secondary: "ดูวิธีใช้งาน",
    chips: ["ลิขสิทธิ์ MIT", "ล็อกอินด้วย Google", "โฮสต์เองด้วย Docker", "ไม่เก็บรูปสลิป"],
    worksWith: "ใช้ได้กับ Claude และผู้ช่วยอื่นที่รองรับ MCP connector",
  },
  mock: {
    example: "ตัวอย่าง",
    assistant: "ผู้ช่วย",
    connected: "เชื่อมต่อเนโม่แล้ว",
    youSaid: "บันทึกสลิปนี้ให้หน่อย",
    assistantSaid: "อ่านสลิปได้แล้ว ข้อมูลถูกต้องไหม",
    amount: "จำนวนเงิน",
    payee: "จ่ายให้",
    date: "วันที่",
    save: "บันทึก",
    edit: "แก้ไข",
    saved: "บันทึกในเนโม่แล้ว",
    month: "เดือนนี้",
    categories: [
      { name: "อาหาร", value: "฿3,240.50", pct: 100 },
      { name: "บิล", value: "฿2,600.00", pct: 80 },
      { name: "เดินทาง", value: "฿1,120.00", pct: 35 },
    ],
    total: "฿6,960.50",
    slipAlt: "สลิปธนาคาร",
  },
  demo: {
    eyebrow: "ดูตัวอย่างการใช้งาน",
    title: "จากรูปถ่ายสู่คำตอบ",
    sub: "ภาพประกอบบทสนทนากับผู้ช่วยที่เชื่อมต่อกับเนโม่",
    points: [
      "แนบสลิปธนาคาร",
      "ตรวจจำนวนเงิน ผู้รับ และวันที่",
      "ตอบตกลงแล้วบันทึกทันที",
      "ถามยอดรวมเมื่อไรก็ได้",
    ],
  },
  how: {
    eyebrow: "วิธีใช้งาน",
    title: "3 ขั้นตอน ไม่ต้องจดสเปรดชีต",
    sub: "เนโม่เชื่อมกับผู้ช่วยที่คุณใช้อยู่แล้ว",
    steps: [
      {
        title: "เชื่อมต่อ",
        body: "วาง URL ของเนโม่ในหน้าตั้งค่า connector ของผู้ช่วย ล็อกอินด้วย Google แล้วกดอนุญาต",
      },
      {
        title: "ถ่ายแล้วยืนยัน",
        body: "แนบรูปสลิป ผู้ช่วยจะอ่านและแสดงจำนวนเงิน ผู้รับ และวันที่ให้ตรวจ เมื่อคุณตอบตกลงจึงบันทึก",
      },
      {
        title: "ถามได้เลย",
        body: "ถามยอดรวมตามหมวดหมู่ เดือน หรือร้านค้า ด้วยภาษาธรรมดา ได้ทุกเมื่อ",
      },
    ],
  },
  features: {
    eyebrow: "ฟีเจอร์",
    title: "ทุกอย่างที่ผู้ช่วยของคุณต้องจำ",
    sub: "เครื่องมือเล็ก ๆ ที่ทำหน้าที่เดียวให้ดี",
    items: [
      {
        icon: "receipt",
        title: "จากสลิปเป็นรายการ",
        body: "แนบสลิปธนาคาร ผู้ช่วยอ่านแล้วแสดงสิ่งที่เจอ และบันทึกต่อเมื่อคุณยืนยันเท่านั้น",
      },
      {
        icon: "chat",
        title: "ถามด้วยภาษาธรรมดา",
        body: "ขอยอดรวมตามหมวดหมู่ เดือน หรือร้านค้าได้ทุกเมื่อ และไม่เอาเงินต่างสกุลมารวมกัน",
      },
      {
        icon: "layers",
        title: "ไม่บันทึกซ้ำ",
        body: "สลิปใบเดิมบันทึกซ้ำไม่ได้ และรายการที่หน้าตาเหมือนกันจะถูกเตือนให้คุณยืนยันก่อน",
      },
      {
        icon: "shield",
        title: "คุณคุมเอง",
        body: "ผู้ช่วยทุกตัวจะแสดงสิ่งที่ต้องการเข้าถึง และเชื่อมต่อได้เมื่อคุณกดอนุมัติเท่านั้น",
      },
      {
        icon: "camera",
        title: "รูปสลิปอยู่กับคุณ",
        body: "ผู้ช่วยของคุณเป็นคนอ่านรูป เนโม่ได้รับเฉพาะจำนวนเงิน ผู้รับ และวันที่ที่คุณยืนยัน",
      },
      {
        icon: "code",
        title: "เปิดและโฮสต์เองได้",
        body: "ลิขสิทธิ์ MIT รันเองด้วย Docker Compose ได้ หรือใช้เนโม่ที่มีคนโฮสต์ให้",
      },
    ],
  },
  modules: {
    eyebrow: "โมดูล",
    title: "กลุ่มโมดูลแห่งเทพเจ้า",
    sub: "แต่ละด้านของชีวิตเป็นโมดูลที่ตั้งชื่อตามเทพเจ้ากรีก เริ่มที่ค่าใช้จ่าย และจะมีอีกหลายอย่างตามมา",
    available: "ใช้งานได้",
    soon: "เร็ว ๆ นี้",
    items: [
      { god: "Plutus", area: "ค่าใช้จ่าย", body: "สลิป ยอดรวม และรายงาน", live: true },
      { god: "Atlas", area: "งาน", body: "งานที่ผู้ช่วยจดไว้ให้คุณ", live: false },
      { god: "Horae", area: "กิจวัตร", body: "กิจวัตรรายวันและรายสัปดาห์", live: false },
      { god: "Hephaestus", area: "ข้อมูลงาน", body: "โน้ตและข้อมูลจากงานของคุณ", live: false },
    ],
  },
  privacy: {
    eyebrow: "ความเป็นส่วนตัว",
    title: "ตรงไปตรงมาเรื่องข้อมูลของคุณ",
    sub: "นี่คือสิ่งที่เนโม่เก็บไว้ และสิ่งที่ไม่เคยเห็น",
    storesTitle: "สิ่งที่เนโม่เก็บ",
    stores: [
      "ชื่อ อีเมล และรูปโปรไฟล์ Google เพื่อใช้ล็อกอิน",
      "รายการค่าใช้จ่ายที่คุณยืนยัน",
      "รายชื่อผู้ช่วยที่คุณอนุมัติ",
    ],
    notTitle: "สิ่งที่เนโม่ไม่เก็บ",
    nots: [
      "รูปสลิปของคุณ",
      "รหัสผ่าน Google ของคุณ",
      "บทสนทนากับผู้ช่วย เนโม่ได้รับเฉพาะคำสั่งเรียกเครื่องมือ",
    ],
    note: "ถ้าใช้เนโม่ที่คนอื่นโฮสต์ ผู้ดูแลระบบเข้าถึงข้อมูลที่เก็บไว้ได้ในทางเทคนิค จึงควรเลือกผู้ที่ไว้ใจได้ ต้องการควบคุมเต็มที่ให้โฮสต์เอง ข้อมูลจะอยู่ในเครื่องของคุณ",
  },
  faq: {
    eyebrow: "คำถามที่พบบ่อย",
    title: "คำถามที่ถามกันบ่อย",
    items: [
      {
        q: "เนโม่คืออะไร",
        a: "เนโม่คือชื่อเรียกง่าย ๆ ของ Nemosyne บริการโอเพนซอร์สที่ให้ผู้ช่วย AI ของคุณมีความจำที่อ่านและเขียนได้ผ่าน MCP เริ่มจากค่าใช้จ่ายของคุณ",
      },
      {
        q: "ทำไมถึงชื่อ Nemosyne",
        a: "มาจาก Mnemosyne เทพีแห่งความจำของกรีก เราย่อเป็น Nemo (เนโม่) ให้พูดง่ายจำง่าย",
      },
      {
        q: "ใช้กับผู้ช่วยตัวไหนได้บ้าง",
        a: "Claude และผู้ช่วยหรือแอปอื่นที่รองรับ remote MCP server พร้อมการล็อกอินแบบ OAuth",
      },
      {
        q: "สลิปถูกอ่านอย่างไร",
        a: "ผู้ช่วยของคุณอ่านรูปด้วยความสามารถของตัวเอง แล้วแสดงจำนวนเงิน ผู้รับ และวันที่ให้คุณตรวจ และบันทึกเมื่อคุณยืนยันเท่านั้น",
      },
      {
        q: "ข้อมูลเก็บไว้ที่ไหน",
        a: "ในฐานข้อมูล Postgres ของผู้ที่โฮสต์เนโม่ ถ้าโฮสต์เอง ข้อมูลจะอยู่ในเครื่องของคุณ",
      },
      {
        q: "โฮสต์เองได้ไหม",
        a: "ได้ เนโม่ใช้ลิขสิทธิ์ MIT และมี Docker Compose ให้พร้อมใช้",
      },
    ],
  },
  cta: {
    title: "ให้ผู้ช่วยของคุณมีความจำ",
    sub: "ล็อกอิน วาง URL เดียวในผู้ช่วยของคุณ แล้วลองถ่ายสลิปใบแรก",
    guide: "อ่านคู่มือการเชื่อมต่อ",
  },
  footer: {
    blurb: "ความจำสำหรับผู้ช่วย AI ของคุณ โอเพนซอร์ส สร้างบน Nemosyne",
    product: "ผลิตภัณฑ์",
    resources: "แหล่งข้อมูล",
    guide: "คู่มือการเชื่อมต่อ",
    source: "ซอร์สโค้ด",
    license: "ลิขสิทธิ์ MIT",
    built: "เนโม่คือชื่อเรียกง่าย ๆ ของ Nemosyne",
  },
  switchLabel: "ภาษา",
};

export const copy: Record<Lang, Copy> = { en, th };
