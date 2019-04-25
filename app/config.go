package app

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// DatabaseConfig stores information about a target database
type DatabaseConfig struct {
	Name         string
	Driver       string
	Uri          string
	Default      bool
	PortForward  bool
	TunnelConfig TunnelConfig
}

// search for the config file and return an error if it doesn't exist
func MustLoadConfig(name string) {
	var suffix = filepath.Ext(name)
	name = strings.TrimSuffix(name, suffix)
	viper.AddConfigPath(".")             // search the current path
	viper.AddConfigPath("/etc/migrant/") // search in the etc path
	viper.SetConfigName(name)
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatal(err)
	}
}

// FindConfig searches for the named db and returns the config for that db if it exists.
func MustFindDBConfig(name string) DatabaseConfig {
	if name == "default!" {
		sm := viper.GetStringMap("databases")

		for k := range sm {
			if viper.GetBool("databases." + k + ".default") {
				name = k
				break
			}
		}

		if name == "default!" {
			log.Fatal("default database not found")
		}
	}

	if viper.Get("databases."+name) == nil {
		log.Fatal("database not found")
	}

	prefix := "databases." + name

	c := DatabaseConfig{
		Name:        name,
		Driver:      viper.GetString(prefix + ".driver"),
		Uri:         viper.GetString(prefix + ".uri"),
		Default:     viper.GetBool(prefix + ".default"),
		PortForward: viper.GetBool(prefix + ".port_forward"),
	}

	if c.PortForward {
		prefix = prefix + ".ssh"
		c.TunnelConfig.Username = viper.GetString(prefix + ".username")
		c.TunnelConfig.Password = viper.GetString(prefix + ".password")
		c.TunnelConfig.Identity = viper.GetString(prefix + ".identity")
		c.TunnelConfig.LocalURI = viper.GetString(prefix + ".local_uri")
		c.TunnelConfig.JumpURI = viper.GetString(prefix + ".jump_uri")
		c.TunnelConfig.RemoteURI = viper.GetString(prefix + ".remote_uri")
	}

	return c
}
