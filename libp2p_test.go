package network

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/libs4go/ipfslog-slf4go" //
	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/file"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console" //
	"github.com/libs4go/smf4go"
	"github.com/stretchr/testify/require"
)

func init() {
	config := scf4go.New()

	err := config.Load(file.New(file.Yaml("./libp2p_test_slf4go.yaml")))

	if err != nil {
		panic(err)
	}

	err = slf4go.Config(config)

	if err != nil {
		panic(err)
	}
}

func createNode(id int) (smf4go.Runnable, error) {
	config := scf4go.New()
	err := config.Load(file.New(file.Yaml(fmt.Sprintf("./libp2p_test_%d.yaml", id))))

	if err != nil {
		return nil, err
	}

	return New(config)

	// if err != nil {
	// 	return nil, err
	// }

	// return node, node.Start()
}

func TestCreate(t *testing.T) {
	n1, err := createNode(1)

	require.NoError(t, err)

	err = n1.Start()

	require.NoError(t, err)

	n2, err := createNode(2)

	require.NoError(t, err)

	err = n2.Start()

	require.NoError(t, err)

	time.Sleep(time.Hour)
}
