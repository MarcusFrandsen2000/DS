package main

import (
	"fmt"
	"sync"
	"time"
)

type fork struct {
	sync.Mutex
}

type philosopher struct {
	id                  int
	leftFork, rightFork *fork
}

func (p philosopher) eat() {
	p.leftFork.Lock()
	p.rightFork.Lock()

	fmt.Printf("Philosopher", p.id, "is eating")
	time.Sleep(10 * time.Second)

	p.leftFork.Unlock()
	p.rightFork.Unlock()

	fmt.Printf("Philosopher", p.id, "is finished eating")
	time.Sleep(10 * time.Second)
}

func main() {
	count := 5

	//creating an array that has got specific fork addresses, and has the length of the counter
	forks := make([]*fork, count)
	//creating the forks and inserting them in to the full array
	for i := 0; i < count; i++ {
		forks[i] = new(fork)
	}

	//creating an array for the philosophers, pointing to an exact address to make sure, we've got the same philosopher object every time.
	philosophers := make([]*philosopher, count)
	//creating the philosophers, and assigning them to two forks each.
	for i := 0; i < count; i++ {
		philosophers[i] = &philosopher{
			id: i, leftFork: forks[i], rightFork: forks[(i+1)%5],
		}

		go philosophers[i].eat()
	}
}
