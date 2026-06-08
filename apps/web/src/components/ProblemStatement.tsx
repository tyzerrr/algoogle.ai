import { ExternalLink } from "lucide-react";
import type { OfficialProblemContent, Problem } from "@/lib/types";

export default function ProblemStatement({
  problem,
  officialContent,
  officialLoading,
  officialError,
  language,
  onLanguageChange,
}: {
  problem: Problem;
  officialContent?: OfficialProblemContent | null;
  officialLoading?: boolean;
  officialError?: string;
  language: "ja" | "en";
  onLanguageChange: (language: "ja" | "en") => void;
}) {
  const canShowEnglish = Boolean(officialContent?.statement);
  const useEnglish = language === "en" && canShowEnglish;
  const statement = useEnglish ? officialContent?.statement ?? problem.statement : problem.statement;
  const examples = useEnglish && officialContent?.examples.length ? officialContent.examples : problem.examples;
  const sourceLabel = useEnglish ? `${officialContent?.source ?? "LeetCode"}本文` : "日本語";
  const images = officialContent?.images ?? [];

  return (
    <div className="statement">
      <div className="statementTop">
        <div className="resultMeta">
          {problem.order_index ? <span className="difficulty">#{problem.order_index}</span> : null}
          <span className="difficulty">{problem.difficulty}</span>
          <span className="tag">{problem.pattern}</span>
          {problem.list_name ? <span className="tag">{problem.list_name}</span> : null}
          <span className="tag">{sourceLabel}</span>
        </div>
        <div className="segmented segmentedTwo statementLanguage" aria-label="問題文の言語">
          <button
            className={`segmentButton ${language === "ja" ? "segmentButtonActive" : ""}`}
            type="button"
            onClick={() => onLanguageChange("ja")}
          >
            日本語
          </button>
          <button
            className={`segmentButton ${language === "en" ? "segmentButtonActive" : ""}`}
            type="button"
            onClick={() => onLanguageChange("en")}
          >
            English
          </button>
        </div>
      </div>
      {officialLoading ? <p className="muted">本家問題文を取得中...</p> : null}
      {officialError && language === "en" ? (
        <p className="muted">本家問題文を取得できないため、日本語表示に戻しています。</p>
      ) : null}
      <p className="statementText">{statement}</p>
      {images.length ? (
        <div className="officialFigures">
          {images.map((image) => (
            <figure className="officialFigure" key={image.url}>
              <img
                src={image.url}
                alt={image.alt || `${problem.title} diagram`}
                loading="lazy"
                referrerPolicy="no-referrer"
              />
              {image.alt ? <figcaption>{image.alt}</figcaption> : null}
            </figure>
          ))}
        </div>
      ) : null}
      {problem.source_url ? (
        <p className="sourceLinkRow">
          <a
            className="sourceLink"
            href={problem.source_url}
            target="_blank"
            rel="noreferrer"
            aria-label="LeetCodeの公式問題を新しいタブで開く"
          >
            公式問題を開く
            <ExternalLink size={16} />
          </a>
        </p>
      ) : null}

      <h3>例</h3>
      {examples.map((example, index) => (
        <div className="example" key={`${example.input}-${index}`}>
          <p>
            <strong>Input:</strong> <code>{example.input}</code>
          </p>
          <p>
            <strong>Output:</strong> <code>{example.output}</code>
          </p>
          {example.explanation ? <p className="muted">{example.explanation}</p> : null}
        </div>
      ))}
    </div>
  );
}
