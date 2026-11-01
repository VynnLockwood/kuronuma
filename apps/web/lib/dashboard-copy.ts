// All text of the dashboard, in English and Thai. The Thai is written for Thai readers, not word for word; the
// pull request lists every Thai string for a native speaker to check.

import type { Lang } from "./web";

export type DashboardCopy = {
  htmlLang: string;
  shell: {
    signIn: { title: string; sub: string; errorTitle: string; button: string; fine: string; back: string };
    unreachable: { title: string; body: string; retry: string };
    loading: string;
    welcome: string;
    signOut: string;
    tabsLabel: string;
    langLabel: string;
    nav: { dashboard: string; connect: string };
    authErrors: Record<string, string>;
  };
  tabs: { overview: string; spending: string };
  overview: {
    thisMonth: string;
    byCategory: string;
    lastMonths: string;
    tasks: string;
    open: string;
    overdue: string;
    dueWeek: string;
    notes: string;
    categories: string;
    categoriesCount: (n: number) => string;
    records: (n: number) => string;
    noSpendingMonth: string;
    noTasks: string;
    noNotes: string;
    categoryHead: string;
    monthHead: string;
    amountHead: string;
    emptyTitle: string;
    emptyBody: string;
    tryAsking: string;
    prompts: string[];
    connect: { title: string; body: string; button: string; guide: string };
  };
  spending: { title: string };
  common: {
    retry: string;
    cancel: string;
    close: string;
    delete: string;
    loading: string;
    errorTitle: string;
    connectLink: string;
    working: string;
  };
};

export const dashboardCopy: Record<Lang, DashboardCopy> = {
  en: {
    htmlLang: "en",
    shell: {
      signIn: {
        title: "Sign in to Kuronuma",
        sub: "Use your Google account to connect your assistant to your own data.",
        errorTitle: "Sign-in didn’t finish",
        button: "Continue with Google",
        fine: "Kuronuma stores your Google name, email and photo to sign you in. Never your password.",
        back: "‹ Back to home",
      },
      unreachable: {
        title: "Can’t reach Kuronuma",
        body: "Kuronuma isn’t responding right now. Please try again in a moment.",
        retry: "Try again",
      },
      loading: "Loading",
      welcome: "Welcome back",
      signOut: "Sign out",
      tabsLabel: "Dashboard sections",
      langLabel: "Language",
      nav: { dashboard: "Dashboard", connect: "Connect" },
      authErrors: {
        access_denied: "Sign-in was cancelled.",
        invalid_flow: "That sign-in attempt expired or was invalid. Please try again.",
        exchange_failed: "Google could not confirm your sign-in. Please try again.",
        email_unverified: "Your Google account email is not verified.",
        not_allowed: "This account is not on the access list yet.",
        server_error: "Something went wrong on our side. Please try again.",
      },
    },
    tabs: { overview: "Overview", spending: "Spending" },
    overview: {
      thisMonth: "This month",
      byCategory: "Spending by category",
      lastMonths: "Last 6 months",
      tasks: "Tasks and to-dos",
      open: "Open",
      overdue: "Overdue",
      dueWeek: "Due this week",
      notes: "Newest notes",
      categories: "Categories",
      categoriesCount: (n) => (n === 1 ? "1 category" : `${n} categories`),
      records: (n) => (n === 1 ? "1 record" : `${n} records`),
      noSpendingMonth: "No spending saved this month yet.",
      noTasks: "No open tasks or to-dos.",
      noNotes: "No notes yet.",
      categoryHead: "Category",
      monthHead: "Month",
      amountHead: "Amount",
      emptyTitle: "Nothing here yet",
      emptyBody: "Connect an assistant, then ask it to save a slip, a task or a note. It will show up here.",
      tryAsking: "Try asking",
      prompts: ["“Save this slip to Kuronuma.”", "“How much did I spend on food this month?”", "“Add a to-do for Friday: send the report.”"],
      connect: {
        title: "Connect your assistant",
        body: "Add Kuronuma to Claude or any assistant that supports MCP connectors. It takes one URL.",
        button: "Connect",
        guide: "Read the setup guide ›",
      },
    },
    spending: { title: "Spending" },
    common: {
      retry: "Try again",
      cancel: "Cancel",
      close: "Close",
      delete: "Delete",
      loading: "Loading",
      errorTitle: "Couldn’t load this",
      connectLink: "Connect an assistant",
      working: "Working…",
    },
  },
  th: {
    htmlLang: "th",
    shell: {
      signIn: {
        title: "เข้าสู่ระบบคุโรนุมะ",
        sub: "ใช้บัญชี Google ของคุณเพื่อเชื่อมต่อผู้ช่วยกับข้อมูลของคุณเอง",
        errorTitle: "เข้าสู่ระบบไม่สำเร็จ",
        button: "ดำเนินการต่อด้วย Google",
        fine: "คุโรนุมะเก็บชื่อ อีเมล และรูปโปรไฟล์ Google ของคุณเพื่อใช้เข้าสู่ระบบเท่านั้น ไม่เก็บรหัสผ่านของคุณ",
        back: "‹ กลับหน้าแรก",
      },
      unreachable: {
        title: "ติดต่อคุโรนุมะไม่ได้",
        body: "ตอนนี้คุโรนุมะยังไม่ตอบสนอง โปรดลองอีกครั้งในอีกสักครู่",
        retry: "ลองอีกครั้ง",
      },
      loading: "กำลังโหลด",
      welcome: "ยินดีต้อนรับกลับ",
      signOut: "ออกจากระบบ",
      tabsLabel: "ส่วนต่าง ๆ ของแดชบอร์ด",
      langLabel: "ภาษา",
      nav: { dashboard: "แดชบอร์ด", connect: "เชื่อมต่อ" },
      authErrors: {
        access_denied: "ยกเลิกการเข้าสู่ระบบแล้ว",
        invalid_flow: "การเข้าสู่ระบบครั้งนี้หมดอายุหรือไม่ถูกต้อง โปรดลองอีกครั้ง",
        exchange_failed: "Google ยืนยันการเข้าสู่ระบบของคุณไม่ได้ โปรดลองอีกครั้ง",
        email_unverified: "อีเมลของบัญชี Google ของคุณยังไม่ได้รับการยืนยัน",
        not_allowed: "บัญชีนี้ยังไม่อยู่ในรายชื่อที่ได้รับอนุญาต",
        server_error: "เกิดข้อผิดพลาดที่ฝั่งเรา โปรดลองอีกครั้ง",
      },
    },
    tabs: { overview: "ภาพรวม", spending: "ค่าใช้จ่าย" },
    overview: {
      thisMonth: "เดือนนี้",
      byCategory: "ค่าใช้จ่ายแยกตามหมวดหมู่",
      lastMonths: "6 เดือนที่ผ่านมา",
      tasks: "งานและสิ่งที่ต้องทำ",
      open: "ยังไม่เสร็จ",
      overdue: "เลยกำหนด",
      dueWeek: "ครบกำหนดใน 7 วัน",
      notes: "โน้ตล่าสุด",
      categories: "หมวดหมู่",
      categoriesCount: (n) => `${n} หมวดหมู่`,
      records: (n) => `${n} รายการ`,
      noSpendingMonth: "เดือนนี้ยังไม่มีค่าใช้จ่ายที่บันทึกไว้",
      noTasks: "ไม่มีงานหรือสิ่งที่ต้องทำที่ค้างอยู่",
      noNotes: "ยังไม่มีโน้ต",
      categoryHead: "หมวดหมู่",
      monthHead: "เดือน",
      amountHead: "จำนวนเงิน",
      emptyTitle: "ยังไม่มีข้อมูล",
      emptyBody: "เชื่อมต่อผู้ช่วย แล้วบอกให้บันทึกสลิป งาน หรือโน้ต ข้อมูลจะมาปรากฏที่นี่",
      tryAsking: "ลองสั่งผู้ช่วยดู",
      prompts: ["“บันทึกสลิปนี้ลงคุโรนุมะ”", "“เดือนนี้ฉันใช้จ่ายค่าอาหารไปเท่าไร”", "“เพิ่มสิ่งที่ต้องทำวันศุกร์: ส่งรายงาน”"],
      connect: {
        title: "เชื่อมต่อผู้ช่วยของคุณ",
        body: "เพิ่มคุโรนุมะใน Claude หรือผู้ช่วยใดก็ได้ที่รองรับ MCP ใช้ URL เดียวเท่านั้น",
        button: "เชื่อมต่อ",
        guide: "อ่านคู่มือการตั้งค่า ›",
      },
    },
    spending: { title: "ค่าใช้จ่าย" },
    common: {
      retry: "ลองอีกครั้ง",
      cancel: "ยกเลิก",
      close: "ปิด",
      delete: "ลบ",
      loading: "กำลังโหลด",
      errorTitle: "โหลดข้อมูลนี้ไม่สำเร็จ",
      connectLink: "เชื่อมต่อผู้ช่วย",
      working: "กำลังดำเนินการ…",
    },
  },
};
