package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/amin-mir/glb"
)

func main() {
	backend1, err := url.Parse("https://localhost:3000")
	if err != nil {
		panic(err)
	}

	backend2, err := url.Parse("https://localhost:3001")
	if err != nil {
		panic(err)
	}

	backend3, err := url.Parse("https://localhost:3002")
	if err != nil {
		panic(err)
	}

	lb, err := glb.NewLeastConns([]*url.URL{backend1, backend2, backend3})
	if err != nil {
		panic(err)
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	go func() {
		for i := 0; i < 10; i++ {
			b, err := lb.Next(ctx1)
			if err != nil {
				fmt.Println("Error on calling lb.Next:", err)
			}
			_ = b
		}
	}()

	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		for i := 0; i < 10; i++ {
			b, err := lb.Next(ctx2)
			if err != nil {
				fmt.Println("Error on calling lb.Next:", err)
			}
			_ = b
		}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel1()
	cancel2()
	time.Sleep(50 * time.Millisecond)
}
