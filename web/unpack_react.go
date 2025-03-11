package web

import (
	"encoding/json"
	"strings"

	"github.com/nextu1337/nxutils/io"
)

type ReactMap struct {
	Version        int         `json:"version"`  // USELESS
	File           string      `json:"file"`     // USELESS
	Mappings       string      `json:"mappings"` // USELESS
	Names          []string    `json:"names"`    // USELESS
	SourceRoot     interface{} `json:"sourceRoot"`
	Sources        []string    `json:"sources"`
	SourcesContent []string    `json:"sourcesContent"`
}

func preparePath(path string, inSrc bool) string {
	pathParts := strings.Split(path, "/")
	prefix := "src/"
	if !inSrc {
		prefix = ""
	}

	if len(pathParts) == 0 {
		return ""
	}

	if len(pathParts) == 1 {
		return prefix + path
	}

	if pathParts[0] == ".." {
		// if pathParts[1] == "webpack" {
		// 	// cba handling it, don't understand
		// 	return ""
		// }

		return strings.Join(pathParts[1:], "/")
	}

	if pathParts[0] == "." {
		return prefix + strings.Join(pathParts[1:], "/")
	}

	return prefix + path
}

// jsonFiles is multiple .map file contents
// inSrc is whether the sources should be in src/ or not (true recommended)
// returns a virtual disk with the files that can be saved as a zip
func UnpackReactMaps(jsonFiles [][]byte, inSrc bool) (*io.Directory, error) {
	var reactData ReactMap
	dir := io.NewVirtualDisk()

	for _, jsonData := range jsonFiles {
		err := json.Unmarshal(jsonData, &reactData)
		if err != nil {
			return nil, err
		}

		for i, file := range reactData.Sources {
			content := []byte(reactData.SourcesContent[i])
			path := preparePath(file, inSrc)

			if path == "" {
				continue
			}

			dir.WriteFile(path, content)
		}
	}

	return dir, nil
}
