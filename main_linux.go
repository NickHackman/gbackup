package main

import (
	"fmt"
	"os"

	"github.com/NickHackman/gbackup/device"
	"github.com/manifoldco/promptui"
)

func main() {
	devices, err := device.Lsblk()
	if err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}

	prompt := promptui.Select{
		Label: "Pick a unmounted device to backup to",
		Items: devices,
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}

	if err = backupDevice(devices[index]); err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}
}
