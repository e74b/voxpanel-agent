package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/specgen"
	"github.com/opencontainers/runtime-spec/specs-go"
	nettypes "go.podman.io/common/libnetwork/types"
)

func handleStartServer(message ControlMessage, routing Routing, app *AppState) {
	uuid := message.Arg["uuid"]
	path, err := os.Getwd()
	app.warnOnError(err, "Failed to fetch working directory")

	dataPath := filepath.Join(path, "data", uuid)
	err = os.MkdirAll(dataPath, os.ModePerm)
	app.logger.Warnf("%s", err)

	image, err := images.GetImage(*app.podmanConnection, "paper-server", nil)
	app.warnOnError(err, "Failed getting podman image")
	spec := specgen.NewSpecGenerator(image.ID, false)
	spec.Remove = new(false)
	dataMount := specs.Mount{
		Source: dataPath,
		Destination: "/data",
		Options: []string{"Z"},
		Type: "bind",
	}
	// TODO: Make a better system of loading default config dirs
	configMount := specs.Mount{
		Source: "/home/e74b/VoxPanel/containers/Servers/config/",
		Destination: "/config",
		Options: []string{"Z"},
		Type: "bind",
	}
	spec.Mounts = append(spec.Mounts, dataMount, configMount)
	port, err := getFreePort()
	app.warnOnError(err, "No free ports found")
	spec.Labels = map[string]string{
		"voxpanel": "true",
		"mc-port": strconv.Itoa(port),
	}
	spec.Stdin = new(true)

	spec.PortMappings = append(spec.PortMappings, nettypes.PortMapping{
		ContainerPort: 25565,
		HostPort: uint16(port),
		HostIP: "0.0.0.0",
	})
	container, err := containers.CreateWithSpec(*app.podmanConnection, spec, nil)
	err = containers.Start(*app.podmanConnection, container.ID, nil)
	app.warnOnError(err, "Failed to create containers")
	containerInfo, err := containers.Inspect(*app.podmanConnection, container.ID, nil)
	app.warnOnError(err, "Inspecting container")

	go containers.Attach(*app.podmanConnection, container.ID, nil, os.Stdout, os.Stdin, nil, nil)

	response := StartServerResponse{
		Status: "ok",
		IPAddress: containerInfo.NetworkSettings.IPAddress,
		ID: container.ID,
		Port: port,
	}

	rpcReply(app, routing, response)
	err = containers.Start(*app.podmanConnection, container.ID, nil)
	app.warnOnError(err, "Failed to start server!")

	go func () {
		containers.Wait(*app.podmanConnection, container.ID, nil)
		app.logger.Info("Cleanup complete.")
	}()
}


func handleStopServer(message ControlMessage, routing Routing, app *AppState) {
	serverId, exists := message.Arg["id"]
	if (!exists) {
		app.logger.Info("Server ID field not present")
		return
	}
	containerExists, err := containers.Exists(*app.podmanConnection, message.Arg["id"], nil)
	app.warnOnError(err, "Checking if container exists")
	if (!containerExists) {
		rpcReply(app, routing, CompletionNotice{
			Complete: true,
			Status: "server not exists",
		})
		return
	}

	rpcReply(app, routing, CompletionNotice{
		Complete: false,
		Status: "waiting",
	})

	err = containers.Stop(*app.podmanConnection, serverId, nil)
	app.warnOnError(err, "Stopping server")

	response := CompletionNotice{Complete: true, Status: "ok"}
	rpcReply(app, routing, response)
}

func handleGetRunning (_ ControlMessage, routing Routing, app *AppState) {
	runningContainers, err := containers.List(*app.podmanConnection, &containers.ListOptions{
		Filters: map[string][]string{"label": {"voxpanel"}},
	})
	app.warnOnError(err, "Enumerating containers")
	strContainerIds := make(map[string]map[string]string)
	for _, container := range runningContainers {
		fmt.Println(container.ID)
		// TODO: Fill in container attributes, with owner ports etc etc
		var attrs = map[string]string {
			"mc-port": container.Labels["mc-port"],
		}
		strContainerIds[container.ID] = attrs 
	}
	rpcReply(app, routing, strContainerIds)
}

