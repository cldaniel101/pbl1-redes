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

var (
	showPing  bool
	pingMutex sync.RWMutex
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	peer := conn.RemoteAddr().String()
	log.Printf("[CLIENT] connected to %s", peer)

	// Goroutine para receber mensagens do servidor
	go func() {
		r := bufio.NewScanner(conn)
		for r.Scan() {
			line := strings.TrimSpace(r.Text())

			if strings.HasPrefix(line, "MSG ") {
				fmt.Printf("Received: %s\n", strings.TrimPrefix(line, "MSG "))
			} else if strings.HasPrefix(line, "PONG ") {
				// Processar resposta PONG e calcular RTT
				timestampStr := strings.TrimPrefix(line, "PONG ")
				if sentTime, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					currentTime := time.Now().UnixMilli()
					rtt := currentTime - sentTime

					// Mostrar RTT apenas se o usuário habilitou via comando /ping
					pingMutex.RLock()
					if showPing {
						fmt.Printf("RTT: %d ms\n", rtt)
					}
					pingMutex.RUnlock()
				} else {
					log.Printf("[CLIENT] error parsing PONG timestamp: %v", err)
				}
			} else if strings.HasPrefix(line, "ACK ") {
				fmt.Printf("Server: %s\n", strings.TrimPrefix(line, "ACK "))
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
		}
	}()

	// Goroutine para ler a entrada do usuário e processar comandos/mensagens
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Digite mensagens para enviar ou comandos começando com '/':")
		fmt.Println("Comandos disponíveis: /ping, /help")

		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if text == "" {
				continue
			}

			// Processar comandos que começam com "/"
			if strings.HasPrefix(text, "/") {
				handleCommand(text, conn)
			} else {
				// Enviar mensagem normal
				fmt.Fprintln(conn, "MSG "+text)
			}
		}
	}()

	// Aguarda indefinidamente
	select {}
}

func handleCommand(command string, conn net.Conn) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/ping":
		pingMutex.Lock()
		showPing = !showPing
		pingMutex.Unlock()

		if showPing {
			fmt.Println("Exibição de RTT ativada. Você verá a latência a cada ping.")
		} else {
			fmt.Println("Exibição de RTT desativada.")
		}

	case "/help":
		fmt.Println("Comandos disponíveis:")
		fmt.Println("  /ping  - Liga/desliga a exibição do RTT (latência)")
		fmt.Println("  /help  - Mostra esta mensagem de ajuda")
		fmt.Println("\nDigite qualquer outra coisa para enviar uma mensagem para outros jogadores.")

	default:
		fmt.Printf("Comando desconhecido: %s. Digite /help para ver os comandos disponíveis.\n", cmd)
	}
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
