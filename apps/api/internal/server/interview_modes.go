package server

import "strings"

type interviewDefaults struct {
	CompanyPreset    string
	InterviewMode    string
	CurrentPhase     string
	TimeLimitSeconds int
	NoRun            bool
	NoAutocomplete   bool
	RequiresPlan     bool
}

func defaultsForInterview(companyPreset, interviewMode string) interviewDefaults {
	companyPreset = normalizeCompanyPreset(companyPreset)
	interviewMode = normalizeInterviewMode(interviewMode)
	defaults := interviewDefaults{
		CompanyPreset:    companyPreset,
		InterviewMode:    interviewMode,
		CurrentPhase:     "clarify",
		TimeLimitSeconds: 2700,
		NoRun:            true,
		NoAutocomplete:   true,
		RequiresPlan:     true,
	}
	if companyPreset == "amazon" {
		defaults.TimeLimitSeconds = 3600
	}
	if interviewMode == "practice" {
		defaults.NoRun = false
		defaults.NoAutocomplete = false
	}
	return defaults
}

func normalizeCompanyPreset(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "google", "meta", "amazon", "generic":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "google"
	}
}

func normalizeInterviewMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "real", "practice":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "real"
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
