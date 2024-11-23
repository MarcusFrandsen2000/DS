package main

import(
	"log"
)

func main(){
	for i := 0; i < 5; i++ {
		clientID := int32(i + 1) // Explicitly cast i + 1 to int32
		client, err := NewClient(clientID, "localhost:50051")

		if err != nil {
			// Handle the error
			log.Println("Error creating client:", err)
			continue
		}
	}

}