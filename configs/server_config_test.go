package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig_Validate(t *testing.T) {
	c := &ServerConfig{
		Listen:      "127.0.0.1",
		Port:        1231,
		LogFilePath: "log/server.log",
		LogOutput:   []string{OutputConsole, OutputFile},
	}
	assert.NoError(t, c.Validate())

	c.Listen = "localhost"
	assert.NoError(t, c.Validate())
	c.Listen = "github.com"
	assert.NoError(t, c.Validate())
	c.Listen = "0.0.0.0"
	assert.NoError(t, c.Validate())
	c.Listen = "192.168.1.1"
	assert.NoError(t, c.Validate())

	c.Listen = "127.0.0.1"

	c.Port = 0
	assert.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'required' tag", c.Validate().Error())
	c.Port = -1
	assert.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'min' tag", c.Validate().Error())
	c.Port = 65536
	assert.Equal(t, "Key: 'ServerConfig.Port' Error:Field validation for 'Port' failed on the 'max' tag", c.Validate().Error())
	c.Port = 65535
	assert.NoError(t, c.Validate())

	c.LogOutput = []string{}
	assert.EqualError(t, c.Validate(), "Key: 'ServerConfig.LogOutput' Error:Field validation for 'LogOutput' failed on the 'min' tag")
	c.LogOutput = []string{OutputConsole, "aaa"}
	assert.EqualError(t, c.Validate(), errInvalidOutput.Error())
	c.LogOutput = []string{OutputFile, OutputConsole}
	c.LogFilePath = ""
	assert.EqualError(t, c.Validate(), errNonFilePath.Error())
	c.LogOutput = []string{OutputFile}
	assert.EqualError(t, c.Validate(), errNonFilePath.Error())
}
