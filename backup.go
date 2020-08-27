package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NickHackman/gbackup/device"
	"github.com/cheggaaa/pb/v3"
	"github.com/jhoonb/archivex"
	"github.com/logrusorgru/aurora"
	"github.com/otiai10/copy"
)

// backupDevice backup a device
func backupDevice(device device.Device) error {
	if err := device.Mount(); err != nil {
		return fmt.Errorf("failed to mount: %v", err)
	}

	config, err := parseConfig(device)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	if err := backup(device, config); err != nil {
		return fmt.Errorf("failed to backup: %v", err)
	}

	fmt.Printf("%s copied files\n", aurora.Green("✔"))

	if config.Zip {
		target := filepath.Join(device.MountpointPath(), config.Name)
		if err = zip(target); err != nil {
			return fmt.Errorf("failed to zip: %v", err)
		}

		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("failed to remove %s: %v", target, err)
		}

		fmt.Printf("%s zipped %s\n", aurora.Green("✔"), target)
	}

	if err := device.Unmount(); err != nil {
		return fmt.Errorf("failed to unmount: %v", err)
	}

	fmt.Printf("%s unmounted %s\n", aurora.Green("✔"), device.MountpointPath())

	return nil
}

// skip to permit skipped with github.com/otiai10/copy.
func skip(files []string) func(src string) (bool, error) {
	return func(src string) (bool, error) {
		for _, file := range files {
			if file == src {
				return true, nil
			}
		}
		return false, nil
	}
}

// backup copies all sources declared in `.gbackup.yml` to a new directory given by
// `.gbackup.yml`'s `name` field rendering a progress bar.
func backup(device device.Device, config *config) error {
	dir := filepath.Join(device.MountpointPath(), config.Name)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	errors := make(chan error)
	progressBar := pb.StartNew(len(config.Backups))

	for _, backup := range config.Backups {
		go func(backup *backupEntity) {
			src := backup.Source
			dest := filepath.Join(dir, backup.Destination)

			if err := copy.Copy(src, dest, copy.Options{Skip: skip(backup.Skip)}); err != nil {
				errors <- fmt.Errorf("failed to copy directory %s to %s: %w", src, dest, err)
			}

			errors <- nil
		}(backup)
	}

	for range config.Backups {
		err := <-errors
		if err != nil {
			fmt.Printf("Failed to copy backup: %v\n", err)
		}
		progressBar.Increment()
	}
	progressBar.Finish()

	return nil
}

// zip zips the directory made by `backup`
func zip(target string) error {
	zip := archivex.ZipFile{}
	defer zip.Close()

	if err := zip.Create(target); err != nil {
		return fmt.Errorf("failed to create zip at %s: %v", target, err)
	}

	if _, err := os.Stat(zip.Name); err != nil {
		return fmt.Errorf("failed to stat %s: %v", zip.Name, err)
	}

	if err := zip.AddAll(target, true); err != nil {
		return fmt.Errorf("failed to zip %s: %v", target, err)
	}

	return nil
}
