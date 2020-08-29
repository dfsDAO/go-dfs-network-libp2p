package network

import (
	"context"

	dht "github.com/dfsdao/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libs4go/errors"
	kcp "github.com/libs4go/libp2p-kcp"
	"github.com/libs4go/scf4go"
	"github.com/libs4go/slf4go"
	"github.com/libs4go/smf4go"
	"github.com/multiformats/go-multiaddr"
)

// ScopeOfAPIError .
const errVendor = "dfs-network-libp2p"

// errors
var (
	ErrInternal    = errors.New("the internal error", errors.WithVendor(errVendor), errors.WithCode(-1))
	ErrDHTBoostrap = errors.New("libp2p dht boostrap addrs not found", errors.WithVendor(errVendor), errors.WithCode(-2))
)

var keydbKey = []byte("libp2p.key")

// Node .
type Node interface {
	smf4go.Runnable
	Host() host.Host
	DHT() *dht.IpfsDHT
}

type libp2pNode struct {
	slf4go.Logger
	host host.Host
	dht  *dht.IpfsDHT
}

// New .
func New(config scf4go.Config) (Node, error) {

	logger := slf4go.Get("dfs-network-libp2p")

	node := &libp2pNode{
		Logger: logger,
	}

	privateKey, err := node.privateKey(config)

	if err != nil {
		return nil, err
	}

	kcp, err := kcp.New(privateKey, kcp.WithTLS())

	if err != nil {
		return nil, err
	}

	addrs, err := node.addrs(config)

	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(addrs...),
		libp2p.Identity(privateKey),
		libp2p.NATPortMap(),
		libp2p.DefaultTransports,
		libp2p.Transport(kcp),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return node.createDHT(config, h)
		}),
		libp2p.EnableAutoRelay(),
	}

	host, err := libp2p.New(context.Background(), opts...)

	if err != nil {
		return nil, errors.Wrap(err, "create libp2p host error")
	}

	logger.I("start libp2p {@id}", host.ID().Pretty())

	node.host = host

	return node, nil
}

func (node *libp2pNode) Host() host.Host {
	return node.host
}

func (node *libp2pNode) DHT() *dht.IpfsDHT {
	return node.dht
}

func (node *libp2pNode) createDHT(config scf4go.Config, host host.Host) (*dht.IpfsDHT, error) {
	var dhtBoostrapAddrs []string

	err := config.Get("libp2p", "dht", "boostrap").Scan(&dhtBoostrapAddrs)

	if err != nil {
		return nil, errors.Wrap(err, "load config libp2p.dht.boostrap error")
	}

	var peers []peer.AddrInfo

	for _, addr := range dhtBoostrapAddrs {

		multiAddr, err := multiaddr.NewMultiaddr(addr)

		if err != nil {
			return nil, errors.Wrap(err, "load config libp2p dht boostrap addr %s error", addr)
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)

		if err != nil {
			return nil, errors.Wrap(err, "load config libp2p dht boostrap error")
		}

		peers = append(peers, *addrInfo)

		node.I("add dht bootstrap addr {@}", addrInfo.String())
	}

	// if len(peers) == 0 {
	// 	return nil, errors.Wrap(ErrDHTBoostrap, "libp2p.dht.boostrap can't be empty")
	// }

	dhOpts := []dht.Option{
		dht.BootstrapPeers(peers...),
	}

	dht, err := dht.New(context.Background(), host, dhOpts...)

	if err != nil {
		return nil, errors.Wrap(err, "create libp2p dht error")
	}

	node.dht = dht

	return dht, nil
}

func (node *libp2pNode) addrs(config scf4go.Config) ([]string, error) {

	var addrs []string

	err := config.Get("libp2p", "listen").Scan(&addrs)

	if err != nil {
		return nil, errors.Wrap(err, "load config libp2p.listen error")
	}

	if len(addrs) == 0 {
		addrs = []string{
			"/ip4/127.0.0.1/udp/1902/kcp",
			"/ip6/::1/udp/1902/kcp",
		}
	}

	return addrs, nil
}

func (node *libp2pNode) privateKey(config scf4go.Config) (crypto.PrivKey, error) {
	libp2pKeyConfig := config.Get("libp2p", "key").String("")

	var privateKey crypto.PrivKey

	if libp2pKeyConfig == "" {
		var err error
		privateKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, 2048)

		if err != nil {
			return nil, errors.Wrap(err, "create libp2p key error")
		}

		buff, err := crypto.MarshalPrivateKey(privateKey)

		if err != nil {
			return nil, errors.Wrap(err, "create libp2p key error")
		}

		libp2pKeyConfig = crypto.ConfigEncodeKey(buff)

		node.I("create new libp2p key: {@key}", libp2pKeyConfig)

	} else {
		buff, err := crypto.ConfigDecodeKey(libp2pKeyConfig)

		if err != nil {
			return nil, errors.Wrap(err, "load libp2p key error")
		}

		privateKey, err = crypto.UnmarshalPrivateKey(buff)

		if err != nil {
			return nil, errors.Wrap(err, "load libp2p key error")
		}
	}

	return privateKey, nil
}

func (node *libp2pNode) Start() error {
	return nil
}
