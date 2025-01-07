package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

var target = flag.String("target", "127.0.0.1:8080", "target server")
var times = flag.Int("times", 0, "number of times to ping")
var interval = flag.Float64("interval", 1, "interval between pings")
var protocol = flag.String("protocol", "udp", "protocol")

var size = flag.Int("size", 512, "size pre packet")

func init() {
	flag.Parse()
}

func ping(conn net.Conn) bool {
	data := make([]byte, *size)
	_, _ = rand.Read(data)

	_ = conn.SetDeadline(time.Now().Add(time.Duration(*interval * float64(time.Second))))
	_, err := conn.Write(data)
	if err != nil {
		log.Println("Error writing:", err)
		return false
	}

	response := make([]byte, *size)
	_, err = conn.Read(response)
	if err != nil {
		log.Println("Error reading:", err)
		return false
	}
	_ = conn.SetDeadline(time.Time{})

	if !bytes.Equal(response, data) {
		log.Println("Response checksum error")
		return false
	}

	return true
}

func sleepContext(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return
	case <-ctx.Done():
		return
	}
}

func main() {
	var conn net.Conn

	delay := time.Duration(*interval * float64(time.Second))
	var err error

	var totalTimes int
	var successTimes int

	var averageRtt float64
	var minRtt float64 = math.MaxFloat64
	var maxRtt float64 = 0.00

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt)
	defer done()

loop:
	for *times == 0 || *times > totalTimes {
		select {
		case <-ctx.Done():
			if conn != nil {
				_ = conn.Close()
			}
			break loop
		default:
			timestamp := time.Now()
			if conn == nil {
				conn, err = net.Dial(*protocol, *target)
			}

			if err != nil {
				log.Println("Error connecting:", err)
			}

			success := false
			if conn != nil {
				success = ping(conn)
			}

			elapsed := time.Since(timestamp)
			totalTimes = totalTimes + 1

			if success {
				successTimes = successTimes + 1
				log.Printf("Ping reply from %s in %s\n", *target, elapsed)

				timeMs := float64(elapsed.Microseconds()) / 1e3
				minRtt = min(timeMs, minRtt)
				maxRtt = max(timeMs, maxRtt)
				averageRtt = averageRtt + timeMs
			}

			if *times == 0 || *times != totalTimes {
				sleepContext(ctx, delay-elapsed)
			}
		}
	}

	fmt.Printf("\n--- %s ping statistics ---\n", *target)
	fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
		totalTimes, successTimes, (1-(successTimes/totalTimes))*100)

	if successTimes > 0 {
		fmt.Printf("round-trip min/avg/max = %s/%s/%s\n",
			formatFloat(minRtt), formatFloat(averageRtt/float64(totalTimes)), formatFloat(maxRtt))
	}
}

func formatFloat(num float64) string {
	s := fmt.Sprintf("%.3f", num)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
