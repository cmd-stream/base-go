package bcln

import (
	"testing"

	"github.com/cmd-stream/base-go"
)

func TestOptions(t *testing.T) {
	var (
		o                                     = Options{}
		wantCallback UnexpectedResultCallback = func(seq base.Seq, result base.Result) {}
	)
	Apply([]SetOption{WithUnexpectedResultCallback(wantCallback)}, &o)

	if o.UnexpectedResultCallback == nil {
		t.Errorf("UnexpectedResultCallback == nil")
	}

}
