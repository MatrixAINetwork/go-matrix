// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package build

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Archive interface {
	// Directory adds a new directory entry to the archive and sets the
	// directory for subsequent calls to Header.
	Directory(name string) error

	// Header adds a new file to the archive. The file is added to the directory
	// set by Directory. The content of the file must be written to the returned
	// writer.
	Header(os.FileInfo) (io.Writer, error)

	// Close flushes the archive and closes the underlying file.
	Close() error
}

func NewArchive(file *os.File) (Archive, string) {
	switch {
	case strings.HasSuffix(file.Name(), ".zip"):
		return NewZipArchive(file), strings.TrimSuffix(file.Name(), ".zip")
	case strings.HasSuffix(file.Name(), ".tar.gz"):
		return NewTarballArchive(file), strings.TrimSuffix(file.Name(), ".tar.gz")
	default:
		return nil, ""
	}
}

// AddFile appends an existing file to an archive.
func AddFile(a Archive, file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()
	fi, err := fd.Stat()
	if err != nil {
		return err
	}
	w, err := a.Header(fi)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, fd); err != nil {
		return err
	}
	return nil
}

// WriteArchive creates an archive containing the given files.
func WriteArchive(name string, files []string) (err error) {
	archfd, err := os.Create(name)
	if err != nil {
		return err
	}

	defer func() {
		archfd.Close()
		// Remove the half-written archive on failure.
		if err != nil {
			os.Remove(name)
		}
	}()
	archive, basename := NewArchive(archfd)
	if archive == nil {
		return fmt.Errorf("unknown archive extension")
	}
	fmt.Println(name)
	if err := archive.Directory(basename); err != nil {
		return err
	}
	for _, file := range files {
		fmt.Println("   +", filepath.Base(file))
		if err := AddFile(archive, file); err != nil {
			return err
		}
	}
	return archive.Close()
}

type ZipArchive struct {
	dir  string
	zipw *zip.Writer
	file io.Closer
}

func NewZipArchive(w io.WriteCloser) Archive {
	return &ZipArchive{"", zip.NewWriter(w), w}
}

func (a *ZipArchive) Directory(name string) error {
	a.dir = name + "/"
	return nil
}

func (a *ZipArchive) Header(fi os.FileInfo) (io.Writer, error) {
	head, err := zip.FileInfoHeader(fi)
	if err != nil {
		return nil, fmt.Errorf("can't make zip header: %v", err)
	}
	head.Name = a.dir + head.Name
	head.Method = zip.Deflate
	w, err := a.zipw.CreateHeader(head)
	if err != nil {
		return nil, fmt.Errorf("can't add zip header: %v", err)
	}
	return w, nil
}

func (a *ZipArchive) Close() error {
	if err := a.zipw.Close(); err != nil {
		return err
	}
	return a.file.Close()
}

type TarballArchive struct {
	dir  string
	tarw *tar.Writer
	gzw  *gzip.Writer
	file io.Closer
}

func NewTarballArchive(w io.WriteCloser) Archive {
	gzw := gzip.NewWriter(w)
	tarw := tar.NewWriter(gzw)
	return &TarballArchive{"", tarw, gzw, w}
}

func (a *TarballArchive) Directory(name string) error {
	a.dir = name + "/"
	return a.tarw.WriteHeader(&tar.Header{
		Name:     a.dir,
		Mode:     0755,
		Typeflag: tar.TypeDir,
	})
}

func (a *TarballArchive) Header(fi os.FileInfo) (io.Writer, error) {
	head, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return nil, fmt.Errorf("can't make tar header: %v", err)
	}
	head.Name = a.dir + head.Name
	if err := a.tarw.WriteHeader(head); err != nil {
		return nil, fmt.Errorf("can't add tar header: %v", err)
	}
	return a.tarw, nil
}

func (a *TarballArchive) Close() error {
	if err := a.tarw.Close(); err != nil {
		return err
	}
	if err := a.gzw.Close(); err != nil {
		return err
	}
	return a.file.Close()
}
