package storagehelper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"

	tracker "github.com/Ethernal-Tech/solana-event-tracker"
	"github.com/gagliardetto/solana-go"
	bolt "go.etcd.io/bbolt"
)

type BoltStorageHandler struct {
	txMode bool
	db     *bolt.DB
}

var (
	slotBucket   = []byte("slot")
	eventsBucket = []byte("events")
)

func NewBoltStorageHandler(path string, txMode bool) (*BoltStorageHandler, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open bolt db: %w", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(slotBucket)
		if err != nil {
			return fmt.Errorf("cannot create the slot bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists(eventsBucket)
		if err != nil {
			return fmt.Errorf("cannot create the events bucket: %w\n", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &BoltStorageHandler{txMode, db}, nil
}

func (b *BoltStorageHandler) Close() {
	b.db.Close()
}

func encodeSlotValue(slot uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, slot)
	return b
}

func decodeSlotValue(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func (b *BoltStorageHandler) ReadSlot() (uint64, error) {
	var retValue uint64

	if err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(slotBucket)
		if bucket == nil {
			return fmt.Errorf("cannot find slot bucket")
		}

		value := bucket.Get([]byte("current"))
		if value == nil {
			retValue = 0
		} else {
			retValue = decodeSlotValue(value)
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return retValue, nil
}

func (b *BoltStorageHandler) StoreSlot(tx tracker.StorageTransaction, slot uint64) error {
	storeFn := func(tx *bolt.Tx) error {
		bucket := tx.Bucket(slotBucket)
		if bucket == nil {
			return fmt.Errorf("cannot find slot bucket")
		}

		return bucket.Put([]byte("current"), encodeSlotValue(slot+1))
	}

	if tx == nil {
		return b.db.Update(storeFn)
	}

	if tx, ok := tx.(*bolt.Tx); ok {
		return storeFn(tx)
	}

	return fmt.Errorf("unknown storage transaction type")
}

func (b *BoltStorageHandler) StoreEvent(
	tx tracker.StorageTransaction,
	slot uint64,
	programID solana.PublicKey,
	eventName string,
	eventData any) error {
	storeFn := func(tx *bolt.Tx) error {
		eventsBucket := tx.Bucket(eventsBucket)
		if eventsBucket == nil {
			return fmt.Errorf("cannot find events bucket")
		}

		slotBucket, err := eventsBucket.CreateBucketIfNotExists(encodeSlotValue(slot))
		if err != nil {
			return fmt.Errorf("cannot create bucket for the slot %v: %w", slot, err)
		}

		programBucket, err := slotBucket.CreateBucketIfNotExists(programID.Bytes())
		if err != nil {
			return fmt.Errorf("cannot create bucket for the program %v: %w", programID, err)
		}

		eventTypeBucket, err := programBucket.CreateBucketIfNotExists([]byte(eventName))
		if err != nil {
			return fmt.Errorf("cannot create bucket for the event type %v: %w", eventName, err)
		}

		value, err := json.Marshal(reflect.ValueOf(eventData).Elem().Interface())
		if err != nil {
			return fmt.Errorf("cannot serialize (marshal) event data: %w", err)
		}

		hash := sha256.Sum256(value)
		key := hash[:]

		return eventTypeBucket.Put(key, value)
	}

	if tx == nil {
		return b.db.Update(storeFn)
	}

	if tx, ok := tx.(*bolt.Tx); ok {
		return storeFn(tx)
	}

	return fmt.Errorf("unknown storage transaction type")
}

func (b *BoltStorageHandler) UseTransactions() bool {
	return b.txMode
}

func (b *BoltStorageHandler) ApplyTransaction(
	slotFn func(tracker.StorageTransaction) error,
	eventFns []func(tracker.StorageTransaction) error) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, fn := range eventFns {
			if err := fn(tx); err != nil {
				return err
			}
		}

		return slotFn(tx)
	})
}

type Event struct {
	Slot      uint64
	ProgramID solana.PublicKey
	EventName string
	EventData any
}

func (b *BoltStorageHandler) GetAllEvents() ([]Event, error) {
	var events []Event

	if err := b.db.View(func(tx *bolt.Tx) error {
		eventsBucket := tx.Bucket(eventsBucket)
		if eventsBucket == nil {
			return fmt.Errorf("cannot find events bucket")
		}

		eventsBucket.ForEach(func(k, v []byte) error {
			events = append(events, Event{
				Slot:      decodeSlotValue(k),
				ProgramID: solana.PublicKeyFromBytes(v),
				EventName: string(v),
				EventData: v,
			})

			return nil
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return events, nil
}
