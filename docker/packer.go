package docker

import (
	"archive/tar"
	"bufio"
	"errors"
	"io"
	"os"
)

var (
	ErrMissingDockerFile = errors.New("Missing Dockerfile in package")
)

// DockerPacker interface
type DockerPacker interface {
	Add(name string, body string)
	AddDockerfile(file string) error
	Pack(w io.Writer) error
	ToFile(fileName string) error
}

type tarPackedFile struct {
	Name string
	Body string
}

type TarPacker struct {
	files      []tarPackedFile
	dockerFile tarPackedFile
}

func NewTarPacker() DockerPacker {
	return &TarPacker{}
}

// Add a file to the tar package
func (p *TarPacker) Add(name string, body string) {
	p.files = append(p.files, tarPackedFile{
		Name: name,
		Body: body,
	})
}

// AddDockerFile adds Dockerfile content to the tar package
func (p *TarPacker) AddDockerfile(content string) error {
	df := tarPackedFile{
		Name: "Dockerfile",
		Body: content,
	}

	p.dockerFile = df

	p.files = append(p.files, df)
	return nil
}

// Pack everything and write it to a writer
func (p *TarPacker) Pack(w io.Writer) error {
	pDockerFile := tarPackedFile{}
	if p.dockerFile == pDockerFile {
		return ErrMissingDockerFile
	}

	tw := tar.NewWriter(w)

	for _, file := range p.files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			return err
		}
	}

	return nil
}

// ToFile writes to a file on the OS
func (p *TarPacker) ToFile(fileName string) error {
	f, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	err = p.Pack(w)

	if err != nil {
		return err
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return err
	}
	w.Flush()

	return nil
}
