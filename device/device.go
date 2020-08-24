package device

// Device: Usb interface requires that a Device on a platform can be
// used as a backup device.
type Device interface {
	// String: gets the full path to the device
	//
	// on Linux:
	// /dev/sda1
	String() string

	// MountpointPath: gets the path to mount the device
	//
	// on Linux:
	// /tmp/dev-sda1
	MountpointPath() string

	// Mounts: the device to the system for read and write creating a
	// directory to mount it to.
	Mount() error

	// Unmounts: the device from the system and cleans up any temporary
	// directories. Unmount should keep attempting to unmount while
	// "device or resource is busy".
	Unmount() error
}
