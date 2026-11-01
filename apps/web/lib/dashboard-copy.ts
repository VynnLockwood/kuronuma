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
