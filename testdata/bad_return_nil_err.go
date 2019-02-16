package testdata

func returnNilErr() error {
	var err error
	if err != nil {
		return err
	}
	return err
}
