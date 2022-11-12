package common

// PanicErr panics if error is not nil
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
