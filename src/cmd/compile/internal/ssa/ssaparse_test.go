package ssa

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa/ssaparser"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TODO: parsing types for use in <bool> <string>.
// TODO: aux creation (Lsym, figure of what types there are there).

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

func TestParseSSA(t *testing.T) {
	c := testConfig(t)
	var s = `
b1:
  (?) v1 = ConstBool <bool> [false]
  (?) v2 = ConstBool <bool> [false]
  If v1 -> b2 b3
b2: <- b1
  Plain -> b4
b3: <- b1
  Plain -> b4
b4: <- b2 b3
  (?) v3 = ConstBool <bool> [false]
  Exit v3
`

	_, s, _ = strings.Cut(s, "\n")
	t.Log(s)

	cc := TestFuncCfg{
		config: c.config,
		tb:     t,
	}

	f := cc.FromSSA(s)
	phiopt(f)
	t.Log(ssaparser.PrintFunc(fprintFunc2(f)))
}

type TestFuncCfg struct {
	config *Config
	tb     testing.TB
	fe     Frontend
}

func (c *TestFuncCfg) FromSSA(ssa string) *Func {
	c.tb.Helper()
	fun, err := ssaparser.ParseSSAFunc(ssa)
	if err != nil {
		c.tb.Fatal(err)
	}

	f := c.config.NewFunc(c.Frontend(), new(Cache))
	f.pass = &emptyPass
	f.cachedLineStarts = newXposmap(map[int]lineRange{0: {0, 100}, 1: {0, 100}, 2: {0, 100}, 3: {0, 100}, 4: {0, 100}})

	blocks := make(map[string]*Block)
	vals := make(map[string]*Value)

	for _, block := range fun.Blocks {
		c.tb.Log(block)
		blockKind, ok := blockKindNames[block.Kind]
		if _, ok := blocks[block.Name]; ok {
			c.tb.Fatalf("block: %v already declared", block.Name)
		}
		if !ok {
			c.tb.Fatalf("block: %v contains invalid kind: %q", block.Name, blockKind)
		}

		b := f.NewBlock(blockKind)
		blocks[block.Name] = b
		if f.Entry == nil {
			f.Entry = b
		}

		for _, val := range block.Values {
			c.tb.Log(val)
			if _, ok := vals[val.Name]; ok {
				c.tb.Fatalf("value: %v already declared", val.Name)
			}
			op, ok := opInfoNames[val.Op]
			if !ok {
				c.tb.Fatalf("block: %v contains value %v with invalid opcode: %q", block.Name, val.Name, val.Op)
			}
			typ := c.mapType(val.Type)
			if typ == nil {
				c.tb.Fatalf("block: %v contains value %v with invalid type: %q", block.Name, val.Name, val.Type)
			}
			intAux, err := c.intAux(val.AuxInt, typ)
			if err != nil {
				c.tb.Fatalf("block: %v %v", block.Name, err)
			}
			v := b.NewValue0(src.NoXPos, op.code, typ)
			vals[val.Name] = v
			v.AuxInt = intAux

			c.tb.Log(v.String())
		}
	}

	for _, block := range fun.Blocks {
		b := blocks[block.Name]
		if len(block.Controls) == 1 {
			b.Controls[0] = vals[block.Controls[0]]
			if b.Controls[0] == nil {
				c.tb.Fatalf("block: %v has a non-existing control val: %v", block.Name, block.Controls[0])
			}
		} else if len(block.Controls) == 2 {
			b.Controls[0] = vals[block.Controls[0]]
			b.Controls[1] = vals[block.Controls[0]]
			if b.Controls[0] == nil {
				c.tb.Fatalf("block: %v has a non-existing control val: %v", block.Name, block.Controls[0])
			}
			if b.Controls[1] == nil {
				c.tb.Fatalf("block: %v has a non-existing control val: %v", block.Name, block.Controls[1])
			}
		} else if len(block.Controls) != 0 {
			c.tb.Fatalf("block: %v has invalid number of controls: %v", block.Name, len(block.Controls))
		}

		for _, v := range block.Succ {
			succ, ok := blocks[v]
			if !ok {
				c.tb.Fatalf("block: %v has invalid succ: %v", block.Name, succ)
			}
			b.AddEdgeTo(succ)
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
					c.tb.Fatalf("value: %v has a non-existing argument: %v", val.Name, arg)
				}
				args = append(args, v)
			}
			v.AddArgs(args...)
		}
	}

	return f
}

func (c *TestFuncCfg) intAux(aux string, typ *types.Type) (int64, error) {
	switch typ {
	case c.config.Types.Bool:
		switch aux {
		case "true":
			return 1, nil
		case "false":
			return 0, nil
		default:
			return 0, fmt.Errorf("invalid aux %q for type %v", aux, typ)
		}
	default:
		panic("unreachable")
	}
}

func (c *TestFuncCfg) mapType(s string) *types.Type {
	switch s {
	case "mem":
		return types.TypeMem
	case "bool":
		return c.config.Types.Bool
	default:
		return nil
	}
}

func (c *TestFuncCfg) Frontend() Frontend {
	if c.fe == nil {
		pkg := types.NewPkg("my/import/path", "path")
		fn := ir.NewFunc(src.NoXPos, src.NoXPos, pkg.Lookup("function"), types.NewSignature(nil, nil, nil))
		fn.DeclareParams(true)
		fn.LSym = &obj.LSym{Name: "my/import/path.function"}
		c.fe = TestFrontend{
			t:    c.tb,
			ctxt: c.config.ctxt,
			f:    fn,
		}
	}
	return c.fe
}

type stringAuxRepr struct {
	AuxInt string
	Aux    string
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
