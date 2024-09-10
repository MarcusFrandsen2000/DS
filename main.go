package main

import {
	"fmt"
	"sync"
	"time"
}

type fork struct {
	sync.Mutex
}

type philosopher struct {
	id int
	leftFork, rightFork * fork
}

func (p philosopher) eat () {
	p.leftFork.Lock()
	p.rightFork.Lock()

	fmt.Printf("Philosopher", p.id, "is eating")
	time.Sleep(10*time.Second)

	p.leftFork.Unlock()
	p.rightFork.Unlock()

	fmt.Printf("Philosopher", p.id, "is finished eating")
	time.Sleep(10 * time.Second)
}

func main() {
	count := 5

	forks := make([] * fork, count)

	for i := 0; i < count; i++ {
		forks[i] = new(fork)
	}

	philosophers := make([] * philosopher, count)

	for i := 0; i < count; i++ {
		philosophers[i] = &philosopher {
			id: i, leftFork: forks[i], rightFork: forks[(i+1) % 5],
		}
		go philosophers[i].eat()
	}
}