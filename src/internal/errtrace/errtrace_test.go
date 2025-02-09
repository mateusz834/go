package errtrace

import "testing"

type err struct{ a [128]byte }

func (e *err) Error() string {
	return "some error"
}

func TestErrorTrace(t *testing.T) {
	e := NewErrorTrace(new(err))
	e = ErrTraceMove(e)
	e = ErrTraceMove(e)
	t.Log(GetTrace(e))
}
