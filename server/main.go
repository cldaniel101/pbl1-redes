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

// Player representa um jogador conectado, com um ID único e sua conexão.
type Player struct {
	ID   string
	Conn net.Conn
}

// Match representa um duelo 1v1 entre dois jogadores.
type Match struct {
	Player1 *Player
	Player2 *Player
}

var (
	// 'matchmakingQueue' é a "sala de espera". É uma lista de jogadores esperando por um oponente.
	matchmakingQueue []*Player
	// 'activeMatches' guarda um registro de todas as partidas que estão em andamento.
	activeMatches = make(map[*Player]*Match)
	// O Mutex continua sendo "cadeado" de segurança para proteger a fila e as partidas.
	mu sync.Mutex
)

// tryCreateMatch é a nossa função "organizadora". Ela verifica a fila de espera.
func tryCreateMatch() {
	// Trancamos o cadeado para mexer na fila com segurança.
	mu.Lock()
	// No final da função, garantimos que o cadeado será destrancado.
	defer mu.Unlock()

	// A condição principal: só criamos uma partida se tivermos 2 ou mais jogadores na fila.
	if len(matchmakingQueue) >= 2 {
		// Pegamos os dois primeiros jogadores da fila.
		player1 := matchmakingQueue[0]
		player2 := matchmakingQueue[1]

		// Removemos esses dois jogadores da fila, pois eles não estão mais esperando.
		matchmakingQueue = matchmakingQueue[2:]

		// Criamos a estrutura da partida com os dois jogadores.
		match := &Match{
			Player1: player1,
			Player2: player2,
		}

		// Guardamos a referência da partida para cada jogador.
		activeMatches[player1] = match
		activeMatches[player2] = match

		// Log para sabermos que a partida foi criada.
		log.Printf("[SERVER] Partida criada entre %s e %s", player1.ID, player2.ID)

		// Avisamos aos jogadores que a partida começou!
		fmt.Fprintln(player1.Conn, "MSG Partida encontrada! Você está jogando contra "+player2.ID)
		fmt.Fprintln(player2.Conn, "MSG Partida encontrada! Você está jogando contra "+player1.ID)

		// LÓGICA DO JOGO...
		// Por enquanto, estou apenas avisando que eles foram pareados.
	}
}

func main() {
	addr := getEnv("LISTEN_ADDR", ":9000")

	log.Printf("[SERVER] starting on %s ...", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[SERVER] listen error: %v", err)
	}
	defer ln.Close()

	// Inicia uma goroutine que fica tentando criar partidas a cada segundo.
	go func() {
		for {
			tryCreateMatch()
			time.Sleep(1 * time.Second) // Espera 1 segundo antes de checar a fila de novo.
		}
	}()

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

	// Criamos um novo jogador para esta conexão.
	// O ID pode ser o endereço de rede ou um nome que o jogador envie.
	currentPlayer := &Player{
		ID:   peer,
		Conn: conn,
	}

	// Lógica de limpeza quando o jogador desconectar.
	defer func() {
		log.Printf("[SERVER] closing %s", peer)
		mu.Lock()

		// Se o jogador estava em uma partida, precisamos notificar o oponente.
		if match, ok := activeMatches[currentPlayer]; ok {
			var opponent *Player
			if match.Player1 == currentPlayer {
				opponent = match.Player2
			} else {
				opponent = match.Player1
			}

			if opponent != nil {
				fmt.Fprintln(opponent.Conn, "MSG Seu oponente desconectou.")
			}
			// Remove a partida da lista de partidas ativas.
			delete(activeMatches, match.Player1)
			delete(activeMatches, match.Player2)
		} else {
			// Se ele não estava em partida, talvez estivesse na fila. Vamos removê-lo.
			for i, p := range matchmakingQueue {
				if p == currentPlayer {
					matchmakingQueue = append(matchmakingQueue[:i], matchmakingQueue[i+1:]...)
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
		log.Printf("[SERVER] <- from %s: %q", peer, line)

		// Vamos criar um novo comando para entrar na fila!
		if line == "CMD FIND_MATCH" {
			log.Printf("[SERVER] Jogador %s entrou na fila de matchmaking.", currentPlayer.ID)
			fmt.Fprintln(currentPlayer.Conn, "ACK Você entrou na fila. Aguardando oponente...")

			mu.Lock()
			// Adiciona o jogador à fila de espera.
			matchmakingQueue = append(matchmakingQueue, currentPlayer)
			mu.Unlock()
			continue
		}

		// Se o jogador estiver em uma partida, as mensagens dele devem ir para o oponente.
		if strings.HasPrefix(line, "MSG ") {
			mu.Lock()
			if match, ok := activeMatches[currentPlayer]; ok {
				var opponent *Player
				if match.Player1 == currentPlayer {
					opponent = match.Player2
				} else {
					opponent = match.Player1
				}

				if opponent != nil {
					// Envia a mensagem para o oponente.
					fmt.Fprintln(opponent.Conn, line)
				}
			}
			mu.Unlock()
			continue
		}

		// Você pode manter a lógica de PING/PONG como está.
		if strings.HasPrefix(line, "PING ") {
			timestamp := strings.TrimPrefix(line, "PING ")
			resp := fmt.Sprintf("PONG %s", timestamp)
			fmt.Fprintln(conn, resp)
			log.Printf("[SERVER] -> to %s: %q (pong response)", peer, resp)
			continue
		}

		// Mensagem de erro para comandos desconhecidos.
		fmt.Fprintln(conn, "ERR unknown command")
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
