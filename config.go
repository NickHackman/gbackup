package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NickHackman/gbackup/device"
	"gopkg.in/yaml.v2"
)

const (
	configPath  = ".gbackup.yml"
	defaultName = "gbackup-2-1-2006"
)

type config struct {
	Name    string
	Backups []*backupEntity
	Zip     bool
}

type backupEntity struct {
	Source      string
	Destination string
	Skip        []string
}

func tildeExpand(source string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	if source == "~" {
		return home, nil
	}

	if strings.HasPrefix(source, "~/") {
		return filepath.Join(home, source[2:]), nil
	}

	return source, nil
}

// newConfig default configuration
func newConfig() *config {
	return &config{
		Name: defaultName,
		Zip:  true,
	}
}

// parseConfig parses the configuration file located at
//
// `$mounted_device_name/.gbackup.yml`
func parseConfig(d device.Device) (*config, error) {
	configFullPath := filepath.Join(d.MountpointPath(), configPath)
	bytes, err := ioutil.ReadFile(configFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", configFullPath, err)
	}

	conf := newConfig()
	if err = yaml.Unmarshal(bytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configFullPath, err)
	}

	conf.Name = time.Now().Format(conf.Name)

	for _, backup := range conf.Backups {
		if backup.Source, err = tildeExpand(backup.Source); err != nil {
			return nil, fmt.Errorf("failed to expand source field: %v", err)
		}
		if backup.Destination == "" {
			_, file := filepath.Split(backup.Source)
			backup.Destination = file
		}
	}

	return conf, nil
}
