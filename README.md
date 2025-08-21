# PBL Sockets: Ping/Pong

## Subir tudo (build + run)
docker compose up --build --scale client=3

## Ver logs
docker compose logs -f server
docker compose logs -f client

## Derrubar
docker compose down -v

## Variáveis úteis
- SERVER_ADDR (client) ex.: server:9000
- PING_INTERVAL_MS (client) ex.: 1000
- LISTEN_ADDR (server) ex.: :9000
