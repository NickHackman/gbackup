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

func (b *backupEntity) dest() string {
	if b.Destination == "" {
		_, file := filepath.Split(b.Source)
		return file
	}
	return b.Destination
}

func (b *backupEntity) src() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	if b.Source == "~" {
		return home, nil
	}

	if strings.HasPrefix(b.Source, "~/") {
		return filepath.Join(home, b.Source[2:]), nil
	}

	return b.Source, nil
}

// newConfig: default configuration
func newConfig() *config {
	return &config{
		Name: defaultName,
		Zip:  true,
	}
}

// parseConfig: parses the configuration file located at
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

	return conf, nil
}
