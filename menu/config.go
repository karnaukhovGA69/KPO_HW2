package menu

import (
	"encoding/json"
	"os"
)

func Load(path string) (Menu, error) {
	f, err := os.Open(path)
	if err != nil {
		return Menu{}, err
	}
	defer f.Close()

	var items []Item
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return Menu{}, err
	}
	return Menu{Items: items}, nil
}
