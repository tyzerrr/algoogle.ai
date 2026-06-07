package server

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CodeFileManager struct {
	workspaceDir string
	publicDir    string
}

func NewCodeFileManager(workspaceDir, publicDir string) *CodeFileManager {
	return &CodeFileManager{
		workspaceDir: workspaceDir,
		publicDir:    publicDir,
	}
}

func (m *CodeFileManager) Ensure(attempt Attempt) (CodeFileResponse, error) {
	if err := os.MkdirAll(m.workspaceDir, 0755); err != nil {
		return CodeFileResponse{}, err
	}
	path := m.internalPath(attempt)
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return CodeFileResponse{}, err
		}
		if err := os.WriteFile(path, []byte(attempt.Code), 0644); err != nil {
			return CodeFileResponse{}, err
		}
	}
	return m.Read(attempt)
}

func (m *CodeFileManager) Read(attempt Attempt) (CodeFileResponse, error) {
	path := m.internalPath(attempt)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m.Ensure(attempt)
		}
		return CodeFileResponse{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return CodeFileResponse{}, err
	}
	return CodeFileResponse{
		Path:      m.publicPath(attempt),
		Content:   string(content),
		UpdatedAt: info.ModTime().UTC().Format(time.RFC3339Nano),
		Size:      info.Size(),
	}, nil
}

func (m *CodeFileManager) Write(attempt Attempt, code string) (CodeFileResponse, error) {
	if err := os.MkdirAll(m.workspaceDir, 0755); err != nil {
		return CodeFileResponse{}, err
	}
	if err := os.WriteFile(m.internalPath(attempt), []byte(code), 0644); err != nil {
		return CodeFileResponse{}, err
	}
	return m.Read(attempt)
}

func (m *CodeFileManager) Decorate(attempt *Attempt) {
	info, err := m.Ensure(*attempt)
	if err != nil {
		return
	}
	attempt.CodeFilePath = info.Path
	attempt.CodeFileUpdatedAt = info.UpdatedAt
}

func (m *CodeFileManager) internalPath(attempt Attempt) string {
	return filepath.Join(m.workspaceDir, m.fileName(attempt))
}

func (m *CodeFileManager) publicPath(attempt Attempt) string {
	return filepath.ToSlash(filepath.Join(m.publicDir, m.fileName(attempt)))
}

func (m *CodeFileManager) fileName(attempt Attempt) string {
	problemID := safeFilePart(attempt.ProblemID)
	attemptID := safeFilePart(attempt.ID)
	if problemID == "" {
		problemID = "problem"
	}
	if attemptID == "" {
		attemptID = "attempt"
	}
	return problemID + "-" + attemptID + ".py"
}

var unsafeFilePart = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func safeFilePart(value string) string {
	value = strings.TrimSpace(value)
	value = unsafeFilePart.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-_")
	return value
}
