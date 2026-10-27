# kalzone-steptest

A throwaway Kalzone service for the step test of adaptive queue rates. Each
message on its `burn` queue runs `SELECT BENCHMARK(BURN_ITERATIONS, SHA2(id, 256))`
on its managed MySQL database, so a queue rate becomes a known CPU load on the
shared MySQL server.

| Variable | Meaning |
|---|---|
| `PLATFORM_DB_DSN` | set by attaching the managed database |
| `PLATFORM_QUEUE_*` | set by Kalzone |
| `BURN_ITERATIONS` | SHA2 rounds per message, 200000 by default |
| `FLOOD_KEY` | secret the `/flood` call must send |

Fill the queue with a backlog, then step its rate:

```sh
curl -X POST -H "Flood-Key: $FLOOD_KEY" "https://steptest.dev.kalz.one/flood?n=20000"
```

Delete the service, its queue and its database when the test is done.
