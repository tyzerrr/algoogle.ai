import type { Metadata } from "next";
import Link from "next/link";
import { Inter, Noto_Sans_JP } from "next/font/google";
import { BrainCircuit, ListChecks, RotateCcw } from "lucide-react";
import ThemeSwitcher from "@/components/ThemeSwitcher";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
});

const notoSansJp = Noto_Sans_JP({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-noto-jp",
});

export const metadata: Metadata = {
  title: "algoogle",
  description: "AI coding interview practice for Google-style SWE interviews",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ja" className={`${inter.variable} ${notoSansJp.variable}`} suppressHydrationWarning>
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
