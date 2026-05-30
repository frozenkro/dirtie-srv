# Deployment Architecture

Declarative description of how `dirtie-srv` and its dependencies are hosted,
built, and deployed.

## Topology

```
10.0.0.1                                            10.0.0.2
+----------------------+   Dedicated Ethernet      +--------------------+
| Control Plane Pi     |<------------------------->| Worker Pi 1        |
|                      |   (static IPs)            |                    |
|  k3s control plane   |                           |  k3s worker        |
|  + Docker Compose    |                           |                    |
|    - influxdb        |                           +--------------------+
|    - postgres        |   Dedicated Ethernet       10.0.0.3
|    - grafana         |<------------------------->+--------------------+
|    - mosquitto       |                           | Worker Pi 2        |
|                      |                           |                    |
|                      |                           |  k3s worker        |
+----------------------+                           |                    |
         ^                                         +--------------------+
         |
         +-- dirtie-srv pod (can land here or on workers)
         |
         +-- Pico dirtie-node (via Wi-Fi, broker = 10.0.0.1:1883)
         |
         +-- Android app / dev laptop (remote, via Tailscale)
```

**Principle:** The three cluster nodes communicate over a dedicated Ethernet
switch using static IPs `10.0.0.1`, `10.0.0.2`, and `10.0.0.3`. Stateful
services (DBs, broker, dashboards) run on the control plane Pi and are reached
by pods and local devices at `10.0.0.1`.

Tailscale is still installed on all nodes, but its primary role is to let the
**control plane Pi** accept inbound traffic from remote clients (Android on
cellular, your dev laptop off-site). Node-to-node and pod-to-service traffic
stays on the `10.0.0.x` wire.

The most durable Pi serves dual roles — k3s control plane and off-cluster data
plane. Stateful services run under Docker Compose on this node. Stateless
workloads (`dirtie-srv`) run on the k3s cluster. Worker Pis are pure k3s nodes
with no local state.

---

## About Tailscale (Remote Access Only)

Tailscale assigns the control plane Pi a **static 100.x.y.z IP**. This is the
address remote clients use to reach the cluster when they are off-LAN.

**When you need it:**

- Android app on cellular data
- Dev laptop connecting from outside the house
- Any client that is not on the same physical network as the Pis

**How to find the control plane Pi's Tailscale IP:**

```bash
# On the control plane Pi
$ tailscale ip -4
100.x.y.z
```

Internal cluster traffic **does not** use this IP. Use `10.0.0.1` for all
ConfigMap values, Docker port bindings, and node references.

---

## Data Plane (`docker-compose.prod.yaml`)

The stateful stack runs on the **control plane Pi** (`10.0.0.1`) via Docker
Compose. It shares the hardware with k3s but runs in a separate orchestrator.
This is intentional — Docker Compose is simpler for single-node stateful
services, and it keeps DB volumes off the Kubernetes SD-card etcd/store path.

```bash
# On the control plane Pi
$ docker compose -f docker-compose.prod.yaml up -d
```

### Services

| Service     | Image                    | Persistence          | Ports  |
|-------------|--------------------------|----------------------|--------|
| `influxdb`  | `influxdb:2`             | Named volume         | 8086   |
| `postgres`  | `postgres:16.4`          | Named volume         | 5432   |
| `grafana`   | `grafana/grafana-oss`    | Named volume         | 3000   |
| `mosquitto` | `eclipse-mosquitto`      | Bind mount (`./mosquitto/`) | 1883, 9001 |

`dirtie-srv` is **not** in this compose file. It lives as a k8s Deployment.

### Networking

Docker Compose creates a bridge network (`dirtie_net`) for inter-container
traffic **on this host only**. It is invisible to k3s pods and worker nodes.

The services publish ports to the **dedicated Ethernet interface** at
`10.0.0.1`. All k3s pods — whether on the control plane or a worker — reach
the databases and broker via the **control plane Pi's static LAN IP**.

```
Worker Pi pod          Control Plane Pi (Docker + k3s + tailscaled)
     |                         |
     | TCP 10.0.0.1:5432       |
     +------------------------> |
                                +--NAT/port-forward--> postgres:5432
                                +--NAT/port-forward--> influxdb:8086
                                +--NAT/port-forward--> mosquitto:1883

Pod on control-plane node
     |
     +--> 10.0.0.1:5432 -------^ (same path, just doesn't leave the box)

Pico on home Wi-Fi
     |
     +--> 10.0.0.1:1883 --------> mosquitto

Android on cellular (via Tailscale)
     |
     +--> 100.x.y.z:8080 ------> Traefik ingress --> dirtie-srv pod
```

**Do not use Docker service names** (e.g., `postgres:5432`) in the k8s
ConfigMap. Use the static LAN IP (`10.0.0.1`).

### Port binding

Bind published ports to the Ethernet IP (`10.0.0.1`). This exposes the services
on the dedicated cluster switch so all nodes and local Wi-Fi devices (Pico)
can reach them. It does **not** expose the ports on the public internet.

```yaml
services:
  influxdb:
    ports:
      - "10.0.0.1:8086:8086"
  postgres:
    ports:
      - "10.0.0.1:5432:5432"
  grafana:
    ports:
      - "10.0.0.1:3000:3000"
  mosquitto:
    ports:
      - "10.0.0.1:1883:1883"
      - "10.0.0.1:9001:9001"
```

**Important:** Docker binds to interfaces by IP address. Using `127.0.0.1`
would break cross-node access. Using `0.0.0.0` would expose the port on all
interfaces (including Wi-Fi / public-facing ones if any).

### Why mosquitto stays in Docker Compose

The Pico dirtie-nodes connect to the MQTT broker over the home Wi-Fi. Keeping
the broker on the most durable node (the control plane Pi) means the broker
address (`10.0.0.1:1883`) is stable and survives worker node reboots. The Pico
connects via Wi-Fi to your home router; as long as the control plane Pi is
reachable from the Wi-Fi subnet, `10.0.0.1` is routable.

If you later move mosquitto into k8s, update the ConfigMap to use the k8s
`Service` DNS name.

### Example `.env` on the control plane Pi

```
INFLUX_USERNAME=admin
INFLUX_PASSWORD=<secret>
INFLUX_ORG=dirtie
INFLUX_DEFAULT_BUCKET=breadcrumbs
INFLUX_TOKEN=<secret>
POSTGRES_DB=dirtie
POSTGRES_USER=dirtie
POSTGRES_PASSWORD=<secret>
```

---

## Cluster Plane (k8s manifests)

### dirtie-srv Deployment

```yaml
# k8s/20-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dirtie-srv
  namespace: dirtie
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dirtie-srv
  template:
    metadata:
      labels:
        app: dirtie-srv
    spec:
      nodeSelector:
        kubernetes.io/arch: arm64
      containers:
        - name: app
          image: ghcr.io/YOURNAME/dirtie-srv:latest
          ports:
            - containerPort: 8080
          envFrom:
            - configMapRef:
                name: dirtie-config
          env:
            - name: INFLUX_TOKEN
              valueFrom:
                secretKeyRef:
                  name: dirtie-secrets
                  key: influx-token
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: dirtie-secrets
                  key: postgres-password
```

### ConfigMap (non-secret env)

```yaml
# k8s/10-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dirtie-config
  namespace: dirtie
data:
  INFLUX_URI: "http://10.0.0.1:8086"
  INFLUX_ORG: "dirtie"
  INFLUX_DEFAULT_BUCKET: "breadcrumbs"
  POSTGRES_SERVER: "10.0.0.1:5432"
  POSTGRES_DB: "dirtie"
  POSTGRES_USER: "dirtie"
  MOSQUITTO_URI: "10.0.0.1:1883"
  APP_HOST: "container"
  ASSETS_DIR: "./assets/"
```

### Secret (sensitive env)

```yaml
# k8s/15-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: dirtie-secrets
  namespace: dirtie
type: Opaque
stringData:
  influx-token: "<token>"
  postgres-password: "<password>"
```

Apply:

```bash
$ kubectl apply -f k8s/10-configmap.yaml
$ kubectl apply -f k8s/15-secret.yaml
$ kubectl apply -f k8s/20-deployment.yaml
$ kubectl apply -f k8s/30-service.yaml
$ kubectl apply -f k8s/40-ingress.yaml
```

---

## Build Pipeline

### Cross-build ARM64 from x86 (laptop / CI)

The `Dockerfile` already sets `CGO_ENABLED=0`, so Go cross-compiles cleanly.

```bash
$ docker buildx build \
  --platform linux/arm64 \
  -t ghcr.io/YOURNAME/dirtie-srv:latest \
  --push .
```

Requirements:
- `docker buildx` with a builder that supports multi-platform (the default
  `docker-container` driver works).
- Logged in to GHCR: `docker login ghcr.io -u YOURNAME`

### Makefile target (optional)

```makefile
# Makefile
IMAGE := ghcr.io/YOURNAME/dirtie-srv

build-arm:
	docker buildx build --platform linux/arm64 -t $(IMAGE):latest --push .
```

### On-cluster update

After the image is pushed, roll out the new deployment:

```bash
$ kubectl rollout restart deployment/dirtie-srv -n dirtie
```

Or use a GitOps tool (Flux, Argo CD) to watch the image tag and reconcile
automatically.

---

## Ingress

The cluster uses the k3s-bundled Traefik. The dirtie-srv API is exposed via
an `IngressRoute` (CRD) or standard `Ingress`.

```yaml
# k8s/40-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dirtie-srv
  namespace: dirtie
spec:
  rules:
    - host: dirtie.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dirtie-srv
                port:
                  number: 8080
```

Access options:

1. **On-LAN / on-switch** — Any device on the home Wi-Fi or the `10.0.0.x`
   switch can reach the ingress at `http://10.0.0.1:8080` (or via
   `dirtie.local` if DNS is configured).

2. **Tailscale (remote clients)** — If the Android phone or dev laptop are on
the same Tailnet, they can reach the ingress through the control plane Pi's
Tailscale IP (`100.x.y.z`). Tailscale DNS or a MagicDNS record can resolve
`dirtie.local` across the Tailnet.

3. **Tailscale Funnel** — Expose a specific port (e.g., 443) to the public
internet via `tailscale funnel`. Useful for demoing the API without requiring
the viewer to install Tailscale. This is handled by the **main node**.

---

## Day-1 Checklist

1. **Wire the nodes** — Connect all 3 Pis to the dedicated Ethernet switch.
   Assign static IPs:
   - Control plane: `10.0.0.1`
   - Worker 1: `10.0.0.2`
   - Worker 2: `10.0.0.3`
2. **Install Tailscale on all nodes** — Authorize them into the same Tailnet.
   Record the control plane Pi's Tailscale IP (`100.x.y.z`) for remote access
   only.
3. **Provision Docker stack** — On the control plane Pi, clone repo, write
   `.env`, ensure `docker-compose.prod.yaml` binds ports to `10.0.0.1`, then
   run `docker compose -f docker-compose.prod.yaml up -d`.
4. **Provision cluster** — Ansible playbooks install k3s on all 3 Pis.
   `kubectl get nodes` shows Ready.
5. **Push image** — `docker buildx build --platform linux/arm64 ... --push`.
6. **Apply manifests** — `kubectl apply -f k8s/` from any node with kubeconfig.
   The ConfigMap already uses `10.0.0.1`.
7. **Verify (on-switch)** — From any Pi or a device on the same network:
   ```bash
   $ curl http://10.0.0.1:8080/health
   $ mosquitto_sub -h 10.0.0.1 -p 1883 -t "test"
   ```
8. **Verify (remote via Tailscale)** — From your dev laptop off-site:
   ```bash
   $ curl http://YOUR_PI_TAILSCALE_IP:8080/health
   ```
9. **Configure Pico / Android** —
   - Pico: Point at `10.0.0.1:1883` for MQTT.
   - Android (on Wi-Fi): Point at `10.0.0.1:8080` for API.
   - Android (on cellular / remote): Point at `YOUR_PI_TAILSCALE_IP:8080` for
     API via Tailscale.

---

## Variables that must stay in sync

| Variable           | Set on              | Consumed by                  |
|--------------------|---------------------|------------------------------|
| `INFLUX_URI`       | ConfigMap           | dirtie-srv                   |
| `POSTGRES_SERVER`  | ConfigMap           | dirtie-srv                   |
| `MOSQUITTO_URI`    | ConfigMap           | dirtie-srv                   |
| `INFLUX_TOKEN`     | Secret              | dirtie-srv                   |
| `POSTGRES_PASSWORD`| Secret              | dirtie-srv                   |
| `MQTT_BROKER_IP`   | Pico                | dirtie-node                  |
| `API_BASE_URL`     | Android             | dirtie-client                |

All database, broker, and on-LAN API addresses point to the **control plane
Pi's static IP (`10.0.0.1`)**. Remote clients use the **control plane Pi's
Tailscale IP**.

Document the mapping in `nodes.md`:

```
| Node                | Role           | Static IP   | Tailscale IP  |
|---------------------|----------------|-------------|---------------|
| pi-control-01       | control + data | 10.0.0.1    | 100.64.1.2    |
| pi-worker-01        | worker         | 10.0.0.2    | 100.64.1.3    |
| pi-worker-02        | worker         | 10.0.0.3    | 100.64.1.4    |
```
