package app_test

import (
	"net/http"
	"testing"

	"bitbucket.org/fantamstick/migrant/app"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("it returns an error if not 3 uris", func(t *testing.T) {
		c := app.TunnelConfig{}
		_, err := app.NewTunnel(c)
		assert.NotNil(t, err, "should return error")
	})

	t.Run("it returns an error if could not find specified identity file", func(t *testing.T) {
		c := app.TunnelConfig{
			RemoteURI: "bogus1",
			JumpURI:   "bogus2",
			LocalURI:  "bogus3",
			Identity:  "./this-file-does-not-exist",
		}

		_, err := app.NewTunnel(c)
		assert.NotNil(t, err, "should return error")
	})

	t.Run("it returns an error if no auth method supplied", func(t *testing.T) {
		c := app.TunnelConfig{
			RemoteURI: "bogus1",
			JumpURI:   "bogus2",
			LocalURI:  "bogus3",
		}
		_, err := app.NewTunnel(c)
		assert.NotNil(t, err, "should return error")
	})

	t.Run("it returns an error if known hosts file not found", func(t *testing.T) {
		c := app.TunnelConfig{
			RemoteURI:  "bogus1",
			JumpURI:    "bogus2",
			LocalURI:   "bogus3",
			Password:   "secret",
			KnownHosts: "./this-file-does-not-exist",
		}
		_, err := app.NewTunnel(c)
		assert.NotNil(t, err, "should return error")
	})
}

func TestTunnel(t *testing.T) {
	t.Run("it creates tunnel through specified host", func(t *testing.T) {
		sshTunnel, _ := app.NewTunnel(app.TunnelConfig{
			LocalURI:   "localhost:9876",
			JumpURI:    "127.0.0.1:8022",
			RemoteURI:  "remote:1234",
			Username:   "root",
			Identity:   "../docker/bastion/id_docker_bastion",
			KnownHosts: "../fixtures/ssh/known_hosts",
		})

		ready := make(chan bool)
		go sshTunnel.Start(ready)

		<-ready

		res, err := http.Get("http://localhost:9876")

		assert.Nil(t, err, "should not return error")
		assert.Equal(t, 200, res.StatusCode, "should return a 200")
	})
}
