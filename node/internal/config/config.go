// Package config reads the node's settings from the environment.
package config


import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/envelope"
)


// --- CONSTANTS ---

const (
	// defaultConnLow is the connection count the manager trims down to.
	defaultConnLow = 32

	// defaultConnHigh is the connection count that starts the trimming.
	defaultConnHigh = 96
)


// --- CODE ---

// Config is every setting the node reads at startup.
type Config struct {
	// Network is the network this node belongs to.
	Network string

	// ListenAddr is the multiaddr the node listens on.
	ListenAddr string

	// GRPCAddr is the host:port the gateway submits events to.
	GRPCAddr string

	// BootstrapAddrs replaces the embedded peer list when not empty.
	BootstrapAddrs []string

	// BoltDBPath is the path of the bbolt file holding the event log.
	BoltDBPath string

	// PeerKeyPath is the file holding the node's libp2p private key.
	PeerKeyPath string

	// ConnLow is the connection count the manager trims down to.
	ConnLow int

	// ConnHigh is the connection count that makes the manager start trimming.
	ConnHigh int

	// MaxEventSize is this node's limit per event, in bytes.
	MaxEventSize int
}


// Load reads the environment and applies the default of each unset variable.
func Load() (Config, error) {
	size, err := loadMaxEventSize()
	if err != nil {
		return Config{}, err
	}

	low, high, err := loadConnLimits()
	if err != nil {
		return Config{}, err
	}

	network := envOr("NETWORK", "ecoa-dev")
	if !envelope.ValidProtocol(network) {
		return Config{}, fmt.Errorf("NETWORK: %q is not ecoa or ecoa-<name>",
		                            network)
	}

	cfg := Config{
		Network:        network,
		ListenAddr:     envOr("LISTEN_ADDR", "/ip4/127.0.0.1/tcp/4001"),
		GRPCAddr:       envOr("GRPC_ADDR", "127.0.0.1:50051"),
		BootstrapAddrs: splitAddrs(os.Getenv("BOOTSTRAP_ADDRS")),
		BoltDBPath:     envOr("BOLT_DB_PATH", "data/bbolt/bolt.db"),
		PeerKeyPath:    envOr("PEER_KEY_PATH", "data/peer.key"),
		ConnLow:        low,
		ConnHigh:       high,
		MaxEventSize:   size,
	}

	return cfg, nil
}


// envOr returns the variable named key, or fallback when it is unset or empty.
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}


// splitAddrs turns a comma-separated list into its items.
func splitAddrs(list string) []string {
	if list == "" {
		return nil
	}

	return strings.Split(list, ",")
}


// loadMaxEventSize reads MAX_EVENT_SIZE, or an error outside the range.
func loadMaxEventSize() (int, error) {
	value := os.Getenv("MAX_EVENT_SIZE")
	if value == "" {
		return envelope.MaxSize, nil
	}

	size, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("MAX_EVENT_SIZE: %w", err)
	}

	if size <= 0 || size > envelope.MaxSize {
		return 0, fmt.Errorf("MAX_EVENT_SIZE: %d is outside 1..%d",
		                     size,
		                     envelope.MaxSize)
	}

	return size, nil
}


// loadConnLimits reads CONN_LOW and CONN_HIGH, or an error for a pair that is
// not 1 <= low <= high.
func loadConnLimits() (int, int, error) {
	low, err := envInt("CONN_LOW", defaultConnLow)
	if err != nil {
		return 0, 0, err
	}

	high, err := envInt("CONN_HIGH", defaultConnHigh)
	if err != nil {
		return 0, 0, err
	}

	if low < 1 || low > high {
		return 0, 0, fmt.Errorf("CONN_LOW %d and CONN_HIGH %d: want "+
		                        "1 <= low <= high",
		                        low,
		                        high)
	}

	return low, high, nil
}


// envInt reads the integer variable named key, or fallback when it is unset.
func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return n, nil
}
