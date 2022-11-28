package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jskoven/dissys_mandatory_5_forreal/replication"
	"google.golang.org/grpc"
)

type bidder struct {
	name string
	replication.ReplicationClient
	replicas map[int32]replication.ReplicationClient
}

func main() {
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Eror on opening file: %s", err)
	}
	log.SetOutput(f)
	fmt.Println("Welcome to the auction!")
	fmt.Println("To see results of current auction, simple type \"result\"")
	fmt.Println("To bid on current auction, simple type \"bid\", press enter and then enter your amount")
	fmt.Println("To start, please write your username:")

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
		log.Printf("Client %s connected to port %d \n", b.name, port)
		fmt.Println()
	}

	for {
		Scanner := bufio.NewScanner(os.Stdin)
		Scanner.Scan()
		MessageToBeSent := strings.ToLower(Scanner.Text())

		switch MessageToBeSent {
		case "result":
			result := b.result()
			if result.Highestbid == 0 {
				fmt.Printf("No one has bid on the auction yet, and the timer hasn't started.\n")
			} else if result.HasEnded {
				fmt.Printf("Auction has been closed. The winner is %s with a bid of %d! \n", result.Winner, result.Highestbid)
			} else {
				fmt.Printf("Current highest bid is %d from user %s \n", result.Highestbid, result.Winner)
			}
		case "bid":
			fmt.Println("How much do you wish to bid?")
			Scanner.Scan()
			toBid := Scanner.Text()
			toBidInInt, err := strconv.Atoi(toBid)
			if err != nil || toBidInInt == 0 {
				fmt.Println("Bid failed, please try again with a whole number")
				continue
			}
			b.bid(toBidInInt)
		case "exit":
			fmt.Println("Leaving auction...")
			break
		}

	}

}

func (b *bidder) bid(bid int) {

	log.Printf("User %s attempting to bid with bid %d\n", b.name, bid)
	confPackage := replication.Confirmation{}
	bp := replication.BidPackage{
		Bid:    int32(bid),
		Bidder: b.name,
	}
	for index, element := range b.replicas {
		if element != nil {
			conf, err := element.Receivebid(context.Background(), &bp)
			if err != nil {
				log.Printf("## Replica number %d is down, skipping it.##\n", index)
			} else {
				confPackage.Confirmation = conf.Confirmation
				confPackage.CurrentPrice = conf.CurrentPrice
				confPackage.CurrentWinner = conf.CurrentWinner
				confPackage.HasEnded = conf.HasEnded
			}
		}
	}
	if !confPackage.HasEnded {
		switch confPackage.Confirmation {
		case true:
			log.Printf("User %s has succesfully bid %d \n", b.name, bid)
			fmt.Printf("You have succesfully bid %d \n", bid)
		case false:
			fmt.Printf("Bid not high enough, current highest bid is: %d from user %s\n", confPackage.CurrentPrice, confPackage.CurrentWinner)
			log.Printf("User %s attempted to bid %d, but bid was not high enough. \n", b.name, bid)
		}
	} else {
		fmt.Println("Auction has ended! See results for more information.")
	}
}

func (b *bidder) result() replication.ResultPackage {
	log.Printf("User %s attempting to request result from replicas.", b.name)
	empty := replication.Empty{}
	result := replication.ResultPackage{}

	for index, element := range b.replicas {
		answer, err := element.Result(context.Background(), &empty)
		if err != nil {
			log.Printf("## Replica number %d is down, skipping it.##\n", index)
		} else {
			result.Highestbid = answer.Highestbid
			result.Winner = answer.Winner
			result.HasEnded = answer.HasEnded
		}
	}

	return result
}
