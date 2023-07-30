package main

import (
	"fmt"
	"time"
)

func main() {
	// 使用方法
	mist := NewMist()
	now := time.Now()
	for i := 0; i < 10000000; i++ {
		mist.Generate()
	}
	fmt.Println(time.Since(now).Milliseconds())
}

/*
1000000 642 639 627
5000000 3060 3100 3131
10000000 6170 6174 6189
*/
