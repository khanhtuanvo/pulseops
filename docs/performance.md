# Performance

Benchmarks for the alert ingestion critical path live in
[`backend/internal/alerting/benchmark_test.go`](../backend/internal/alerting/benchmark_test.go).

## Running

```bash
# Pure CPU benchmark (no dependencies)
go test -bench=BenchmarkFingerprint -benchmem ./internal/alerting/...

# Full suite — the dedup + handler benchmarks need a local MongoDB replica set.
# They skip automatically if Mongo is unreachable.
MONGODB_TEST_URI='mongodb://localhost:27017/?replicaSet=rs0' \
  go test -bench=. -benchmem -benchtime=10s ./internal/alerting/...
```

The webhook-handler benchmark resets the per-API-key rate limiter (100/min)
outside the timed region so it measures the request→MongoDB-write path rather
than rate-limit rejections.

## Results

### BenchmarkFingerprint — captured

```
goos: linux  goarch: amd64  cpu: 13th Gen Intel(R) Core(TM) i7-13620H
BenchmarkFingerprint-4    8726893    263.4 ns/op    240 B/op    6 allocs/op
```

SHA-256 fingerprinting costs ~260 ns and is never the bottleneck; ingestion
latency is dominated by the MongoDB round-trips below.

### BenchmarkFingerprintDedup / BenchmarkWebhookHandler — pending capture

These require a MongoDB replica set and have **not yet been run on a real host**
(the dev sandbox has no Mongo). Capture them with the full-suite command above
and paste the numbers here. Expected shape (illustrative only — replace with
measured values):

```
BenchmarkFingerprint-8        10000000    120 ns/op
BenchmarkFingerprintDedup-8      50000   28000 ns/op   (~28 µs, one dup-insert + one read)
BenchmarkWebhookHandler-8        10000  150000 ns/op   (~150 µs, request → incident write)
```

A sub-millisecond handler write path (well under the "sub-2-second end-to-end"
target, which also includes change-stream propagation and WebSocket fan-out) is
what these numbers are meant to demonstrate once captured.
