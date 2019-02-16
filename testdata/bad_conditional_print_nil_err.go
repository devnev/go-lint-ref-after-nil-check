package testdata

import "fmt"

func printNilErrConditionally() {
	var err error
	var expected bool
	if err != nil {
		return
	}
	if !expected {
		fmt.Println(err)
	}
}
