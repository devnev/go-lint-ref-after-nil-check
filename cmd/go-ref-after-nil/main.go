package main

import (
	"flag"
	"fmt"
	"github.com/devnev/go-lint-ref-after-nil-check"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var opts struct {
	verbose bool
	machine bool
	fix bool
	exit bool
	source int
}

func main() {
	parseFlags()
	var haveFails bool
	for _, path := range flag.Args() {
		inf, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Failed to stat %s: %v", path, err)
		}
		fset := token.NewFileSet()
		if inf.IsDir() {
			pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
			if err != nil {
				log.Fatalf("Failed to parse %s: %v", path, err)
			}
			var files []*ast.File
			for _, pkg := range pkgs {
				for _, file := range pkg.Files {
					files = append(files, file)
				}
			}
			sort.Slice(files, func(i, j int) bool {
				return fset.Position(files[i].Pos()).Filename < fset.Position(files[j].Pos()).Filename
			})
			for _, file := range files {
				if checkFile(fset, file) {
					haveFails = true
				}
			}
		} else {
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				log.Fatalf("Failed to parse %s: %v", path, err)
			}
			if checkFile(fset, file) {
				haveFails = true
			}
		}
	}
	if opts.verbose {
		if haveFails {
			log.Printf("Check failed on some inputs.")
		} else {
			log.Printf("Checks passed on all inputs.")
		}
	}
	if haveFails && opts.exit {
		os.Exit(2)
	}
}

func parseFlags() {
	flag.BoolVar(&opts.verbose, "verbose", false, "If set, log every file as it is checked to stderr.")
	flag.BoolVar(&opts.machine, "machine", false, "If set, limit output to machine-readable file:line:col format.")
	flag.BoolVar(&opts.fix, "fix", false, `If set, replace bad references with "nil" in input files.`)
	flag.BoolVar(&opts.exit, "exit-code", false, "If set, exit with code 2 if there were failures. Exit 0 by default.")
	flag.IntVar(&opts.source, "source-context", 2, `If >= 0 (and -machine is unset), print the offending line of source code and given number of lines of context`)
	flag.Parse()
}

func checkFile(fset *token.FileSet, file *ast.File) bool {
	if opts.verbose {
		log.Printf("Checking %s", fset.Position(file.Pos()).Filename)
	}

	fails := nilref.Check(file)
	fails = removeSkipped(fails, fset, file)

	lastFile, lastLine := "", -1
	for _, fail := range fails {
		pos := fset.Position(fail.Pos())
		if opts.machine {
			fmt.Println(pos.String())
			continue
		}
		if lastFile == pos.Filename && lastLine == pos.Line {
			continue
		}
		fmt.Printf("Referenced identifier %s will always be nil at %s\n", fail.Name, pos.String())
		if opts.source >= 0 {
			printSource(pos, opts.source)
		}
	}

	if opts.fix {
		nilref.Fix(fails)
		fname := fset.Position(file.Pos()).Filename
		out, err := ioutil.TempFile(filepath.Dir(fname), filepath.Base(fname))
		if err != nil {
			log.Fatalf("Failed to create replacement file for %q: %s", fname, err.Error())
		}
		err = format.Node(out, fset, file)
		if err != nil {
			log.Fatalf("Failed to write fixed source to %s: %s", out.Name(), err.Error())
		}
		err = os.Rename(out.Name(), fname)
		if err != nil {
			log.Fatalf("Failed to replace old file %q with fixed file %q: %s", fname, out.Name(), err.Error())
		}
	}

	return len(fails) > 0
}

func printSource(pos token.Position, lines int) {
	f, err := os.Open(pos.Filename)
	if err != nil {
		log.Printf("Failed to re-open source for %s: %s", pos.Filename, err.Error())
		return
	}
	start := pos.Offset - 1024
	if start < 0 {
		start = 0
	}
	buf := make([]byte, 4*1024)
	n, err := f.ReadAt(buf, int64(start))
	if err != nil && err != io.EOF {
		log.Printf("Failed to re-read source for %s: %s", pos.Filename, err.Error())
		return
	}
	buf = buf[:n]
	lineStart, lineEnd := pos.Offset-start, pos.Offset-start
	for i := 0; i <= lines && lineStart >= 0; i++ {
		lineStart = strings.LastIndex(string(buf[:lineStart]), "\n")
	}
	for i := 0; i <= lines && lineEnd >= 0; i++ {
		lineEnd = lineEnd + 1 + strings.Index(string(buf[lineEnd+1:]), "\n")
	}
	if lineEnd > 0 {
		buf = buf[:lineEnd]
	}
	buf = buf[lineStart+1:]
	fmt.Println(string(buf))
}

func removeSkipped(fails []*ast.Ident, fset *token.FileSet, file *ast.File) []*ast.Ident {
	skipLines := map[int]struct{}{}
	for _, cg := range file.Comments {
		last := cg.List[len(cg.List)-1]
		if last.Text == "// nolint" {
			pos := fset.Position(last.Pos())
			skipLines[pos.Line+1] = struct{}{}
		}
	}

	next := 0
	for _, fail := range fails {
		pos := fset.Position(fail.Pos())
		if _, ok := skipLines[pos.Line]; ok {
			continue
		}
		fails[next] = fail
		next++
	}

	return fails[:next]
}
