package buid_test

import (
	"buid"
	"testing"
)

func BenchmarkBUID(b *testing.B) {
	buidGenerator, err := buid.NewDefaultBUID()
	if err != nil {
		b.Fatal(err)
	}
	buidGenerator.Start()

	for i := 0; i < b.N; i++ {
		_ = buidGenerator.GetUID()
	}
}
