import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Connect your assistant",
  description: "Add Kuronuma to Claude or any MCP-compatible assistant with one URL.",
};

export default function ConnectLayout({ children }: { children: React.ReactNode }) {
  return children;
}
