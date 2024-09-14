package main

import (
	"fmt"
	"sync"
	"time"
)

var waitingGroup sync.WaitGroup

type philosopher struct {
	id                int
	leftFork          chan int
	rightFork         chan int
	numberOfTimeEaten int
	requestEat        chan int
	startEat          chan int
	finishEat         chan int
}

/*
Each philosopher repeatedly attempts to eat until they’ve eaten three times. They request to eat via the requestEat channel. The host grants permission by sending 1 on the startEat channel, the philosopher picks up the forks if and only if a message is stored in the channel. The philospher then eats and then returns the forks (sending back to the fork channels). After finishing, they notify the host via the finishEat channel. When not eating the philosophers think.
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
			<-p.leftFork
			<-p.rightFork

			fmt.Printf("Philosopher %d is eating\n", p.id)
			time.Sleep(time.Second)
			p.numberOfTimeEaten++
			fmt.Printf("Philosopher %d is thinking\n", p.id)
			p.leftFork <- 1
			p.rightFork <- 1
			p.finishEat <- 1
		}
	}
}

/*
Initializes the system by setting up the philosophers and forks, creating buffered channels for each fork. It creates the philosophers, each with its own left and right fork channels. The host function is run concurrently to manage the dining process, while each philosopher’s eat() function is executed as a goroutine. The sync.WaitGroup ensures the program waits for all philosophers to eat three times before terminating.
*/
func main() {
	count := 5
	waitingGroup.Add(count)

	forks := make([]chan int, count)
	philosophers := make([]*philosopher, count)
	requestEat := make(chan int)
	startEat := make(chan int)
	finishEat := make(chan int)

	for i := 0; i < count; i++ {
		forks[i] = make(chan int, 1)
		forks[i] <- 1
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
