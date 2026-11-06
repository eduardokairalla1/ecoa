package node


import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eduardokairalla1/EcoaNetwork/internal/orphan"
	"github.com/eduardokairalla1/EcoaNetwork/internal/store"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/envelope"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/fixturetest"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/identity"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/schema"
	"github.com/eduardokairalla1/EcoaNetwork/protocol/go/testevent"
	bolt "go.etcd.io/bbolt"
)


// --- CONSTANTS ---

// network is the network every test node belongs to.
const network = "ecoa-dev"


// fixtures is the directory of the behavioural cases.
const fixtures = "../../../protocol/fixtures"


// peerQuota is the orphans one peer may hold in a test node.
const peerQuota = 3


// --- GLOBALS ---

// now is the local clock of every test node.
var now = time.Date(2026, 8, 5, 17, 0, 0, 0, time.UTC)


// gateway is the arrival the tests hand events through.
var gateway = Arrival{Kind: Gateway}


// --- CODE ---

// clock reads now, so a test moves every clock of the node by setting it.
func clock() time.Time {
	return now
}


// newNode returns a node over a fresh store, and the store's file path.
func newNode(t *testing.T) (*Node, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "bolt.db")

	s, err := store.New(path)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}

	t.Cleanup(func() { s.Close() })

	schemas, err := schema.Load()
	if err != nil {
		t.Fatalf("schema.Load: %v", err)
	}

	rules := envelope.Rules{
		Network: network,
		MaxSize: envelope.MaxSize,
		Now:     clock,
	}

	n := &Node{
		store:      s,
		rules:      rules,
		schemas:    schemas,
		identities: identity.NewTracker(),
		orphans:    orphan.New(10, peerQuota, time.Hour, clock),
	}

	return n, path
}


// review returns the fields of a valid review on network.
func review() testevent.Fields {
	return testevent.Fields{
		Protocol:  network,
		Version:   1,
		EventType: "review.created",
		CreatedAt: "2026-08-05T17:00:00Z",
		Payload: `{"entity":"cnpj:12345678000199",` +
		         `"rating":4,"body":"ok"}`,
	}
}


// sign returns fields signed by the key of seed, or stops the test.
func sign(t *testing.T, seed byte, fields testevent.Fields) []byte {
	t.Helper()

	key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, 32))

	raw, err := testevent.Sign(key, fields)
	if err != nil {
		t.Fatalf("testevent.Sign: %v", err)
	}

	return raw
}


// withField returns raw with the top-level field name set to value.
func withField(t *testing.T, raw []byte, name string, value any) []byte {
	t.Helper()

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	fields[name] = encoded

	edited, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return edited
}


// field returns the string value of the top-level field name of raw.
func field(t *testing.T, raw []byte, name string) string {
	t.Helper()

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	value, _ := fields[name].(string)

	return value
}


// stored reads every entry in the events bucket of the file at path.
func stored(t *testing.T, n *Node, path string) map[string][]byte {
	t.Helper()

	// the store holds the file lock until closed
	if err := n.store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatalf("bolt.Open: %v", err)
	}

	defer db.Close()

	entries := map[string][]byte{}
	err = db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("events")).ForEach(func(k, v []byte) error {
			entries[string(k)] = bytes.Clone(v)
			return nil
		})
	})

	if err != nil {
		t.Fatalf("read events: %v", err)
	}

	return entries
}


func TestReceiveStoresAValidEventByteForByte(t *testing.T) {
	// Arrange
	n, path := newNode(t)
	raw := sign(t, 1, review())

	// Act
	outcome, err := n.Receive(raw, gateway)

	// Assert
	if outcome != Stored || err != nil {
		t.Fatalf("Receive = %s, %v, want stored", outcome, err)
	}

	entries := stored(t, n, path)
	id := field(t, raw, "eventId")

	if len(entries) != 1 || !bytes.Equal(entries[id], raw) {
		t.Fatalf("stored %q, want only %s = %q", entries, id, raw)
	}
}


func TestReceiveStoresAnUnknownEventType(t *testing.T) {
	// Arrange
	n, _ := newNode(t)
	opaque := review()
	opaque.EventType = "thing.happened"
	opaque.Payload = `{"anything":[1,2]}`

	// Act
	outcome, err := n.Receive(sign(t, 1, opaque), gateway)

	// Assert
	if outcome != Stored || err != nil {
		t.Fatalf("Receive = %s, %v, want stored", outcome, err)
	}
}


func TestReceiveDropsAReplay(t *testing.T) {
	// Arrange
	n, path := newNode(t)
	first := sign(t, 1, review())

	if outcome, err := n.Receive(first, gateway); outcome != Stored {
		t.Fatalf("first Receive = %s, %v, want stored", outcome, err)
	}

	// same eventId with a signature that does not verify
	other := field(t, sign(t, 2, review()), "signature")
	replay := withField(t, first, "signature", other)

	// Act
	outcome, err := n.Receive(replay, gateway)

	// Assert
	if outcome != Duplicate || err != nil {
		t.Fatalf("Receive = %s, %v, want duplicate", outcome, err)
	}

	id := field(t, first, "eventId")
	if entries := stored(t, n, path); !bytes.Equal(entries[id], first) {
		t.Fatalf("stored %q, want the first %s", entries[id], id)
	}
}


func TestReceiveIgnoresAnotherNetwork(t *testing.T) {
	// Arrange
	n, path := newNode(t)
	other := review()
	other.Protocol = "ecoa-other"

	// Act
	outcome, _ := n.Receive(sign(t, 1, other), gateway)

	// Assert
	if outcome != Ignored {
		t.Fatalf("Receive = %s, want ignored", outcome)
	}

	if entries := stored(t, n, path); len(entries) != 0 {
		t.Fatalf("stored %q, want nothing", entries)
	}
}


func TestReceiveRejects(t *testing.T) {
	badRating := review()
	badRating.Payload = `{"entity":"cnpj:12345678000199",` +
	                    `"rating":9,"body":"ok"}`

	future := review()
	future.CreatedAt = "2026-08-05T17:06:00Z"

	device := review()
	deviceKey := "did:key:z6MkiTBz1ymuepAQ4HEHYSF1H8quG5GLVVQR3djdX3mDooWp"
	device.DeviceKey = &deviceKey

	valid := sign(t, 1, review())
	other := sign(t, 2, review())

	cases := []struct {
		name string
		raw  []byte
	}{
		{"too large", bytes.Repeat([]byte(" "), envelope.MaxSize+1)},
		{"not json", []byte(`not json`)},
		{"future createdAt", sign(t, 1, future)},
		{"bad payload", sign(t, 1, badRating)},
		{"tampered body", bytes.Replace(valid, []byte(`"ok"`),
		                                []byte(`"ko"`), 1)},
		{"forged eventId", withField(t, valid, "eventId",
		                             field(t, other, "eventId"))},
		{"other signature", withField(t, valid, "signature",
		                              field(t, other, "signature"))},
		{"device key", sign(t, 1, device)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Arrange
			n, path := newNode(t)

			// Act
			outcome, err := n.Receive(c.raw, gateway)

			// Assert
			if outcome != Rejected || err == nil {
				t.Fatalf("Receive = %s, %v, want rejected", outcome, err)
			}

			if entries := stored(t, n, path); len(entries) != 0 {
				t.Fatalf("stored %q, want nothing", entries)
			}
		})
	}
}


func TestReceiveFailsWhenTheStoreIsClosed(t *testing.T) {
	// Arrange
	n, _ := newNode(t)
	if err := n.store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	// Act
	outcome, err := n.Receive(sign(t, 1, review()), gateway)

	// Assert
	if outcome != Failed || err == nil {
		t.Fatalf("Receive = %s, %v, want failed", outcome, err)
	}
}


// did returns the author of the events sign signs with seed.
func did(seed byte) string {
	key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, 32))
	public, _ := key.Public().(ed25519.PublicKey)

	return identity.FromKey(public).String()
}


// rotation returns the fields of a rotation to the key of seed.
func rotation(seed byte, createdAt string) testevent.Fields {
	return testevent.Fields{
		Protocol:  network,
		Version:   1,
		EventType: "identity.key_rotated",
		CreatedAt: createdAt,
		Payload:   `{"successor":"` + did(seed) + `"}`,
	}
}


func TestReceiveRejectsWhatARevokedKeySignsAfterwards(t *testing.T) {
	// Arrange
	n, _ := newNode(t)

	revocation := review()
	revocation.EventType = "identity.key_revoked"
	revocation.Payload = `{"reason":"compromised"}`

	// signed before the revocation, but received after it
	backdated := review()
	backdated.CreatedAt = "2026-03-01T10:00:00Z"

	before := sign(t, 1, review())
	n.Receive(before, gateway)
	n.Receive(sign(t, 1, revocation), gateway)

	// Act
	outcome, err := n.Receive(sign(t, 1, backdated), gateway)

	// Assert
	var revoked *RevokedKeyError
	if outcome != Rejected || !errors.As(err, &revoked) {
		t.Fatalf("Receive = %s, %v, want rejected by a revoked key",
		         outcome, err)
	}
}


func TestReceiveRejectsARotationAfterTheFirstOne(t *testing.T) {
	// Arrange
	n, _ := newNode(t)
	n.Receive(sign(t, 1, rotation(2, "2026-01-01T10:00:00Z")), gateway)

	// Act
	outcome, err := n.Receive(
		sign(t, 1, rotation(3, "2026-07-01T10:00:00Z")), gateway,
	)

	// Assert
	var observed *identity.SuccessorObservedError
	if outcome != Rejected || !errors.As(err, &observed) {
		t.Fatalf("Receive = %s, %v, want a successor already observed",
		         outcome, err)
	}
}


func TestReceiveAcceptsOnlyOneOfManyParallelRotations(t *testing.T) {
	// Arrange
	n, _ := newNode(t)
	const rivals = 20

	var wg sync.WaitGroup
	var storedCount atomic.Int32

	// Act
	for seed := byte(2); seed < 2+rivals; seed++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			raw := sign(t, 1, rotation(seed, "2026-01-01T10:00:00Z"))
			if outcome, _ := n.Receive(raw, gateway); outcome == Stored {
				storedCount.Add(1)
			}
		}()
	}

	wg.Wait()

	// Assert
	if got := storedCount.Load(); got != 1 {
		t.Fatalf("stored %d rotations of one key, want 1", got)
	}
}


// endorsement returns the fields of a vote on the event target.
func endorsement(target string) testevent.Fields {
	return testevent.Fields{
		Protocol:  network,
		Version:   1,
		EventType: "review.endorsed",
		CreatedAt: "2026-08-05T17:00:00Z",
		Payload:   `{"target":"` + target + `"}`,
	}
}


// expectBuffered stops the test when the buffer differs from c.
func expectBuffered(t *testing.T, n *Node, c fixturetest.Case) {
	t.Helper()

	if _, ok := c.Expect["buffered"]; ok {
		if got, want := n.orphans.Len(), len(c.List("buffered")); got != want {
			t.Fatalf("buffered %d events, want %d", got, want)
		}
	}

	if c.Expect["perPeerCapEnforced"] == true && n.orphans.Len() > peerQuota {
		t.Fatalf("buffered %d events of one peer, quota is %d",
		         n.orphans.Len(), peerQuota)
	}

	if after, ok := c.Expect["afterBound"].(map[string]any); ok {
		now = now.Add(2 * time.Hour)
		want, _ := after["buffered"].([]any)

		if got := n.orphans.Len(); got != len(want) {
			t.Fatalf("after the bound buffered %d, want %d", got, len(want))
		}
	}
}


// expectStored stops the test when the log differs from the stored of c.
func expectStored(t *testing.T,
                  n *Node,
                  path string,
                  c fixturetest.Case,
                  ids map[string]string) {
	t.Helper()

	if _, ok := c.Expect["stored"]; !ok {
		return
	}

	var want []string
	for _, id := range c.List("stored") {
		want = append(want, ids[id])
	}

	var got []string
	for id := range stored(t, n, path) {
		got = append(got, id)
	}

	slices.Sort(want)
	slices.Sort(got)

	if !slices.Equal(got, want) {
		t.Fatalf("stored %v, want %v", got, want)
	}
}


func TestOrphanFixtures(t *testing.T) {
	file := fixturetest.LoadFrom(t, fixtures, "orphan-buffer")

	fixturetest.Run(t, file, func(t *testing.T, c fixturetest.Case) {
		// Arrange
		n, path := newNode(t)
		ids := map[string]string{}
		saved := now
		t.Cleanup(func() { now = saved })

		// Act
		for i, e := range c.Events(t) {
			from := peer.ID(c.Arrival[i].Author)
			n.Receive(e.Raw, Arrival{Kind: Gossip, Peer: from})

			ids[e.ID] = e.Env.EventID
		}

		// Assert
		expectBuffered(t, n, c)
		expectStored(t, n, path, c, ids)
	})
}


func TestReceiveHoldsAnOrphanWithoutStoringIt(t *testing.T) {
	// Arrange
	n, path := newNode(t)
	raw := sign(t, 1, endorsement(fixturetest.EventID("e1")))

	// Act
	outcome, err := n.Receive(raw, gateway)

	// Assert
	if outcome != Orphaned || err != nil {
		t.Fatalf("Receive = %s, %v, want orphaned", outcome, err)
	}

	if n.orphans.Len() != 1 || len(stored(t, n, path)) != 0 {
		t.Fatal("the orphan was stored, or not held")
	}
}


func TestReceiveKeepsAnotherPeersOrphanWhenOnePeerFloods(t *testing.T) {
	// Arrange
	n, _ := newNode(t)
	bob := Arrival{Kind: Gossip, Peer: "bob"}
	mallory := Arrival{Kind: Gossip, Peer: "mallory"}

	n.Receive(sign(t, 1, endorsement(fixturetest.EventID("wanted"))), bob)

	// Act
	for i := range 2 * peerQuota {
		target := fixturetest.EventID("never" + string(rune('a'+i)))
		n.Receive(sign(t, 2, endorsement(target)), mallory)
	}

	// Assert
	if got := n.orphans.Len(); got != 1+peerQuota {
		t.Fatalf("buffered %d, want bob's 1 and mallory's %d", got, peerQuota)
	}

	if n.orphans.Release(fixturetest.EventID("wanted")) == nil {
		t.Fatal("the orphan of another peer was evicted")
	}
}


func TestReceiveReleasesAChainOfOrphans(t *testing.T) {
	// Arrange
	n, _ := newNode(t)
	mallory := Arrival{Kind: Gossip, Peer: "mallory"}

	first := fixturetest.Case{Arrival: []fixturetest.Arrival{
		{ID: "e1", Author: "alice", Type: "review.created",
		 Payload: []byte(`{"entity":"cnpj:1","rating":2,"body":"x"}`)},
		{ID: "v1", Author: "bob", Type: "review.endorsed",
		 Payload: []byte(`{"target":"e1"}`)},
		{ID: "v2", Author: "carol", Type: "review.endorsed",
		 Payload: []byte(`{"target":"v1"}`)},
	}}
	events := first.Events(t)

	n.Receive(events[2].Raw, mallory)
	n.Receive(events[1].Raw, mallory)

	// Act
	outcome, _ := n.Receive(events[0].Raw, mallory)

	// Assert
	if outcome != Stored || n.orphans.Len() != 0 {
		t.Fatalf("Receive = %s with %d held, want the chain released",
		         outcome, n.orphans.Len())
	}
}
