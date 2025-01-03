package main

import (
	"flag"
	"log"
	"net"
	"time"
)

var target = flag.String("target", "127.0.0.1:8080", "target server")
var times = flag.Int("times", 0, "number of times to ping")

func init() {
	flag.Parse()
}

func ping(conn net.Conn) bool {
	_, err := conn.Write([]byte("ping"))
	if err != nil {
		log.Println("Error writing:", err)
		return false
	}
	_, err = conn.Read(make([]byte, 4))
	if err != nil {
		log.Println("Error reading:", err)
		return false
	}

	return true
}

func main() {
	var conn net.Conn
	var err error
	for {
		timestamp := time.Now()
		if conn == nil {
			conn, err = net.Dial("udp", *target)
		}

		if err != nil {
			log.Println("Error connecting:", err)
		}

		success := false
		if conn != nil {
			success = ping(conn)
		}

		elapsed := time.Since(timestamp)

		if success {
			log.Printf("Ping reply from %s in %s\n", *target, elapsed)
		}
		time.Sleep(time.Second - elapsed)
	}
}
