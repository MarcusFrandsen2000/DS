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
	eats                int
	requestEat          chan int
	startEat            chan int
	finishEat           chan int
}

var gw sync.WaitGroup

func (p philosopher) eat() {
	if p.eats == 3 {
		fmt.Printf("Philosopher %d has eaten %d meals\n", p.id, p.eats)
		gw.Done()
		return
	}

	p.requestEat <- p.id

	if eatingAllowed := <-p.startEat; eatingAllowed == 1 {
		p.leftFork.Lock()
		p.rightFork.Lock()

		fmt.Printf("Philosopher %d is eating\n", p.id)
		time.Sleep(3 * time.Second)
		p.eats++
		fmt.Printf("Philosopher %d is finished eating\n", p.id)
		p.leftFork.Unlock()
		p.rightFork.Unlock()

		p.finishEat <- p.id
	}
}

func main() {
	count := 5
	gw.Add(count)

	forks := make([]*fork, count)
	philosophers := make([]*philosopher, count)
	requestEat := make(chan int)
	startEat := make(chan int)
	finishEat := make(chan int)

	for i := 0; i < count; i++ {
		forks[i] = new(fork)
	}

	for i := 0; i < count; i++ {
		philosophers[i] = &philosopher{
			id:         i,
			leftFork:   forks[i],
			rightFork:  forks[(i+1)%5],
			eats:       0,
			requestEat: requestEat,
			startEat:   startEat,
			finishEat:  finishEat,
		}
	}

	go host(requestEat, startEat, finishEat)
	for _, p := range philosophers {
		go p.eat()
	}

	gw.Wait()
	fmt.Printf("All philosophers have eaten")
}

func host(requestEat, startEat, finishEat chan int) {
	philosopherID := make(map[int]string)

	defer func() {
		fmt.Print("Hi")
		return
	}()
	for {
		select {
		case id := <-requestEat:
			if len(philosopherID) < 2 {
				philosopherID[id] = "Eating"
				startEat <- 1
			} else {
				startEat <- 0
			}
		case id := <-finishEat:
			fmt.Print("What up?\n")
			delete(philosopherID, id)
		}
	}
}
