"use client";

import { useEffect, useState } from "react";

// Renders a Mermaid source to SVG, re-rendering on theme changes.
// Extracted verbatim from the former WhiteboardPanel: dynamic import, strict
// security, theme derived from the documentElement theme, cancellation-safe.
export function useMermaid(source: string, enabled: boolean): { svg: string; error: string } {
  const [svg, setSvg] = useState("");
  const [error, setError] = useState("");
  const [themeTick, setThemeTick] = useState(0);

  useEffect(() => {
    function handleThemeChange() {
      setThemeTick((current) => current + 1);
    }
    window.addEventListener("algosensei-theme-change", handleThemeChange);
    return () => window.removeEventListener("algosensei-theme-change", handleThemeChange);
  }, []);

  useEffect(() => {
    if (!enabled) {
      setSvg("");
      setError("");
      return undefined;
    }
    let cancelled = false;
    async function renderMermaid() {
      try {
        const mermaidModule = await import("mermaid");
        const mermaid = mermaidModule.default;
        const theme = document.documentElement.dataset.theme === "light" ? "default" : "dark";
        mermaid.initialize({ startOnLoad: false, securityLevel: "strict", theme });
        const id = `mermaid-${Date.now()}-${Math.random().toString(36).slice(2)}`;
        const { svg: rendered } = await mermaid.render(id, source);
        if (!cancelled) {
          setSvg(rendered);
          setError("");
        }
      } catch (err) {
        if (!cancelled) {
          setSvg("");
          setError((err as Error).message);
        }
      }
    }
    // Debounce: mermaid.render is expensive and the composer feeds keystrokes here.
    const timer = window.setTimeout(() => {
      void renderMermaid();
    }, 250);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [source, enabled, themeTick]);

  return { svg, error };
}
