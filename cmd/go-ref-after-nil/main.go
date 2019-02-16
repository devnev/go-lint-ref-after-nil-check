package main

import (
	"flag"
	"fmt"
	"go-err-after-nil"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

var verbose = flag.Bool("verbose", false, "If set, print every file as it is checked.")

func main() {
	flag.Parse()
	var haveFails bool
	for _, path := range flag.Args() {
		inf, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Failed to stat %s: %v", path, err)
		}
		fset := token.NewFileSet()
		if inf.IsDir() {
			pkgs, err := parser.ParseDir(fset, path, nil, 0)
			if err != nil {
				log.Fatalf("Failed to parse %s: %v", path, err)
			}
			for _, pkg := range pkgs {
				for _, file := range pkg.Files {
					if checkFile(fset, file) {
						haveFails = true
					}
				}
			}
		} else {
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				log.Fatalf("Failed to parse %s: %v", path, err)
			}
			if checkFile(fset, file) {
				haveFails = true
			}
		}
	}
	if *verbose {
		if haveFails {
			log.Printf("Check failed on some inputs.")
		} else {
			log.Printf("Checks passed on all inputs.")
		}
	}
	if haveFails {
		os.Exit(1)
	}
}

func checkFile(fset *token.FileSet, file *ast.File) bool {
	if *verbose {
		log.Printf("Checking %s", fset.Position(file.Pos()).Filename)
	}
	fails := nilref.Check(file)
	for _, fail := range fails {
		pos := fset.Position(fail.Pos())
		fmt.Printf("Reference after nil check at %s", pos.String())
	}
	return len(fails) > 0
}
