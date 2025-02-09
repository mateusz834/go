package errtrace

import (
	"internal/abi"
	"runtime"
	"unsafe"
)

// TODO: add an internal error type for: fmt.Errorf("%v", err), so that it also preserves the error trace (implicit unwrap).

type _error struct {
	err         unsafe.Pointer
	traceLength uintptr
	// trace [traceLength]uintptr
}

func (e *_error) trace() []uintptr {
	return unsafe.Slice(
		(*uintptr)(unsafe.Pointer(unsafe.Add(
			unsafe.Pointer(e),
			unsafe.Offsetof(e.traceLength)+unsafe.Sizeof(e.traceLength),
		))),
		e.traceLength,
	)
}

func newError(err unsafe.Pointer, traceLength uintptr) *_error {
	e := (*_error)(mallocgc(unsafe.Sizeof(_error{})+(traceLength*unsafe.Sizeof(uintptr(0))), abi.TypeFor[_error]()))
	e.err = err
	e.traceLength = traceLength
	return e
}

func Get(err error) []uintptr {
	if err == nil {
		return nil
	}
	return (*_error)(ifaceOf(&err).data).trace()
}

func Use(err error) error {
	if err == nil {
		return nil
	}
	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: (*_error)(ifaceOf(&err).data).err,
	}))
}

func callerPC() uintptr {
	pc, _, _, _ := runtime.Caller(2)
	return pc
}

func New(err error) error {
	traceErr := newError(ifaceOf(&err).data, 1)
	traceErr.trace()[0] = callerPC()
	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: unsafe.Pointer(traceErr),
	}))
}

func Move(err error) error {
	if err == nil {
		return nil
	}

	e := (*_error)(ifaceOf(&err).data)
	// TODO: add sentinel error, that trace is truncated.
	if e.traceLength >= 256 {
		return err
	}

	e0 := newError(e.err, e.traceLength+1)
	copy(e0.trace(), e.trace())
	e0.trace()[len(e0.trace())-1] = callerPC()

	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: unsafe.Pointer(e0),
	}))
}

type iface struct {
	tab  *abi.ITab
	data unsafe.Pointer
}

func ifaceOf[T any](ep *T) *iface {
	return (*iface)(unsafe.Pointer(ep))
}

//go:linkname mallocgc
func mallocgc(size uintptr, typ *abi.Type) unsafe.Pointer

//go:linkname typedmemmove
func typedmemmove(typ *abi.Type, dst, src unsafe.Pointer)

//go:linkname newobject
func newobject(typ *abi.Type) unsafe.Pointer
