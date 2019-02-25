package testdata

import "fmt"

func goAndDeferPrintNilErr() error {
	var err error
	if err != nil {
		return err
	}
	defer func() {
		fmt.Println(err)
	}()
	go func() {
		fmt.Println(err)
	}()
	return nil
}
