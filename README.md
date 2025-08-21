# Projeto de Sockets: Chat em Go

Este repositório contém uma aplicação cliente-servidor desenvolvida em Go para demonstrar a comunicação via sockets TCP. O ambiente é totalmente containerizado com Docker e Docker Compose para facilitar a execução e o teste.

## Funcionalidades

- **Conexão Cliente/Servidor**: Comunicação estável via Sockets TCP.
- **Mecanismo de Ping/Pong**: O cliente envia mensagens "PING" em intervalos regulares e o servidor responde com "PONG", permitindo o cálculo do RTT (Round-Trip Time).
- **Chat em Tempo Real**: Um sistema de chat simples que parea os dois primeiros clientes que entram na mesma sala, retransmitindo as mensagens entre eles.
- **Ambiente Containerizado**: Uso de Docker e Docker Compose para criar um ambiente de execução consistente e isolado.

## Tecnologias Utilizadas

-   **Go**: Linguagem de programação para o desenvolvimento do cliente e do servidor.
-   **Docker**: Para a criação das imagens dos contêineres do cliente e do servidor.
-   **Docker Compose**: Para orquestrar e gerenciar os contêineres do ambiente.

## Como Executar

Certifique-se de ter o Docker e o Docker Compose instalados em sua máquina.

1.  **Clone o repositório** (ou garanta que todos os arquivos estejam no mesmo diretório).

2.  **Construa as imagens e inicie os contêineres:**
    O comando a seguir irá construir as imagens (se não existirem), iniciar 1 contêiner para o servidor e 2 contêineres para os clientes em modo "detached" (em segundo plano).

    ```bash
    docker compose up --build --scale client=2 -d
    ```

## Testando o Chat Interativo

Para testar o chat, você precisa de uma sessão interativa com cada um dos clientes. Para isso, você precisará de duas janelas de terminal.

1.  **Liste os contêineres em execução** para obter seus nomes:
    ```bash
    docker ps
    ```
    A saída mostrará os nomes dos contêineres, como `pbl1-redes-client-1` e `pbl1-redes-client-2`.

2.  **Abra dois terminais.**

3.  **No primeiro terminal**, conecte-se ao primeiro cliente:
    ```bash
    docker attach pbl1-redes-client-1
    ```

4.  **No segundo terminal**, conecte-se ao segundo cliente:
    ```bash
    docker attach pbl1-redes-client-2
    ```

5.  Agora, tudo o que você digitar em um terminal aparecerá no outro.

> **Importante**: Para sair da sessão `attach` sem derrubar o contêiner, use a combinação de teclas `Ctrl+P` e em seguida `Ctrl+Q`.

## Comandos Úteis

-   **Visualizar os logs de todos os serviços em tempo real:**
    ```bash
    docker compose logs -f
    ```

-   **Derrubar o ambiente** (para e remove os contêineres):
    ```bash
    docker compose down -v
    ```

## Variáveis de Ambiente

É possível customizar a execução através de variáveis de ambiente no arquivo `docker-compose.yml`.

-   `SERVER_ADDR` (cliente): Endereço do servidor ao qual o cliente deve se conectar. Ex: `server:9000`.
-   `PING_INTERVAL_MS` (cliente): Intervalo em milissegundos para o envio de PINGs. Ex: `1000`.
-   `LISTEN_ADDR` (servidor): Endereço e porta em que o servidor escutará por conexões. Ex: `:9000`.
