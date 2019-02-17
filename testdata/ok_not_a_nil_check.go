package testdata

import "fmt"

func notANilCheck() {
	var err, err2 error
	if err != err2 {
		return
	}
	fmt.Println(err, err2)
}
