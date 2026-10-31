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
    panelTitle: string;
    panelNote: string;
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
  mission: { eyebrow: string; title: string; sub: string; items: { title: string; body: string }[]; note: string };
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
    title: "Nemo (เนโม่): a memory for your AI assistant",
    description:
      "Nemo, short for Nemosyne, is an open-source memory for AI assistants. Connect once through MCP and your assistant can remember what matters. Spending is the first module.",
  },
  nav: [
    { id: "features", label: "Features" },
    { id: "how", label: "How it works" },
    { id: "modules", label: "Modules" },
    { id: "mission", label: "Mission" },
    { id: "privacy", label: "Privacy" },
    { id: "faq", label: "FAQ" },
  ],
  signIn: "Sign in with Google",
  openApp: "Open dashboard",
  hero: {
    eyebrow: "Open-source memory for AI",
    title: ["Your assistant’s memory,", "owned by you."],
    sub: "Nemo (short for Nemosyne) gives any assistant that supports MCP a memory it can read and write. You approve what it can see. It remembers your spending, tasks, to-dos and notes, and more kinds of data are on the way.",
    primary: "Get started",
    secondary: "See how it works",
    chips: ["MIT licensed", "Google sign-in", "Self-host with Docker", "Memory you own"],
    worksWith: "Works with Claude and other assistants that support MCP connectors.",
    panelTitle: "Memory",
    panelNote: "Every assistant you approve reads the same memory.",
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
    slipAlt: "A bank slip",
  },
  demo: {
    eyebrow: "Example: Spending",
    title: "From a photo to a memory",
    sub: "One example from the Spending module: a chat with an assistant connected to Nemo.",
    points: [
      "Attach a bank slip",
      "Check the amount, payee and date",
      "Say yes and it becomes a memory",
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
        title: "Let it remember",
        body: "Your assistant saves what you confirm. Today that means spending records, such as a bank slip you checked.",
      },
      {
        title: "Just ask",
        body: "Ask your assistant what it remembers, in plain language. Totals by category, month or merchant work today.",
      },
    ],
  },
  features: {
    eyebrow: "Features",
    title: "Memory your assistant can actually use",
    sub: "Small, careful tools that do one job well.",
    items: [
      {
        icon: "chat",
        title: "One memory, any assistant",
        body: "Connect once. Every assistant you approve reads and writes the same memory.",
      },
      {
        icon: "shield",
        title: "You stay in control",
        body: "Every assistant shows what it wants to access and connects only when you approve.",
      },
      {
        icon: "receipt",
        title: "Spending, first",
        body: "Attach a bank slip. Your assistant reads it, shows what it found, and saves only after you confirm.",
      },
      {
        icon: "layers",
        title: "No double entries",
        body: "The same slip can’t be saved twice, and payments that look identical are flagged for you to confirm.",
      },
      {
        icon: "camera",
        title: "Photos stay with you",
        body: "Your assistant reads the image. Nemo only receives the fields you confirmed.",
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
    sub: "Each kind of memory is a module named after a Greek god. Spending and tasks are live; more are on the way.",
    available: "Available",
    soon: "Coming soon",
    items: [
      { god: "Plutus", area: "Spending", body: "Slips, totals and reports.", live: true },
      { god: "Atlas", area: "Tasks and notes", body: "Tasks, to-dos and notes your assistant keeps for you.", live: true },
      { god: "Horae", area: "Routines", body: "Daily and weekly routines.", live: false },
      { god: "Hephaestus", area: "Work data", body: "Notes and data from your work.", live: false },
    ],
  },
  mission: {
    eyebrow: "Mission",
    title: "Get better at every kind of data",
    sub: "Spending and tasks are where Nemo starts. Our mission is to make it a good memory for the other kinds of data you want your assistant to keep.",
    items: [
      { title: "Routines (Horae)", body: "Daily and weekly habits and schedules." },
      { title: "Work data (Hephaestus)", body: "Notes and data from your work." },
      { title: "Better at what it keeps", body: "Smarter organisation of what is stored. Categories and payee rules already group spending; more is planned." },
    ],
    note: "These are goals, not release dates. A module is called available only when it works.",
  },
  privacy: {
    eyebrow: "Privacy",
    title: "Honest about your data",
    sub: "Here is exactly what Nemo keeps, and what it never sees.",
    storesTitle: "What Nemo stores",
    stores: [
      "Your Google name, email and photo, to sign you in",
      "The memories you confirm, such as spending records",
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
        a: "Nemo is the friendly name of Nemosyne, an open-source memory for AI assistants. Your assistant reads and writes it through MCP. Spending is the first module.",
      },
      {
        q: "Why “Nemosyne”?",
        a: "It comes from Mnemosyne, the Greek goddess of memory. We shorten it to Nemo (เนโม่) so it is easy to say.",
      },
      {
        q: "What can it remember today?",
        a: "The spending records you confirm, and the tasks, to-dos and notes your assistant saves for you. Routines and work data are planned modules and are not available yet.",
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
    sub: "Sign in, paste one URL into your assistant, and let it start remembering.",
    guide: "Read the connection guide",
  },
  footer: {
    blurb: "A memory for your AI assistant. Open source, built on Nemosyne.",
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
      "เนโม่ (Nemo ชื่อย่อของ Nemosyne) คือความจำแบบโอเพนซอร์สสำหรับผู้ช่วย AI เชื่อมต่อครั้งเดียวผ่าน MCP ผู้ช่วยของคุณก็จดจำสิ่งสำคัญได้ เริ่มจากค่าใช้จ่ายเป็นโมดูลแรก",
  },
  nav: [
    { id: "features", label: "ฟีเจอร์" },
    { id: "how", label: "วิธีใช้งาน" },
    { id: "modules", label: "โมดูล" },
    { id: "mission", label: "พันธกิจ" },
    { id: "privacy", label: "ความเป็นส่วนตัว" },
    { id: "faq", label: "คำถามที่พบบ่อย" },
  ],
  signIn: "เข้าสู่ระบบด้วย Google",
  openApp: "เปิดแดชบอร์ด",
  hero: {
    eyebrow: "ความจำแบบโอเพนซอร์สสำหรับ AI",
    title: ["ความจำของผู้ช่วย AI", "คุณเป็นเจ้าของเอง"],
    sub: "เนโม่ (Nemo ชื่อย่อของ Nemosyne) ให้ผู้ช่วยที่รองรับ MCP มีความจำที่อ่านและเขียนได้ คุณเป็นคนอนุมัติว่าผู้ช่วยเห็นอะไรได้บ้าง จำค่าใช้จ่าย งาน สิ่งที่ต้องทำ และโน้ตของคุณได้ และจะมีข้อมูลประเภทอื่นตามมา",
    primary: "เริ่มใช้งาน",
    secondary: "ดูวิธีใช้งาน",
    chips: ["ลิขสิทธิ์ MIT", "ล็อกอินด้วย Google", "โฮสต์เองด้วย Docker", "ความจำที่คุณเป็นเจ้าของ"],
    worksWith: "ใช้ได้กับ Claude และผู้ช่วยอื่นที่รองรับ MCP connector",
    panelTitle: "ความจำ",
    panelNote: "ผู้ช่วยทุกตัวที่คุณอนุมัติอ่านความจำเดียวกัน",
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
    slipAlt: "สลิปธนาคาร",
  },
  demo: {
    eyebrow: "ตัวอย่าง: ค่าใช้จ่าย",
    title: "จากรูปถ่ายสู่ความจำ",
    sub: "ตัวอย่างจากโมดูลค่าใช้จ่าย: บทสนทนากับผู้ช่วยที่เชื่อมต่อกับเนโม่",
    points: [
      "แนบสลิปธนาคาร",
      "ตรวจจำนวนเงิน ผู้รับ และวันที่",
      "ตอบตกลงแล้วกลายเป็นความจำ",
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
        title: "ให้ผู้ช่วยจำ",
        body: "ผู้ช่วยบันทึกสิ่งที่คุณยืนยัน ตอนนี้คือรายการค่าใช้จ่าย เช่นสลิปธนาคารที่คุณตรวจแล้ว",
      },
      {
        title: "ถามได้เลย",
        body: "ถามผู้ช่วยว่าจำอะไรไว้บ้างด้วยภาษาธรรมดา ตอนนี้ถามยอดรวมตามหมวดหมู่ เดือน หรือร้านค้าได้",
      },
    ],
  },
  features: {
    eyebrow: "ฟีเจอร์",
    title: "ความจำที่ผู้ช่วยใช้ได้จริง",
    sub: "เครื่องมือเล็ก ๆ ที่ทำหน้าที่เดียวให้ดี",
    items: [
      {
        icon: "chat",
        title: "ความจำเดียว ผู้ช่วยหลายตัว",
        body: "เชื่อมต่อครั้งเดียว ผู้ช่วยทุกตัวที่คุณอนุมัติอ่านและเขียนความจำเดียวกัน",
      },
      {
        icon: "shield",
        title: "คุณคุมเอง",
        body: "ผู้ช่วยทุกตัวจะแสดงสิ่งที่ต้องการเข้าถึง และเชื่อมต่อได้เมื่อคุณกดอนุมัติเท่านั้น",
      },
      {
        icon: "receipt",
        title: "เริ่มที่ค่าใช้จ่าย",
        body: "แนบสลิปธนาคาร ผู้ช่วยอ่านแล้วแสดงสิ่งที่เจอ และบันทึกต่อเมื่อคุณยืนยันเท่านั้น",
      },
      {
        icon: "layers",
        title: "ไม่บันทึกซ้ำ",
        body: "สลิปใบเดิมบันทึกซ้ำไม่ได้ และรายการที่หน้าตาเหมือนกันจะถูกเตือนให้คุณยืนยันก่อน",
      },
      {
        icon: "camera",
        title: "รูปภาพอยู่กับคุณ",
        body: "ผู้ช่วยของคุณเป็นคนอ่านรูป เนโม่ได้รับเฉพาะข้อมูลที่คุณยืนยัน",
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
    sub: "ความจำแต่ละประเภทเป็นโมดูลที่ตั้งชื่อตามเทพเจ้ากรีก ค่าใช้จ่ายและงานใช้งานได้แล้ว และจะมีอีกหลายอย่างตามมา",
    available: "ใช้งานได้",
    soon: "เร็ว ๆ นี้",
    items: [
      { god: "Plutus", area: "ค่าใช้จ่าย", body: "สลิป ยอดรวม และรายงาน", live: true },
      { god: "Atlas", area: "งานและโน้ต", body: "งาน สิ่งที่ต้องทำ และโน้ตที่ผู้ช่วยจดไว้ให้คุณ", live: true },
      { god: "Horae", area: "กิจวัตร", body: "กิจวัตรรายวันและรายสัปดาห์", live: false },
      { god: "Hephaestus", area: "ข้อมูลงาน", body: "โน้ตและข้อมูลจากงานของคุณ", live: false },
    ],
  },
  mission: {
    eyebrow: "พันธกิจ",
    title: "เก่งขึ้นกับข้อมูลทุกประเภท",
    sub: "ค่าใช้จ่ายและงานคือจุดเริ่มต้น พันธกิจของเนโม่คือทำให้เป็นความจำที่ดีสำหรับข้อมูลประเภทอื่นที่คุณอยากให้ผู้ช่วยจำ",
    items: [
      { title: "กิจวัตร (Horae)", body: "กิจวัตรและตารางรายวันรายสัปดาห์" },
      { title: "ข้อมูลงาน (Hephaestus)", body: "โน้ตและข้อมูลจากงานของคุณ" },
      { title: "เก่งขึ้นกับสิ่งที่เก็บ", body: "จัดระเบียบสิ่งที่เก็บไว้ให้ฉลาดขึ้น หมวดหมู่และกฎผู้รับเงินช่วยจัดกลุ่มค่าใช้จ่ายได้แล้ว และมีแผนทำเพิ่มอีก" },
    ],
    note: "เหล่านี้คือเป้าหมาย ไม่ใช่วันวางจำหน่าย เราจะเรียกโมดูลว่าใช้งานได้เมื่อมันใช้งานได้จริงเท่านั้น",
  },
  privacy: {
    eyebrow: "ความเป็นส่วนตัว",
    title: "ตรงไปตรงมาเรื่องข้อมูลของคุณ",
    sub: "นี่คือสิ่งที่เนโม่เก็บไว้ และสิ่งที่ไม่เคยเห็น",
    storesTitle: "สิ่งที่เนโม่เก็บ",
    stores: [
      "ชื่อ อีเมล และรูปโปรไฟล์ Google เพื่อใช้ล็อกอิน",
      "ความจำที่คุณยืนยัน เช่น รายการค่าใช้จ่าย",
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
        a: "เนโม่คือชื่อเรียกง่าย ๆ ของ Nemosyne ความจำแบบโอเพนซอร์สสำหรับผู้ช่วย AI ผู้ช่วยของคุณอ่านและเขียนผ่าน MCP โมดูลแรกคือค่าใช้จ่าย",
      },
      {
        q: "ทำไมถึงชื่อ Nemosyne",
        a: "มาจาก Mnemosyne เทพีแห่งความจำของกรีก เราย่อเป็น Nemo (เนโม่) ให้พูดง่ายจำง่าย",
      },
      {
        q: "ตอนนี้จำอะไรได้บ้าง",
        a: "รายการค่าใช้จ่ายที่คุณยืนยัน รวมถึงงาน สิ่งที่ต้องทำ และโน้ตที่ผู้ช่วยจดไว้ให้ ส่วนกิจวัตรและข้อมูลงานเป็นโมดูลที่วางแผนไว้ ยังใช้งานไม่ได้",
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
    sub: "ล็อกอิน วาง URL เดียวในผู้ช่วยของคุณ แล้วให้เริ่มจำได้เลย",
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
