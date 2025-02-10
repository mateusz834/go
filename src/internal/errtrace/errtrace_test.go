package errtrace

import "testing"

type err struct{ a [128]byte }

func (e *err) Error() string {
	return "some error"
}

func TestErrorTrace(t *testing.T) {
	var er err
	e := New(&er)

	e = Move(e)
	e = Move(e)
	e = Move(e)
	e = Move(e)
	t.Log(Get(e))
	t.Log(e.Error())

	a(&e)

	v, ok := Use(e).(*err)
	t.Logf("%p %v", v, ok)
	t.Logf("%p %v", &er, ok)
}

func a(t *error) error {
	return *t
}
