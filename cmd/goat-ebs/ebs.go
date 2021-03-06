package main

import (
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/sevagh/goat/pkg/awsutil"
	"github.com/sevagh/goat/pkg/driveutil"
	"github.com/sevagh/goat/pkg/fsutil"
	"github.com/sevagh/goat/pkg/raidutil"
)

//GoatEbs runs Goat for your EBS volumes - attach, mount, mkfs, etc.
func GoatEbs(dryRun bool, debug bool) {
	log.Printf("WELCOME TO GOAT")
	log.Printf("1: COLLECTING EC2 INFO")
	ec2Instance := awsutil.GetEC2InstanceData()

	log.Printf("2: COLLECTING EBS INFO")
	ec2Instance.FindEbsVolumes()

	log.Printf("3: ATTACHING EBS VOLS")
	ec2Instance.AttachEbsVolumes(dryRun)

	log.Printf("4: MOUNTING ATTACHED VOLS")

	if len(ec2Instance.Vols) == 0 {
		log.Warn("Empty vols, nothing to do")
		os.Exit(0)
	}

	for volName, vols := range ec2Instance.Vols {
		prepAndMountDrives(volName, vols, dryRun)
	}
}

func prepAndMountDrives(volName string, vols []awsutil.EbsVol, dryRun bool) {
	driveLogger := log.WithFields(log.Fields{"vol_name": volName, "vols": vols})

	mountPath := vols[0].MountPath
	desiredFs := vols[0].FsType
	raidLevel := vols[0].RaidLevel

	if volName == "" {
		driveLogger.Info("No volume name given, not performing further actions")
		return
	}

	if driveutil.DoesDriveExist("/dev/disk/by-label/GOAT-" + volName) {
		driveLogger.Info("Label already exists, jumping to mount phase")
	} else {
		var driveName string
		if len(vols) == 1 {
			driveLogger.Info("Single drive, no RAID")
			driveName = vols[0].AttachedName
		} else {
			if raidLevel == -1 {
				driveLogger.Info("Raid level not provided, not performing further actions")
				return
			}
			driveLogger.Info("Creating RAID array")
			driveNames := []string{}
			for _, vol := range vols {
				driveNames = append(driveNames, vol.AttachedName)
			}
			driveName = raidutil.CreateRaidArray(driveNames, volName, raidLevel, dryRun)
		}

		if desiredFs == "" {
			driveLogger.Info("Desired filesystem not provided, not performing further actions")
			return
		}

		driveLogger.Info("Checking for existing filesystem")

		if err := fsutil.CheckFilesystem(driveName, desiredFs, volName, dryRun); err != nil {
			driveLogger.Fatalf("Checking for existing filesystem: %v", err)
		}
		if err := fsutil.CreateFilesystem(driveName, desiredFs, volName, dryRun); err != nil {
			driveLogger.Fatalf("Error when creating filesystem: %v", err)
		}
	}

	if mountPath == "" {
		driveLogger.Info("Mount point not provided, not performing further actions")
		return
	}

	driveLogger.Info("Checking if something already mounted at %s", mountPath)
	if isMounted, err := fsutil.IsMountpointAlreadyMounted(mountPath); err != nil {
		driveLogger.Fatalf("Error when checking mount point for existing mounts: %v", err)
	} else {
		if isMounted {
			driveLogger.Fatalf("Something already mounted at %s", mountPath)
		}
	}

	if !dryRun {
		if err := os.MkdirAll(mountPath, 0777); err != nil {
			driveLogger.Fatalf("Couldn't mkdir: %v", err)
		}
	}

	driveLogger.Info("Appending fstab entry")
	if err := fsutil.AppendToFstab("GOAT-"+volName, desiredFs, mountPath, dryRun); err != nil {
		driveLogger.Fatalf("Couldn't append to fstab: %v", err)
	}

	driveLogger.Info("Now mounting")
	if err := fsutil.Mount(mountPath, dryRun); err != nil {
		driveLogger.Fatalf("Couldn't mount: %v", err)
	}

	driveLogger.Info("Now persisting mdadm conf")
	if err := raidutil.PersistMdadm(); err != nil {
		driveLogger.Fatalf("Couldn't persist mdadm conf: %v", err)
	}
}
