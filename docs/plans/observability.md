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

### [AI AGENT] `k8s/60-alloy-daemonset.yaml` — Ship Pod Logs to Loki

Create a new DaemonSet manifest with:

1. **RBAC**
   - `ServiceAccount` in `dirtie` namespace
   - `ClusterRole`: `get`, `list`, `watch` on `pods`, `nodes`, `namespaces`
   - `ClusterRoleBinding`

2. **DaemonSet spec**
   - Image: `grafana/alloy:latest` (multi-arch ARM64)
   - Volume mounts:
     - `hostPath` `/var/log`
     - `hostPath` `/var/log/pods`
     - `hostPath` `/run/k3s/containerd` (for containerd CRI)
   - ConfigMap holding `alloy-config.alloy`

3. **Alloy config (`alloy-config.alloy`)**
   ```alloy
   loki.source.kubernetes "pods" {
     targets    = discovery.kubernetes.pods.targets
     forward_to = [loki.write.local_loki.receiver]
   }

   discovery.kubernetes_pods "pods" {
     namespaces = ["dirtie"]
   }

   loki.write "local_loki" {
     endpoint {
       url = "http://10.0.0.1:3100/loki/api/v1/push"
     }
   }
   ```

Apply with `kubectl apply -f k8s/60-alloy-daemonset.yaml`.

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

|- Alloy per node: ~30–60 MB RAM.
|- Alloy per node: ~30–60 MB RAM.
|- k3s uses containerd, not Docker. Alloy should mount `/var/log/pods` and `/run/k3s/containerd` from the host. Check exact host paths with `ls /run/k3s/containerd/` on a worker node.

---

## Rollout Order

1. `[AI AGENT]` Patch `docker-compose.prod.yaml` and `k8s/10-configmap.yaml`
2. `[USER]` Deploy data plane: `docker compose up -d` on rpic1
3. `[USER]` `kubectl apply -f k8s/` (ConfigMap + new Alloy manifest)
4. `[AI AGENT]` Implement `logdumpsvc.go` Loki pusher
5. `[USER]` Rollout restart app, verify Grafana datasource, test log flow

---

## Architecture Evolution: Decoupling Observability

**Current State:** Observability stack (Loki) runs on `rpi1` alongside the application.

**Proposed Exploration: Dedicated Monitoring Node**
Moving the observability stack to a separate server (e.g., a dedicated "monitoring" Pi or an existing Linux machine) provides:

- **Resource Isolation:** Prevents Prometheus/Loki RAM usage from causing OOM kills on the main application service.
- **Resilience:** The monitoring stack remains operational even if `rpi1` is down or being redeployed.
- **Centralization:** A single endpoint to monitor multiple different projects/servers across the homelab.

**Challenges to Evaluate:**
- **Networking:** Requires stable connectivity and appropriate firewall/port rules between `rpi-app` and `rpi-monitor`.
- **Complexity:** Managing an additional OS, updates, and backups for a second node.