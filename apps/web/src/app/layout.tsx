import type { Metadata } from "next";
import Link from "next/link";
import { BrainCircuit, ListChecks, RotateCcw } from "lucide-react";
import ThemeSwitcher from "@/components/ThemeSwitcher";
import "./globals.css";

export const metadata: Metadata = {
  title: "algoogle",
  description: "AI coding interview practice for Google-style SWE interviews",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ja" suppressHydrationWarning>
      <body>
        <header className="topbar">
          <Link href="/" className="brand" aria-label="algoogle home">
            <BrainCircuit size={22} />
            <span>algoogle</span>
          </Link>
          <nav className="nav">
            <Link href="/problems">
              <ListChecks size={17} />
              問題
            </Link>
            <Link href="/review">
              <RotateCcw size={17} />
              復習
            </Link>
            <ThemeSwitcher />
          </nav>
        </header>
        {children}
      </body>
    </html>
  );
}
