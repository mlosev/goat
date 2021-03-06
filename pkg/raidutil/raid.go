package raidutil

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"

	"github.com/sevagh/goat/pkg/driveutil"
	"github.com/sevagh/goat/pkg/execute"
)

//CreateRaidArray runs the appropriate mdadm command for the given list of EbsVol that should be raided together. It takes dryRun as a boolean, where it tells you which mdadm it would have run
func CreateRaidArray(driveNames []string, volName string, raidLevel int, dryRun bool) string {
	raidLogger := log.WithFields(log.Fields{"vol_name": volName, "drives": driveNames})

	raidLogger.Info("Mounting raid drives")

	var raidDriveName string
	var err error
	raidLogger.Info("Searching for unused RAID drive name")
	if raidDriveName, err = driveutil.RandRaidDriveNamePicker(); err != nil {
		raidLogger.Fatalf("Couldn't select unused RAID drive name: %v", err)
	}

	cmd := "mdadm"

	nameString := "--name='GOAT-" + volName + "'"

	var args []string
	args = []string{
		"--create",
		raidDriveName,
		"--level=" + strconv.Itoa(raidLevel),
		nameString,
		"--raid-devices=" + strconv.Itoa(len(driveNames)),
	}

	args = append(args, driveNames...)
	raidLogger.Infof("RAID: Creating RAID drive: %s %s", cmd, args)
	if dryRun {
		return raidDriveName
	}
	if _, err := execute.Command(cmd, args); err != nil {
		raidLogger.Fatalf("Error when executing mdadm command: %v", err)
	}

	return raidDriveName
}

//PersistMdadm dumps the current mdadm config to /etc/mdadm.conf
func PersistMdadm() error {
	cmd := "mdadm"

	args := []string{
		"--verbose",
		"--detail",
		"--scan",
	}

	log.Infof("Persisting mdadm settings: %s %s", cmd, args)

	var out execute.CommandOut
	var err error
	if out, err = execute.Command(cmd, args); err != nil {
		log.Fatalf("Error when executing mdadm command: %v", err)
	}

	f, err := os.OpenFile("/etc/mdadm.conf", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(out.Stdout); err != nil {
		return err
	}
	return nil
}
