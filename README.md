# Rate Limiter (Go + Redis)

Middleware HTTP que limita requisições por **IP** ou por **token** (header `API_KEY`), com bloqueio temporário ao exceder o limite.

## Regras
- Sem token: limite por IP (`RATE_LIMIT_IP` req/s).
- Com header `API_KEY: <TOKEN>`: vale **somente** o limite do token (`TOKEN_LIMITS` ou `RATE_LIMIT_TOKEN`), que tem precedência sobre o IP.
- Ao exceder: `429` com a mensagem `you have reached the maximum number of requests or actions allowed within a certain time frame`, e o IP/token fica bloqueado por `BLOCK_DURATION`.

## Configuração (`.env` ou variáveis de ambiente)
| Variável | Padrão | Descrição |
|---|---|---|
| `PORT` | 8080 | Porta HTTP |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | localhost:6379 / – / 0 | Conexão Redis |
| `RATE_LIMIT_IP` | 10 | req/s por IP |
| `RATE_LIMIT_TOKEN` | 100 | req/s para tokens sem limite específico |
| `TOKEN_LIMITS` | – | `token:limite,token:limite` |
| `BLOCK_DURATION` | 5m | Tempo de bloqueio (ex.: `30s`, `5m`) |

## Executar
```sh
docker compose up --build        # app em http://localhost:8080 + Redis
curl -i localhost:8080
curl -i -H "API_KEY: abc123" localhost:8080
```

## Testes
```sh
docker compose run --rm test     # ou: go test ./...
```
Cobrem limite por IP, precedência Token > IP, bloqueio, a estratégia Redis (miniredis) e o middleware (429 + mensagem).

## Arquitetura
```
cmd/server            composição (main)
internal/config       leitura de env/.env
internal/limiter      regras de negócio + interface Store (Strategy)
internal/storage      implementações: redis.go, memory.go
internal/middleware   adaptador net/http
```
### Trocar a estratégia de persistência
Implemente `limiter.Store`:
```go
Hit(ctx context.Context, key string, limit int, blockFor time.Duration) (allowed bool, err error)
```
e troque a linha `storage.NewRedis(rdb)` em `cmd/server/main.go`. Nada mais muda: o limiter e o middleware só conhecem a interface.
