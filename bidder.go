package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jskoven/dissys_mandatory_5_forreal/replication"
	"google.golang.org/grpc"
)

type bidder struct {
	name string
	replication.ReplicationClient
	replicas map[int32]replication.ReplicationClient
}

func main() {

	log.Println("Please insert username:")
	Scanner := bufio.NewScanner(os.Stdin)
	Scanner.Scan()
	username := Scanner.Text()

	b := &bidder{
		replicas: make(map[int32]replication.ReplicationClient),
		name:     username,
	}
	for i := 0; i < 3; i++ {
		port := int32(9000) + int32(i)
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}

		defer conn.Close()

		c := replication.NewReplicationClient(conn)
		b.replicas[port] = c
		log.Printf("Client connected to port %d \n", port)
	}

	for {
		Scanner := bufio.NewScanner(os.Stdin)
		Scanner.Scan()
		MessageToBeSent := Scanner.Text()

		switch MessageToBeSent {
		case "result":
			empty := replication.Empty{}
			ap := replication.ResultPackage{}
			for _, replica := range b.replicas {
				if replica != nil {
					resultPointer, _ := replica.Result(context.Background(), &empty)
					ap.Highestbid = resultPointer.Highestbid
					ap.Winner = resultPointer.Winner
				}
			}
			if ap.Winner != "None have won yet!" {
				log.Printf("Auction has been closed. The winner is %s with a bid of %d!", ap.Winner, ap.Highestbid)
			} else {
				log.Printf("Current highest bid is %d from user %s", ap.Highestbid, ap.Winner)
			}
		case "bid":
			log.Println("How much do you wish to bid?")
			Scanner.Scan()
			toBid := Scanner.Text()
			toBidInInt, err := strconv.Atoi(toBid)
			if err != nil {
				log.Println("Bid failed, please try again with a whole number")
				continue
			}
			b.bid(toBidInInt)
		case "exit":
			log.Printf("Leaving auction...")
			break
		}

	}

}

func (b *bidder) bid(bid int) {

	confPackage := replication.Confirmation{}
	bp := replication.BidPackage{
		Bid:    int32(bid),
		Bidder: b.name,
	}
	for _, element := range b.replicas {
		fmt.Println("Looped")
		if element != nil {
			conf, err := element.Receivebid(context.Background(), &bp)
			if err != nil {

			} else {
				confPackage.Confirmation = conf.Confirmation
				confPackage.CurrentPrice = conf.CurrentPrice
				confPackage.CurrentWinner = conf.CurrentWinner
			}
		}
	}

	switch confPackage.Confirmation {
	case true:
		log.Printf("%s has bid %d \n", b.name, bid)
	case false:
		log.Printf("Bid not high enough, current highest bid is: %d from user %s", confPackage.CurrentPrice, confPackage.GetCurrentWinner)
	}
}

func (b *bidder) result() replication.ResultPackage {
	empty := replication.Empty{}
	result := replication.ResultPackage{}

	for _, element := range b.replicas {
		fmt.Println("Looped")
		answer, err := element.Result(context.Background(), &empty)
		if err != nil {
			fmt.Println("yo wtf")
		}
		result.Highestbid = answer.Highestbid
		result.Winner = answer.Winner
	}

	return result
}
