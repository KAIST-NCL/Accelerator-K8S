package main

import (
	"os"
	"syscall"
	"github.com/fsnotify/fsnotify"
	k8sPluginApi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"log"
)

func main(){
	manager,err := initializeAccManager()
	if err != nil {
		log.Println("Cannot initialize manager")
		os.Exit(1)
	}
	defer manager.Shutdown()

	go manager.HealthCheck()

	restart := true
	var devicePlugins []*AccDevicePlugin
	for {
		//Restart all device plugin servers
		if restart {
			if devicePlugins != nil {
				for _,dp := range devicePlugins{
					dp.StopServer()
				}
			}
			devicePlugins = initializeAccDevicePlugins(manager)
			for _,dp := range devicePlugins{
				if err := dp.Serve(); err != nil{
					log.Println("Cannot serve device plugin servers")
				}else{
					restart = false
				}
			}
		}

		select {
		case <-manager.restartChan:
			log.Println("Restarting device plugins")
			restart = true
		case e := <-manager.dpWatcher.Events:
			if e.Name == k8sPluginApi.KubeletSocket && e.Op&fsnotify.Create == fsnotify.Create{
				restart = true
			}
		case e := <-manager.dpWatcher.Errors:
			log.Println("Error watching file : "+e.Error())
			os.Exit(1)
		case s := <-manager.osWatcher :
			switch s {
			case syscall.SIGHUP:
				log.Println("Restarting device plugins")
				restart = true
			default:
				log.Println("Shutting down")
				for _,dp := range devicePlugins{
					dp.StopServer()
				}
				os.Exit(0)
			}
		}
	}
}
