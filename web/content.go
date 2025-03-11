package web

import (
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
	PB_IIS     = "iis"
	PB_APACHE  = "apache"
	PB_NGINX   = "nginx"
	PB_PHP     = "php"
	PB_PHP_OLD = "php_old" // PHP 5.x and older, a category on its own because of the large number of vulnerabilities
)

// ST - Service Type
const (
	ST_OPENDIR  = "open" // Directory listing
	ST_DBA      = "dba"  // Database Administration System (phpMyAdmin, phpPgAdmin, etc.)
	ST_JENKINS  = "jenkins"
	ST_WEBMAIL  = "webmail" // Webmail login page (Roundcube, SquirrelMail, etc.) TODO: implement
	ST_ASPNET   = "aspnet"  // ASP.NET errors/web services, can possibly be IIS shortname scanned
	ST_REACT    = "react"   // React App (create-react-app), could have map files
	ST_GITLAB   = "gitlab"  // TODO: implement
	ST_FORGEJO  = "forgejo"
	ST_JIRA     = "jira"
	ST_SNRS     = "snrs"   // Synerise API
	ST_MSLOGIN  = "msl"    // Microsoft Login page
	ST_GMLOGIN  = "gml"    // Google Mail Login page TODO: implement
	ST_CFACCESS = "cfa"    // Cloudflare Access login page
	ST_NGINX    = "nginx"  // Nginx default page
	ST_APACHE   = "apache" // Apache default page TODO: implement
	ST_IIS      = "iis"    // IIS default page TODO: implement
)

// Try grabbing all emails from the HTML
func GetEmailsFromHTML(html string) []string {
	re := regexp.MustCompile("(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])")
	return re.FindAllString(html, -1)
}

// TODO: Implement
func RecognizePBFromHeaders(headers map[string]string) {
	// server, X-Powered-By,
}

func RecognizeContentFromHTML(html string) []string {
	return detectContentFromHTML(html, []string{})
}

func detectContentFromHTML(html string, skipArr []string) []string {
	if len(skipArr) > 100 {
		return skipArr
	}

	next := func(whatToSkip string) []string {
		return detectContentFromHTML(html, append(skipArr, whatToSkip))
	}

	not := func(v string) bool {
		return !slices.Contains(skipArr, v)
	}

	switch true {
	case not(ST_OPENDIR) && strings.Contains(html, `<title>Index of /`):
		return next(ST_OPENDIR)
	case not(ST_ASPNET) && isASP(html):
		return next(ST_ASPNET)
	case not(ST_REACT) && isReact(html):
		return next(ST_REACT)
	case not(ST_DBA) && isDBA(html):
		return next(ST_DBA)
	case not(ST_JENKINS) && strings.HasPrefix(getTitleFromHTML(html), "Jenkins"):
		return next(ST_JENKINS)
	case not(ST_JIRA) && strings.Contains(html, "id=\"jira\""):
		return next(ST_JIRA)
	case not(ST_FORGEJO) && strings.Contains(html, "href=\"https://forgejo.org\""):
		return next(ST_FORGEJO)
	case not(ST_SNRS) && strings.HasPrefix(getTitleFromHTML(html), "Synerise "):
		return next(ST_SNRS)
	case not(ST_CFACCESS) && strings.HasSuffix(getTitleFromHTML(html), "Cloudflare Access"):
		return next(ST_CFACCESS)
	case not(ST_NGINX) && strings.Contains(html, "Welcome to nginx!"):
		return next(ST_NGINX)
	case not(ST_APACHE) && strings.Contains(html, "Apache2 Debian Default Page"): // TODO: Improve
		return next(ST_APACHE)
	case not(ST_MSLOGIN) && strings.Contains(html, " Copyright (C) Microsoft Corporation. All rights reserved."):
		return next(ST_MSLOGIN)
	}

	return skipArr
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
	}

	return slices.Contains(checks, true)
}

var isReact = func(html string) bool {
	re := regexp.MustCompile(`<(?:script[^>]*src|link[^>]*href)=["']((?=[^"']*(?:\/|^)[^\/]+\.[0-9a-f]{8}\.)[^"']+)["'][^>]*>`)

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
