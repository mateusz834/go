package errtrace

import (
	"internal/abi"
	"internal/goarch"
	"runtime"
	"sync"
	"unsafe"
)

type TraceError[T error] struct {
	err   T
	trace []uintptr
}

func GetTrace(e error) []uintptr {
	if e == nil {
		return nil
	}

	// TODO: make sure that error is a traceError[T], otherwise skip (i.e. return e).

	data := ifaceOf(&e).data

	return *(*[]uintptr)(unsafe.Add(data, goarch.PtrSize))
}

var traceErrTypes sync.Map // map[abi.TypeOf(T)]abi.TypeOf(TraceError[T])

func NewErrorTrace[T error](e T) error {
	var err error = e
	typ := ifaceOf(&err).tab.Type
	data := ifaceOf(&err).data

	newData := newobject(typ)
	typedmemmove(typ, newData, data)

	var te any = TraceError[T]{}
	traceErrTypes.Store(typ, efaceOf(&te)._type)

	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: data,
	}))
}

func ErrTraceMove(e error) error {
	if e == nil {
		return nil
	}

	// TODO: make sure that error is a traceError[T], otherwise skip (i.e. return e).

	typ := ifaceOf(&e).tab.Type
	data := ifaceOf(&e).data

	if typ == nil {
		panic("her")
	}

	traceErrType, ok := traceErrTypes.Load(typ)
	if !ok {
		panic("internal/errtrace: internal error")
	}

	newData := newobject(traceErrType.(*abi.Type))
	typedmemmove(typ, newData, data)

	// TODO: find offset through type.
	trace := (*[]uintptr)(unsafe.Add(newData, goarch.PtrSize))

	pc, _, _, _ := runtime.Caller(1)
	*trace = append((*trace)[:len(*trace)], pc)

	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&e).tab,
		data: newData,
	}))
}

type eface struct {
	_type *abi.Type
	data  unsafe.Pointer
}

func efaceOf(ep *any) *eface {
	return (*eface)(unsafe.Pointer(ep))
}

type iface struct {
	tab  *abi.ITab
	data unsafe.Pointer
}

func ifaceOf[T any](ep *T) *iface {
	return (*iface)(unsafe.Pointer(ep))
}

//go:linkname typedmemmove
func typedmemmove(typ *abi.Type, dst, src unsafe.Pointer)

//go:linkname newobject
func newobject(typ *abi.Type) unsafe.Pointer
