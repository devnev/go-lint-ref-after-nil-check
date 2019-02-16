package testdata

import "fmt"

func printNilErr() {
	var err error
	if err != nil {
		return
	}
	fmt.Println(err)
}
