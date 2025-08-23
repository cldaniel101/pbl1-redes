package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	rooms = make(map[string][]net.Conn)
	mu    sync.Mutex
)

func main() {
	addr := getEnv("LISTEN_ADDR", ":9000")

	log.Printf("[SERVER] starting on %s ...", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[SERVER] listen error: %v", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[SERVER] accept error: %v", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	peer := conn.RemoteAddr().String()
	log.Printf("[SERVER] connection from %s", peer)
	var currentRoom string

	defer func() {
		log.Printf("[SERVER] closing %s", peer)
		mu.Lock()
		if room, ok := rooms[currentRoom]; ok {
			for i, c := range room {
				if c == conn {
					rooms[currentRoom] = append(room[:i], room[i+1:]...)
					break
				}
			}
		}
		mu.Unlock()
		conn.Close()
	}()

	r := bufio.NewScanner(conn)
	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		nowMs := time.Now().UnixMilli()
		log.Printf("[SERVER] <- from %s: %q", peer, line)

		if strings.HasPrefix(line, "PING ") {
			timestamp := strings.TrimPrefix(line, "PING ")
			resp := fmt.Sprintf("PONG %s", timestamp)
			fmt.Fprintln(conn, resp)
			log.Printf("[SERVER] -> to %s: %q (pong response)", peer, resp)
			continue
		}

		if strings.HasPrefix(line, "CMD JOIN ") {
			roomName := strings.TrimPrefix(line, "CMD JOIN ")
			currentRoom = roomName
			mu.Lock()
			rooms[roomName] = append(rooms[roomName], conn)
			mu.Unlock()
			resp := fmt.Sprintf("ACK %d", nowMs)
			fmt.Fprintln(conn, resp)
			log.Printf("[SERVER] -> to %s: %q (ack to command)", peer, resp)
			continue
		}

		if strings.HasPrefix(line, "MSG ") {
			msg := strings.TrimPrefix(line, "MSG ")
			mu.Lock()
			if room, ok := rooms[currentRoom]; ok {
				for _, client := range room {
					if client != conn {
						fmt.Fprintln(client, "MSG "+msg)
					}
				}
			}
			mu.Unlock()
			continue
		}

		resp := fmt.Sprintf("ERR unknown msg at %d", nowMs)
		fmt.Fprintln(conn, resp)
		log.Printf("[SERVER] -> to %s: %q", peer, resp)
	}

	if err := r.Err(); err != nil {
		log.Printf("[SERVER] read error from %s: %v", peer, err)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
