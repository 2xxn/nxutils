package web

import (
	"testing"

	"github.com/nextu1337/nxutils/io"
)

func TestWriteFileEnv(t *testing.T) {
	vd := io.NewVirtualDisk()
	err := writeSymfonyEnv("https://it20.phpinfo.eu/84/", vd)
	if err != nil {
		t.Error(err)
	}

	if len(vd.Files) == 0 {
		t.Error("No files written")
	}

	t.Log("Files written:", len(vd.Files))
	for _, file := range vd.Files {
		t.Logf("File: %s, Size: %d\n", file.Name, len(file.Content))
	}

	if vd.GetFile("phpinfo.html") == nil {
		t.Error("phpinfo.html not written")
	}

	if vd.GetFile("server.env") == nil {
		t.Error("server.env not written")
	}

	// vd.SaveAsZip("symfony_test.zip") // optional

	// .env is optional, empty on the test site
}
