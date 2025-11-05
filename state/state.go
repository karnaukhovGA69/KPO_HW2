package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const fileName = ".state.json"

type appState struct {
	CurrentAccountID string `json:"current_account_id"`
}

func path() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, fileName)
}

func LoadAccountID() (string, error) {
	b, err := os.ReadFile(path())
	if err != nil {
		return "", err
	}
	var s appState
	if err := json.Unmarshal(b, &s); err != nil {
		return "", err
	}
	return s.CurrentAccountID, nil
}

func SaveAccountID(id string) error {
	b, err := json.MarshalIndent(appState{CurrentAccountID: id}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path(), b, 0644)
}
