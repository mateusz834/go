package ssa

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa/ssaparser"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestParseSSA(t *testing.T) {
	c := testConfig(t)

	var s = `
b1:

    (?) v1 = InitMem <mem>
    (?) v7 = Arg <bool> {a} (a[bool])
    (?) v8 = Arg <bool> {b} (b[bool])
    (?) v10 = ConstBool <bool> [true]
If v7 -> b2 b3

b2: <- b1 b3
    (?) v11 = Phi <bool> v10 v8
    (?) v13 = MakeResult <bool,mem> v11 v1
Ret v13

b3: <- b1
Plain -> b2
`

	_, s, _ = strings.Cut(s, "\n")
	t.Log(s)

	cc := TestFuncCfg{
		config: c.config,
		tb:     t,
		aux: map[string]Aux{
			"a": &obj.LSym{Name: "a"},
			"b": &obj.LSym{Name: "b"},
			"c": &obj.LSym{Name: "c"},
			"d": &obj.LSym{Name: "d"},
		},
	}

	f := cc.BuildSSAFunc(s).f
	phiopt(f)
	fuseLate(f)
	t.Log(ssaparser.PrintFunc(fprintFunc2(f)))
	checkFunc(f)
}

type opInfoCode struct {
	info opInfo
	code Op
}

var (
	blockKindNames = make(map[string]BlockKind)
	opInfoNames    = make(map[string]opInfoCode)
)

func init() {
	for k, v := range blockString {
		blockKindNames[v] = BlockKind(k)
	}
	for k, v := range opcodeTable {
		opInfoNames[v.name] = opInfoCode{
			info: v,
			code: Op(k),
		}
	}
}

func defaultTypes(c Types) map[string]*types.Type {
	return map[string]*types.Type{
		"bool":     c.Bool,
		"int8":     c.Int8,
		"int16":    c.Int16,
		"int32":    c.Int32,
		"int64":    c.Int64,
		"uint8":    c.UInt8,
		"uint16":   c.UInt16,
		"uint32":   c.UInt32,
		"uint64":   c.UInt64,
		"int":      c.Int,
		"float32":  c.Float32,
		"float64":  c.Float64,
		"uint":     c.UInt,
		"uintptr":  c.Uintptr,
		"string":   c.String,
		"*byte":    c.BytePtr,
		"*int32":   c.Int32Ptr,
		"*uint32":  c.UInt32Ptr,
		"*int":     c.IntPtr,
		"*uintptr": c.UintptrPtr,
		"*float32": c.Float32Ptr,
		"*float64": c.Float64Ptr,
		"**byte":   c.BytePtrPtr,
	}
}

func (t *TestFuncCfg) typ(typ string) *types.Type {
	if typ == "mem" {
		return types.TypeMem
	}

	tt := t.types[typ]
	if tt != nil {
		return tt
	}

	// Handle mem tuple types <int,mem>.
	b, a, ok := strings.Cut(typ, ",")
	if ok && strings.Count(typ, ",") == 1 {
		b := t.types[b]
		if b == nil || a != "mem" {
			return nil
		}
		return types.NewTuple(b, types.TypeMem)
	}

	return nil
}

type TestFuncCfg struct {
	config *Config
	fe     Frontend
	tb     testing.TB

	aux   map[string]Aux
	types map[string]*types.Type
}

type TestFunc struct {
	cfg *TestFuncCfg

	f      *Func
	blocks map[string]*Block
	values map[string]*Value
}

func (t *TestFunc) blockName(b *Block) string {
	for k, v := range t.blocks {
		if v == b {
			return k
		}
	}
	return fmt.Sprintf("b%v", b.ID)
}

func (t *TestFunc) valueName(v *Value) string {
	for k, vv := range t.values {
		if vv == v {
			return k
		}
	}
	return fmt.Sprintf("v%v", v.ID)
}

func (c *TestFuncCfg) BuildSSAFunc(ssa string) *TestFunc {
	c.tb.Helper()
	c.types = defaultTypes(c.config.Types)
	tf, err := c.buildSSAFunc(ssa)
	if err != nil {
		c.tb.Fatalf("failed to build SSA func: %v", err)
	}
	checkFunc(tf.f)
	return tf
}

func (c *TestFuncCfg) buildSSAFunc(ssa string) (*TestFunc, error) {
	fun, err := ssaparser.ParseSSAFunc(ssa)
	if err != nil {
		return nil, fmt.Errorf("failed parsing SSA: %v", err)
	}

	// TODO: check unused vals and blocks.

	f := c.config.NewFunc(c.Frontend(), new(Cache))
	f.pass = &emptyPass
	f.cachedLineStarts = newXposmap(map[int]lineRange{0: {0, 100}, 1: {0, 100}, 2: {0, 100}, 3: {0, 100}, 4: {0, 100}})

	blocks := make(map[string]*Block)
	vals := make(map[string]*Value)

	// Create all blocks and values, partially initializing them.
	for _, block := range fun.Blocks {
		if _, ok := blocks[block.Name]; ok {
			return nil, fmt.Errorf("block: %v already declared", block.Name)
		}

		blockKind, ok := blockKindNames[block.Kind]
		if !ok {
			return nil, fmt.Errorf("block: %v unrecognized kind %q", block.Name, block.Kind)
		}

		b := f.NewBlock(blockKind)
		blocks[block.Name] = b

		// Treat first block as entry.
		if f.Entry == nil {
			f.Entry = b
			if len(block.Preds) != 0 {
				return nil, fmt.Errorf("entry block: %v has unexpected predecessors", block.Name)
			}
		}

		for _, val := range block.Values {
			if _, ok := vals[val.Name]; ok {
				return nil, fmt.Errorf("value: %v already declared", val.Name)
			}
			op, ok := opInfoNames[val.Op]
			if !ok {
				return nil, fmt.Errorf("value: %v unrecognized opcode: %q", val.Name, val.Op)
			}
			typ := c.typ(val.Type)
			if typ == nil {
				return nil, fmt.Errorf("value: %v unrecognized type: %q", val.Name, val.Type)
			}
			intAux, aux, err := c.mapAux(op.code, val.AuxInt, val.Aux)
			if err != nil {
				return nil, fmt.Errorf("value: %v  invalid aux: %v", val.Name, err)
			}
			v := b.NewValue0(src.NoXPos, op.code, typ)
			vals[val.Name] = v
			v.AuxInt = intAux
			v.Aux = aux
		}
	}

	// Populate (*Block).Controls.
	for _, block := range fun.Blocks {
		b := blocks[block.Name]
		if len(block.Controls) <= 2 && len(block.Controls) != 0 {
			b.Controls[0] = vals[block.Controls[0]]
			if b.Controls[0] == nil {
				return nil, fmt.Errorf("block: %v references a non-existing control val: %v", block.Name, block.Controls[0])
			}
			b.Controls[0].Uses++
			if len(block.Controls) == 2 {
				b.Controls[1] = vals[block.Controls[0]]
				if b.Controls[1] == nil {
					return nil, fmt.Errorf("block: %v references a non-existing control val: %v", block.Name, block.Controls[1])
				}
				b.Controls[1].Uses++
			}
		} else if len(block.Controls) != 0 {
			return nil, fmt.Errorf("block: %v references an unexpected amount of control values", block.Name)
		}
	}

	// Populate successors and predecessors.
	for _, block := range fun.Blocks {
		b := blocks[block.Name]
		for _, v := range block.Succ {
			succ, ok := blocks[v]
			if !ok {
				return nil, fmt.Errorf("block: %v references a non-existing block in its successors: %v", block.Name, succ)
			}
			b.Succs = append(b.Succs, Edge{succ, -1})
		}

		for _, v := range block.Preds {
			pred, ok := blocks[v]
			if !ok {
				return nil, fmt.Errorf("block: %v references a non-existing block in its predecessors: %v", block.Name, pred)
			}
			b.Preds = append(b.Preds, Edge{pred, -1})
		}
	}

	blockName := func(b *Block) string {
		name := ""
		for k, v := range blocks {
			if b == v {
				name = k
				break
			}
		}
		return name
	}

	// Fill reverse edges indexes.
	for _, b := range f.Blocks {
		for i := range b.Succs {
			s := &b.Succs[i]
			s.i = slices.IndexFunc(s.b.Preds, func(e Edge) bool { return e.b == b })
			if s.i == -1 {
				return nil, fmt.Errorf(
					"block: %v has a successor block %v, but %v is missing a corresponding predecessor",
					blockName(b), blockName(s.b), blockName(s.b),
				)
			}
		}
		for i := range b.Preds {
			p := &b.Preds[i]
			p.i = slices.IndexFunc(p.b.Succs, func(e Edge) bool { return e.b == b })
			if p.i == -1 {
				return nil, fmt.Errorf(
					"block: %v has a predecessor block %v, but %v is missing a corresponding successor",
					blockName(b), blockName(p.b), blockName(p.b),
				)
			}
		}
	}

	// Assign value arguments.
	for _, block := range fun.Blocks {
		for _, val := range block.Values {
			v := vals[val.Name]
			var args []*Value
			for _, arg := range val.Args {
				v, ok := vals[arg]
				if !ok {
					return nil, fmt.Errorf("value: %v references a non-existing value: %v", val.Name, arg)
				}
				args = append(args, v)
			}
			v.AddArgs(args...)
		}
	}

	return &TestFunc{
		f:      f,
		blocks: blocks,
		values: vals,
	}, nil
}

func intAux[T int8 | int16 | int32 | int64](auxInt, aux string) (int64, Aux, error) {
	if aux != "" {
		return 0, nil, fmt.Errorf("unexpected aux: %q", aux)
	}
	size := -1
	switch any(new(T)).(type) {
	case int8:
		size = 8
	case int16:
		size = 16
	case int32:
		size = 32
	case int64:
		size = 64
	}
	val, err := strconv.ParseInt(auxInt, 10, size)
	if err != nil {
		return 0, nil, fmt.Errorf("unexpected IntAux: %q", auxInt)
	}
	return int64(T(val)), nil, nil
}

func (c *TestFuncCfg) mapAux(op Op, auxInt, aux string) (int64, Aux, error) {
	switch opcodeTable[op].auxType {
	case auxNone:
		return 0, nil, nil
	case auxBool:
		if aux != "" {
			return 0, nil, fmt.Errorf("unexpected aux: %q", aux)
		}
		if auxInt == "false" {
			return 0, nil, nil
		} else if auxInt == "true" {
			return 1, nil, nil
		}
		return 0, nil, fmt.Errorf("unexpected IntAux: %q", auxInt)
	case auxInt8:
		return intAux[int8](auxInt, aux)
	case auxInt16:
		return intAux[int16](auxInt, aux)
	case auxInt32:
		return intAux[int32](auxInt, aux)
	case auxInt64:
		return intAux[int64](auxInt, aux)
	case auxSym:
		if v, ok := c.aux[aux]; ok {
			return 0, v, nil
		}
		return 0, nil, nil
	case auxSymOff:
		return 0, nil, nil
	default:
		panic(fmt.Sprintf("unexpected auxType: %v", opcodeTable[op].auxType))
	}
}

func (c *TestFuncCfg) Frontend() Frontend {
	if c.fe == nil {
		pkg := types.NewPkg("my/import/path", "path")
		fn := ir.NewFunc(src.NoXPos, src.NoXPos, pkg.Lookup("function"), types.NewSignature(nil, nil, nil))
		fn.DeclareParams(true)
		fn.LSym = &obj.LSym{Name: "my/import/path.function"}
		c.fe = TestFrontend{
			ctxt: c.config.ctxt,
			f:    fn,
			t:    c.tb,
		}
	}
	return c.fe
}

// TODO: we need differentiate aliases from real blocks IDs. and  disallow real block IDs as aliases.
func (c *TestFuncCfg) SSA(tf *TestFunc) string {
	f := &ssaparser.Func{}
	for _, block := range tf.f.Blocks {
		b := &ssaparser.Block{
			Name: tf.blockName(block),
			Kind: block.Kind.String(),
		}
		f.Blocks = append(f.Blocks, b)

		for _, e := range block.Preds {
			b.Preds = append(b.Preds, tf.blockName(e.b))
		}
		for _, e := range block.Succs {
			b.Succ = append(b.Succ, tf.blockName(e.b))
		}
		for _, c := range block.ControlValues() {
			b.Controls = append(b.Controls, tf.valueName(c))
		}

		for _, value := range block.Values {
			v := &ssaparser.Value{
				Pos:    "",
				Name:   tf.valueName(value),
				Op:     value.Op.String(),
				Type:   value.Type.String(),
				Aux:    "",
				AuxInt: "",
				Args:   []string{},
			}

			for k, vv := range tf.cfg.aux {
				if vv == value.Aux {
					v.Aux = k
				}
			}

			b.Values = append(b.Values, v)
		}
	}
	return ssaparser.PrintFunc(f)
}

type stringAuxRepr struct {
	AuxInt string
	Aux    string
}

func fillAux(v *Value, v2 *ssaparser.Value) {
	switch opcodeTable[v.Op].auxType {
	case auxBool:
		if v.AuxInt == 0 {
			v2.AuxInt = "false"
		} else {
			v2.AuxInt = "true"
		}
	case auxInt8:
		v2.AuxInt = strconv.FormatInt(int64(v.AuxInt8()), 10)
	case auxInt16:
		v2.AuxInt = strconv.FormatInt(int64(v.AuxInt16()), 10)
	case auxInt32:
		v2.AuxInt = strconv.FormatInt(int64(v.AuxInt32()), 10)
	case auxInt64, auxInt128:
		v2.AuxInt = strconv.FormatInt(v.AuxInt, 10)
	case auxUInt8:
		v2.AuxInt = strconv.FormatUint(uint64(v.AuxUInt8()), 10)
	case auxString:
		v2.Aux = fmt.Sprintf("%q", v.Aux)
	case auxSym, auxCall, auxTyp:
		if v.Aux != nil {
			return stringAuxRepr{Aux: fmt.Sprintf("%v", v.Aux)}
		}
		return stringAuxRepr{}
	case auxSymOff, auxCallOff, auxTypSize, auxNameOffsetInt8:
		var aux stringAuxRepr
		if v.Aux != nil {
			aux.Aux = fmt.Sprintf("%v", v.Aux)
		}
		if v.AuxInt != 0 || opcodeTable[v.Op].auxType == auxNameOffsetInt8 {
			aux.AuxInt = fmt.Sprintf("%v", v.AuxInt)
		}
		return aux
	case auxSymValAndOff:
		var aux stringAuxRepr
		if v.Aux != nil {
			aux.Aux = fmt.Sprintf("%v", v.Aux)
		}
		aux.AuxInt = fmt.Sprintf("%s", v.AuxValAndOff())
		return aux
	case auxNone:
	default:
		// If you see this, add a case above instead.
		return stringAuxRepr{AuxInt: fmt.Sprintf("auxtype=%d AuxInt=%d Aux=%v", opcodeTable[v.Op].auxType, v.AuxInt, v.Aux)}
	}
}

func auxAsString(v *Value) stringAuxRepr {
	switch opcodeTable[v.Op].auxType {
	case auxBool:
		if v.AuxInt == 0 {
			return stringAuxRepr{AuxInt: "false"}
		} else {
			return stringAuxRepr{AuxInt: "true"}
		}
	case auxInt8:
		return stringAuxRepr{AuxInt: strconv.FormatInt(int64(v.AuxInt8()), 10)}
	case auxInt16:
		return stringAuxRepr{AuxInt: strconv.FormatInt(int64(v.AuxInt16()), 10)}
	case auxInt32:
		return stringAuxRepr{AuxInt: strconv.FormatInt(int64(v.AuxInt32()), 10)}
	case auxInt64, auxInt128:
		return stringAuxRepr{AuxInt: strconv.FormatInt(v.AuxInt, 10)}
	case auxUInt8:
		return stringAuxRepr{AuxInt: strconv.FormatUint(uint64(v.AuxUInt8()), 10)}
	case auxARM64BitField:
		lsb := v.AuxArm64BitField().lsb()
		width := v.AuxArm64BitField().width()
		return stringAuxRepr{AuxInt: fmt.Sprintf("lsb=%d,width=%d", lsb, width)}
	case auxFloat32, auxFloat64:
		return stringAuxRepr{AuxInt: fmt.Sprintf("%g", v.AuxFloat())}
	case auxString:
		return stringAuxRepr{Aux: fmt.Sprintf("%q", v.Aux)}
	case auxSym, auxCall, auxTyp:
		if v.Aux != nil {
			return stringAuxRepr{Aux: fmt.Sprintf("%v", v.Aux)}
		}
		return stringAuxRepr{}
	case auxSymOff, auxCallOff, auxTypSize, auxNameOffsetInt8:
		var aux stringAuxRepr
		if v.Aux != nil {
			aux.Aux = fmt.Sprintf("%v", v.Aux)
		}
		if v.AuxInt != 0 || opcodeTable[v.Op].auxType == auxNameOffsetInt8 {
			aux.AuxInt = fmt.Sprintf("%v", v.AuxInt)
		}
		return aux
	case auxSymValAndOff:
		var aux stringAuxRepr
		if v.Aux != nil {
			aux.Aux = fmt.Sprintf("%v", v.Aux)
		}
		aux.AuxInt = fmt.Sprintf("%s", v.AuxValAndOff())
		return aux
	case auxCCop:
		return stringAuxRepr{AuxInt: fmt.Sprintf("%s", Op(v.AuxInt))}
	case auxS390XCCMask, auxS390XRotateParams:
		return stringAuxRepr{Aux: fmt.Sprintf("%v", v.Aux)}
	case auxFlagConstant:
		return stringAuxRepr{Aux: fmt.Sprintf("%s", flagConstant(v.AuxInt))}
	case auxNone:
		return stringAuxRepr{}
	default:
		// If you see this, add a case above instead.
		return stringAuxRepr{AuxInt: fmt.Sprintf("auxtype=%d AuxInt=%d Aux=%v", opcodeTable[v.Op].auxType, v.AuxInt, v.Aux)}
	}
}
