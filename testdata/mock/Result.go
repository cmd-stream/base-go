package mock

import (
	"github.com/ymz-ncnk/mok"
)

func NewResult() Result {
	return Result{mok.New("Result")}
}

type Result struct {
	*mok.Mock
}

func (m Result) RegisterLastOne(fn func() (lastOne bool)) Result {
	m.Register("LastOne", fn)
	return m
}

func (m Result) LastOne() (lastOne bool) {
	vals, err := m.Call("LastOne")
	if err != nil {
		panic(err)
	}
	lastOne = vals[0].(bool)
	return
}
