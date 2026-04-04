package web

import (
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// Content Recognized Website
type CRWebsite struct {
	URL             string
	PoweredBy       string
	DetectedContent []string
	Emails          []string
}

// PB - Powered By
const (
	PB_IIS        = "iis"
	PB_APACHE     = "apache"
	PB_NGINX      = "nginx"
	PB_PHP        = "php"
	PB_PHP_OLD    = "php_old" // PHP 5.x and older, a category on its own because of the large number of vulnerabilities
	PB_CLOUDFLARE = "cloudflare"
	PB_CLOUDFRONT = "cloudfront"
	PB_UNKNOWN    = "unknown"
)

// ST - Service Type
const (
	ST_WORDPRESS = "wordpress"
	ST_JOOMLA    = "joomla"
	ST_DRUPAL    = "drupal"
	ST_OPENDIR   = "open" // Directory listing
	ST_DBA       = "dba"  // Database Administration System (phpMyAdmin, phpPgAdmin, etc.)
	ST_JENKINS   = "jenkins"
	ST_ASPNET    = "aspnet" // ASP.NET errors/web services, can possibly be IIS shortname scanned
	ST_REACT     = "react"  // React App (create-react-app), could have map files
	ST_GITLAB    = "gitlab"
	ST_FORGEJO   = "forgejo"
	ST_JIRA      = "jira"
	ST_SNRS      = "snrs"   // Synerise API
	ST_MSLOGIN   = "msl"    // Microsoft Login page
	ST_GMLOGIN   = "gml"    // Google Mail Login page
	ST_CFACCESS  = "cfa"    // Cloudflare Access login page
	ST_NGINX     = "nginx"  // Nginx default page
	ST_APACHE    = "apache" // Apache default page
	ST_IIS       = "iis"    // IIS default page
	// ST_SYMFONY   = "symfony" // Symfony TODO: implement
)

// Try grabbing all emails from the HTML
func GetEmailsFromHTML(html string) []string {
	re := regexp.MustCompile("(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])")
	return re.FindAllString(html, -1)
}

func RecognizePBFromHeaders(headers map[string]string) []string {
	normalized := map[string]string{}
	for key, value := range headers {
		normalized[strings.ToLower(key)] = strings.ToLower(value)
	}

	detected := []string{}
	seen := map[string]struct{}{}

	add := func(pb string) {
		if _, exists := seen[pb]; exists {
			return
		}

		detected = append(detected, pb)
		seen[pb] = struct{}{}
	}

	detectFromValue := func(value string) {
		if value == "" {
			return
		}

		rules := []struct {
			pb      string
			needles []string
		}{
			{pb: PB_CLOUDFLARE, needles: []string{"cloudflare"}},
			{pb: PB_CLOUDFRONT, needles: []string{"cloudfront"}},
			{pb: PB_IIS, needles: []string{"iis", "asp.net"}},
			{pb: PB_APACHE, needles: []string{"apache"}},
			{pb: PB_NGINX, needles: []string{"nginx"}},
		}

		for _, rule := range rules {
			for _, needle := range rule.needles {
				if strings.Contains(value, needle) {
					add(rule.pb)
					break
				}
			}
		}

		if strings.Contains(value, "php") {
			if strings.Contains(value, "php/5.") || strings.Contains(value, "php/4.") {
				add(PB_PHP_OLD)
				return
			}

			add(PB_PHP)
		}
	}

	for _, key := range []string{"server", "x-powered-by", "via", "x-cache"} {
		detectFromValue(normalized[key])
	}

	if _, ok := normalized["cf-ray"]; ok {
		add(PB_CLOUDFLARE)
	}

	if _, ok := normalized["cf-cache-status"]; ok {
		add(PB_CLOUDFLARE)
	}

	if _, ok := normalized["x-amz-cf-id"]; ok {
		add(PB_CLOUDFRONT)
	}

	if _, ok := normalized["x-amz-cf-pop"]; ok {
		add(PB_CLOUDFRONT)
	}

	if len(detected) == 0 {
		return []string{PB_UNKNOWN}
	}

	return detected
}

func RecognizeContentFromHTML(html string) []string {
	return detectContentFromHTMLIter(html)
}

func detectContentFromHTMLIter(html string) []string {
	detected := []string{}
	skip := map[string]struct{}{}

	check := func(name string, cond bool) {
		if cond {
			if _, exists := skip[name]; !exists {
				detected = append(detected, name)
				skip[name] = struct{}{}
			}
		}
	}

	title := getTitleFromHTML(html)

	check(ST_OPENDIR, strings.Contains(html, `<title>Index of `))
	check(ST_ASPNET, isASP(html))
	check(ST_REACT, isReact(html))
	check(ST_DBA, isDBA(html))
	check(ST_JENKINS, strings.HasPrefix(title, "Jenkins"))
	check(ST_JIRA, strings.Contains(html, `id="jira"`))
	check(ST_FORGEJO, strings.Contains(html, `href="https://forgejo.org"`))
	check(ST_SNRS, strings.HasPrefix(title, "Synerise "))
	check(ST_CFACCESS, strings.HasSuffix(title, "Cloudflare Access"))
	check(ST_NGINX, strings.Contains(html, "Welcome to nginx!"))
	check(ST_APACHE, strings.Contains(html, "Apache2 Debian Default Page"))
	check(ST_GMLOGIN, strings.Contains(html, "Sign in - Google Accounts"))
	check(ST_IIS, strings.Contains(title, "IIS Windows Server"))
	check(ST_GITLAB, strings.Contains(html, "GitLab"))
	check(ST_MSLOGIN, strings.Contains(html, " Copyright (C) Microsoft Corporation. All rights reserved."))
	check(ST_WORDPRESS, strings.Contains(html, "wp-content"))
	check(ST_JOOMLA, strings.Contains(html, "content=\"Joomla! - Open Source Content Management\""))
	check(ST_JOOMLA, strings.Contains(html, "/using-joomla/"))
	check(ST_DRUPAL, strings.Contains(html, "data-drupal-"))

	return detected
}

func ListOpenDirFilesRecursive(openDirBaseUrl string, relativePath string) ([]string, error) {
	var files []string
	TO_SKIP := []string{"Name", "Last modified", "Size", "Description", "Parent Directory"}

	req, err := http.Get(openDirBaseUrl + relativePath)
	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	content, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`<a href=\"([^\"]+)\">(.*?)<\/a>`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if (len(match) < 3) || (match[2] == "../" || match[1] == "../") {
			continue
		}

		if slices.Contains(TO_SKIP, match[2]) {
			continue
		}

		if strings.HasSuffix(match[1], "/") {
			subFiles, err := ListOpenDirFilesRecursive(openDirBaseUrl, relativePath+match[1])
			if err != nil {
				continue
			}

			files = append(files, subFiles...)
		} else {
			files = append(files, relativePath+match[1])
		}
	}

	return files, nil
}

func HTMLDecode(s string) string {
	toReplace := map[string]string{
		"&lt;":   "<",
		"&gt;":   ">",
		"&amp;":  "&",
		"&quot;": "\"",
		"&apos;": "'",
		"&nbsp;": " ",
	}

	for k, v := range toReplace {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

// #region Helper functions
func getTitleFromHTML(html string) string {
	re := regexp.MustCompile(`<title>(.*?)</title>`)
	match := re.FindStringSubmatch(html)

	if len(match) > 1 {
		return match[1]
	}

	return ""
}

// #region Content recognition functions
var isASP = func(html string) bool {
	checks := []bool{
		strings.Contains(html, "System.Web.HttpException"),
		strings.Contains(html, "System.UriFormatException"),
		strings.Contains(html, "<%@ WebService"),

		strings.Contains(html, "aspnetForm"),
		regexp.MustCompile(`href\s*=\s*"(?:[\w\-/\.]+)\.(?:aspx|ashx|asmx)"`).MatchString(html),
		// ASP.NET error pages often contain the error message in the body, and a 200 OK status code, which is a pain to detect without parsing the HTML
		strings.Contains(html, "<title>403 - Forbidden: Access is denied.</title>"),
		strings.Contains(html, "<title>404 - File or directory not found.</title>"),
	}

	return slices.Contains(checks, true)
}

var isReact = func(html string) bool {
	re := regexp.MustCompile(`<(?:script[^>]*src|link[^>]*href)=["']([^"']*\.[0-9a-f]{8}\.[^"']+)["'][^>]*>`)

	return re.MatchString(html)
}

var isDBA = func(html string) bool {
	title := getTitleFromHTML(html)
	checks := []bool{
		strings.Contains(title, "phpMyAdmin"),
		strings.Contains(title, "phpPgAdmin"),
		strings.Contains(title, "MySQL"),
		strings.Contains(title, "Mariadb"),
		strings.Contains(title, "PostgreSQL"),
		strings.Contains(title, "Mongodb"),
		strings.Contains(title, "RockMongo"),
	}

	return slices.Contains(checks, true)
}

// #endregion
// #endregion
