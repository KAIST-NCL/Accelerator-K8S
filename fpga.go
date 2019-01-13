package main

import (
	"io/ioutil"
	"github.com/gogo/protobuf/proto"
	pb "./device_proto"
	k8sPluginApi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"strings"
	"os"
	"path"
	"log"
	"github.com/fsnotify/fsnotify"
	"os/signal"
	"syscall"
)

const (
	USR_LIST_DEFAULT = "/etc/accelerator-docker/device.pbtxt"
	STAT_LIST_DEFAULT = "/etc/accelerator-docker/stat.pb"
	accManagerName = "acc-manager"
)

type FpgaManager struct {
	userListPath string
	statListPath string

	restartChan chan interface{}
	dpWatcher *fsnotify.Watcher
	osWatcher chan os.Signal
	healthWatcher *fsnotify.Watcher
}

func initializeFpgaManager() (*FpgaManager,error){
	//TODO : fetch list files from ACC-Manager
	/*accManagerPath, err := exec.LookPath(accManagerName)
	if err != nil {
		log.Println("No "+accManagerName+" found\nFirst install Accelerator-Docker")
	}
	listPaths, err := exec.Command(accManagerPath,"paths").Output()
	if err != nil {
		log.Println("Cannot execute "+accManagerName)
	}
	fmt.Println(string(listPaths))*/

	tmp_user_list_path := USR_LIST_DEFAULT
	tmp_stat_list_path := STAT_LIST_DEFAULT

	//Handle change of device plugin path
	tmpDpWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Cannot initialize file watcher : "+err.Error())
		return nil, err
	}
	err = tmpDpWatcher.Add(k8sPluginApi.DevicePluginPath)
	if err != nil {
		tmpDpWatcher.Close()
		log.Println("Cannot watch ["+k8sPluginApi.DevicePluginPath+"] : "+err.Error())
		return nil, err
	}

	//Handle OS interrupts
	tmpOsWatcher := make(chan os.Signal,1)
	signal.Notify(tmpOsWatcher, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	//Health Watcher
	tmpHealthWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Cannot initialize file watcher for health check : "+err.Error())
		return nil, err
	}

	if err := tmpHealthWatcher.Add(tmp_user_list_path); err != nil {
		log.Println("Cannot watch ["+tmp_user_list_path+"] : "+err.Error())
		return nil, err
	}
	if err := tmpHealthWatcher.Add(tmp_stat_list_path); err != nil {
		log.Println("Cannot watch ["+tmp_stat_list_path+"] : "+err.Error())
		return nil, err
	}

	return &FpgaManager{
		userListPath: tmp_user_list_path,
		statListPath: tmp_stat_list_path,

		restartChan: make(chan interface{}),
		dpWatcher: tmpDpWatcher,
		osWatcher: tmpOsWatcher,
		healthWatcher: tmpHealthWatcher,
	},nil
}

func (m *FpgaManager) getDevices() map[string][]*pb.Device{
	statListPath := m.statListPath
	usrListPath := m.userListPath

	statListTxt, err := ioutil.ReadFile(statListPath)
	if err != nil {
		log.Println("cannot read ["+statListPath+"] : "+err.Error())
	}
	statList := &pb.DeviceList{}
	if err := proto.Unmarshal(statListTxt, statList); err != nil {
		log.Println("cannot parse ["+ statListPath +"] : "+err.Error())
	}

	usrListTxt, err := ioutil.ReadFile(usrListPath)
	if err != nil {
		log.Println("cannot read ["+usrListPath+"] : "+err.Error())
	}
	usrList := &pb.DeviceList{}
	if err := proto.UnmarshalText(string(usrListTxt), usrList); err != nil {
		log.Println("cannot parse ["+ usrListPath +"] : "+err.Error())
	}

	devices := make(map[string][]*pb.Device)
	for _, elem := range usrList.GetDevices(){
		name := strings.Trim(strings.ToLower(*elem.Name)," ")
		dev := proto.Clone(elem).(*pb.Device)
		for _,elemStat := range statList.GetDevices(){
			nameStat := strings.Trim(strings.ToLower(*elemStat.Name)," ")
			if nameStat == name {
				if _,err := os.Stat(path.Join("/proc",string(*elemStat.Pid))); os.IsNotExist(err) {
					tmpStatus := pb.Device_IDLE
					dev.Status = &tmpStatus
					dev.Pid = proto.Int32(0)
				}else{
					dev.Status = elemStat.Status
					dev.Pid = proto.Int32(*elemStat.Pid)
				}

				if _,ok := devices[name]; !ok {
					devices[name] = []*pb.Device{}
				}
				devices[name] = append(devices[name],dev)
				break
			}
		}
	}

	return devices
}

func (m *FpgaManager) HealthCheck() {
	//TODO: Cleanup file watcher
	//TODO: fix from per dp to overall


	/*for{
		select {
		case <-watcher.Events:
			var tmpDevs []*k8sPluginApi.Device
			devs := getDevices()
			if _,ok := devs[dp.resName]; !ok {
				log.Println("Resource ["+dp.resName+"] not exists anymore")
				dp.health <- nil
				return
			}
			tmpDevs = convertDeviceVar(dp.resName,devs[dp.resName])
			dp.devs = tmpDevs
			dp.health <- dp.devs

		case e := <-watcher.Errors:
			log.Println("Error while watching status files: "+e.Error())
		}
	}*/
}

func (m *FpgaManager) Shutdown(){
	if m.dpWatcher != nil {
		m.dpWatcher.Close()
	}
	if m.healthWatcher != nil {
		m.healthWatcher.Close()
	}
}
