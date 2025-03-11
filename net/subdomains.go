package net

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

type crtShResponse struct {
	CommonName string `json:"common_name"`
	NameValue  string `json:"name_value"`
}

func parseCrtShDomain(domain string) []string {
	var parsed []string
	domains := strings.Split(domain, "\n")

	for _, domain := range domains {
		if strings.HasPrefix(domain, "*.") {
			domain, _ = strings.CutPrefix(domain, "*.")
		}

		parsed = append(parsed, domain)
	}

	return parsed
}

// This has a 50/50 chance of returning a 500 error, crt.sh is a bit unreliable
func GetSubdomains(domain string) ([]string, error) {
	var value []*crtShResponse
	var domainsArr []string

	resp, err := http.Get("https://crt.sh/?output=json&q=" + domain)
	if err != nil {
		return domainsArr, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return domainsArr, err
	}

	json.Unmarshal(data, &value)

	for _, v := range value {
		domainsArr = append(domainsArr, parseCrtShDomain(v.CommonName)...)
		domainsArr = append(domainsArr, parseCrtShDomain(v.NameValue)...)
	}

	return removeDuplicate(domainsArr), nil
}
