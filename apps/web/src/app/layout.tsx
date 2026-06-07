import type { Metadata } from "next";
import Link from "next/link";
import { BrainCircuit, ListChecks, RotateCcw } from "lucide-react";
import "./globals.css";

export const metadata: Metadata = {
  title: "AlgoSensei",
  description: "AI coding interview practice for Google-style SWE interviews",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ja">
      <body>
        <header className="topbar">
          <Link href="/" className="brand" aria-label="AlgoSensei home">
            <BrainCircuit size={22} />
            <span>AlgoSensei</span>
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
          </nav>
        </header>
        {children}
      </body>
    </html>
  );
}
