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

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 9000
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	log.Println()
	log.Println()
	log.Printf("### Logs for chat session started at: %s ###", time.Now())
	log.Printf("Starting server...")

	if err != nil {
		log.Printf("Failed to listen on port :%d, Error: %s", ownPort, err)
	}

	log.Printf("Listening on port :%d", ownPort)

	grpcserver := grpc.NewServer()

	server := Biddinghouse{}
	server.currentbid = 0
	server.itemprice = 500
	server.winner = "None have won yet!"
	replication.RegisterReplicationServer(grpcserver, &server)

	err = grpcserver.Serve(listener)
	if err != nil {
		log.Printf("Failed to serve with listener, error: %s", err)
	}

}

type Biddinghouse struct {
	currentbid int32
	itemprice  int32
	winner     string
	mux        sync.Mutex
	replication.UnimplementedReplicationServer
}

func (b *Biddinghouse) Receivebid(ctx context.Context, bid *replication.BidPackage) (con *replication.Confirmation, err error) {
	bidFromBidder := bid.Bid
	conpackage := &replication.Confirmation{}
	if bidFromBidder > b.currentbid {
		//Increase bid
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
	}
	return conpackage, nil
}

func (b *Biddinghouse) Result(ctx context.Context, empty *replication.Empty) (res *replication.ResultPackage, err error) {
	resultPackage := &replication.ResultPackage{}
	if b.currentbid < b.itemprice {
		resultPackage.Highestbid = b.currentbid
	} else {
		resultPackage.Winner = b.winner
	}
	return resultPackage, nil

}
