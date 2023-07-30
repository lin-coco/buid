package buid_test

import (
	"buid"
	"fmt"
	_ "testing"
)

func ExampleBUID() {
	buidGenerator, err := buid.NewDefaultBUID()
	if err != nil {
		panic(err)
	}
	buidGenerator.Start()

	ids := buidGenerator.Gets(10000000)
	buidGenerator.Exit()
	i := 1
	for i <= 100 {
		fmt.Println(ids[10000000-i])
		i++
	}
	fmt.Println(len(ids))
	// Output:
}
