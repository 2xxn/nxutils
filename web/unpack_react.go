package web

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/nextu1337/nxutils/io"
)

type ReactMap struct {
	Version        int      `json:"version"`    // USELESS
	File           string   `json:"file"`       // USELESS
	Mappings       string   `json:"mappings"`   // USELESS
	Names          []string `json:"names"`      // USELESS
	SourceRoot     string   `json:"sourceRoot"` // USELESS
	Sources        []string `json:"sources"`
	SourcesContent []string `json:"sourcesContent"`
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

// Grabs internal links from a React HTML file, adding .map to the end has a chance of working
// returns a list of links (without .map or url, only relative path)
func GrabReactLinksFromHTML(html string) []string {
	var links []string

	// Thanks ChatGPT for the regex!
	re := regexp.MustCompile(`<(?:script[^>]*src|link[^>]*href)=["']((?=[^"']*(?:\/|^)[^\/]+\.[0-9a-f]{8}\.)[^"']+)["'][^>]*>"`)
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
