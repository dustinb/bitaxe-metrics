# Bitaxe Metrics

Scans the network for Bitaxe miners and exposes metrics in Prometheus format.  The default polling interval is 10 seconds.

# Running

Modify the prometheus.yml to point to the IP where the Go program will be run, set to `host.docker.internal` by default. Then `docker compose up` to start Prometheus.

```
docker compose up
```

Run the Go program to scan the network and start collecting metrics. The metrics are exposed on port 8077, http://localhost:8077/metrics.

```
go run main.go
```
# Example Query

Browse to http://localhost:9090 to do some queries. Average hash rate over 10m period for all bitaxe miners

```
avg_over_time(hash_rate[10m])
```

![Image](./image.png)
