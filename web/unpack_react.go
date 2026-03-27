package web

import (
	"encoding/json"
	"path"
	"regexp"
	"strings"

	"github.com/2xxn/nxutils/io"
)

type ReactMap struct {
	// Version        int      `json:"version"`
	// File           string   `json:"file"`
	// Mappings       string   `json:"mappings"`
	// Names          []string `json:"names"`
	// SourceRoot     string   `json:"sourceRoot"`
	// commented out until i find a use for these
	Sources        []string `json:"sources"`
	SourcesContent []string `json:"sourcesContent"`
}

func preparePath(p string, inSrc bool) string {
	if p == "" {
		return ""
	}

	cleanPath := path.Clean(p)

	for strings.HasPrefix(cleanPath, "..") {
		cleanPath = strings.TrimPrefix(cleanPath, "../")
		cleanPath = path.Clean(cleanPath)
		if cleanPath == "." {
			cleanPath = ""
			break
		}
	}

	cleanPath = strings.TrimPrefix(cleanPath, "./")

	if cleanPath == "" {
		return ""
	}

	if inSrc {
		cleanPath = path.Join("src", cleanPath)
	}

	return cleanPath
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

// Grabs internal links from a React HTML file, adding .map to the end has a chance of working
// returns a list of links (without .map or url, only relative path)
func GrabReactLinksFromHTML(html string) []string {
	var links []string

	// Thanks ChatGPT for the regex!
	re := regexp.MustCompile(`<(?:script[^>]*src|link[^>]*href)=["']([^"']*\.[0-9a-f]{8}\.[^"']+)["'][^>]*>`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		url := match[1]
		if url[:4] == "http" {
			continue // skip external links
		}

		links = append(links, url)
	}

	return links
}
