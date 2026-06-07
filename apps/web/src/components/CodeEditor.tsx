"use client";

import dynamic from "next/dynamic";
import { useEffect, useState } from "react";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
  ssr: false,
  loading: () => <div className="empty">エディタを読み込んでいます...</div>,
});

export default function CodeEditor({
  value,
  onChange,
  noAutocomplete = false,
}: {
  value: string;
  onChange: (value: string) => void;
  noAutocomplete?: boolean;
}) {
  const [editorTheme, setEditorTheme] = useState("vs");

  useEffect(() => {
    function syncEditorTheme() {
      const theme = document.documentElement.dataset.theme;
      setEditorTheme(theme === "dark" || theme === "netflix" ? "vs-dark" : "vs");
    }
    syncEditorTheme();
    window.addEventListener("algosensei-theme-change", syncEditorTheme);
    return () => window.removeEventListener("algosensei-theme-change", syncEditorTheme);
  }, []);

  return (
    <div className="editorFrame">
      <MonacoEditor
        height="100%"
        language="python"
        theme={editorTheme}
        value={value}
        onChange={(next) => onChange(next ?? "")}
        options={{
          minimap: { enabled: false },
          fontSize: 14,
          lineHeight: 22,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 4,
          wordWrap: "on",
          padding: { top: 14, bottom: 14 },
          quickSuggestions: noAutocomplete ? false : true,
          suggestOnTriggerCharacters: !noAutocomplete,
          acceptSuggestionOnCommitCharacter: !noAutocomplete,
          parameterHints: { enabled: !noAutocomplete },
          wordBasedSuggestions: noAutocomplete ? "off" : "matchingDocuments",
        }}
      />
    </div>
  );
}
