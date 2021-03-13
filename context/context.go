package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	minLatency = 10
	maxLatency = 5000
	timeout    = 3000
)

func main() {
	// small program that searches flight routes
	// we are going to use a mock backend/database

	// the purpose of this is to show how the context can be used to propaate
	// cancellation signals across go routines
	rootCtx := context.Background()
	ctxWithTimeout, cancel := context.WithTimeout(rootCtx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// listen for interrupt signal
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig

		// NOW cancel
		fmt.Println("aborting due to interrupt...")
		cancel()
	}()

	res, err := Search(ctxWithTimeout, "nyc", "london")
	if err != nil {
		fmt.Println("got error", err)
		return
	}
	fmt.Println("got results: ", res)
}

func Search(ctx context.Context, from, to string) ([]string, error) {
	// slowSearch
	// watch for when ctx.Done() is closed
	result := make(chan []string)
	go func() {
		r, _ := slowSearch(from, to)
		result <- r
		close(result)
	}()

	// wait for 2 events: either of one will be the result
	for {
		select {
		case dst := <-result:
			return dst, nil
		case <-ctx.Done():
			return []string{}, ctx.Err()
		}
	}

	return []string{}, nil

}

// SlowSearch is a very slow function
func slowSearch(from, to string) ([]string, error) {
	// sleep for a random period b/w 10 and 5000 ms

	rand.Seed(time.Now().Unix())
	latency := rand.Intn(maxLatency-minLatency+1) - minLatency
	time.Sleep(time.Duration(latency) * time.Millisecond)
	fmt.Printf("started to SlowSearch for %s-%s takes %dms...", from, to, latency)
	return []string{from + "-" + to + "-british airways-11am", from + "-" + to + "-delta airlines-12am"}, nil
}
