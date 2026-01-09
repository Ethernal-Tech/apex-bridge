package storagehelper

import (
	tracker "github.com/Ethernal-Tech/solana-event-tracker"
	"github.com/gagliardetto/solana-go"
)

type Event struct {
	Slot      uint64
	ProgramID solana.PublicKey
	EventName string
	EventData any
}
type StorageHelper struct {
	Events []Event
	Slot   uint64
}

func NewStorage() *StorageHelper {
	return &StorageHelper{
		Events: make([]Event, 0),
		Slot:   0,
	}
}

func (b *StorageHelper) Close() {}

func (b *StorageHelper) ReadSlot() (uint64, error) {
	return b.Slot, nil
}

func (b *StorageHelper) StoreSlot(tx tracker.StorageTransaction, slot uint64) error {
	b.Slot = slot
	return nil
}

func (b *StorageHelper) StoreEvent(
	tx tracker.StorageTransaction,
	slot uint64,
	programID solana.PublicKey,
	eventName string,
	eventData any) error {
	b.Events = append(b.Events, Event{
		Slot:      slot,
		ProgramID: programID,
		EventName: eventName,
		EventData: eventData,
	})
	return nil
}

func (b *StorageHelper) UseTransactions() bool {
	return false
}

func (b *StorageHelper) ApplyTransaction(
	slotFn func(tracker.StorageTransaction) error,
	eventFns []func(tracker.StorageTransaction) error) error {
	return nil
}
