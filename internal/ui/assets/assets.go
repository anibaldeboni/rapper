package assets

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed figlets/*
var figlets embed.FS

type Figlets map[string][]byte

func LoadAllFiglets() (Figlets, error) {
	files, err := figlets.ReadDir("figlets")
	if err != nil {
		return nil, fmt.Errorf("Error reading figlet files: %w", err)
	}

	var figletMap = make(map[string][]byte)
	for _, file := range files {
		if !file.IsDir() {
			font, err := figlets.ReadFile("figlets/" + file.Name())
			if err != nil {
				return nil, err
			}
			figletMap[strings.TrimSuffix(file.Name(), ".flf")] = font
		}
	}

	return figletMap, nil
}
