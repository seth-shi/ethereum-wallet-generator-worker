package utils

func MustError(err error) {
	if err != nil {
		panic(err)
	}
}
