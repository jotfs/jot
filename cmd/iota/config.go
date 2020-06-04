package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

const configFileEnvVar = "IOTA_CONFIG_FILE"

type config struct {
	Profile []profile
}

type profile struct {
	Name     string `toml:"name"`
	Endpoint string `toml:"endpoint"`
}

// loadConfig loads a profile from a config file. Returns an error if the config file
// could not be found, or the file does not contain the given profile.
func loadConfig(cfgName string, profileName string) (profile, error) {
	if cfgName == "" {
		cfgName = getConfigFile()
	}
	if cfgName == "" {
		return profile{}, errors.New("please specify the path to the config file")
	}

	b, err := ioutil.ReadFile(cfgName)
	if err != nil {
		return profile{}, fmt.Errorf("unable to read config: %v", err)
	}

	var cfg config
	if _, err := toml.Decode(string(b), &cfg); err != nil {
		return profile{}, fmt.Errorf("unable to parse config file %s: %v", cfgName, err)
	}

	for _, p := range cfg.Profile {
		if p.Name == profileName {
			return p, nil
		}
	}
	return profile{}, fmt.Errorf("profile %q not found in %s", profileName, cfgName)
}

// getConfigFile returns the location of the CLI config file.
func getConfigFile() string {
	if name := os.Getenv(configFileEnvVar); name != "" {
		return name
	}
	if isOneOf(runtime.GOOS, []string{"linux", "darwin", "freebsd", "openbsd", "netbsd"}) {
		home, err := homedir.Dir()
		if err != nil {
			return ""
		}
		name := filepath.Join(home, ".iota", "config.toml")
		if checkFile(name) {
			return name
		}
		return ""
	} else if runtime.GOOS == "windows" {
		// TODO: try to return default windows config dir
	}

	return ""
}

// checkFile returns true if path name exists and is not a directory.
func checkFile(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isOneOf(s string, vals []string) bool {
	for _, v := range vals {
		if v == s {
			return true
		}
	}
	return false
}
