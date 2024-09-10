package main

import (
	"fmt"
	"sync"
	"time"
)

var waitingGroup sync.WaitGroup

type fork struct {
	sync.Mutex
}

type philosopher struct {
	id                  int
	leftFork, rightFork *fork
	numberOfTimeEaten   int
	requestEat          chan int
	startEat            chan int
	finishEat           chan int
}

func (p philosopher) eat() {
	for {
		if p.numberOfTimeEaten == 3 {
			fmt.Printf("Philosopher %d has eaten %d meals\n", p.id, p.numberOfTimeEaten)
			waitingGroup.Done()
			return
		}

		p.requestEat <- 1

		if eatingAllowed := <-p.startEat; eatingAllowed == 1 {
			p.leftFork.Lock()
			p.rightFork.Lock()

			fmt.Printf("Philosopher %d is eating\n", p.id)
			time.Sleep(time.Second)
			p.numberOfTimeEaten++
			fmt.Printf("Philosopher %d is thinking\n", p.id)
			p.leftFork.Unlock()
			p.rightFork.Unlock()

			p.finishEat <- 1
		}
	}
}

func main() {
	count := 5
	waitingGroup.Add(count)

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
			id:                i + 1,
			leftFork:          forks[i],
			rightFork:         forks[(i+1)%5],
			numberOfTimeEaten: 0,
			requestEat:        requestEat,
			startEat:          startEat,
			finishEat:         finishEat,
		}
	}

	go host(requestEat, startEat, finishEat)
	for _, p := range philosophers {
		go p.eat()
	}

	waitingGroup.Wait()
	fmt.Printf("All philosophers have eaten")
}

func host(requestEat, startEat, finishEat chan int) {
	numberOfPeopleEating := 0

	defer func() {
		return
	}()
	for {
		select {
		case incrementAmount := <-requestEat:
			if numberOfPeopleEating < 2 {
				numberOfPeopleEating = numberOfPeopleEating + incrementAmount
				startEat <- 1
			} else {
				startEat <- 0
			}
		case decrementAmount := <-finishEat:
			numberOfPeopleEating = numberOfPeopleEating - decrementAmount
		}
	}
}

/*
The host function limits the number of concurrent eaters, which prevents contention and possible deadlock.

Channels are used for synchronization, ensuring that philosophers can only proceed when allowed.

The numberOfPeopleEating counter accurately reflects the number of philosophers eating, and forks are managed to avoid deadlock.
*/
