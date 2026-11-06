package p2p


import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)


// --- CONSTANTS ---

const (
	// keyMode is the permission of the peer key file: owner only.
	keyMode = 0o600

	// resourceMemory is the memory, in bytes, the resource manager scales its
	// limits against.
	resourceMemory = 256 << 20

	// resourceFDs is the file descriptor count the resource manager scales
	// its limits against.
	resourceFDs = 512
)


// --- GLOBALS ---

// dialTimeout bounds each connection attempt at startup, so a peer that is
// down delays the start by this much and no more.
var dialTimeout = 5 * time.Second


// DefaultBootstrapAddrs is the bootstrap list shipped with the node. Empty
// until a public node exists; the operator overrides it from configuration.
var DefaultBootstrapAddrs []string


// --- CODE ---

// Config is what the node needs to join the network.
type Config struct {
	// Network is the network name, such as "ecoa-dev". It names the topic.
	Network string

	// Listen is the multiaddr this node listens on.
	Listen string

	// BootstrapAddrs is the list of peer multiaddrs to connect to on startup.
	BootstrapAddrs []string

	// KnownAddrs are the peers of the last session, dialed before the
	// bootstrap list.
	KnownAddrs []string

	// OnConnected is called with the multiaddr of each peer this node dials
	// and reaches. It is how the caller remembers peers without p2p knowing
	// the store.
	OnConnected func(addr string)

	// KeyPath is the file holding the peer's private key. Empty makes the
	// peer ID last only as long as the process.
	KeyPath string

	// ConnLow is the connection count the manager trims down to.
	ConnLow int

	// ConnHigh is the connection count that makes the manager start trimming.
	ConnHigh int
}


// Peer is this node's presence on the network.
type Peer struct {
	host  host.Host
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}


// New builds the peer: host, bootstrap connections, Gossipsub and the
// subscription. It does not start receiving events; call Run for that.
func New(ctx context.Context, cfg Config) (*Peer, error) {
	opts, err := hostOptions(cfg)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	if cfg.OnConnected != nil {
		h.Network().Notify(remember(cfg.OnConnected))
	}

	connectBootstrap(ctx, h, dialOrder(cfg))

	topic, sub, err := joinTopic(ctx, h, cfg.Network)
	if err != nil {
		h.Close()
		return nil, err
	}

	return &Peer{host: h, topic: topic, sub: sub}, nil
}


// hostOptions returns the libp2p options: address, identity and limits.
func hostOptions(cfg Config) ([]libp2p.Option, error) {
	cm, err := connmgr.NewConnManager(cfg.ConnLow, cfg.ConnHigh)
	if err != nil {
		return nil, fmt.Errorf("connection manager: %w", err)
	}

	limits := rcmgr.DefaultLimits
	libp2p.SetDefaultServiceLimits(&limits)

	scaled := limits.Scale(resourceMemory, resourceFDs)

	rm, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(scaled))
	if err != nil {
		return nil, fmt.Errorf("resource manager: %w", err)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.Listen),
		libp2p.ConnectionManager(cm),
		libp2p.ResourceManager(rm),
	}

	if cfg.KeyPath == "" {
		return opts, nil
	}

	key, err := loadOrCreateKey(cfg.KeyPath)
	if err != nil {
		return nil, err
	}

	return append(opts, libp2p.Identity(key)), nil
}


// loadOrCreateKey reads the private key at path, or creates the file with a
// new one when it does not exist.
func loadOrCreateKey(path string) (crypto.PrivKey, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return createKey(path)
	}

	if err != nil {
		return nil, fmt.Errorf("read peer key %s: %w", path, err)
	}

	key, err := crypto.UnmarshalPrivateKey(raw)
	if err != nil {
		return nil, fmt.Errorf("peer key %s is corrupt: %w", path, err)
	}

	return key, nil
}


// createKey generates an Ed25519 key and writes it to path, never over a file.
func createKey(path string) (crypto.PrivKey, error) {
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	raw, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, keyMode)
	if err != nil {
		return nil, fmt.Errorf("create peer key %s: %w", path, err)
	}

	_, err = file.Write(raw)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		return nil, fmt.Errorf("write peer key %s: %w", path, err)
	}

	return key, nil
}


// bootstrapList returns the configured peers, or else the embedded list.
func bootstrapList(cfg Config) []string {
	if len(cfg.BootstrapAddrs) > 0 {
		return cfg.BootstrapAddrs
	}

	return DefaultBootstrapAddrs
}


// dialOrder returns the peers of the last session, then the bootstrap list.
func dialOrder(cfg Config) []string {
	return append(slices.Clone(cfg.KnownAddrs), bootstrapList(cfg)...)
}


// remember reports each peer this node dialed, since an inbound connection
// comes from a port nobody can dial back.
func remember(onConnected func(addr string)) *network.NotifyBundle {
	return &network.NotifyBundle{
		ConnectedF: func(_ network.Network, c network.Conn) {
			if c.Stat().Direction != network.DirOutbound {
				return
			}

			onConnected(c.RemoteMultiaddr().String() +
			            "/p2p/" + c.RemotePeer().String())
		},
	}
}


// connectBootstrap dials every address, each under its own timeout. A peer
// being down must not stop this node from starting, so a failure is logged
// and skipped.
func connectBootstrap(ctx context.Context, h host.Host, addrs []string) {
	for _, addr := range addrs {
		info, err := peer.AddrInfoFromString(addr)
		if err != nil {
			slog.Warn("bootstrap unusable", "addr", addr, "err", err)
			continue
		}

		dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
		err = h.Connect(dialCtx, *info)
		cancel()

		if err != nil {
			slog.Warn("bootstrap unreachable", "addr", addr, "err", err)
		}
	}
}


// Addrs returns this node's full multiaddrs, ready to be used as another
// node's bootstrap entries.
func (p *Peer) Addrs() []string {
	var addrs []string
	for _, a := range p.host.Addrs() {
		addrs = append(addrs, a.String()+"/p2p/"+p.host.ID().String())
	}
	return addrs
}


// Close shuts down the host.
func (p *Peer) Close() error {
	return p.host.Close()
}
