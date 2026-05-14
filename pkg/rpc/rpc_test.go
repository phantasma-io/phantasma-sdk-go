package rpc_test

import (
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := rpc.NewRPCMainnet()
	assert.NotNil(t, client)
}
