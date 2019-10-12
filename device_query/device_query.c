#include <stdio.h>
#include <string.h>
#include <cndev.h>

int getCardCount() {
    // init libcndev
    cndevRet_t ret = cndevInit(0);
    if (CNDEV_SUCCESS != ret) {
	return -1;
    }

    //get card count
    cndevCardInfo_t cardInfo;
    cardInfo.version = CNDEV_VERSION_1;
    ret = cndevGetDeviceCount(&cardInfo);
    if (ret == CNDEV_ERROR_LOW_DRIVER_VERSION) {
    	printf("api version is low");
	return -1;
    }

    int cardCount = cardInfo.Number;
    cndevRelease();
    return cardCount;
}

int getDeviceUtil(int card) {
    cndevRet_t ret = cndevInit(0);
    cndevUtilizationInfo_t utilInfo;
    utilInfo.version = CNDEV_VERSION_2;
    cndevCheckErrors(cndevGetDeviceUtilizationInfo(&utilInfo, card));
    cndevRelease();
    return utilInfo.BoardUtilization;
}

int getCoreCount(int card) {
    cndevRet_t ret = cndevInit(0);
    cndevCardCoreCount_t cardCoreCount;
    cardCoreCount.version = CNDEV_VERSION_2;
    cndevCheckErrors(cndevGetCoreCount(&cardCoreCount, card));
    cndevRelease();
    return cardCoreCount.count;
}

int getCoreUtilInfo(int card, int coreId) {
    cndevRet_t ret = cndevInit(0);
    cndevUtilizationInfo_t utilInfo;
    utilInfo.version = CNDEV_VERSION_2;
    cndevCheckErrors(cndevGetDeviceUtilizationInfo(&utilInfo, card));
    cndevRelease();
    return utilInfo.CoreUtilization[coreId];
}
