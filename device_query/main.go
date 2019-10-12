package device_query

// #cgo CFLAGS: -I./device_query
// #cgo LDFLAGS: -L./device_query -ldeviceq
// #cgo LDFLAGS: -L./device_query -lcndev
// #include "device_query.h"
// #include "cndev.h"
import "C"
import "fmt"

func main() {
    count := C.getCardCount()
    fmt.Printf("%d\n", count)

    deviceUtil := C.getDeviceUtil(0)
    fmt.Printf("device Util is %d%%\n", deviceUtil)

    coreCount := C.getCoreCount(0)
    fmt.Printf("card core count is %d\n", coreCount)

    coreUtilInfo := C.getCoreUtilInfo(0,0)
    fmt.Printf("Util CORE%d:%d%%\n",0, coreUtilInfo)
    
    unuseCoreCount := getUnuseCoreCount(0)
    fmt.Printf("unuse core count is %d \n", unuseCoreCount)
}

func getUnuseCoreCount(card int) int {
    ret := 0
    coreCount := int(C.getCoreCount(0))
    for i:=0; i<coreCount; i++ {
        coreUtilInfo := C.getCoreUtilInfo(C.int(card), C.int(i))
	if coreUtilInfo == 0 {
	    ret++
	}
    }
    return ret
}

// get cambricon card count
func GetCardCount() int {
    return int(C.getCardCount())
}

// get specific card util Utilization rate integrate
func GetDeviceUtil(cardId int) int {
    return int(C.getDeviceUtil(C.int(cardId)))
}

// get specific card core number
func GetCoreCount(cardId int) int {

    return int(C.getCoreCount(C.int(cardId)))
}

// get specific core on specific card Utilization rate integrate
func GetCoreUtil(cardId int, coreId int) int {

    return int(C.getCoreUtilInfo(C.int(cardId), C.int(coreId)))
}

// get the count of utilization rate integrate is 0 on specific card
func GetUnuseCoreCount(card int) int {
    return getUnuseCoreCount(card)
}