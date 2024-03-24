package utils

import (
	"fmt"
	"time"
)

func MustError(err error) {
	if err != nil {
		panic(err)
	}
}

func ShowIfError(err error) {
	if err != nil {
		fmt.Printf("%s:%s", time.Now().Format(time.DateTime), err.Error())
	}
}
