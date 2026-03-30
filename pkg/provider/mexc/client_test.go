package mexc

import (
	"testing"

	"github.com/mdnmdn/bits/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestMEXCClient_ID(t *testing.T) {
	cfg := config.MEXCConfig{}
	client := NewClient(cfg)
	assert.Equal(t, "mexc", client.ID())
}

func TestMEXCClient_Capabilities(t *testing.T) {
	cfg := config.MEXCConfig{}
	client := NewClient(cfg)
	caps := client.Capabilities()
	assert.NotNil(t, caps)
	assert.True(t, len(caps) > 0)
}
