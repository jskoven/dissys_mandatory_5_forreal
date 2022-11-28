package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jskoven/dissys_mandatory_5_forreal/replication"

	"google.golang.org/grpc"
)

var (
	timestamp time.Time
	timeLimit time.Time
)

func main() {
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Eror on opening file: %s", err)
	}
	log.SetOutput(f)
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 9000
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	log.Printf("Starting server...")

	if err != nil {
		log.Printf("Failed to listen on port :%d, Error: %s", ownPort, err)
	}

	log.Printf("Listening on port :%d", ownPort)

	grpcserver := grpc.NewServer()

	server := Biddinghouse{}
	server.currentbid = 0
	server.winner = "none yet"
	server.id = ownPort
	server.bidders = make(map[string]string)
	replication.RegisterReplicationServer(grpcserver, &server)

	err = grpcserver.Serve(listener)
	if err != nil {
		log.Printf("Replica #%d:  Failed to serve with listener, error: %s\n", server.id, err)

	}
	timeLimit = timestamp.Add(time.Duration(500) * time.Minute)

}

type Biddinghouse struct {
	currentbid int32
	itemprice  int32
	winner     string
	mux        sync.Mutex
	replication.UnimplementedReplicationServer
	id      int32
	bidders map[string]string
}

func (b *Biddinghouse) Receivebid(ctx context.Context, bid *replication.BidPackage) (con *replication.Confirmation, err error) {
	bidFromBidder := bid.Bid
	conpackage := &replication.Confirmation{}
	if b.currentbid == 0 {
		timestamp = time.Now()
		timeLimit = timestamp.Add(time.Duration(1) * time.Minute)
	}
	if timeLimit.Before(time.Now()) {
		log.Printf("Replica #%d:  Timelimit reached, auction ending.\n", b.id)
		conpackage.HasEnded = true
	} else if bidFromBidder > b.currentbid {
		//Increase bid
		log.Printf("Replica #%d:  Increasing bid to %d\n", b.id, bidFromBidder)
		b.mux.Lock()
		b.currentbid = bid.Bid
		b.winner = bid.Bidder
		conpackage.CurrentWinner = b.winner
		conpackage.CurrentPrice = b.currentbid
		b.mux.Unlock()
		conpackage.Confirmation = true
	} else {
		//Bid not high enough
		conpackage.Confirmation = false
		b.mux.Lock()
		conpackage.CurrentWinner = b.winner
		conpackage.CurrentPrice = b.currentbid
		b.mux.Unlock()
	}
	if b.bidders[bid.Bidder] == "" {
		b.bidders[bid.Bidder] = bid.Bidder
		log.Printf("Replica #%d:  New user detected, adding \"%s\" to list of users.", b.id, bid.Bidder)
	}
	return conpackage, nil
}

func (b *Biddinghouse) Result(ctx context.Context, empty *replication.Empty) (res *replication.ResultPackage, err error) {
	log.Printf("Replica #%d: Result requested from user, answering with result message.\n", b.id)
	resultPackage := &replication.ResultPackage{}
	resultPackage.Winner = b.winner
	resultPackage.Highestbid = b.currentbid
	if timeLimit.Before(time.Now()) {
		resultPackage.HasEnded = true
	} else {
		resultPackage.HasEnded = false
	}
	return resultPackage, nil

}
