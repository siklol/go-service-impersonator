package docker_test

import (
	"github.com/siklol/go-service-impersonator/docker"
	"log"
	"os"
	"testing"
)

func TestTarPacker_Pack(t *testing.T) {
	// given
	fileName := "/tmp/test.tar"

	// when
	packer := docker.NewTarPacker()
	packer.Add("first_file.txt", "Das steht in der ersten Datei")
	packer.Add("seondFile.txt", "Blablabla\nBlablabla")
	packer.AddDockerfile(`FROM node:latest
MAINTAINER Christian Lück <christian@lueck.tv>

RUN npm install -g json-server

WORKDIR /data
VOLUME /data

ADD db.json /data

EXPOSE 8080
CMD ["json-server", "/data/db.json", "-p", "8080"]`)
	err := packer.ToFile(fileName)

	if err != nil {
		log.Fatal(err)
		t.FailNow()
	}

	defer os.Remove(fileName)
}
