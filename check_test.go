package nilref

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCheck_withTestData(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "testdata", nil, 0)
	if err != nil {
		t.Fatalf("Failed to parse test data: %s", err)
	}
	if len(pkgs) > 1 {
		t.Fatalf("Got multiple packages from testdata. Expected 1, got %d", len(pkgs))
	}
	var pkg *ast.Package
	for _, pkg = range pkgs {
	}
	for fname, f := range pkg.Files {
		t.Run(fname, func(t *testing.T) {
			fails := Check(f)
			for i, failure := range fails {
				var b strings.Builder
				err := ast.Fprint(&b, fset, failure, nil)
				if err != nil {
					t.Fatalf("failed to format failure: %s", err)
				}
				t.Logf("failure %d for %s:\n%s", i, fname, b.String())
			}
			if strings.HasPrefix(fname, "testdata/bad") {
				if len(fails) == 0 {
					t.Errorf("Expected failures on testdata file %s, got zero failures", fname)
				}
			} else if strings.HasPrefix(fname, "testdata/ok") {
				if len(fails) != 0 {
					t.Errorf("Unexpected failures on testdata file %s. Expected zero, got %d", fname, len(fails))
				}
			} else {
				t.Fatalf("unexpected filename %q, missing prefix bad/ok", fname)
			}
		})
	}
}
