package buid_test

import (
	"buid"
	"fmt"
	"testing"
	"time"
)

func TestBuid(t *testing.T) {
	buidGenerator, err := buid.NewBUID(&buid.Config{
		Bits: [3]uint{40, 12, 11},
	})
	if err != nil {
		t.Fatal(err)
	}
	buidGenerator.Start()
	for i := 0; i < 100; i++ {
		fmt.Println(buidGenerator.GetUID())
	}

	buidGenerator.Exit()
}

func TestComputer(t *testing.T) {
	a := time.Now().UnixMilli()
	b := int64(1577808000000)
	fmt.Println(a - b)
}
