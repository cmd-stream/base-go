package mock

import (
	"sync"

	"github.com/ymz-ncnk/mok"
)

func NewClientKeepaliveDelegate() ClientKeepaliveDelegate {
	return ClientKeepaliveDelegate{
		ClientDelegate: NewClientDelegate(),
		Mock:           mok.New("ClientKeepaliveDelegate"),
	}
}

type ClientKeepaliveDelegate struct {
	ClientDelegate
	*mok.Mock
}

func (m ClientKeepaliveDelegate) RegisterKeepalive(
	fn func(muSn *sync.Mutex)) ClientKeepaliveDelegate {
	m.Register("Keepalive", fn)
	return m
}

func (m ClientKeepaliveDelegate) Keepalive(muSn *sync.Mutex) {
	_, err := m.Call("Keepalive", muSn)
	if err != nil {
		panic(err)
	}
}
