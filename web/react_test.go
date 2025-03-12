package web

import (
	"io"
	"net/http"
	"slices"
	"testing"
)

func TestGrabReactLinksFromHTML(t *testing.T) {
	html := `<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="description" content="Web site created using create-react-app"><script defer="defer" src="/static/js/main.6d790b08.js"></script><link href="/static/css/main.89e1fca1.css" rel="stylesheet"><script src="chrome-extension://abongimigdegepeildlddiiliadolokj/dist/inapp.js"></script></head></html>`
	links := GrabReactLinksFromHTML(html)

	if len(links) == 0 {
		t.Error("No links found")
	}

	t.Log("Links found:", len(links))
}

func TestReactUnpack(t *testing.T) {
	url := "http://138.199.150.107:3000"
	req, err := http.Get(url)
	if err != nil {
		t.Fatal("Site is probably down, this test unit must be updated")
	}

	defer req.Body.Close()

	content, err := io.ReadAll(req.Body)
	if err != nil {
		t.Error(err)
	}

	contents := RecognizeContentFromHTML(string(content))
	if !slices.Contains(contents, ST_REACT) {
		t.Error("React not found")
	}

	reactLinks := GrabReactLinksFromHTML(string(content))
	if len(reactLinks) == 0 {
		t.Error("React links not found")
	}

	maps := [][]byte{}
	for _, relativePath := range reactLinks {
		t.Log("Downloading", relativePath)

		resp, err := http.Get(url + relativePath + ".map")
		if err != nil {
			t.Error(err)
		}

		defer resp.Body.Close()

		reactContent, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		maps = append(maps, reactContent)
	}

	dir, err := UnpackReactMaps(maps, true)

	if err != nil {
		t.Error(err)
	}

	t.Log("React maps unpacked, base dir files/directories: ", len(dir.Files)+len(dir.Directories))
	// dir.SaveAsZip("react.zip") // optional
	// TestReactUnpack(t)
}
