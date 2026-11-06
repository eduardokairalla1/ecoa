package store


// --- GLOBALS ---

// eventsBucket holds every stored event, raw, keyed by its eventId.
var eventsBucket = []byte("events")


// peersBucket holds the multiaddr of every peer this node has connected to.
var peersBucket = []byte("peers")
