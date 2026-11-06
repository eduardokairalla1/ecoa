package store


// --- GLOBALS ---

// eventsBucket holds every stored event, raw, keyed by its eventId.
var eventsBucket = []byte("events")


// cursorsBucket holds, per peer ID, the last sync cursor received from it.
var cursorsBucket = []byte("cursors")
