package testdata

import "fmt"

func printAssignedErr() {
	var err error
	if err != nil {
		return
	}
	err = fmt.Errorf("boo")
	fmt.Println(err)
}
