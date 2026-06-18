# Observability Plan

## Goal
Aggregate device log dumps and k3s pod logs into Grafana via Loki.

## Current Issues
- Grafana port binds to `3000:3000` (all interfaces) instead of `10.0.0.1:3000`
- `logdumpsvc.go` is a no-op
- k3s pod logs are ephemeral

---

## Changes

### [AI AGENT] `docker-compose.prod.yaml` — Add Loki + Fix Grafana Binding

**Fix Grafana:**
```yaml
ports:
  - "10.0.0.1:3000:3000"
```

**Add Loki:**
```yaml
  loki:
    image: grafana/loki:latest
    ports:
      - "10.0.0.1:3100:3100"
    volumes:
      - loki-data:/loki
    command: -config.file=/etc/loki/local-config.yaml
```

**Add volume:**
```yaml
volumes:
  loki-data:
```

Notes:
- `grafana/loki:latest` is multi-arch (ARM64 ok).
- Default `local-config.yaml` runs single-instance. Data persists to `/loki`.

---

### [USER] `k8s/10-configmap.yaml` — Inject `LOKI_URI`

Add to `data:`
```yaml
  LOKI_URI: "http://10.0.0.1:3100"
```

Then `kubectl apply -f k8s/10-configmap.yaml` and rollout restart.

---

### [AI AGENT] `internal/services/logdumpsvc.go` — Loki Push Client

Implement a lightweight HTTP client (no SDK needed).

Push payload shape (Loki v1):
```json
{
  "streams": [
    {
      "stream": {
        "mac_addr": "aa:bb:cc...",
        "contract": "uuid",
        "source": "device"
      },
      "values": [
        ["<unix_epoch_ns>", "log line 1"],
        ["<unix_epoch_ns>", "log line 2"]
      ]
    }
  ]
}
```

Rules:
- POST to `${LOKI_URI}/loki/api/v1/push`
- Labels are **low cardinality**: `mac_addr`, `contract`, `source` only.
- Timestamps must be nanosecond strings.
- Expect `204 No Content` on success.
- Keep retry logic minimal (fail open). Do not block MQTT handler.

---

### [AI AGENT] `k8s/60-promtail-daemonset.yaml` — Ship Pod Logs to Loki

Create a new DaemonSet manifest with:

1. **RBAC**
   - `ServiceAccount` in `dirtie` namespace
   - `ClusterRole`: `get`, `list` on `pods`, `nodes`
   - `ClusterRoleBinding`

2. **DaemonSet spec**
   - Image: `grafana/promtail:latest` (multi-arch ARM64)
   - Volume mounts:
     - `hostPath` `/var/log`
     - `hostPath` `/var/lib/docker/containers` (or containerd equivalent path on k3s)
     - `hostPath` `/var/log/pods`
   - ConfigMap holding `promtail-config.yaml`

3. **Promtail config key bits**
   ```yaml
   clients:
     - url: http://10.0.0.1:3100/loki/api/v1/push
   scrape_configs:
     - job_name: kubernetes-pods
       kubernetes_sd_configs:
         - role: pod
       pipeline_stages:
         - docker: {}
       relabel_configs:
         - source_labels: [__meta_kubernetes_namespace]
           target_label: namespace
         - source_labels: [__meta_kubernetes_pod_name]
           target_label: pod
         - source_labels: [__meta_kubernetes_container_name]
           target_label: container
   ```

Apply with `kubectl apply -f k8s/60-promtail-daemonset.yaml`.

---

### [USER] Grafana UI — Add Loki Datasource

1. Browse to `http://10.0.0.1:3000`
2. Configuration → Data Sources → Add → Loki
3. URL: `http://loki:3100` (inside Docker Compose network)
4. Save & Test

---

## Verification Checklist

| Check | How |
|---|---|
| Loki healthy | `curl http://10.0.0.1:3100/ready` |
| Device logs in Grafana | Explore → Loki → `{source="device",mac_addr="..."}` |
| Pod logs in Grafana | Explore → Loki → `{namespace="dirtie"}` |
| Grafana bound to `10.0.0.1` only | `ss -tlnp` on rpic1 should show `10.0.0.1:3000` |

---

## ARM64 / Pi Notes

- Loki idle: ~50–100 MB RAM.
- Promtail per node: ~20–40 MB RAM.
- Retention: default local config grows until disk runs out. If the Pi SD card is small, add a `--limits.retention=168h` tweak (7 days) later.
- k3s uses containerd, not Docker. Promtail should mount `/var/log/pods` and `/run/containerd` (or `/run/k3s/containerd`) from the host. Check exact host paths with `ls /run/k3s/containerd/` on a worker node.

---

## Rollout Order

1. `[AI AGENT]` Patch `docker-compose.prod.yaml` and `k8s/10-configmap.yaml`
2. `[USER]` Deploy data plane: `docker compose up -d` on rpic1
3. `[USER]` `kubectl apply -f k8s/` (ConfigMap + new Promtail manifest)
4. `[AI AGENT]` Implement `logdumpsvc.go` Loki pusher
5. `[USER]` Rollout restart app, verify Grafana datasource, test log flow
