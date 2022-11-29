# dissys_mandatory_5_forreal

To start the program:
The program needs 3 servers (replicas) to start. To start them, while in the directory of the handin, please write:
"Go run server.go 0", "Go run server.go 1" and "Go run server.go 2" in 3 separate terminals (specifying the ports, port 9000, 9001, 9002)

This starts all the replicas. Bidder(s) can now be started, simply write "Go run bidder.go". Instructions on how to use the program will
then be printed in the terminal. The user will be prompted to write their username.

To simulate the program being resilient to a crash, one of the terminals running one of the servers can simply be closed, either by just
closing the terminal or pressing Ctrl+C in the terminal. 

If another auction needs to be started, the servers needs to be closed and started again. No functionality to restart auctions have been implemented.
