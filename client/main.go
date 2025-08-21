package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	addr := getEnv("SERVER_ADDR", "server:9000")
	for {
		log.Printf("[CLIENT] dialing %s ...", addr)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("[CLIENT] dial error: %v (retrying in 1s)", err)
			time.Sleep(time.Second)
			continue
		}
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	peer := conn.RemoteAddr().String()
	log.Printf("[CLIENT] connected to %s", peer)

	go func() {
		r := bufio.NewScanner(conn)
		for r.Scan() {
			line := strings.TrimSpace(r.Text())
			log.Printf("[CLIENT] <- %q", line)
			if strings.HasPrefix(line, "MSG ") {
				fmt.Printf("Received: %s\n", strings.TrimPrefix(line, "MSG "))
			}
		}
		if err := r.Err(); err != nil {
			log.Printf("[CLIENT] read error: %v", err)
		}
		log.Printf("[CLIENT] server closed connection")
	}()

	fmt.Fprintln(conn, "CMD JOIN room-1")

	// Goroutine para ler a entrada do usuário e enviar mensagens
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text != "" {
				fmt.Fprintln(conn, "MSG "+text)
			}
		}
	}()

	// Aguarda indefinidamente, pois não há mais envio periódico
	select {}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
