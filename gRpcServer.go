package main

import (
	"path"
	"google.golang.org/grpc"
	k8sPluginApi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	pb "./device_proto"
	"os"
	"net"
	"golang.org/x/net/context"
	"time"
	"log"
	"strings"
	"fmt"
)

func dial(sockPath string, timeout time.Duration) (*grpc.ClientConn, error){
	ctx, cancel := context.WithTimeout(context.Background(),timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx,sockPath,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration)(net.Conn,error){
		return net.DialTimeout("unix",addr,timeout)
	},))
	if err != nil {
		return nil,err
	}
	return conn,nil
}

func convertDeviceVar(dType string, devices []*pb.Device) []*k8sPluginApi.Device {
	devs := []*k8sPluginApi.Device{}
	if devices == nil {
		return nil
	}
	for i, elem := range devices{
		statusTmp := k8sPluginApi.Healthy
		if *elem.Status == pb.Device_USED{
			statusTmp = k8sPluginApi.Unhealthy
		}
		dev := &k8sPluginApi.Device{
			ID: dType+"_"+string(i),
			Health: statusTmp,
		}
		devs = append(devs,dev)
	}
	return devs
}

type FpgaDevicePlugin struct {
	resName string
	devs []*k8sPluginApi.Device
	socket string

	stop   chan interface{}
	health chan []*k8sPluginApi.Device

	server *grpc.Server
}

func getSocketAddr(dType string) string {
	return path.Join(k8sPluginApi.DevicePluginPath,dType)+".sock"
}

func initializeFpgaDevicePlugins(m *FpgaManager) []*FpgaDevicePlugin {
	plugins := []*FpgaDevicePlugin{}
	devices := m.getDevices()
	for key,val := range devices{
		plugins = append(plugins,initializeFpgaDevicePlugin(key,val))
	}

	return plugins
}

func initializeFpgaDevicePlugin(dType string, devices []*pb.Device) *FpgaDevicePlugin {
	devs := convertDeviceVar(dType,devices)
	return &FpgaDevicePlugin{
		resName: "fpga.k8s/"+strings.Trim(strings.ToLower(dType)," "),
		devs:	devs,
		socket:	getSocketAddr(dType),

		stop:	make(chan interface{}),
		health:	make(chan []*k8sPluginApi.Device),
	}
}

func (dp *FpgaDevicePlugin) Serve() error {
	if err := dp.StartServer(); err != nil {
		log.Println("Cannot start gRPC server : "+err.Error())
		return err
	}

	if err := dp.Register(); err != nil {
		dp.StopServer()
		log.Println("Cannot register gRPC server to kubelet : "+err.Error())
		return err
	}
	return nil
}

func (dp *FpgaDevicePlugin) StartServer() error {
	if err := dp.Cleanup(); err != nil {
		log.Println("Cannot cleanup server socket : "+err.Error())
		return err
	}
	sock, err := net.Listen("unix",dp.socket)
	if err != nil {
		log.Println("Cannot create socket : "+err.Error())
		return err
	}
	dp.server = grpc.NewServer([]grpc.ServerOption{}...)
	k8sPluginApi.RegisterDevicePluginServer(dp.server,dp)

	go dp.server.Serve(sock)
	conn, err := dial(dp.socket,5*time.Second)
	if err != nil {
		dp.Cleanup()
		log.Println("Cannot connect to created socket : "+err.Error())
		return err
	}
	conn.Close()

	return nil
}

func (dp *FpgaDevicePlugin) StopServer() error {
	if dp.server != nil {
		dp.server.Stop()
		dp.server = nil
		close(dp.stop)
	}
	return dp.Cleanup()
}

func (dp *FpgaDevicePlugin) Cleanup() error {
	if err := os.Remove(dp.socket); err != nil && !os.IsNotExist(err){
		return err
	}
	return nil
}

func (dp *FpgaDevicePlugin) Register() error {
	conn, err := dial(k8sPluginApi.KubeletSocket, 5*time.Second)
	if err != nil {
		log.Println("Cannot dial kubelet socket : "+err.Error())
		return err
	}
	defer conn.Close()

	client := k8sPluginApi.NewRegistrationClient(conn)
	log.Println("Trying to register resource ["+dp.resName+"]")
	_, err = client.Register(context.Background(),&k8sPluginApi.RegisterRequest{
		Version:		k8sPluginApi.Version,
		Endpoint:		path.Base(dp.socket),
		ResourceName:	dp.resName,
	})
	if err != nil {
		log.Println("Cannot register gRPC server to kubelet : "+err.Error())
		return err
	}
	log.Println("Done register resource ["+dp.resName+"]")
	return nil
}

/*
	Kubernetes Device Plugin Server Implementation
*/
func (dp *FpgaDevicePlugin) GetDevicePluginOptions(c context.Context, e *k8sPluginApi.Empty) (*k8sPluginApi.DevicePluginOptions, error){
	return &k8sPluginApi.DevicePluginOptions{},nil
}

func (dp *FpgaDevicePlugin) ListAndWatch(e *k8sPluginApi.Empty, s k8sPluginApi.DevicePlugin_ListAndWatchServer) error{
	log.Println("List and Watch called for ["+dp.resName+"]")
	s.Send(&k8sPluginApi.ListAndWatchResponse{Devices: dp.devs})
	for {
		select{
		case <-dp.stop:
			return nil
		case <-dp.health:
			s.Send(&k8sPluginApi.ListAndWatchResponse{Devices: dp.devs})
		}
	}
	return nil
}

func (dp *FpgaDevicePlugin) Allocate(c context.Context, reqs *k8sPluginApi.AllocateRequest) (*k8sPluginApi.AllocateResponse,error){
	devs := dp.devs
	resps := k8sPluginApi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		resp := k8sPluginApi.ContainerAllocateResponse{
			Envs: map[string]string{
				"ACC_VISIBLE_DEVICES": strings.Join(req.DevicesIDs,","),
			},
		}

		for _,id := range req.DevicesIDs {
			res := false
			for _,dev := range devs {
				if dev.ID == id {
					res = true
					break
				}
			}
			if !res {
				return nil, fmt.Errorf("unknown device : %s",id)
			}
		}
		resps.ContainerResponses = append(resps.ContainerResponses,&resp)
	}
	return &resps,nil
}

func (dp *FpgaDevicePlugin) PreStartContainer(c context.Context, req *k8sPluginApi.PreStartContainerRequest) (*k8sPluginApi.PreStartContainerResponse, error){
	return &k8sPluginApi.PreStartContainerResponse{},nil
}