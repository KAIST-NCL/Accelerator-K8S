package main

import (
	pb "./device_proto"
	"fmt"
)

func hashStr(str string) uint32 {
	h := uint32(37)
	for _,c := range str{
		h = (h * uint32(54059)) ^ (uint32(c) * uint32(76963))
	}
	return h % 86969
}

func generateDeviceId(dev *pb.Device) string{
	hash := uint32(0)
	for _,path := range dev.GetDeviceFile(){
		hash += hashStr(path)
	}
	for _,path := range dev.GetLibrary(){
		hash += hashStr(path)
	}
	return fmt.Sprint(hash)
}