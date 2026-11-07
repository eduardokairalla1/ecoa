package e2e


import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/eduardokairalla1/EcoaNetwork/internal/protobuf"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/testevent"
)


// --- CODE ---

// TestMain builds the node binary once, unless the tests run on an image.
func TestMain(m *testing.M) {
	if os.Getenv("ECOA_IMAGE") == "" {
		dir, err := os.MkdirTemp("", "ecoa-e2e")
		if err != nil {
			fmt.Fprintln(os.Stderr, "tempdir:", err)
			os.Exit(1)
		}

		binary = filepath.Join(dir, "ecoa")

		build := exec.Command("go", "build", "-o", binary, "../cmd/ecoa")
		build.Stderr = os.Stderr

		if err := build.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "build:", err)
			os.Exit(1)
		}
	}

	code := m.Run()

	if binary != "" {
		os.RemoveAll(filepath.Dir(binary))
	}

	os.Exit(code)
}


// cluster starts three nodes that all know the first one.
func cluster(t *testing.T) (*testNode, *testNode, *testNode) {
	t.Helper()

	a := startNode(t, "a", nil)
	b := startNode(t, "b", []string{a.p2pAddr})
	c := startNode(t, "c", []string{a.p2pAddr, b.p2pAddr})

	return a, b, c
}


// signed returns an event of eventType by the key of seed, made now.
func signed(t *testing.T, seed byte, eventType string, payload string) []byte {
	t.Helper()

	key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, 32))

	raw, err := testevent.Sign(key, testevent.Fields{
		Protocol:  "ecoa-dev",
		Version:   1,
		EventType: eventType,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	return raw
}


// review returns a valid review by the key of seed with a body of its own.
func review(t *testing.T, seed byte, body string) []byte {
	t.Helper()

	return signed(t, seed, "review.created",
	              `{"entity":"cnpj:12345678000199","rating":4,"body":"`+
	              body+`"}`)
}


// endorsement returns a vote by the key of seed on the event target.
func endorsement(t *testing.T, seed byte, target string) []byte {
	t.Helper()

	return signed(t, seed, "review.endorsed", `{"target":"`+target+`"}`)
}


// idOf returns the eventId of raw.
func idOf(t *testing.T, raw []byte) string {
	t.Helper()

	var fields struct {
		EventID string `json:"eventId"`
	}

	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("eventId: %v", err)
	}

	return fields.EventID
}


// holds reports whether the node logged every id as stored.
func holds(n *testNode, ids ...string) bool {
	stored := n.logs.stored()

	for _, id := range ids {
		if !stored[id] {
			return false
		}
	}

	return true
}


// reachAll waits until every node holds every id.
func reachAll(t *testing.T, ids []string, nodes ...*testNode) {
	t.Helper()

	for _, n := range nodes {
		eventually(t, "node "+n.name+" to store the events", func() bool {
			return holds(n, ids...)
		})
	}
}


// sameSet stops the test unless the nodes stored exactly the same eventIds.
func sameSet(t *testing.T, nodes ...*testNode) {
	t.Helper()

	want := nodes[0].logs.stored()

	for _, n := range nodes[1:] {
		got := n.logs.stored()

		if len(got) != len(want) || !holds(n, keys(want)...) {
			t.Fatalf("node %s stored %v, node %s stored %v",
			         n.name, keys(got), nodes[0].name, keys(want))
		}
	}
}


// keys returns the keys of set.
func keys(set map[string]bool) []string {
	var out []string
	for key := range set {
		out = append(out, key)
	}

	return out
}


func TestEventPublishedToOneNodeReachesAllThree(t *testing.T) {
	// Arrange
	a, b, c := cluster(t)
	raw := review(t, 1, "propagates")

	// Act
	resp := b.publish(raw)

	// Assert
	if resp.Outcome != pb.Outcome_ACCEPTED {
		t.Fatalf("publish = %s %q, want ACCEPTED", resp.Outcome, resp.Reason)
	}

	reachAll(t, []string{idOf(t, raw)}, a, b, c)
}


func TestInvalidEventIsStoredByNone(t *testing.T) {
	// Arrange
	a, b, c := cluster(t)
	tampered := bytes.Replace(review(t, 1, "ok"), []byte(`"ok"`),
	                          []byte(`"ko"`), 1)
	valid := review(t, 2, "after")

	// Act
	resp := a.publish(tampered)
	a.publish(valid)

	// Assert
	if resp.Outcome != pb.Outcome_REJECTED {
		t.Fatalf("publish = %s, want REJECTED", resp.Outcome)
	}

	// the valid event published after it has crossed the network, so the
	// invalid one had every chance to
	reachAll(t, []string{idOf(t, valid)}, a, b, c)

	for _, n := range []*testNode{a, b, c} {
		if holds(n, idOf(t, tampered)) {
			t.Errorf("node %s stored the invalid event", n.name)
		}
	}
}


func TestNodeThatWasDownCatchesUpOnReturning(t *testing.T) {
	// Arrange
	a, b, c := cluster(t)
	first := review(t, 1, "before")
	a.publish(first)
	reachAll(t, []string{idOf(t, first)}, a, b, c)

	c.stop()

	var missed []string

	for i := range 3 {
		raw := review(t, byte(2+i), fmt.Sprintf("while down %d", i))
		a.publish(raw)

		missed = append(missed, idOf(t, raw))
	}

	reachAll(t, missed, a, b)

	// Act
	c.restart()

	// Assert
	reachAll(t, missed, c)
	sameSet(t, a, b, c)
}


func TestEventThatArrivesBeforeItsTargetIsStoredEverywhere(t *testing.T) {
	// Arrange
	a, b, c := cluster(t)
	target := review(t, 1, "target")
	vote := endorsement(t, 2, idOf(t, target))

	// Act
	first := a.publish(vote)
	b.publish(target)

	// Assert
	if first.Outcome != pb.Outcome_ORPHAN {
		t.Fatalf("publish = %s, want ORPHAN", first.Outcome)
	}

	reachAll(t, []string{idOf(t, target), idOf(t, vote)}, a, b, c)
}
