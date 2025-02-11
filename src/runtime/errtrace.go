package runtime

import (
	"internal/abi"
	"unsafe"
)

type traceError struct {
	err         unsafe.Pointer
	traceLength uintptr
	// trace [traceLength]uintptr
}

func (e *traceError) trace() []uintptr {
	return unsafe.Slice(
		(*uintptr)(unsafe.Pointer(unsafe.Add(
			unsafe.Pointer(e),
			unsafe.Offsetof(e.traceLength)+unsafe.Sizeof(e.traceLength),
		))),
		e.traceLength,
	)
}

func newError(err unsafe.Pointer, traceLength uintptr) *traceError {
	e := (*traceError)(mallocgc(unsafe.Sizeof(traceError{})+(traceLength*unsafe.Sizeof(uintptr(0))), abi.TypeFor[traceError](), true))
	e.err = err
	e.traceLength = traceLength
	return e
}

func errTraceGet(err error) []uintptr {
	if err == nil {
		return nil
	}
	return (*traceError)(ifaceOf(&err).data).trace()
}

func errTraceNew(err error) error {
	fatal("errTraceNew")
	traceErr := newError(ifaceOf(&err).data, 1)
	traceErr.trace()[0] = errTraceCallerPC()
	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: unsafe.Pointer(traceErr),
	}))
}

// TODO: needs to work on any interface.
func errTraceUse(err error) error {
	if err == nil {
		return nil
	}
	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: (*traceError)(ifaceOf(&err).data).err,
	}))
}

func errTraceMove(err error) error {
	if err == nil {
		return nil
	}

	e := (*traceError)(ifaceOf(&err).data)
	// TODO: add sentinel error, that trace is truncated.
	if e.traceLength >= 256 {
		return err
	}

	e0 := newError(e.err, e.traceLength+1)
	copy(e0.trace(), e.trace())
	e0.trace()[len(e0.trace())-1] = errTraceCallerPC()

	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: unsafe.Pointer(e0),
	}))
}

func errTraceCallerPC() uintptr {
	pc, _, _, _ := Caller(2)
	return pc
}

func ifaceOf[T any](ep *T) *iface {
	return (*iface)(unsafe.Pointer(ep))
}
