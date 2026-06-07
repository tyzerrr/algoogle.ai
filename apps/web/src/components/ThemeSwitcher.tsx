"use client";

import { Clapperboard, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import type { ThemeMode } from "@/lib/types";

const THEME_OPTIONS: { id: ThemeMode; label: string; icon: typeof Sun }[] = [
  { id: "light", label: "Light", icon: Sun },
  { id: "dark", label: "Dark", icon: Moon },
  { id: "netflix", label: "Netflix", icon: Clapperboard },
];

const STORAGE_KEY = "algosensei-theme";

function applyTheme(theme: ThemeMode) {
  document.documentElement.dataset.theme = theme;
  window.localStorage.setItem(STORAGE_KEY, theme);
  window.dispatchEvent(new CustomEvent("algosensei-theme-change", { detail: theme }));
}

function readTheme(): ThemeMode {
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === "dark" || stored === "netflix") return stored;
  return "light";
}

export default function ThemeSwitcher() {
  const [theme, setTheme] = useState<ThemeMode>("light");

  useEffect(() => {
    const stored = readTheme();
    setTheme(stored);
    applyTheme(stored);
  }, []);

  function selectTheme(nextTheme: ThemeMode) {
    setTheme(nextTheme);
    applyTheme(nextTheme);
  }

  return (
    <div className="themeSwitcher" aria-label="Theme mode">
      {THEME_OPTIONS.map((option) => {
        const Icon = option.icon;
        return (
          <button
            className={`themeButton ${theme === option.id ? "themeButtonActive" : ""}`}
            key={option.id}
            type="button"
            title={`${option.label} mode`}
            aria-label={`${option.label} mode`}
            onClick={() => selectTheme(option.id)}
          >
            <Icon size={16} />
          </button>
        );
      })}
    </div>
  );
}
