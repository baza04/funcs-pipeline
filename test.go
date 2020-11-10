package main

import (
	"fmt"
	"strings"
)

func main() {
	arr := make([]string, 3)
	f1(arr)
	fmt.Println(strings.Join(arr, ""))
}

func f1(arr []string) {
	r := '2'
	for i := 0; i < 3; i++ {
		arr[i] = string(r)
		r--
	}
}
