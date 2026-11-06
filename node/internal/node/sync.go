package node


import (
	"context"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/peer"
)


// --- CODE ---

// syncWith catches up with p and logs why it could not.
func (n *Node) syncWith(ctx context.Context, p peer.ID) {
	if err := n.catchUp(ctx, p); err != nil {
		slog.Warn("catch up stopped", "peer", p, "reason", err)
	}
}


// catchUp pulls every page p accepted since the cursor kept for it.
func (n *Node) catchUp(ctx context.Context, p peer.ID) error {
	cursor, err := n.store.Cursor(p.String())
	if err != nil {
		return err
	}

	for {
		page, err := n.peer.GetEventsSince(ctx, p, cursor)
		if err != nil {
			return err
		}

		if page.HasMore && page.NextCursor <= cursor {
			return fmt.Errorf("cursor stuck at %d with more to send", cursor)
		}

		if err := n.receivePage(page.Events, p); err != nil {
			return err
		}

		cursor = page.NextCursor

		if err := n.store.SetCursor(p.String(), cursor); err != nil {
			return err
		}

		if !page.HasMore {
			return nil
		}
	}
}


// receivePage hands each event of a page to Receive, as an arrival from p.
func (n *Node) receivePage(events [][]byte, p peer.ID) error {
	arrival := Arrival{Kind: Sync, Peer: p}

	for _, raw := range events {
		outcome, err := n.Receive(raw, arrival)

		// A rejected event is the peer's fault; a failed one is ours, and
		// moving the cursor past it would lose it.
		if outcome == Failed {
			return fmt.Errorf("receive: %w", err)
		}
	}

	return nil
}
