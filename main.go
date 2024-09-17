package main

import (
	"fmt"
	"sync"
	"time"
)

var waitingGroup sync.WaitGroup

type philosopher struct {
	id                int
	leftFork          chan bool
	rightFork         chan bool
	numberOfTimeEaten int
	requestEat        chan int
	startEat          chan int
	finishEat         chan int
}

/*
Each philosopher repeatedly attempts to eat until they’ve eaten three times. They request to eat via the requestEat channel. The host grants permission by sending 1 on the startEat channel. Once permission is granted, the philosopher picks up the left and right forks by sending a signal to their respective fork channels. After eating, the philosopher returns the forks by receiving from the fork channels. Finally, they notify the host via the finishEat channel. When not eating, the philosophers think.
*/
func (p philosopher) eat() {
	for {
		if p.numberOfTimeEaten == 3 {
			fmt.Printf("Philosopher %d has eaten %d meals\n", p.id, p.numberOfTimeEaten)
			waitingGroup.Done()
			return
		}

		p.requestEat <- 1

		if eatingAllowed := <-p.startEat; eatingAllowed == 1 {
			p.leftFork <- true
			p.rightFork <- true

			fmt.Printf("Philosopher %d is eating\n", p.id)
			time.Sleep(time.Second)
			p.numberOfTimeEaten++
			fmt.Printf("Philosopher %d is thinking\n", p.id)
			<-p.leftFork
			<-p.rightFork
			p.finishEat <- 1
		}
	}
}

func fork(forkChan chan bool) {
	for {
		<-forkChan
		forkChan <- true
	}
}

/*
Initializes the system by setting up philosophers and forks. It creates a channel for each fork, which is handled by the fork goroutines. Each philosopher is initialized with a left and right fork channel, and the host function is started concurrently to manage how many philosophers can eat at a time. Each philosopher’s eat() function runs as a goroutine. The sync.WaitGroup ensures the program waits for all philosophers to eat three times before terminating.
*/
func main() {
	count := 5
	waitingGroup.Add(count)

	forks := make([]chan bool, count)
	philosophers := make([]*philosopher, count)
	requestEat := make(chan int)
	startEat := make(chan int)
	finishEat := make(chan int)

	for i := 0; i < count; i++ {
		forks[i] = make(chan bool, 1)
		go fork(forks[i])
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

/*
The host function controls how many philosophers can eat simultaneously. It listens for signals from the philosophers via the requestEat channel. If fewer than two philosophers are eating, the host allows the philosopher to eat by sending 1 on the startEat channel and then increments the count of philosophers eating. Otherwise, it sends 0 to deny permission. The finishEat channel decrements the count of philosophers eating when they finish their meal.
*/
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
The system avoids deadlock by limiting the number of philosophers allowed to eat simultaneously. The host function ensures that at most two philosophers can eat at the same time. This prevents the system from deadlocking, since only two philosophers can eat at once, and therefore at least one pair of forks is always free for the next philosopher to eat.
*/
