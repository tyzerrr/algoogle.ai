"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Bot, Loader2 } from "lucide-react";
import AIChat from "@/components/AIChat";
import ProblemStatement from "@/components/ProblemStatement";
import ReviewPanel from "@/components/ReviewPanel";
import TestResultPanel from "@/components/TestResultPanel";
import ConfirmDialog from "@/components/ui/ConfirmDialog";
import ErrorNotice from "@/components/ui/ErrorNotice";
import { SkeletonBlock } from "@/components/ui/Skeleton";
import CodePane from "@/components/interview/CodePane";
import InterviewHeader from "@/components/interview/InterviewHeader";
import InterviewSettingsDialog, {
  AI_PROVIDER_OPTIONS,
  COMPANY_OPTIONS,
  MODE_OPTIONS,
  type InterviewSettings,
} from "@/components/interview/InterviewSettingsDialog";
import SessionStrip from "@/components/interview/SessionStrip";
import WorkspaceTabs, {
  type TestBadgeTone,
  type WorkspaceTabId,
} from "@/components/interview/WorkspaceTabs";
import { useInterviewSession } from "@/hooks/useInterviewSession";
import { useOfficialContent } from "@/hooks/useOfficialContent";
import { activePhase } from "@/lib/interviewPhase";

type PendingChange =
  | { kind: "new-interview" }
  | { kind: "settings"; settings: InterviewSettings };

function settingsSummary(settings: InterviewSettings) {
  const provider = AI_PROVIDER_OPTIONS.find((o) => o.id === settings.aiProvider)?.label ?? "";
  const company = COMPANY_OPTIONS.find((o) => o.id === settings.companyPreset)?.label ?? "";
  const mode = MODE_OPTIONS.find((o) => o.id === settings.interviewMode)?.label ?? "";
  return `設定: ${provider} · ${company} · ${mode}`;
}

function testBadgeTone(status: string | undefined, passed: boolean | undefined): TestBadgeTone | null {
  if (status === undefined) return null;
  if (status === "disabled" || status === "not_configured") return "gray";
  return passed ? "green" : "red";
}

export default function ProblemDetailPage() {
  const params = useParams<{ id: string }>();
  const problemId = params.id;
  const session = useInterviewSession(problemId);
  const official = useOfficialContent(problemId);

  const [problemLanguage, setProblemLanguage] = useState<"ja" | "en">("ja");
  const [activeTab, setActiveTab] = useState<WorkspaceTabId>("code");
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [pendingChange, setPendingChange] = useState<PendingChange | null>(null);

  const { problem, attempt } = session;

  useEffect(() => {
    setProblemLanguage("ja");
  }, [problem?.id]);

  const currentSettings: InterviewSettings = {
    aiProvider: session.aiProvider,
    companyPreset: session.companyPreset,
    interviewMode: session.interviewMode,
  };

  async function executeChange(change: PendingChange) {
    setSettingsOpen(false);
    if (change.kind === "settings") {
      await session.resetAttempt(change.settings);
    } else {
      await session.resetAttempt();
    }
    setActiveTab("code");
  }

  function requestChange(change: PendingChange) {
    if (!session.hasProgress) {
      void executeChange(change);
      return;
    }
    setPendingChange(change);
  }

  async function handleRun() {
    await session.runCode();
    setActiveTab("test");
  }

  async function handleReview() {
    await session.requestReview();
    setActiveTab("review");
  }

  if (session.loading) {
    return (
      <main className="pageWide">
        <div className="pageHeader">
          <div style={{ width: "100%" }}>
            <p className="eyebrow">Interview room</p>
            <SkeletonBlock lines={2} />
          </div>
        </div>
        <SkeletonBlock lines={3} />
        <div className="interviewShell" style={{ marginTop: 14 }}>
          <section className="pane">
            <div className="paneBody">
              <SkeletonBlock lines={6} />
            </div>
          </section>
          <section className="pane">
            <div className="paneBody">
              <SkeletonBlock lines={6} />
            </div>
          </section>
        </div>
      </main>
    );
  }

  if (session.error && !problem) {
    return (
      <main className="pageWide">
        <ErrorNotice message={`面接画面を読み込めませんでした: ${session.error}`} />
      </main>
    );
  }

  if (!problem) return null;

  const phase = activePhase({
    currentPhase: attempt?.current_phase,
    hasRun: Boolean(session.runResult),
    hasReview: Boolean(session.review),
  });

  const confirmCopy =
    pendingChange?.kind === "settings"
      ? { title: "設定を変更しますか？", confirmLabel: "変更して再開" }
      : { title: "新しい面接を始めますか？", confirmLabel: "新しく始める" };

  return (
    <main className="pageWide">
      <InterviewHeader
        problem={problem}
        remainingSeconds={session.remainingSeconds}
        settingsSummary={settingsSummary(currentSettings)}
        onOpenSettings={() => setSettingsOpen(true)}
        onNewInterview={() => requestChange({ kind: "new-interview" })}
        disabled={!session.canAct}
      />

      {session.error ? (
        <div style={{ marginBottom: 12 }}>
          <ErrorNotice message={session.error} />
        </div>
      ) : null}

      <SessionStrip attempt={attempt} activePhase={phase} syncState={session.syncState} />

      <div className="interviewShell">
        <section className="pane">
          <div className="paneHeader">
            <h2>問題</h2>
          </div>
          <div className="paneBody">
            <ProblemStatement
              problem={problem}
              officialContent={official.content}
              officialLoading={official.loading}
              officialError={official.error}
              language={problemLanguage}
              onLanguageChange={setProblemLanguage}
              onRetryOfficial={official.retry}
            />
          </div>

          <div className="paneHeader">
            <h2>
              <Bot size={18} /> AI面接官
            </h2>
            {session.busy === "chat" ? <Loader2 className="spin" size={17} /> : null}
          </div>
          <div className="paneBody">
            <AIChat
              messages={session.messages}
              loading={session.busy === "chat"}
              loadingLabel={session.chatLoadingLabel}
              disabled={!attempt}
              onSend={session.sendMessage}
              onSubmitArtifact={session.submitArtifact}
              artifactBusy={session.busy === "artifact"}
              onActivity={session.markActivity}
            />
          </div>
        </section>

        <section className="pane editorPane">
          <WorkspaceTabs
            active={activeTab}
            onChange={setActiveTab}
            testBadge={testBadgeTone(session.runResult?.status, session.runResult?.passed)}
            reviewBadge={Boolean(session.review)}
            codePanel={
              <CodePane
                code={session.code}
                onCodeChange={session.onCodeChange}
                codeFile={session.codeFile}
                noAutocomplete={Boolean(attempt?.no_autocomplete)}
                noRun={Boolean(attempt?.no_run)}
                running={session.busy === "run"}
                reviewing={session.busy === "review"}
                canAct={session.canAct}
                onRun={handleRun}
                onReview={handleReview}
              />
            }
            testPanel={
              <div className="paneBody">
                <TestResultPanel result={session.runResult} />
              </div>
            }
            reviewPanel={
              <div className="paneBody">
                <ReviewPanel review={session.review} />
              </div>
            }
          />
        </section>
      </div>

      <InterviewSettingsDialog
        open={settingsOpen}
        aiProvider={session.aiProvider}
        companyPreset={session.companyPreset}
        interviewMode={session.interviewMode}
        disabled={!session.canAct}
        onRequestChange={(settings) => requestChange({ kind: "settings", settings })}
        onClose={() => setSettingsOpen(false)}
      />

      <ConfirmDialog
        open={pendingChange !== null}
        title={confirmCopy.title}
        description="現在のattemptのコードとチャットは新しいattemptに引き継がれません。"
        confirmLabel={confirmCopy.confirmLabel}
        onConfirm={() => {
          if (pendingChange) void executeChange(pendingChange);
          setPendingChange(null);
        }}
        onCancel={() => setPendingChange(null)}
      />
    </main>
  );
}
