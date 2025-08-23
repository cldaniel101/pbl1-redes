# Jogo de Cartas Multiplayer

Este repositório contém o desenvolvimento de um jogo de cartas online multiplayer, concebido como o primeiro grande desafio de uma startup fictícia de jogos indie fundada por estudantes de Engenharia de Computação. O projeto foca em duelos táticos e coleção de cartas, com toda a lógica de jogo, estado dos jogadores e comunicação gerenciada por um servidor centralizado.

A aplicação cliente-servidor é desenvolvida em Go e a comunicação é realizada via sockets TCP. O ambiente é totalmente containerizado com Docker e Docker Compose para facilitar a execução, o teste e a implantação.

## Funcionalidades Principais

- **Servidor Centralizado**: Gerencia a lógica do jogo, o estado dos jogadores e a comunicação entre eles em tempo real.

- **Comunicação via Sockets**: A comunicação entre cliente e servidor é bidirecional e implementada exclusivamente com a biblioteca nativa de sockets TCP, sem o uso de frameworks externos.

- **Partidas 1v1**: O sistema permite que múltiplos jogadores se conectem simultaneamente e os pareia em duelos únicos, garantindo que um jogador não enfrente vários oponentes ao mesmo tempo.

- **Visualização de Atraso**: Sistema implementado de PING/PONG que permite aos jogadores visualizar em tempo real a latência (RTT - Round-Trip Time) de sua comunicação com o servidor, exibindo valores como `RTT: 3 ms` no console.

- **Sistema de Pacotes de Cartas**: Uma mecânica central para adquirir novas cartas é a abertura de pacotes. O servidor gerencia um "estoque" global e trata requisições concorrentes de forma justa para evitar duplicação ou perda de cartas.

- **Chat em Tempo Real**: Sistema de comunicação entre jogadores baseado em salas, permitindo coordenação e interação social durante as partidas.

- **Ambiente Containerizado**: Todos os componentes são desenvolvidos e testados em contêineres Docker, permitindo a fácil execução e escalabilidade para testes.

## Tecnologias Utilizadas

- **Go**: Linguagem de programação para o desenvolvimento do cliente e do servidor.
- **Docker**: Para a criação das imagens dos contêineres do cliente e do servidor.
- **Docker Compose**: Para orquestrar e gerenciar os contêineres da aplicação.

## Como Executar

Certifique-se de ter o Docker e o Docker Compose instalados em sua máquina.

1. **Clone o repositório** (ou garanta que todos os arquivos estejam no mesmo diretório).

2. **Construa as imagens e inicie os contêineres:**
   O comando a seguir irá construir as imagens, iniciar 1 contêiner para o servidor e 2 contêineres para os clientes em modo "detached" (em segundo plano).

   ```bash
   docker compose up --build --scale client=2 -d
   ```

## Testando a Aplicação

Para interagir com a aplicação, você pode se conectar a uma sessão interativa com cada um dos clientes. Para isso, você precisará de duas janelas de terminal.

1. **Liste os contêineres em execução** para obter seus nomes:
   ```bash
   docker ps
   ```
   A saída mostrará os nomes dos contêineres, como `pbl1-redes-client-1` e `pbl1-redes-client-2`.

2. **Abra dois terminais.**

3. **No primeiro terminal**, conecte-se ao primeiro cliente:
   ```bash
   docker attach pbl1-redes-client-1
   ```

4. **No segundo terminal**, conecte-se ao segundo cliente:
   ```bash
   docker attach pbl1-redes-client-2
   ```

5. **Agora você pode**:
   - **Monitorar a latência**: Observe as mensagens `RTT: X ms` que aparecem automaticamente a cada 2 segundos, mostrando a qualidade da conexão com o servidor
   - **Comunicar entre jogadores**: Digite mensagens que serão transmitidas para outros jogadores na mesma sala
   - **Testar a conectividade**: Verifique a estabilidade da comunicação cliente-servidor através dos logs detalhados

### Exemplo de Saída do Cliente:
```
[CLIENT] connected to 172.20.0.2:9000
[CLIENT] <- "ACK 1755907668723"
RTT: 3 ms
[CLIENT] RTT calculated: 3 ms
Received: Pronto para o duelo!
RTT: 2 ms
```

> **Importante**: Para sair da sessão `attach` sem derrubar o contêiner, use a combinação de teclas `Ctrl+P` e em seguida `Ctrl+Q`.

## Comandos Úteis

- **Visualizar os logs de todos os serviços em tempo real:**
  ```bash
  docker compose logs -f
  ```

- **Visualizar apenas os logs do servidor (monitorar PING/PONG e conexões):**
  ```bash
  docker attach pbl_server
  ```

- **Derrubar o ambiente** (para e remove os contêineres):
  ```bash
  docker compose down -v
  ```

## Variáveis de Ambiente

É possível customizar a execução através de variáveis de ambiente no arquivo `docker-compose.yml`.

- `SERVER_ADDR` (cliente): Endereço do servidor ao qual o cliente deve se conectar. Ex: `server:9000`.
- `PING_INTERVAL_MS` (cliente): Intervalo em milissegundos para o envio de PINGs para medição de latência. Padrão: `2000` (2 segundos).
- `LISTEN_ADDR` (servidor): Endereço e porta em que o servidor escutará por conexões. Ex: `:9000`.

## Arquitetura da Aplicação

### Fluxo de Comunicação:

1. **Inicialização**: Cliente conecta ao servidor e entra automaticamente na sala de jogo
2. **Monitoramento de Latência**: Cliente envia PING com timestamp → Servidor responde PONG → Cliente calcula e exibe RTT
3. **Comunicação**: Mensagens e comandos de jogo são transmitidos em tempo real via sockets TCP
4. **Gestão de Estado**: Servidor mantém o estado centralizado de todos os jogadores e partidas

### Protocolo de Mensagens:

- `CMD JOIN <sala>`: Cliente entra em uma sala de jogo
- `PING <timestamp>`: Solicitação de latência com timestamp para monitoramento de qualidade da conexão
- `PONG <timestamp>`: Resposta de latência ecoando o timestamp original
- `MSG <texto>`: Mensagem de comunicação entre jogadores
- `ACK <timestamp>`: Confirmação de comando executado pelo servidor

## Status do Desenvolvimento

**Fase Atual: Infraestrutura Base** ✅
- [x] Comunicação cliente-servidor via sockets TCP
- [x] Sistema de monitoramento de latência (PING/PONG RTT)
- [x] Chat em tempo real entre jogadores
- [x] Containerização completa com Docker
- [x] Sistema de salas básico

**Próximas Fases:**
- [ ] Implementação da lógica de duelos 1v1
- [ ] Sistema de cartas e baralhos
- [ ] Mecânica de pacotes de cartas
- [ ] Interface de jogo mais elaborada
- [ ] Sistema de ranking e progressão

---

*Este projeto representa a implementação de uma infraestrutura robusta para jogos online multiplayer, demonstrando conceitos fundamentais de programação de redes, sistemas distribuídos e desenvolvimento de jogos.*