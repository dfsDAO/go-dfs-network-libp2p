package network

import (
	"testing"

	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/file"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console" //
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

func TestCreate(t *testing.T) {
	config := scf4go.New()
	err := config.Load(file.New(file.Yaml("./libp2p_test.yaml")))

	require.NoError(t, err)

	_, err = New(config)

	require.NoError(t, err)
}
