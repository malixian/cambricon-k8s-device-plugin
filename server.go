// Copyright (c) 2017, NVIDIA CORPORATION. All rights reserved.

package main

 import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	 "strconv"
 )

const (
	resourceName           = "cambricon.com/mlu-core"
	serverSock             = pluginapi.DevicePluginPath + "cambricon.sock"
	envDisableHealthChecks = "DP_DISABLE_HEALTHCHECKS"
	allHealthChecks        = "xids"
	maxCardNum             = 4
)
var cardNames = [...]string {"/dev/cambricon_c10Dev0", "/dev/cambricon_c10Dev1", "/dev/cambricon_c10Dev2", "/dev/cambricon_c10Dev3"}

// CambriconDevicePlugin implements the Kubernetes device plugin API
type CambriconDevicePlugin struct {
	devs   []*pluginapi.Device
	socket string

	stop   chan interface{}
	health chan *pluginapi.Device

	server *grpc.Server
}

// NewCambriconDevicePlugin returns an initialized CambriconDevicePlugin
func NewCambriconDevicePlugin() *CambriconDevicePlugin {
	return &CambriconDevicePlugin{
		devs:   getDevices(),
		socket: serverSock,

		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}

func (m *CambriconDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the gRPC server of the device plugin
func (m *CambriconDevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

	// Wait for server by launching a blocking connection
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	go m.healthcheck()

	return nil
}

// Stop stops the gRPC server
func (m *CambriconDevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}

	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// Register the device plugin for the given resourceName with Kubelet.
func (m *CambriconDevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (m *CambriconDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			// TODO: find solution for recover from the Unhealthy state.
			d.Health = pluginapi.Unhealthy
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
		}
	}
}

func (m *CambriconDevicePlugin) unhealthy(dev *pluginapi.Device) {
	m.health <- dev
}

// Allocate which return list of devices.
func (m *CambriconDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := m.devs
	hostPath := "/dev/cambricon_c10Dev0"
	containerPath := "/dev/cambricon_c10Dev0"
	responses := pluginapi.AllocateResponse{}
	devices := make([]*pluginapi.DeviceSpec, 0)
	for _, req := range reqs.ContainerRequests {
		log.Printf("deviceIDs is %v", req.DevicesIDs)
		for _, id := range req.DevicesIDs {
			if !deviceExists(devs, id) {
				return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
			}
		}

        if len(strings.Split(req.DevicesIDs[0], "-")) > 1 {
        	i, err := strconv.Atoi(strings.Split(req.DevicesIDs[0], "-")[2])
        	if err == nil {
				hostPath = cardNames[i]
				containerPath = hostPath
				log.Printf("host path is %s", hostPath)
			}
		}

		devices = append(devices, &pluginapi.DeviceSpec{
                                HostPath:       hostPath,
                                ContainerPath:  containerPath,
                                Permissions:    "rw",
                        })
		response := pluginapi.ContainerAllocateResponse{
			Devices: devices,
		}

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
		}
	return &responses, nil
}


func (m *CambriconDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (m *CambriconDevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (m *CambriconDevicePlugin) healthcheck() {
	disableHealthChecks := strings.ToLower(os.Getenv(envDisableHealthChecks))
	if disableHealthChecks == "all" {
		disableHealthChecks = allHealthChecks
	}

//	ctx, cancel := context.WithCancel(context.Background())

	var xids chan *pluginapi.Device
//	if !strings.Contains(disableHealthChecks, "xids") {
//		xids = make(chan *pluginapi.Device)
//		go watchXIDs(ctx, m.devs, xids)
//	}

	for {
		select {
		case <-m.stop:
//			cancel()
			return
		case dev := <-xids:
			m.unhealthy(dev)
		}
	}
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *CambriconDevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		log.Printf("Could not start device plugin: %s", err)
		return err
	}
	log.Println("Starting to serve on", m.socket)

	err = m.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		log.Printf("Could not register device plugin: %s", err)
		m.Stop()
		return err
	}
	log.Println("Registered device plugin with Kubelet")

	return nil
}
