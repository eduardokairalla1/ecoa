package store


import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)


// ErrDuplicate is returned by AddEvent for an id already stored.
var ErrDuplicate = errors.New("store: event already stored")


// StoredEvent is an event of the log with the sequence it was received at.
type StoredEvent struct {
	// Sequence is the local receive sequence number of the event.
	Sequence uint64

	// ID is the eventId of the event.
	ID string

	// Raw is the event, byte for byte as it was received.
	Raw []byte
}


// AddEvent stores raw, byte for byte, under id, never overwriting an id. It
// returns the next local sequence number, which it assigns to the event along
// with receivedAt.
func (s *Store) AddEvent(id string,
                         raw []byte,
                         receivedAt time.Time) (uint64, error) {
	var sequence uint64

	err := s.db.Update(func(tx *bolt.Tx) error {
		var err error

		sequence, err = addEvent(tx, id, raw, receivedAt)
		return err
	})

	if err != nil {
		return 0, err
	}

	return sequence, nil
}


// addEvent writes the event into events, log and meta, all within tx.
func addEvent(tx *bolt.Tx,
              id string,
              raw []byte,
              receivedAt time.Time) (uint64, error) {
	events := tx.Bucket(eventsBucket)

	if events.Get([]byte(id)) != nil {
		return 0, ErrDuplicate
	}

	if err := events.Put([]byte(id), raw); err != nil {
		return 0, err
	}

	log := tx.Bucket(logBucket)

	sequence, err := log.NextSequence()
	if err != nil {
		return 0, err
	}

	if err := log.Put(sequenceKey(sequence), []byte(id)); err != nil {
		return 0, err
	}

	meta := binary.BigEndian.AppendUint64(sequenceKey(sequence),
	                                      uint64(receivedAt.UnixNano()))

	return sequence, tx.Bucket(metaBucket).Put([]byte(id), meta)
}


// sequenceKey returns the log key of a sequence number, big-endian so that
// byte order is sequence order.
func sequenceKey(sequence uint64) []byte {
	return binary.BigEndian.AppendUint64(nil, sequence)
}


// GetEvent returns the event stored under id, or nil when there is none.
func (s *Store) GetEvent(id string) ([]byte, error) {
	var raw []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		// copy of the stored bytes, nil when id is absent
		raw = bytes.Clone(tx.Bucket(eventsBucket).Get([]byte(id)))
		return nil
	})

	return raw, err
}


// EventsSince returns up to limit events received after the sequence cursor,
// in the order they were received.
func (s *Store) EventsSince(cursor uint64, limit int) ([]StoredEvent, error) {
	var found []StoredEvent

	err := s.db.View(func(tx *bolt.Tx) error {
		events := tx.Bucket(eventsBucket)
		log := tx.Bucket(logBucket).Cursor()

		key, id := log.Seek(sequenceKey(cursor))

		for ; key != nil && len(found) < limit; key, id = log.Next() {
			sequence := binary.BigEndian.Uint64(key)

			// Seek lands on the cursor itself when the log has it
			if sequence == cursor {
				continue
			}

			found = append(found, StoredEvent{
				Sequence: sequence,
				ID:       string(id),
				Raw:      bytes.Clone(events.Get(id)),
			})
		}

		return nil
	})

	return found, err
}
