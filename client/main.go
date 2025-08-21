package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	addr := getEnv("SERVER_ADDR", "server:9000")
	pingEveryMs := getEnvInt("PING_INTERVAL_MS", 1000)

	for {
		log.Printf("[CLIENT] dialing %s ...", addr)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Printf("[CLIENT] dial error: %v (retrying in 1s)", err)
			time.Sleep(time.Second)
			continue
		}
		handleConn(conn, time.Duration(pingEveryMs)*time.Millisecond)
	}
}

func handleConn(conn net.Conn, interval time.Duration) {
	defer conn.Close()
	peer := conn.RemoteAddr().String()
	log.Printf("[CLIENT] connected to %s", peer)

	// map para guardar ts por seq e medir RTT quando chegar o PONG
	var mu sync.Mutex
	sent := make(map[int64]int64) // seq -> ts(ms)
	var seq int64 = 0

	// leitor assíncrono
	go func() {
		r := bufio.NewScanner(conn)
		for r.Scan() {
			line := strings.TrimSpace(r.Text())
			now := time.Now().UnixMilli()
			log.Printf("[CLIENT] <- %q", line)

			if strings.HasPrefix(line, "PONG") {
				tsStr := parseKV(line, "ts")
				seqStr := parseKV(line, "seq")
				ts, _ := strconv.ParseInt(tsStr, 10, 64)
				seqVal, _ := strconv.ParseInt(seqStr, 10, 64)

				mu.Lock()
				start, ok := sent[seqVal]
				delete(sent, seqVal)
				mu.Unlock()

				if ok {
					rtt := now - start
					log.Printf("[CLIENT] RTT seq=%d ~ %d ms", seqVal, rtt)
				} else if ts > 0 {
					rtt := now - ts
					log.Printf("[CLIENT] RTT(heuristic) ~ %d ms", rtt)
				}
			}
		}
		if err := r.Err(); err != nil {
			log.Printf("[CLIENT] read error: %v", err)
		}
		log.Printf("[CLIENT] server closed connection")
	}()

	// Envia um comando exemplo (como se fosse "jogo/sala")
	fmt.Fprintln(conn, "CMD JOIN room-1")

	// Loop de PING
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		seq++
		ts := time.Now().UnixMilli()
		msg := fmt.Sprintf("PING ts=%d seq=%d", ts, seq)

		mu.Lock()
		sent[seq] = ts
		mu.Unlock()

		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Printf("[CLIENT] write error: %v", err)
			return
		}
		log.Printf("[CLIENT] -> %q", msg)
	}
}

func parseKV(s, key string) string {
	parts := strings.Fields(s)
	for _, p := range parts {
		if strings.HasPrefix(p, key+"=") {
			return strings.TrimPrefix(p, key+"=")
		}
	}
	return ""
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
