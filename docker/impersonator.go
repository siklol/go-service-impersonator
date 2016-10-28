package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
	"log"
	"os"
	"time"
)

var (
	tags = map[string][]string{
		"rest": []string{"impersonator-rest", "impersonator-rest:0.1"},
	}
	tmpTarFile     = "/tmp/impersonator_%s.tar"
	restDockerfile = `FROM node:latest
MAINTAINER Christian Lück <christian@lueck.tv>

RUN npm install -g json-server

WORKDIR /data
VOLUME /data

ADD db.json /data

EXPOSE 8080
CMD ["json-server", "/data/db.json", "-p", "8080"]`
)

type Impersonator struct {
	cli      *client.Client
	ctx      context.Context
	tempFile string
}

func NewImpersonator() (*Impersonator, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return &Impersonator{}, err
	}

	for tagType, tagNames := range tags {
		if len(tagNames) == 0 {
			return &Impersonator{}, errors.New(fmt.Sprintf("Not enough tags defined for %s", tagType))
		}
	}

	return &Impersonator{
		cli:      cli,
		ctx:      context.Background(),
		tempFile: fmt.Sprintf(tmpTarFile, time.Now().Format(time.RFC3339Nano)),
	}, nil
}

// RESTService creates a docker image, container and starts it up
func (i *Impersonator) RESTService(resourceMap map[string]interface{}) error {
	jsonResource, err := json.Marshal(resourceMap)

	if err != nil {
		return err
	}

	packer := NewTarPacker()
	packer.Add("db.json", string(jsonResource))
	packer.AddDockerfile(restDockerfile)

	if err = packer.ToFile(i.tempFile); err != nil {
		return err
	}

	reader, err := os.Open(i.tempFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	dockerContext, relDockerFile, err := builder.GetContextFromReader(reader, "Dockerfile")
	if err != nil {
		return err
	}

	log.Println(relDockerFile)
	resp, err := i.cli.ImageBuild(i.ctx, dockerContext, types.ImageBuildOptions{
		Tags:       tags["rest"],
		Dockerfile: relDockerFile,
	})

	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	log.Println(buf.String())

	os.Remove(i.tempFile)

	var internalPort nat.Port
	internalPort = "8080/tcp"

	respC, err := i.cli.ContainerCreate(
		i.ctx,
		&container.Config{
			Image: tags["rest"][0],
			ExposedPorts: nat.PortSet{
				internalPort: struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				internalPort: []nat.PortBinding{{HostPort: "8080"}},
			},
		},
		&network.NetworkingConfig{},
		"",
	)

	i.cli.ContainerStart(i.ctx, respC.ID, types.ContainerStartOptions{})

	return nil
}
