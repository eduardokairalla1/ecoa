// Package node decides what happens to each event the node receives.
package node


import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/eduardokairalla1/EcoaNetwork/internal/p2p"
	"github.com/eduardokairalla1/EcoaNetwork/internal/store"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/envelope"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/event"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/schema"
)


// --- CONSTANTS ---

const (
	// Stored is an event that passed every check and is now in the log.
	Stored Outcome = iota + 1

	// Duplicate is an eventId already in the log.
	Duplicate

	// Ignored is an event of another network or version.
	Ignored

	// Rejected is an event that failed a check.
	Rejected

	// Failed is an event the node could not judge, for a fault of its own.
	Failed
)


// --- GLOBALS ---

// outcomeNames is the log name of each outcome.
var outcomeNames = map[Outcome]string{
	Stored:    "stored",
	Duplicate: "duplicate",
	Ignored:   "ignored",
	Rejected:  "rejected",
	Failed:    "failed",
}


// --- CODE ---

// Node is the one path every arrival is handed to.
type Node struct {
	store *store.Store
	peer  *p2p.Peer

	// rules are the size, network and clock this node opens envelopes by
	rules envelope.Rules

	// schemas validate payloads by eventType
	schemas *schema.Set
}


// Outcome is what the node did with one received event.
type Outcome int


// New returns a node that judges events by the rules of cfg.Network.
func New(ctx context.Context,
         s *store.Store,
         cfg p2p.Config,
         maxEventSize int) (*Node, error) {
	schemas, err := schema.Load()
	if err != nil {
		return nil, err
	}

	p, err := p2p.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	rules := envelope.Rules{
		Network: cfg.Network,
		MaxSize: maxEventSize,
		Now:     time.Now,
	}

	return &Node{store: s, peer: p, rules: rules, schemas: schemas}, nil
}


// Run pulls events from the topic until ctx is cancelled.
func (n *Node) Run(ctx context.Context) {
	for {
		raw, from, err := n.peer.ReceiveEvent(ctx)
		if err != nil {
			return
		}

		n.Receive(raw, Arrival{Kind: Gossip, Peer: from})
	}
}


// Receive validates one event and stores it, however it arrived.
func (n *Node) Receive(raw []byte, arrival Arrival) (Outcome, error) {
	env, err := envelope.Open(raw, n.rules)
	if _, ok := errors.AsType[*envelope.OtherNetworkError](err); ok {
		return Ignored, err
	}

	if err != nil {
		return report(arrival, Rejected, err)
	}

	stored, err := n.store.GetEvent(env.EventID)
	if err != nil {
		return report(arrival, Failed, err)
	}

	if stored != nil {
		return Duplicate, nil
	}

	verified, err := event.Verify(env, n.schemas)
	if err != nil {
		return report(arrival, Rejected, err)
	}

	_, err = n.store.AddEvent(verified.ID(), verified.Raw(), n.rules.Now())
	if errors.Is(err, store.ErrDuplicate) {
		return Duplicate, nil
	}

	if err != nil {
		return report(arrival, Failed, err)
	}

	slog.Info("stored", "arrival", arrival.Kind, "bytes", len(raw))
	return Stored, nil
}


// report logs an event the node rejected or failed to store.
func report(arrival Arrival, outcome Outcome, err error) (Outcome, error) {
	level := slog.LevelInfo
	if outcome == Failed {
		level = slog.LevelWarn
	}

	slog.Log(context.Background(), level, outcome.String(),
	         "arrival", arrival.Kind,
	         "peer", arrival.Peer,
	         "reason", err)

	return outcome, err
}


// Publish waits for the mesh first, since Gossipsub drops unroutable messages.
func (n *Node) Publish(ctx context.Context, raw []byte) error {
	if err := n.peer.WaitForPeers(ctx); err != nil {
		return err
	}

	return n.peer.Publish(ctx, raw)
}


func (n *Node) Addrs() []string {
	return n.peer.Addrs()
}


func (n *Node) Close() error {
	return n.peer.Close()
}


// String returns the log name of the outcome.
func (o Outcome) String() string {
	return outcomeNames[o]
}
