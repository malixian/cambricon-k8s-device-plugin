
package main

import (
	"log"
	"strings"
	"crypto/rand"
	"encoding/hex"
	cndev "cambricon-k8s-device-plugin/device_query"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"strconv"
	"fmt"
	"math"
)

const (
        deviceGlob = "/dev/cambricon*"
	deviceIdPrefix = "cambricon-mlu-"
)

func check(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func randomId() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return deviceIdPrefix + hex.EncodeToString(bytes), nil
}

func enrichDevices(devs []string) map[string]string {
	devMaps := make(map[string]string)
	for _, v := range devs {
		if id, err := randomId(); err != nil {
			log.Panicln("Error when generate id for", v)
			break
		} else {
			devMaps[id] = v
		}
	}
	return devMaps
}

func cleanDevices(names []string) []string {
	devs := make([]string, 0)
	for _, v := range names {
		nv := strings.TrimSpace(v)
		if len(nv) > 0 {
			devs = append(devs, nv)
		}
	}
	return devs
}

func setDeviceID(cardID int, coreID int) string {
	return deviceIdPrefix + strconv.Itoa(cardID) + "-" + strconv.Itoa(coreID)
}

func getDevices() ([]*pluginapi.Device) {
	/*
	Only send cambricon card number
	out, err := exec.Command("sh", "-c", fmt.Sprintf("ls %s", deviceGlob)).Output()
	check(err)
	names := strings.Split(string(out), "\n")
	devices := cleanDevices(names)
	log.Println("The output is ", devices)
	log.Println("The length of out is ", len(devices))
	devMaps := enrichDevices(devices)
	log.Println("The dev map is", devMaps)

	var devs []*pluginapi.Device
	for id, _ := range devMaps {
		devs = append(devs, &pluginapi.Device{
			ID:     id,
			Health: pluginapi.Healthy,
		})
	}

	return devs, devMaps
	*/

	// send cambricon core number
	cardCount := cndev.GetCardCount()
	log.Printf("cambricon card count is %d", cardCount)
	var devs []*pluginapi.Device
	for i := 0; i < cardCount; i++ {
		// GetUnuseCoreCount method is too slow
		//unuseCore := cndev.GetUnuseCoreCount(i)
		deviceUtil := cndev.GetDeviceUtil(i)
		unuseCore := int(math.Floor((1 - float64(deviceUtil)*1.0/100) * 32))
		log.Printf("card NO%d has unuse core is %d", i, unuseCore)
		for j := 0; j < unuseCore; j++ {
			id := setDeviceID(i, j)
			devs = append(devs, &pluginapi.Device{
				ID: id,
				Health: pluginapi.Healthy,
			})
		}
	}
	return devs

}

func GetMaxUtilCard() string {
	cardCount := cndev.GetCardCount()
	maxUtil := -1
	useCardName := ""
	for i:=0; i<cardCount; i++ {
		tmpUtil := cndev.GetDeviceUtil(i)
		// 可能与yaml算mp、dp的设备不统一
		if tmpUtil >= 90 {
			continue
		}
		if maxUtil < tmpUtil {
			maxUtil = tmpUtil
			useCardName = setCardName(i)
		}
	}
	return useCardName
}

func setCardName(cardIndex int) string {
	return fmt.Sprintf("/dev/cambricon_c10Dev%d", cardIndex)
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		if d.ID == id {
			return true
		}
	}
	return false
}
