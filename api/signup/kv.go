package signup

import (
	"log"

	"github.com/mantil-io/mantil.go"
)

type kv struct {
	r *kvOps
	a *kvOps
	k *kvOps
}

func (k *kv) Registrations() *kvOps {
	if k.r == nil {
		k.r = newKvOps(registrationPartition)
	}
	return k.r
}

func (k *kv) Activations() *kvOps {
	if k.a == nil {
		k.a = newKvOps(activationPartition)
	}
	return k.a
}

func (k *kv) Keys() *kvOps {
	if k.k == nil {
		k.k = newKvOps(keysPartition)
	}
	return k.k
}

func newKvOps(partition string) *kvOps {
	kv, err := mantil.NewKV(partition)
	if err != nil {
		log.Printf("mantil.NewKV failed: %s", err)
	}
	return &kvOps{connectError: err, kv: kv}
}

type kvOps struct {
	connectError error
	kv           *mantil.KV
}

func (k *kvOps) Put(id string, rec interface{}) error {
	if k.connectError != nil {
		return errInternal
	}
	if err := k.kv.Put(id, rec); err != nil {
		return errInternal
	}
	return nil
}

func (k *kvOps) Get(id string, rec interface{}) error {
	if k.connectError != nil {
		return errInternal
	}
	if err := k.kv.Get(id, &rec); err != nil {
		return err
	}
	return nil
}
