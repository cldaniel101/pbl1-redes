package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	// "strconv"
	"strings"
	"time"
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
	defer func() {
		log.Printf("[SERVER] closing %s", peer)
		conn.Close()
	}()

	r := bufio.NewScanner(conn)
	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		nowMs := time.Now().UnixMilli()
		log.Printf("[SERVER] <- from %s: %q", peer, line)

		// Protocolinho simples:
		// PING ts=<ms> seq=<n>
		// CMD <qualquer-coisa>
		if strings.HasPrefix(line, "PING") {
			// extraia ts= e seq=
			ts := parseKV(line, "ts")
			seq := parseKV(line, "seq")
			resp := fmt.Sprintf("PONG ts=%s seq=%s server_ts=%d", ts, seq, nowMs)
			fmt.Fprintln(conn, resp)
			log.Printf("[SERVER] -> to %s: %q", peer, resp)
			continue
		}

		if strings.HasPrefix(line, "CMD ") {
			// Apenas acusa recebimento. Ex.: "CMD JOIN room-1"
			resp := fmt.Sprintf("ACK %d", nowMs)
			fmt.Fprintln(conn, resp)
			log.Printf("[SERVER] -> to %s: %q (ack to command)", peer, resp)
			continue
		}

		// Mensagem desconhecida
		resp := fmt.Sprintf("ERR unknown msg at %d", nowMs)
		fmt.Fprintln(conn, resp)
		log.Printf("[SERVER] -> to %s: %q", peer, resp)
	}

	if err := r.Err(); err != nil {
		log.Printf("[SERVER] read error from %s: %v", peer, err)
	}
}

func parseKV(s, key string) string {
	// procura por " key=valor"
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
