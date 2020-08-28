package network

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libs4go/errors"
	kcp "github.com/libs4go/libp2p-kcp"
	"github.com/libs4go/scf4go"
	"github.com/libs4go/smf4go"
)

var keydbKey = []byte("libp2p.key")

type libp2pNode struct {
	host host.Host
}

// New .
func New(config scf4go.Config) (smf4go.Service, error) {

	libp2pKeyConfig := config.Get("libp2p", "key").String("")

	var privateKey crypto.PrivKey

	if libp2pKeyConfig == "" {
		var err error
		privateKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, 2048)

		if err != nil {
			return nil, errors.Wrap(err, "create libp2p key error")
		}
	} else {
		buff, err := crypto.ConfigDecodeKey(libp2pKeyConfig)

		if err != nil {
			return nil, errors.Wrap(err, "load libp2p key error")
		}

		privateKey, err = crypto.UnmarshalPrivateKey(buff)
	}

	kcp, err := kcp.New(privateKey, kcp.WithTLS())

	if err != nil {
		return nil, err
	}

	var addrs []string

	err = config.Get("libp2p", "listen").Scan(&addrs)

	if err != nil {
		return nil, errors.Wrap(err, "load config libp2p.listen error")
	}

	if len(addrs) == 0 {
		addrs = []string{
			"/ip4/127.0.0.1/udp/1902/kcp",
			"/ip6/::1/udp/1902/kcp",
		}
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(addrs...),
		libp2p.Identity(privateKey),
		libp2p.DisableRelay(),
		libp2p.Transport(kcp),
	}

	host, err := libp2p.New(context.Background(), opts...)

	return &libp2pNode{
		host: host,
	}, nil
}
