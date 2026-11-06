package store


// --- GLOBALS ---

// eventsBucket holds every stored event, raw, keyed by its eventId.
var eventsBucket = []byte("events")


// logBucket lists the eventIds in the order this node received them, keyed by
// the local sequence number (uint64, big-endian).
var logBucket = []byte("log")


// metaBucket holds what this node observed about an event, keyed by its
// eventId: the sequence number and the local time of receipt.
var metaBucket = []byte("meta")


// tombstonesBucket holds a pruned event's author, keyed by its eventId. The
// value is the pruning time (uint64 Unix nanoseconds, big-endian) followed by
// the author. A v1 node never prunes, so it never writes here.
var tombstonesBucket = []byte("tombstones")


// buckets is every bucket New creates.
var buckets = [][]byte{
	eventsBucket,
	logBucket,
	metaBucket,
	tombstonesBucket,
}
