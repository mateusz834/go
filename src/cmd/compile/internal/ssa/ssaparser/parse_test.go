package ssaparser

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

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
			got = append(got, '\n')

			if !ok {
				// Test file has only the SSA part, generate the expect part.
				if err := os.WriteFile(path, slices.Concat(ssa, []byte("======\n"), got), 0666); err != nil {
					t.Fatal(err)
				}
				return
			}

			if !bytes.Equal(expect, got) {
				t.Errorf("got:\n%s\nwant:\n%s", got, expect)
				//t.Fatalf("diff:\n%s", diff.Diff("expect", expect, "got", got))
			}
		})
	}
}
