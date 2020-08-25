package device

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

// LinuxDevice A Device given by `lsblk` on Linux
type LinuxDevice struct {
	Name       string
	Ro         bool
	Mountpoint string
	Type       string
	Children   []*LinuxDevice
}

// String returns the full path to the device
//
// Effectively `/dev/Name`
func (d *LinuxDevice) String() string {
	return fmt.Sprintf("/dev/%s", d.Name)
}

// Lsblk lists devices using `lsblk` filtering out mounted, Read only, and crypt devices
//
// `TYPE` must be "part"
//
// Converts
//
// NAME                MAJ:MIN RM   SIZE RO TYPE  MOUNTPOINT
// sda                   8:0    0 232.9G  0 disk
// └─sda1                8:1    0 232.9G  0 part
// nvme0n1             259:0    0 238.5G  0 disk
// ├─nvme0n1p1         259:1    0     1G  0 part  /boot/efi
// └─nvme0n1p2         259:2    0 237.5G  0 part
//   └─root            254:0    0 237.5G  0 crypt
//     ├─Volumes-swap  254:1    0     8G  0 lvm   [SWAP]
//     ├─Volumes-nixos 254:2    0    60G  0 lvm   /
//     └─Volumes-home  254:3    0 169.5G  0 lvm   /home
//
//
// To
//
// [ sda1 ]
func Lsblk() ([]*LinuxDevice, error) {
	out, err := exec.Command("lsblk", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run `lsblk`: %w", err)
	}

	type lsblkOutput struct {
		Blockdevices []*LinuxDevice
	}

	var output lsblkOutput

	err = json.Unmarshal(out, &output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output of `lsblk --json`: %w", err)
	}

	var unmountedDevices []*LinuxDevice
	for _, device := range output.Blockdevices {
		for _, child := range device.Children {
			if child.Mountpoint == "" && !child.Ro && child.Type == "part" && child.Children == nil {
				unmountedDevices = append(unmountedDevices, child)
			}
		}
	}
	return unmountedDevices, nil
}

// Mount execs the "mount" command on Linux, differing from golang.org/x/sys/unix because
// mount (as seen in `$ man mount 8`) guesses the filesystem and other things about the device.
func (d *LinuxDevice) Mount() error {
	mountPath := d.MountpointPath()
	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return fmt.Errorf("failed to mount - create directory failed: %w", err)
	}
	return exec.Command("mount", d.String(), d.MountpointPath()).Run()
}

// Unmount uses the golang.org/x/sys/unix Unmount in order to unmount and cleanup
func (d *LinuxDevice) Unmount() error {
	mountPath := d.MountpointPath()

	for {
		err := unix.Unmount(mountPath, 0)
		if err == nil {
			break
		}

		if !strings.HasSuffix(err.Error(), "device or resource busy") {
			return fmt.Errorf("failed to unmount %s: %v", mountPath, err)
		}
	}
	return os.Remove(mountPath)
}

// MountpointPath mounts a device to `/tmp/` on Unix of the name of
//
// `/tmp/dev-Name`
func (d *LinuxDevice) MountpointPath() string {
	return fmt.Sprintf("/tmp/dev-%s", d.Name)
}
