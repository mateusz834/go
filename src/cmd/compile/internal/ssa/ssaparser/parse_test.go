package ssaparser

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const test = `
b1:
  (?) v1 = ConstBool <bool> [false]
  (?) v2 = InitMem <mem> v1 v1
  (?) v2 = InitMem <mem> v1 v1
  If v1 -> b2 b3
b2: <- b1
  (?) v22 = Phi <int8> v1 v16
  (?) v2 = InitMem <mem> v1 v1
  Plain -> b4
b3: <- b1
  (?) v2 = InitMem <mem> v1 v1
  Plain -> b4
b4: <- b2 b3
  (?) v2 = InitMem <mem> [1] {main} v1 v1
  Exit v2
`

func TestParse(t *testing.T) {
	const testdir = "./testdata/"
	files, err := os.ReadDir(testdir)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".gossa" {
			continue
		}
		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join(testdir, file.Name())
			c, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			ssa, expect, ok := bytes.Cut(c, []byte("======\n"))
			f, err := ParseSSAFunc(string(ssa))
			if err != nil {
				t.Fatal(err)
			}

			got, err := json.MarshalIndent(f, "", "    ")
			if err != nil {
				t.Fatal(err)
			}

			if !ok {
				// Test file has only the SSA part, generate the expect part.
				if err := os.WriteFile(path, slices.Concat(ssa, []byte("======\n"), got), 0666); err != nil {
					t.Fatal(err)
				}
				return
			}

			if !bytes.Equal(expect, got) {
				t.Fatalf("got:\n%v\nwant:\n%v", got, expect)
			}
		})
	}

	f, err := ParseSSAFunc(test)
	if err != nil {
		t.Fatal(err)
	}

	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", out)
}
