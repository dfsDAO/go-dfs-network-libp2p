package mock

import (
	"context"
	"fmt"
	"strings"
	"time"

	network "github.com/dfsdao/go-dfs-network-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libs4go/errors"
	_ "github.com/libs4go/ipfslog-slf4go" //
	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/memory"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console" //
)

// ScopeOfAPIError .
const errVendor = "dfs-network-libp2p-mock"

// errors
var (
	ErrInternal = errors.New("the internal error", errors.WithVendor(errVendor), errors.WithCode(-1))
	ErrParams   = errors.New("params error", errors.WithVendor(errVendor), errors.WithCode(-2))
)

const bootstrapConfigJSON = `
{
	"libp2p":{
		"listen":[
			"/ip4/0.0.0.0/udp/{{port}}/kcp"
		]
	}
}
`

const slf4goConfigJSON = `
{
	"default":{
		"backend":"console",
		"level":"debug"
	},
	"logger":{
		"addrutil": {
			"backend":"null",
			"level":"debug"
		},
		"swarm2": {
			"backend":"null",
			"level":"debug"
		},
		"net/identify": {
			"backend":"null",
			"level":"debug"
		},
		"nat":{
			"backend":"null",
			"level":"debug"
		}
	},
	"backend":{
		"console":{
			"formatter":{
				"timestamp":"15:04:05",
				"output":"@t @l @s @m"
			}
		}
	}
}
`

var log = slf4go.Get("dfs-network-libp2p-mock")

// Sharding the dfs network sharding
type Sharding interface {
	Nodes() []network.Node
}

// dfsSharding .
type dfsSharding struct {
	nodes []network.Node
}

func (sharding *dfsSharding) Nodes() []network.Node {
	return sharding.nodes
}

func setupLogger() error {
	config := scf4go.New()

	err := config.Load(memory.New(memory.Data(slf4goConfigJSON, "json")))

	if err != nil {
		return err
	}

	err = slf4go.Config(config)

	return err
}

func init() {
	err := setupLogger()

	if err != nil {
		panic(err)
	}
}

func setupBootstrapConfig(index int) (scf4go.Config, error) {
	config := scf4go.New()

	jsondata := strings.ReplaceAll(bootstrapConfigJSON, "{{port}}", fmt.Sprintf("%d", index+4000))

	err := config.Load(memory.New(memory.Data(jsondata, "json")))

	if err != nil {
		return nil, err
	}

	return config, nil
}

// SetupBootstrapSharding .
func SetupBootstrapSharding(name string, size int) (Sharding, error) {

	if size < 2 {
		return nil, errors.Wrap(ErrParams, "size must > 2")
	}

	var nodes []network.Node

	for i := 0; i < size; i++ {
		log.I("create bootstrap({@name}) node {@node}", name, i)

		config, err := setupBootstrapConfig(i)

		if err != nil {
			return nil, err
		}

		node, err := network.New(config)

		if err != nil {
			return nil, err
		}

		log.I("create bootstrap({@name}) node {@node} -- success", name, i)

		nodes = append(nodes, node)
	}

	node := nodes[0]

	addrInfo := peer.AddrInfo{
		ID:    node.Host().ID(),
		Addrs: node.Host().Addrs(),
	}

	for i := 1; i < size; i++ {
		nodes[i].DHT().Host().Connect(context.Background(), addrInfo)
	}

	return &dfsSharding{
		nodes: nodes,
	}, nil
}

// LoopRoutingTable .
func LoopRoutingTable(node network.Node, duration time.Duration) {

	ticker := time.NewTicker(duration)

	defer ticker.Stop()

	for range ticker.C {

		log.I("{@id} list dht peer", node.Host().ID().Pretty())
		for _, peer := range node.DHT().RoutingTable().ListPeers() {
			log.I("find peer {@peer}", peer.Pretty())
		}

		log.I("{@id} list dht peer -- complete", node.Host().ID().Pretty())
	}
}
