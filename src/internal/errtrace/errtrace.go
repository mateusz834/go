package errtrace

import (
	"internal/abi"
	"runtime"
	"sync"
	"unsafe"
)

// TODO: add an internal error type for: fmt.Errorf("%v", err), so that it also preserves the error trace (implicit unwrap).

// TODO: remove.
type Error[T any] struct {
	err   T
	trace []uintptr
}

type Err struct {
	err         unsafe.Pointer
	traceLength uintptr

	// trace [traceLength]uintptr
}

// TODO: remove.
var traceErrTypes sync.Map // map[abi.TypeOf(T)]abi.TypeOf(Error[T])

func Get(e error) []uintptr {
	if e == nil {
		return nil
	}
	typ := ifaceOf(&e).tab.Type
	data := ifaceOf(&e).data

	traceErrTypeAny, ok := traceErrTypes.Load(typ)
	if !ok {
		panic("internal/errtrace: internal error")
	}
	traceErrType := traceErrTypeAny.(*abi.Type)

	if traceErrType.Kind_&abi.KindMask != abi.Struct {
		panic("internal/errtrace: internal error")
	}

	st := (*abi.StructType)(unsafe.Pointer(traceErrType))
	if len(st.Fields) != 2 {
		panic("internal/errtrace: internal error")
	}

	return *(*[]uintptr)(unsafe.Add(data, st.Fields[1].Offset))
}

// TODO
//func test() {
//	// TODO: every interface use would need an Use? Without interface to interface conversion.
//	var e error
//	e = fmt.Errorf("test %w", e)
//	_ = e
//}

// TODO: should work for all interfaces.
func Use(e error) error {
	if e == nil {
		return nil
	}

	itab := ifaceOf(&e).tab
	data := ifaceOf(&e).data

	if itab.Type.IsDirectIface() {
		return *(*error)(unsafe.Pointer(&iface{
			tab:  itab,
			data: (*Error[unsafe.Pointer])(data).err,
		}))
	}

	return e
}

func New[T error](e T) error {
	var err error = e
	traceErrTypes.Store(abi.TypeFor[T](), abi.TypeFor[Error[T]]())
	return *(*error)(unsafe.Pointer(&iface{
		tab:  ifaceOf(&err).tab,
		data: unsafe.Pointer(&Error[T]{err: e}),
	}))
}

func Move(e error) error {
	if e == nil {
		return nil
	}

	typ := ifaceOf(&e).tab.Type
	data := ifaceOf(&e).data

	traceErrTypeAny, ok := traceErrTypes.Load(typ)
	if !ok {
		panic("internal/errtrace: internal error")
	}
	traceErrType := traceErrTypeAny.(*abi.Type)

	newData := newobject(traceErrType)
	typedmemmove(traceErrType, newData, data)

	if traceErrType.Kind_&abi.KindMask != abi.Struct {
		panic("internal/errtrace: internal error")
	}

	st := (*abi.StructType)(unsafe.Pointer(traceErrType))
	if len(st.Fields) != 2 {
		panic("internal/errtrace: internal error")
	}

	trace := (*[]uintptr)(unsafe.Add(newData, st.Fields[1].Offset))

	// TODO: add an sentinel pc??
	if len(*trace) >= 256 {
		return e
	}

	// TODO: if ok is false, add an sentinel pc??
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
