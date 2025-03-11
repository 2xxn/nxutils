package io

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
)

// File represents a file in the virtual disk.
type File struct {
	Name    string
	Content []byte
}

// Directory represents a directory in the virtual disk.
type Directory struct {
	Path        string
	Directories []*Directory
	Files       []*File
}

// NewVirtualDisk creates a new virtual disk with the root directory.
func NewVirtualDisk() *Directory {
	return &Directory{Path: "/"}
}

// CompressAsZip returns []byte which, when saved, is a valid zip archive.
func (d *Directory) CompressAsZip() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	var addDir func(dir *Directory, base string) error
	addDir = func(dir *Directory, base string) error {
		// If we are in a subdirectory, create a directory entry.
		// (Zip entries for directories should end with a '/'.)
		if base != "" {
			h := &zip.FileHeader{
				Name:   strings.TrimSuffix(base, "/") + "/",
				Method: zip.Deflate,
			}

			h.SetMode(0755)
			if _, err := zipWriter.CreateHeader(h); err != nil {
				return err
			}
		}

		// Write the files in the current directory.
		for _, file := range dir.Files {
			filePath := base + file.Name
			h := &zip.FileHeader{
				Name:   filePath,
				Method: zip.Deflate,
			}
			h.SetMode(0644)
			w, err := zipWriter.CreateHeader(h)
			if err != nil {
				return err
			}
			if _, err := w.Write(file.Content); err != nil {
				return err
			}
		}

		// Recurse into subdirectories.
		for _, subdir := range dir.Directories {
			newBase := base + subdir.Path
			if !strings.HasSuffix(newBase, "/") {
				newBase += "/"
			}

			if err := addDir(subdir, newBase); err != nil {
				return err
			}
		}
		return nil
	}

	// Start at the root. If d.path is "/" we use an empty base.
	base := ""
	if d.Path != "/" && d.Path != "" {
		base = d.Path
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
	}
	if err := addDir(d, base); err != nil {
		// Handle error as appropriate.
		return []byte{}
	}

	zipWriter.Close()
	return buf.Bytes()
}

// SaveAsZip writes the zip archive to the given file name.
func (d *Directory) SaveAsZip(fileName string) error {
	zipped := d.CompressAsZip()

	// Write the zip data to file.
	err := os.WriteFile(fileName, zipped, 0644)
	return err
}

// Will resolve things such as 'node_modules/prop-types/index.js' or 'components/BookVisit.js' and create directories if not existent
func (d *Directory) WriteFile(relativePath string, content []byte) {
	parts := strings.Split(relativePath, "/")
	currentDir := d

	for i := 0; i < len(parts)-1; i++ {
		dirName := parts[i]
		if dirName == "" {
			continue
		}

		var nextDir *Directory
		for _, subDir := range currentDir.Directories {
			if subDir.Path == dirName {
				nextDir = subDir
				break
			}
		}

		if nextDir == nil {
			nextDir = &Directory{
				Path: dirName,
			}
			currentDir.Directories = append(currentDir.Directories, nextDir)
		}
		currentDir = nextDir
	}

	fileName := parts[len(parts)-1]
	if fileName == "" {
		return
	}

	for _, f := range currentDir.Files {
		if f.Name == fileName {
			f.Content = content
			return
		}
	}

	newFile := &File{
		Name:    fileName,
		Content: content,
	}
	currentDir.Files = append(currentDir.Files, newFile)
}

func (d *Directory) DeleteFile(relativePath string) {
	parts := strings.Split(relativePath, "/")
	currentDir := d

	for i := 0; i < len(parts)-1; i++ {
		dirName := parts[i]
		if dirName == "" {
			continue
		}

		var nextDir *Directory
		for _, subDir := range currentDir.Directories {
			if subDir.Path == dirName {
				nextDir = subDir
				break
			}
		}
		if nextDir == nil {
			return
		}
		currentDir = nextDir
	}

	fileName := parts[len(parts)-1]
	if fileName == "" {
		return
	}

	for i, f := range currentDir.Files {
		if f.Name == fileName {
			currentDir.Files = append(currentDir.Files[:i], currentDir.Files[i+1:]...)
			return
		}
	}
}

func (d *Directory) GetFile(relativePath string) *File {
	parts := strings.Split(relativePath, "/")
	currentDir := d

	for i := 0; i < len(parts)-1; i++ {
		dirName := parts[i]
		if dirName == "" {
			continue
		}

		var nextDir *Directory
		for _, subDir := range currentDir.Directories {
			if subDir.Path == dirName {
				nextDir = subDir
				break
			}
		}

		if nextDir == nil {
			return nil
		}
		currentDir = nextDir
	}

	fileName := parts[len(parts)-1]
	if fileName == "" {
		return nil
	}

	for _, f := range currentDir.Files {
		if f.Name == fileName {
			return f
		}
	}

	return nil
}

func (d *Directory) GetDirectory(relativePath string) *Directory {
	parts := strings.Split(relativePath, "/")
	currentDir := d

	for i := 0; i < len(parts); i++ {
		dirName := parts[i]
		if dirName == "" {
			continue
		}

		var nextDir *Directory
		for _, subDir := range currentDir.Directories {
			if subDir.Path == dirName {
				nextDir = subDir
				break
			}
		}

		if nextDir == nil {
			return nil
		}
		currentDir = nextDir
	}

	return currentDir
}

func (d *Directory) CreateDirectory(relativePath string) {
	parts := strings.Split(relativePath, "/")
	currentDir := d

	for i := 0; i < len(parts); i++ {
		dirName := parts[i]
		if dirName == "" {
			continue
		}

		var nextDir *Directory
		for _, subDir := range currentDir.Directories {
			if subDir.Path == dirName {
				nextDir = subDir
				break
			}
		}

		if nextDir == nil {
			nextDir = &Directory{
				Path: dirName,
			}
			currentDir.Directories = append(currentDir.Directories, nextDir)
		}
		currentDir = nextDir
	}
}
