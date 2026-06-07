package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

func LoadSeedProblems() ([]Problem, error) {
	data, err := readFirstExisting(seedCandidates())
	if err != nil {
		return nil, err
	}
	var problems []Problem
	if err := json.Unmarshal(data, &problems); err != nil {
		return nil, err
	}
	return problems, nil
}

func seedCandidates() []string {
	candidates := []string{}
	if configured := os.Getenv("SEED_FILE"); configured != "" {
		candidates = append(candidates, configured)
	}
	candidates = append(candidates, "seed/problems.json")
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(file), "../../seed/problems.json"))
	}
	return candidates
}

func readFirstExisting(paths []string) ([]byte, error) {
	var lastErr error
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
