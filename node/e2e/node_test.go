package e2e


import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/eduardokairalla1/EcoaNetwork/internal/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


// --- CONSTANTS ---

// waitFor is how long a node may take to start or to catch up.
const waitFor = 30 * time.Second


// --- GLOBALS ---

// binary is the node executable the tests start, built once by TestMain.
var binary string


// --- CODE ---

// logSink keeps the JSON log lines a node wrote, across restarts.
type logSink struct {
	mu    sync.Mutex
	lines []map[string]any
}


// Write splits p into log lines and keeps each one that is JSON.
func (l *logSink) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, line := range bytes.Split(p, []byte("\n")) {
		var entry map[string]any
		if json.Unmarshal(line, &entry) == nil {
			l.lines = append(l.lines, entry)
		}
	}

	return len(p), nil
}


// count returns how many lines the sink holds.
func (l *logSink) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.lines)
}


// first returns field of the first line from index from on whose msg is msg.
func (l *logSink) first(msg string, field string, from int) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, entry := range l.lines[from:] {
		if entry["msg"] == msg {
			value, _ := entry[field].(string)

			return value
		}
	}

	return ""
}


// stored returns the eventId of every event the node logged as stored.
func (l *logSink) stored() map[string]bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	ids := map[string]bool{}

	for _, entry := range l.lines {
		if id, ok := entry["eventId"].(string); ok && entry["msg"] == "stored" {
			ids[id] = true
		}
	}

	return ids
}


// testNode is one node of the cluster, a process or a container.
type testNode struct {
	t    *testing.T
	name string

	// dir holds the database and the peer key, so a restart keeps both
	dir  string
	logs *logSink
	cmd  *exec.Cmd

	// bootstrap are the peers this node dials on start
	bootstrap []string

	// p2pAddr and grpcAddr are what the running node logged it listens on
	p2pAddr  string
	grpcAddr string
}


// startNode starts a node that dials bootstrap, and waits until it listens.
func startNode(t *testing.T, name string, bootstrap []string) *testNode {
	t.Helper()

	n := &testNode{
		t:         t,
		name:      name,
		dir:       t.TempDir(),
		logs:      &logSink{},
		bootstrap: bootstrap,
	}

	n.start()
	t.Cleanup(n.stop)

	return n
}


// command returns the process that runs the node, image or binary.
func (n *testNode) command() *exec.Cmd {
	env := []string{
		"NETWORK=ecoa-dev",
		"LISTEN_ADDR=/ip4/127.0.0.1/tcp/0",
		"GRPC_ADDR=127.0.0.1:0",
		"BOOTSTRAP_ADDRS=" + strings.Join(n.bootstrap, ","),
	}

	image := os.Getenv("ECOA_IMAGE")
	if image == "" {
		env = append(env,
		             "BOLT_DB_PATH="+filepath.Join(n.dir, "bolt.db"),
		             "PEER_KEY_PATH="+filepath.Join(n.dir, "peer.key"))

		cmd := exec.Command(binary)
		cmd.Env = append(os.Environ(), env...)

		return cmd
	}

	args := []string{"run", "--rm", "--name", n.container(),
	                 "--network", "host", "--user", userID(),
	                 "-v", n.dir + ":/data"}

	for _, kv := range env {
		args = append(args, "-e", kv)
	}

	return exec.Command("docker", append(args, image)...)
}


// container is the name of the container of this node.
func (n *testNode) container() string {
	return fmt.Sprintf("ecoa-e2e-%d-%s", os.Getpid(), n.name)
}


// userID is the user and group the container runs as.
func userID() string {
	return fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
}


// start runs the node and waits for the addresses it listens on.
func (n *testNode) start() {
	n.t.Helper()

	// the lines of an earlier run stay in the sink; the new addresses follow
	from := n.logs.count()

	n.cmd = n.command()
	n.cmd.Stdout = n.logs
	n.cmd.Stderr = n.logs

	if err := n.cmd.Start(); err != nil {
		n.t.Fatalf("start %s: %v", n.name, err)
	}

	eventually(n.t, n.name+" listening", func() bool {
		n.p2pAddr = n.logs.first("listening", "addr", from)
		n.grpcAddr = n.logs.first("grpc", "addr", from)

		return n.p2pAddr != "" && n.grpcAddr != ""
	})
}


// stop ends the node and waits for it to exit. It is safe to call twice.
func (n *testNode) stop() {
	if n.cmd == nil {
		return
	}

	if os.Getenv("ECOA_IMAGE") != "" {
		exec.Command("docker", "stop", "-t", "5", n.container()).Run()
	} else {
		n.cmd.Process.Signal(os.Interrupt)
	}

	n.cmd.Wait()
	n.cmd = nil
}


// restart stops the node and starts it again over the same database and key.
func (n *testNode) restart() {
	n.t.Helper()

	n.stop()
	n.start()
}


// publish submits raw through the gRPC port and returns the node's outcome.
func (n *testNode) publish(raw []byte) *pb.PublishResponse {
	n.t.Helper()

	conn, err := grpc.NewClient(n.grpcAddr, grpc.WithTransportCredentials(
		insecure.NewCredentials()))
	if err != nil {
		n.t.Fatalf("dial %s: %v", n.name, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()

	resp, err := pb.NewNodeServiceClient(conn).Publish(
		ctx, &pb.Event{Envelope: raw})
	if err != nil {
		n.t.Fatalf("publish to %s: %v", n.name, err)
	}

	return resp
}


// eventually polls cond until it holds, or stops the test.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(waitFor)

	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}

		time.Sleep(50 * time.Millisecond)
	}
}
