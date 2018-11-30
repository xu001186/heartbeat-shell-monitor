package docker

import (
	"archive/tar"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func (d *DockerClient) copyFilesToContainer(filePath string, dstPath string, mode string) error {
	err := d.CheckClient()
	if err != nil {
		return err
	}
	if _, err = os.Stat(filePath); err != nil {
		return err
	}

	defer d.Close()
	_, name := filepath.Split(filePath)

	if _, err := d.Run("", "mkdir", "-p", dstPath); err != nil {
		return err
	}
	if err := tarFile(filePath); err != nil {
		return err
	}
	id, _, err := d.getContainerIDAndState()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	fromFile, err := os.Open(filePath + ".tar")
	if err != nil {
		return err
	}

	err = d.dockerClient.CopyToContainer(ctx, id, dstPath, fromFile, types.CopyToContainerOptions{AllowOverwriteDirWithFile: true})
	if err != nil {
		return err
	}
	if _, err := d.Run("", "chmod", mode, filepath.Join(dstPath, name+".tar")); err != nil {
		return err
	}
	if _, err := d.Run("", "mv", filepath.Join(dstPath, name+".tar"), filepath.Join(dstPath, name)); err != nil {
		return err
	}

	return os.Remove(filePath + ".tar")
}

func tarFile(filePath string) error {
	fileBody, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	_, name := filepath.Split(filePath)

	tarFile, err := os.OpenFile(filePath+".tar", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(tarFile)

	hdr := &tar.Header{
		Name: name + `.tar`,
		Mode: 0600,
		Size: int64(len(fileBody)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(fileBody)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	return tarFile.Close()

}

func (d *DockerClient) getContainerIDAndState() (id string, state string, err error) {
	var containerDetail types.Container
	containerDetail, err = d.getContainerBriefDetails(true)
	if err != nil {
		return
	}
	id = containerDetail.ID
	state = containerDetail.State
	err = nil
	return
}

func (d *DockerClient) getContainerBriefDetails(all bool) (types.Container, error) {
	err := d.CheckClient()
	if err != nil {
		return types.Container{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	filter := filters.NewArgs()
	for _, value := range d.Filter {
		kv := strings.Split(value, ":")
		if len(kv) == 2 {
			filter.Add(kv[0], kv[1])
		}
	}
	d.actionMutex.RLock()
	containers, err := d.dockerClient.ContainerList(ctx, types.ContainerListOptions{All: all, Filters: filter})
	d.actionMutex.RUnlock()
	if err != nil {
		return types.Container{}, err
	}
	if len(containers) == 0 {
		return types.Container{}, fmt.Errorf("Container %v doesnot exist", d.Filter)
	}
	if len(containers) > 1 {
		return types.Container{}, fmt.Errorf("Return more than 1 container with filter %v", d.Filter)
	}
	return containers[0], err
}

func (d *DockerClient) inspectContainer() (types.ContainerJSON, error) {
	err := d.CheckClient()
	if err != nil {
		return types.ContainerJSON{}, err
	}

	id, _, err := d.getContainerIDAndState()

	if err != nil {
		return types.ContainerJSON{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()
	d.actionMutex.RLock()
	containerJSON, err := d.dockerClient.ContainerInspect(ctx, id)
	d.actionMutex.RUnlock()
	return containerJSON, err
}
