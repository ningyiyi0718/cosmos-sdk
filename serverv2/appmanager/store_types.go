package appmanager

import "cosmossdk.io/core/store"

type Store interface {
	NewBlockWithVersion(version uint64) (ReadonlyStore, error)
	ReadonlyWithVersion(version uint64) (ReadonlyStore, error)
	CommitChanges(changes []ChangeSet) (Hash, error)
}

type ChangeSet struct {
	Key, Value []byte
	Remove     bool
}

type BranchStore interface {
	ReadonlyStore
	store.KVStore
	ApplyChangeSets(changes []ChangeSet) error
	ChangeSets() ([]ChangeSet, error)
}

type Iterator interface {
	Next() error
	Valid() bool
	Key() []byte
	Value() []byte
	Close() error
}

type ReadonlyStore interface {
	Get([]byte) ([]byte, error)
	Iterate(start, end []byte) Iterator // consider removing iterate?
}
