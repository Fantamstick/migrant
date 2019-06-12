package app

import (
	"fmt"
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
	User         string
	Pass         string
	Port         string
	Host         string
	Prms         string
	Default      bool
	PortForward  bool
	TunnelConfig TunnelConfig
}

// TunnelConfig specifies parameters for a Tunnel.
type TunnelConfig struct {
	LocalURI                string
	LocalHost               string
	LocalPort               string
	JumpURI                 string
	JumpHost                string
	JumpPort                string
	RemoteURI               string
	RemoteHost              string
	RemotePort              string
	Username                string
	Password                string
	Identity                string
	KnownHosts              string
	InsecureHostKeyChecking bool
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
		User:        viper.GetString(prefix + ".user"),
		Pass:        viper.GetString(prefix + ".pass"),
		Host:        viper.GetString(prefix + ".host"),
		Port:        viper.GetString(prefix + ".port"),
		Prms:        viper.GetString(prefix + ".prms"),
		Default:     viper.GetBool(prefix + ".default"),
		PortForward: viper.GetBool(prefix + ".port_forward"),
	}

	if c.PortForward {
		prefix = prefix + ".ssh"
		c.TunnelConfig.Username = viper.GetString(prefix + ".username")
		c.TunnelConfig.Password = viper.GetString(prefix + ".password")
		c.TunnelConfig.Identity = viper.GetString(prefix + ".identity")
		c.TunnelConfig.LocalURI = viper.GetString(prefix + ".local_uri")
		c.TunnelConfig.LocalHost = viper.GetString(prefix + ".local_host")
		c.TunnelConfig.LocalPort = viper.GetString(prefix + ".local_port")
		c.TunnelConfig.JumpURI = viper.GetString(prefix + ".jump_uri")
		c.TunnelConfig.JumpHost = viper.GetString(prefix + ".jump_host")
		c.TunnelConfig.JumpPort = viper.GetString(prefix + ".jump_port")
		c.TunnelConfig.RemoteURI = viper.GetString(prefix + ".remote_uri")
		c.TunnelConfig.RemoteHost = viper.GetString(prefix + ".remote_host")
		c.TunnelConfig.RemotePort = viper.GetString(prefix + ".remote_port")
	}

	return c
}

// resolve any secrets in the uri and compile uri components into a single uri
func resolveDatabaseUri(c *DatabaseConfig) {

	// first try to use any uri value that was injected
	if c.Uri != "" {
		c.Uri = NeedSecret(c.Uri)
		return
	}

	// otherwise try to make the uri from the components
	c.User = NeedSecret(c.User)
	c.Pass = NeedSecret(c.Pass)
	c.Host = NeedSecret(c.Host)
	c.Port = NeedSecret(c.Port)

	if c.User == "" || c.Pass == "" || c.Host == "" || c.Port == "" {
		panic("not enough components to make a db uri - you need user, pass, host, and port")
	}

	c.Uri = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Pass, c.Host, c.Port, c.Name)

	if c.Prms != "" {
		c.Uri = c.Uri + "?" + c.Prms
	}
}

// resolve any secrets in the tunnel uri and compile uri components into a single uri
func resolveTunnelURIs(c *TunnelConfig) {
	resolve := func(uri, host, port string) string {
		if uri != "" {
			return NeedSecret(uri)
		}

		host = NeedSecret(host)
		port = NeedSecret(port)

		if port == "" || host == "" {
			panic("local uri does not have enough components - need host and port")
		}

		return host + ":" + port
	}

	c.LocalURI = resolve(c.LocalURI, c.LocalHost, c.LocalPort)
	c.RemoteURI = resolve(c.RemoteURI, c.RemoteHost, c.RemotePort)
	c.JumpURI = resolve(c.JumpURI, c.JumpHost, c.JumpPort)
}
