package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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

	// Goroutine para receber mensagens do servidor
	go func() {
		r := bufio.NewScanner(conn)
		for r.Scan() {
			line := strings.TrimSpace(r.Text())
			log.Printf("[CLIENT] <- %q", line)

			if strings.HasPrefix(line, "MSG ") {
				fmt.Printf("Received: %s\n", strings.TrimPrefix(line, "MSG "))
			} else if strings.HasPrefix(line, "PONG ") {
				// Processar resposta PONG e calcular RTT
				timestampStr := strings.TrimPrefix(line, "PONG ")
				if sentTime, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					currentTime := time.Now().UnixMilli()
					rtt := currentTime - sentTime
					fmt.Printf("RTT: %d ms\n", rtt)
					log.Printf("[CLIENT] RTT calculated: %d ms", rtt)
				} else {
					log.Printf("[CLIENT] error parsing PONG timestamp: %v", err)
				}
			}
		}
		if err := r.Err(); err != nil {
			log.Printf("[CLIENT] read error: %v", err)
		}
		log.Printf("[CLIENT] server closed connection")
	}()

	// Envia o comando para entrar na fila de matchmaking
	fmt.Fprintln(conn, "CMD FIND_MATCH")

	// Goroutine para enviar PINGs periódicos
	pingInterval := getPingInterval()
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for range ticker.C {
			timestamp := time.Now().UnixMilli()
			pingMsg := fmt.Sprintf("PING %d", timestamp)
			fmt.Fprintln(conn, pingMsg)
			log.Printf("[CLIENT] -> %q", pingMsg)
		}
	}()

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

	// Aguarda indefinidamente
	select {}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getPingInterval() time.Duration {
	intervalStr := getEnv("PING_INTERVAL_MS", "2000")
	if intervalMs, err := strconv.Atoi(intervalStr); err == nil {
		return time.Duration(intervalMs) * time.Millisecond
	}
	return 2 * time.Second // fallback padrão
}
