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
  tabs: { overview: string; spending: string; tasks: string };
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
  tasks: {
    title: string;
    filters: { kind: string; status: string; category: string; tag: string; search: string; allKinds: string; allStatuses: string; allCategories: string; clear: string };
    kinds: { task: string; todo: string; note: string };
    statuses: { open: string; done: string; archived: string };
    priorities: { low: string; normal: string; high: string };
    overdue: string;
    due: (date: string) => string;
    complete: string;
    reopen: string;
    archive: string;
    completeLabel: (title: string) => string;
    reopenLabel: (title: string) => string;
    archiveLabel: (title: string) => string;
    emptyNone: string;
    emptyFiltered: string;
    fullNote: (n: number) => string;
    gone: string;
    updateFailed: string;
    form: {
      addButton: string;
      addTitle: string;
      editTitle: string;
      kind: string;
      title: string;
      body: string;
      bodyHint: string;
      due: string;
      priority: string;
      category: string;
      categoryHint: string;
      tags: string;
      tagsHint: string;
      save: string;
      saving: string;
    };
    saved: string;
    savedNewCategory: (name: string) => string;
    updated: string;
    markedDone: string;
    reopened: string;
    archived: string;
    deleted: string;
    nothingToChange: string;
    edit: string;
    editLabel: (title: string) => string;
    deleteLabel: (title: string) => string;
    deleteTitle: string;
    deleteBody: (title: string) => string;
    deleteFailed: string;
  };
  spending: {
    title: string;
    totals: string;
    totalsNote: string;
    period: (from: string, to: string) => string;
    groupBy: string;
    groups: { category: string; month: string; merchant: string };
    recordsHead: string;
    totalHead: string;
    noTotals: string;
    records: string;
    from: string;
    to: string;
    category: string;
    allCategories: string;
    clearFilters: string;
    fullNote: (n: number) => string;
    emptyNone: string;
    emptyFiltered: string;
    deleteLabel: (merchant: string, amount: string) => string;
    deleteTitle: string;
    deleteBody: (merchant: string, amount: string) => string;
    deleteFailed: string;
    add: {
      title: string;
      amount: string;
      currency: string;
      merchant: string;
      date: string;
      category: string;
      categoryHint: string;
      note: string;
      submit: string;
      saving: string;
      saved: string;
      savedNewCategory: (name: string) => string;
      alreadySaved: string;
      dupTitle: string;
      dupBody: string;
      saveAnyway: string;
      dontSave: string;
    };
  };
  bills: {
    tabsLabel: string;
    records: string;
    tab: string;
    month: string;
    prevMonth: string;
    nextMonth: string;
    totals: string;
    planned: string;
    paid: string;
    counts: (paid: number, unpaid: number) => string;
    noTotals: string;
    listTitle: string;
    empty: string;
    statusPaid: string;
    statusUnpaid: string;
    paidOn: (amount: string, date: string) => string;
    endsFrom: (month: string) => string;
    markPaid: string;
    markPaidLabel: (name: string, amount: string) => string;
    marking: string;
    markedPaid: string;
    edit: string;
    editLabel: (name: string) => string;
    end: string;
    endLabel: (name: string) => string;
    deleteLabel: (name: string) => string;
    deleteTitle: string;
    deleteBody: (name: string) => string;
    actionFailed: string;
    form: {
      addTitle: string;
      editTitle: (name: string) => string;
      name: string;
      amount: string;
      currency: string;
      currencyLocked: string;
      category: string;
      categoryHint: string;
      startMonth: string;
      startMonthHint: string;
      add: string;
      save: string;
      saving: string;
      saved: string;
      savedNewCategory: (name: string) => string;
      cancel: string;
      nothingToChange: string;
    };
    endForm: {
      title: (name: string) => string;
      from: string;
      hint: string;
      confirm: string;
      ending: string;
      ended: string;
    };
  };
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
    tabs: { overview: "Overview", spending: "Spending", tasks: "Tasks and notes" },
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
    tasks: {
      title: "Tasks and notes",
      filters: {
        kind: "Type",
        status: "Status",
        category: "Category",
        tag: "Tag",
        search: "Search",
        allKinds: "All types",
        allStatuses: "All statuses",
        allCategories: "All categories",
        clear: "Clear filters",
      },
      kinds: { task: "Task", todo: "To-do", note: "Note" },
      statuses: { open: "Open", done: "Done", archived: "Archived" },
      priorities: { low: "Low priority", normal: "Normal priority", high: "High priority" },
      overdue: "Overdue",
      due: (date) => `Due ${date}`,
      complete: "Mark done",
      reopen: "Reopen",
      archive: "Archive",
      completeLabel: (t) => `Mark done: “${t}”`,
      reopenLabel: (t) => `Reopen “${t}”`,
      archiveLabel: (t) => `Archive “${t}”`,
      emptyNone: "Nothing saved yet.",
      emptyFiltered: "No items match these filters.",
      fullNote: (n) => `Showing the first ${n} items. Use the filters to narrow it down.`,
      gone: "That item was already changed or removed, so the list has been refreshed.",
      updateFailed: "The change could not be saved.",
      form: {
        addButton: "Add item",
        addTitle: "Add an item",
        editTitle: "Edit item",
        kind: "Type",
        title: "Title",
        body: "Details",
        bodyHint: "Line breaks aren’t kept.",
        due: "Due date",
        priority: "Priority",
        category: "Category",
        categoryHint: "Pick one of yours or type a new name.",
        tags: "Tags",
        tagsHint: "Separate with commas. Up to 10.",
        save: "Save",
        saving: "Saving…",
      },
      saved: "Saved.",
      savedNewCategory: (name) => `Saved. “${name}” is a new category.`,
      updated: "Changes saved.",
      markedDone: "Marked as done.",
      reopened: "Reopened.",
      archived: "Archived.",
      deleted: "Deleted.",
      nothingToChange: "Nothing to change.",
      edit: "Edit",
      editLabel: (t) => `Edit “${t}”`,
      deleteLabel: (t) => `Delete “${t}”`,
      deleteTitle: "Delete this item?",
      deleteBody: (t) => `“${t}” will be permanently deleted. You can archive it instead to keep it out of the way.`,
      deleteFailed: "The item could not be deleted.",
    },
    spending: {
      title: "Spending",
      totals: "Totals",
      totalsNote: "Totals cover the dates you pick, or this month if you pick none (the last 12 months when grouped by month). They don’t follow the category filter.",
      period: (from, to) => `${from} – ${to}`,
      groupBy: "Group by",
      groups: { category: "Category", month: "Month", merchant: "Merchant" },
      recordsHead: "Records",
      totalHead: "Total",
      noTotals: "No spending in this period.",
      records: "Records",
      from: "From",
      to: "To",
      category: "Category",
      allCategories: "All categories",
      clearFilters: "Clear filters",
      fullNote: (n) => `Showing the newest ${n} records. Use the dates to narrow it down.`,
      emptyNone: "No spending saved yet.",
      emptyFiltered: "No records match these filters.",
      deleteLabel: (m, a) => `Delete ${m}, ${a}`,
      deleteTitle: "Delete this record?",
      deleteBody: (m, a) => `${a} at ${m} will be permanently deleted. This can’t be undone.`,
      deleteFailed: "The record could not be deleted.",
      add: {
        title: "Add a record",
        amount: "Amount",
        currency: "Currency",
        merchant: "Merchant",
        date: "Date",
        category: "Category",
        categoryHint: "Pick one of yours or type a new name.",
        note: "Note (optional)",
        submit: "Save record",
        saving: "Saving…",
        saved: "Saved.",
        savedNewCategory: (name) => `Saved. “${name}” is a new category.`,
        alreadySaved: "This slip was already saved, so nothing new was added.",
        dupTitle: "This looks like a duplicate",
        dupBody: "A record with the same amount, currency, date and merchant already exists. Is this a separate payment?",
        saveAnyway: "Save anyway",
        dontSave: "Don’t save",
      },
    },
    bills: {
      tabsLabel: "Spending sections",
      records: "Records",
      tab: "Bills",
      month: "Month",
      prevMonth: "Previous month",
      nextMonth: "Next month",
      totals: "Totals for the month",
      planned: "Planned",
      paid: "Paid",
      counts: (p, u) => `${p} paid, ${u} not yet paid`,
      noTotals: "No bills count in this month.",
      listTitle: "Fixed monthly bills",
      empty: "No bills in this month. Add one below.",
      statusPaid: "Paid",
      statusUnpaid: "Unpaid",
      paidOn: (a, d) => `${a} paid on ${d}`,
      endsFrom: (m) => `Ends from ${m}`,
      markPaid: "Mark paid",
      markPaidLabel: (n, a) => `Mark ${n}, ${a}, as paid`,
      marking: "Saving…",
      markedPaid: "Marked as paid.",
      edit: "Edit",
      editLabel: (n) => `Edit ${n}`,
      end: "End",
      endLabel: (n) => `End ${n}`,
      deleteLabel: (n) => `Delete ${n}`,
      deleteTitle: "Delete this bill?",
      deleteBody: (n) => `${n} will be permanently deleted. This can’t be undone.`,
      actionFailed: "That didn’t work",
      form: {
        addTitle: "Add a bill",
        editTitle: (n) => `Edit ${n}`,
        name: "Name",
        amount: "Monthly amount",
        currency: "Currency",
        currencyLocked: "The currency can’t be changed.",
        category: "Category",
        categoryHint: "Pick one of yours or type a new name.",
        startMonth: "First month",
        startMonthHint: "Leave empty to start this month.",
        add: "Add bill",
        save: "Save changes",
        saving: "Saving…",
        saved: "Saved.",
        savedNewCategory: (n) => `Saved. “${n}” is a new category.`,
        cancel: "Cancel",
        nothingToChange: "Nothing to change.",
      },
      endForm: {
        title: (n) => `End ${n}`,
        from: "First month it no longer counts",
        hint: "Earlier months still show the bill.",
        confirm: "End bill",
        ending: "Ending…",
        ended: "The bill has ended.",
      },
    },
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
    tabs: { overview: "ภาพรวม", spending: "ค่าใช้จ่าย", tasks: "งานและโน้ต" },
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
    tasks: {
      title: "งานและโน้ต",
      filters: {
        kind: "ประเภท",
        status: "สถานะ",
        category: "หมวดหมู่",
        tag: "แท็ก",
        search: "ค้นหา",
        allKinds: "ทุกประเภท",
        allStatuses: "ทุกสถานะ",
        allCategories: "ทุกหมวดหมู่",
        clear: "ล้างตัวกรอง",
      },
      kinds: { task: "งาน", todo: "สิ่งที่ต้องทำ", note: "โน้ต" },
      statuses: { open: "ยังไม่เสร็จ", done: "เสร็จแล้ว", archived: "เก็บถาวร" },
      priorities: { low: "ความสำคัญต่ำ", normal: "ความสำคัญปกติ", high: "ความสำคัญสูง" },
      overdue: "เลยกำหนด",
      due: (date) => `ครบกำหนด ${date}`,
      complete: "ทำเครื่องหมายว่าเสร็จ",
      reopen: "เปิดใหม่",
      archive: "เก็บถาวร",
      completeLabel: (t) => `ทำเครื่องหมายว่าเสร็จ: “${t}”`,
      reopenLabel: (t) => `เปิดใหม่ “${t}”`,
      archiveLabel: (t) => `เก็บถาวร “${t}”`,
      emptyNone: "ยังไม่มีอะไรที่บันทึกไว้",
      emptyFiltered: "ไม่พบรายการที่ตรงกับตัวกรอง",
      fullNote: (n) => `แสดง ${n} รายการแรก ใช้ตัวกรองเพื่อจำกัดให้แคบลง`,
      gone: "รายการนี้ถูกเปลี่ยนหรือถูกลบไปแล้ว จึงรีเฟรชรายการให้แล้ว",
      updateFailed: "บันทึกการเปลี่ยนแปลงไม่สำเร็จ",
      form: {
        addButton: "เพิ่มรายการ",
        addTitle: "เพิ่มงานหรือโน้ต",
        editTitle: "แก้ไขรายการ",
        kind: "ประเภท",
        title: "ชื่อเรื่อง",
        body: "รายละเอียด",
        bodyHint: "ระบบไม่เก็บการขึ้นบรรทัดใหม่",
        due: "วันครบกำหนด",
        priority: "ความสำคัญ",
        category: "หมวดหมู่",
        categoryHint: "เลือกจากหมวดหมู่ของคุณ หรือพิมพ์ชื่อใหม่",
        tags: "แท็ก",
        tagsHint: "คั่นด้วยเครื่องหมายจุลภาค ได้สูงสุด 10 แท็ก",
        save: "บันทึก",
        saving: "กำลังบันทึก…",
      },
      saved: "บันทึกแล้ว",
      savedNewCategory: (name) => `บันทึกแล้ว “${name}” เป็นหมวดหมู่ใหม่`,
      updated: "บันทึกการเปลี่ยนแปลงแล้ว",
      markedDone: "ทำเครื่องหมายว่าเสร็จแล้ว",
      reopened: "เปิดใหม่แล้ว",
      archived: "เก็บถาวรแล้ว",
      deleted: "ลบแล้ว",
      nothingToChange: "ไม่มีอะไรที่เปลี่ยน",
      edit: "แก้ไข",
      editLabel: (t) => `แก้ไข “${t}”`,
      deleteLabel: (t) => `ลบ “${t}”`,
      deleteTitle: "ลบรายการนี้ไหม",
      deleteBody: (t) => `“${t}” จะถูกลบถาวร หากอยากเก็บไว้แต่ไม่ให้เกะกะ ให้เก็บถาวรแทน`,
      deleteFailed: "ลบรายการนี้ไม่สำเร็จ",
    },
    spending: {
      title: "ค่าใช้จ่าย",
      totals: "ยอดรวม",
      totalsNote: "ยอดรวมครอบคลุมช่วงวันที่ที่คุณเลือก ถ้าไม่เลือกจะเป็นเดือนนี้ (หรือ 12 เดือนที่ผ่านมาเมื่อจัดกลุ่มตามเดือน) และไม่ขึ้นกับตัวกรองหมวดหมู่",
      period: (from, to) => `${from} – ${to}`,
      groupBy: "จัดกลุ่มตาม",
      groups: { category: "หมวดหมู่", month: "เดือน", merchant: "ร้านค้า" },
      recordsHead: "รายการ",
      totalHead: "รวม",
      noTotals: "ไม่มีค่าใช้จ่ายในช่วงนี้",
      records: "รายการ",
      from: "ตั้งแต่",
      to: "ถึง",
      category: "หมวดหมู่",
      allCategories: "ทุกหมวดหมู่",
      clearFilters: "ล้างตัวกรอง",
      fullNote: (n) => `แสดง ${n} รายการล่าสุด ใช้ช่วงวันที่เพื่อจำกัดให้แคบลง`,
      emptyNone: "ยังไม่มีค่าใช้จ่ายที่บันทึกไว้",
      emptyFiltered: "ไม่พบรายการที่ตรงกับตัวกรอง",
      deleteLabel: (m, a) => `ลบ ${m}, ${a}`,
      deleteTitle: "ลบรายการนี้ไหม",
      deleteBody: (m, a) => `${a} ที่ ${m} จะถูกลบถาวร และเรียกคืนไม่ได้`,
      deleteFailed: "ลบรายการนี้ไม่สำเร็จ",
      add: {
        title: "เพิ่มรายการ",
        amount: "จำนวนเงิน",
        currency: "สกุลเงิน",
        merchant: "ร้านค้า",
        date: "วันที่",
        category: "หมวดหมู่",
        categoryHint: "เลือกจากหมวดหมู่ของคุณ หรือพิมพ์ชื่อใหม่",
        note: "บันทึกเพิ่มเติม (ไม่บังคับ)",
        submit: "บันทึกรายการ",
        saving: "กำลังบันทึก…",
        saved: "บันทึกแล้ว",
        savedNewCategory: (name) => `บันทึกแล้ว “${name}” เป็นหมวดหมู่ใหม่`,
        alreadySaved: "สลิปนี้บันทึกไว้แล้ว จึงไม่ได้เพิ่มรายการใหม่",
        dupTitle: "ดูเหมือนรายการซ้ำ",
        dupBody: "มีรายการที่จำนวนเงิน สกุลเงิน วันที่ และร้านค้าเหมือนกันอยู่แล้ว นี่เป็นการจ่ายเงินคนละครั้งใช่ไหม",
        saveAnyway: "บันทึกอยู่ดี",
        dontSave: "ไม่บันทึก",
      },
    },
    bills: {
      tabsLabel: "ส่วนของค่าใช้จ่าย",
      records: "รายการ",
      tab: "บิล",
      month: "เดือน",
      prevMonth: "เดือนก่อนหน้า",
      nextMonth: "เดือนถัดไป",
      totals: "ยอดรวมของเดือน",
      planned: "ที่ตั้งไว้",
      paid: "จ่ายแล้ว",
      counts: (p, u) => `จ่ายแล้ว ${p} ยังไม่จ่าย ${u}`,
      noTotals: "เดือนนี้ไม่มีบิลที่นับ",
      listTitle: "บิลประจำเดือน",
      empty: "เดือนนี้ไม่มีบิล เพิ่มได้ด้านล่าง",
      statusPaid: "จ่ายแล้ว",
      statusUnpaid: "ยังไม่จ่าย",
      paidOn: (a, d) => `จ่าย ${a} เมื่อ ${d}`,
      endsFrom: (m) => `สิ้นสุดตั้งแต่ ${m}`,
      markPaid: "ทำเครื่องหมายว่าจ่ายแล้ว",
      markPaidLabel: (n, a) => `ทำเครื่องหมายว่าจ่าย ${n} ${a} แล้ว`,
      marking: "กำลังบันทึก…",
      markedPaid: "บันทึกว่าจ่ายแล้ว",
      edit: "แก้ไข",
      editLabel: (n) => `แก้ไข ${n}`,
      end: "สิ้นสุด",
      endLabel: (n) => `สิ้นสุด ${n}`,
      deleteLabel: (n) => `ลบ ${n}`,
      deleteTitle: "ลบบิลนี้ไหม",
      deleteBody: (n) => `${n} จะถูกลบถาวร และเรียกคืนไม่ได้`,
      actionFailed: "ดำเนินการไม่สำเร็จ",
      form: {
        addTitle: "เพิ่มบิล",
        editTitle: (n) => `แก้ไข ${n}`,
        name: "ชื่อ",
        amount: "จำนวนเงินต่อเดือน",
        currency: "สกุลเงิน",
        currencyLocked: "เปลี่ยนสกุลเงินไม่ได้",
        category: "หมวดหมู่",
        categoryHint: "เลือกจากหมวดหมู่ของคุณ หรือพิมพ์ชื่อใหม่",
        startMonth: "เดือนแรก",
        startMonthHint: "เว้นว่างไว้เพื่อเริ่มเดือนนี้",
        add: "เพิ่มบิล",
        save: "บันทึกการเปลี่ยนแปลง",
        saving: "กำลังบันทึก…",
        saved: "บันทึกแล้ว",
        savedNewCategory: (n) => `บันทึกแล้ว “${n}” เป็นหมวดหมู่ใหม่`,
        cancel: "ยกเลิก",
        nothingToChange: "ไม่มีอะไรที่เปลี่ยน",
      },
      endForm: {
        title: (n) => `สิ้นสุด ${n}`,
        from: "เดือนแรกที่ไม่นับอีกต่อไป",
        hint: "เดือนก่อนหน้ายังแสดงบิลนี้อยู่",
        confirm: "สิ้นสุดบิล",
        ending: "กำลังดำเนินการ…",
        ended: "สิ้นสุดบิลแล้ว",
      },
    },
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
