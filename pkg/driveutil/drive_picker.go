package driveutil

import (
	"fmt"
	"strconv"
)

//RandDriveNamePicker returns a /dev/xvd[b-z] string, whichever is the first that doesn't exist
func RandDriveNamePicker() (string, error) {
	ctr := 0
	deviceName := "/dev/xvd"
	runes := []rune("bcdefghijklmnopqrstuvwxyz")
	for {
		if ctr >= len(runes) {
			return "", fmt.Errorf("Ran out of drive names")
		}
		if !DoesDriveExist(deviceName + string(runes[ctr])) {
			break
		}
		ctr++
	}
	return deviceName + string(runes[ctr]), nil
}

//RandRaidDriveNamePicker returns a /dev/md[0-9] string, whichever is the first that doesn't exist
func RandRaidDriveNamePicker() (string, error) {
	ctr := 0
	deviceName := "/dev/md"
	maxDevices := 99
	for {
		if ctr >= maxDevices {
			return "", fmt.Errorf("Ran out of raid drive names")
		}
		if !DoesDriveExist(deviceName + strconv.FormatInt(int64(ctr), 10)) {
			break
		}
		ctr++
	}
	return deviceName + strconv.FormatInt(int64(ctr), 10), nil
}
