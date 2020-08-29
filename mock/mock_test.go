package mock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateBoostrapSharding(t *testing.T) {
	sharding, err := SetupBootstrapSharding("B1", 3)

	require.NoError(t, err)

	require.Equal(t, len(sharding.Nodes()), 3)

	// dht := sharding.Nodes()[0].DHT()

	// err = <-dht.ForceRefresh()

	require.NoError(t, err)

	LoopRoutingTable(sharding.Nodes()[0], time.Second*2)
}
