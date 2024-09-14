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
eat allows a philosopher to request to eat, wait for permission, lock the forks, eat, and then release the forks.

- If the philosopher has eaten 3 times, it prints a message and signals completion.
- Otherwise, it requests permission to eat.
- If permitted, it locks the left and right forks, then eats, increments the count of meals eaten, prints status messages, and unlocks the forks.
- Finally, it signals that it has finished eating.
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
main initializes the system with philosophers and forks, sets up communication channels, and starts the host and philosopher goroutines.

- It creates a specified number of forks and philosophers.
- Initializes channels for communication between philosophers and the host.
- Starts the host goroutine to manage eating requests and completions.
- Starts goroutines for each philosopher to simulate their eating and thinking process.
- Waits for all philosophers to complete their eating before printing a final message.
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
host manages the synchronization of eating requests from philosophers.

- It tracks the number of philosophers currently eating with the `numberOfPeopleEating` counter.
- Processes requests from philosophers to start eating. Allows up to 2 philosophers to eat simultaneously.
- Updates the count of eating philosophers based on requests and finishes.
- Sends signals to philosophers about whether they can start eating or not.
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
The host function limits the number of concurrent eaters, which prevents contention and possible deadlock.

Channels are used for synchronization, ensuring that philosophers can only proceed when allowed.

The numberOfPeopleEating counter accurately reflects the number of philosophers eating, and forks are managed to avoid deadlock.
*/
