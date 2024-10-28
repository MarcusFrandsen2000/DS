HOW TO RUN OUR CHITTYCHAT PROGRAM: 

1. Clone the repository and select the branch called "mandatory-activity-3". 
2. Open the program and cd to the Server directory. 
3. From the Server directory, run the command: go run chittychatserver.go 
4. This command will open 3 terminal windows (clients) where the following command will run automatically "go run chittychatclient.go <username>". 
<username> are different from each terminal specified as User1, User2 and User3. 
5. Now the clients can write messages through the chittychat, and they will be broadcasted in each of the terminal windows. 
6. If a client wants to exit the program, it can do so by writing "exit", and a message about the client leaving the chittychat will be broadcasted. 