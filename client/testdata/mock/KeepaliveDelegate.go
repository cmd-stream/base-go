package mock

import (
	"sync"

	"github.com/ymz-ncnk/mok"
)

type KeepaliveFn func(muSn *sync.Mutex)

func NewKeepaliveDelegate() KeepaliveDelegate {
	return KeepaliveDelegate{
		Delegate: NewDelegate(),
		Mock:     mok.New("KeepaliveDelegate"),
	}
}

type KeepaliveDelegate struct {
	Delegate
	*mok.Mock
}

func (m KeepaliveDelegate) RegisterKeepalive(fn KeepaliveFn) KeepaliveDelegate {
	m.Register("Keepalive", fn)
	return m
}

func (m KeepaliveDelegate) Keepalive(muSn *sync.Mutex) {
	_, err := m.Call("Keepalive", muSn)
	if err != nil {
		panic(err)
	}
}
