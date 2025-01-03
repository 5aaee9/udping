package main

import (
	"context"
	"flag"
	"github.com/libp2p/go-reuseport"
	"log"
	"os"
	"os/signal"
)

var server = flag.String("server", ":8080", "Server address")

func init() {
	flag.Parse()
}

func listenPacket(ctx context.Context) {
	packet, err := reuseport.ListenPacket("udp", *server)
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 65565)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			length, from, err := packet.ReadFrom(buffer)
			if err != nil {
				log.Printf("Error reading from packet: %v", err)
				continue
			}
			log.Printf("Read %d bytes from %s", length, from)
			_, err = packet.WriteTo(buffer[:length], from)
			if err != nil {
				log.Printf("Error writing to packet: %v", err)
			}
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go listenPacket(ctx)
	<-ctx.Done()
}
