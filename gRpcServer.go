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

func convertDeviceVar(devices []*pb.Device) []*k8sPluginApi.Device {
	devs := []*k8sPluginApi.Device{}
	if devices == nil {
		return nil
	}
	for _, dev := range devices{
		statusTmp := k8sPluginApi.Healthy
		if *dev.Status != pb.Device_IDLE{
			statusTmp = k8sPluginApi.Unhealthy
		}
		dev := &k8sPluginApi.Device{
			ID: dev.GetId(),
			Health: statusTmp,
		}
		devs = append(devs,dev)
	}
	return devs
}

type AccDevicePlugin struct {
	resName string
	resNameModified string
	devs []*k8sPluginApi.Device
	socket string

	stop   chan interface{}
	health chan []*k8sPluginApi.Device

	server *grpc.Server
}

func getSocketAddr(devType string) string {
	return path.Join(k8sPluginApi.DevicePluginPath,strings.Replace(devType,"/","_",-1))+".sock"
}

func initializeAccDevicePlugins(m *AccManager) []*AccDevicePlugin {
	plugins := []*AccDevicePlugin{}
	accs := m.getAccelerators()
	for _, acc := range accs{
		plugins = append(plugins,initializeAccDevicePlugin(acc))
	}

	return plugins
}

func initializeAccDevicePlugin(acc *pb.Accelerator) *AccDevicePlugin {
	devs := convertDeviceVar(acc.Devices)
	return &AccDevicePlugin{
		resName: acc.GetType(),
		resNameModified: modifyResName(acc.GetType()),
		devs:	devs,
		socket:	getSocketAddr(acc.GetType()),

		stop:	make(chan interface{}),
		health:	make(chan []*k8sPluginApi.Device),
	}
}

func (dp *AccDevicePlugin) Serve() error {
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

func (dp *AccDevicePlugin) StartServer() error {
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

func (dp *AccDevicePlugin) StopServer() error {
	if dp.server != nil {
		dp.server.Stop()
		dp.server = nil
		close(dp.stop)
	}
	return dp.Cleanup()
}

func (dp *AccDevicePlugin) Cleanup() error {
	if err := os.Remove(dp.socket); err != nil && !os.IsNotExist(err){
		return err
	}
	return nil
}

func (dp *AccDevicePlugin) Register() error {
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
func (dp *AccDevicePlugin) GetDevicePluginOptions(c context.Context, e *k8sPluginApi.Empty) (*k8sPluginApi.DevicePluginOptions, error){
	return &k8sPluginApi.DevicePluginOptions{},nil
}

func (dp *AccDevicePlugin) ListAndWatch(e *k8sPluginApi.Empty, s k8sPluginApi.DevicePlugin_ListAndWatchServer) error{
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

func (dp *AccDevicePlugin) Allocate(c context.Context, reqs *k8sPluginApi.AllocateRequest) (*k8sPluginApi.AllocateResponse,error){
	devs := dp.devs
	resps := new(k8sPluginApi.AllocateResponse)
	for _, req := range reqs.ContainerRequests {
		log.Println("Allocate ["+strings.Join(req.DevicesIDs,",")+"]")
		resp := new(k8sPluginApi.ContainerAllocateResponse)
		resp.Envs = make(map[string]string)
		resp.Envs["ACC_VISIBLE_DEVICES"] = strings.Join(req.DevicesIDs,",")

		resp.Envs["ACC_VISIBLE_DEVICES_+"+dp.resNameModified] = strings.Join(req.DevicesIDs,",")

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
		resps.ContainerResponses = append(resps.ContainerResponses,resp)
	}
	return resps,nil
}

func (dp *AccDevicePlugin) PreStartContainer(c context.Context, req *k8sPluginApi.PreStartContainerRequest) (*k8sPluginApi.PreStartContainerResponse, error){
	return &k8sPluginApi.PreStartContainerResponse{},nil
}
