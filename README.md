HOW TO RUN OUR AUCTION PROGRAM: 

1. Clone the repository and select the branch called "mandatory-activity-5". 
2. Open the program and cd to the program directory. 
3. From the program directory, run the command: go run Program.go 
4. This command will initiate the program.
4a. It will firstly initiate the Backup Server (If you are on windows you have to press "Allow").
4b. It will secondly initiate the Primary Server (If you are on windows you have to press "Allow").
4c. It will finally initiate 4 seperate terminals, which each represents a client who can bid at the auction.
5. Logs will show the flow of the auction.
6. Go to the terminal responsible for the Primary Server and press "ctrl+c". This will terminate the Server and force the Backup Server to take over.
7. The Backup Server will automatically be promoted to Primary Server and the auction and logs will continue from where it was before the crash happened.

Side note:
Step 6 and 7 can work vice versa as the Primary Server can handle if the Backup Server should crash.