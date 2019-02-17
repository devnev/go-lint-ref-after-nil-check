package testdata

import "fmt"

func printNilErrNolint() {
	var err error
	if err != nil {
		return
	}
	// This file is "bad" because nolint is not handled by the Check() method.
	// nolint
	fmt.Println(err)
}
