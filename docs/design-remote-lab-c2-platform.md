# SCRT Remote Lab & C2 Platform — Design Document

**Date**: 2026-03-28
**Status**: Approved — ready for incremental implementation
**Approach**: A (embedded serve mode) → B (split binary) → C (agent model)

---

## Understanding Summary

- **What**: A deployment architecture for `scrt` that adds remote lab capability and lays the foundation for a C2 platform
- **Why**: Personal security research tool (CTFs, HTB, bug bounty) that needs to operate in any environment — local workstation, cloud VM, k8s cluster, air-gapped
- **Who**: Solo operator today; designed to grow toward small team (2–5) and multi-tenant without a rewrite
- **Key constraints**: Docker socket access required; TTY sessions need real terminal; statically linked binary; root/post-privesc assumed for daemon-less operation; Linux only
- **Non-goals (now)**: Rootless containers, Windows/macOS daemon-less, full C2 implant implementation, multi-tenant RBAC

---

## Assumptions

1. Target OS is Linux
2. Post-privesc root is available when Docker is absent
3. HTTPS termination is handled at the edge (Caddy in Compose; nginx ingress + cert-manager in k8s)
4. Air-gapped environments pre-pull images or receive the binary with a bundled OCI tarball
5. "Web UI" in Phase 1 is a read-only status dashboard — no terminal emulator yet
6. The `docker` CLI binary is present when Docker backend is active (TTY sessions)

---

## Architecture: Approach A — Embedded Serve Mode

The `scrt` binary gains a `serve` subcommand. Everything else is unchanged.

```
scrt (binary)
├── existing commands: start, enter, stop, destroy, backup, pull, import, list, config, version
├── serve  ← new: HTTP API + web UI
└── internal/
    ├── backend/       ← NEW: Backend interface + tier detection
    ├── api/           ← NEW: HTTP handlers + embedded web UI
    ├── container/     ← existing (DockerBackend lives here, renamed from Manager)
    ├── tui/           ← existing, unchanged
    ├── config/        ← existing, unchanged
    └── project/       ← existing, unchanged
```

### Progression path

| Phase | Shape | Trigger |
|---|---|---|
| A | Single binary with `serve` subcommand | Now |
| B | `scrt` CLI + `scrtd` daemon, shared `internal/` | When multi-user or process isolation is needed |
| C | `scrt` operator + `scrt agent` with encrypted comms | When C2 operator/agent split is required |

The `backend.Backend` interface is the seam that makes A→B→C a refactor, not a rewrite.

---

## Backend Abstraction (`internal/backend`)

### Interface

```go
// Backend abstracts container runtime operations across tiers.
// Implementations: DockerBackend, ContainerdBackend, OCIBackend.
type Backend interface {
    Start(ctx context.Context, p container.RunParams) error
    Enter(ctx context.Context, project, shell string) error
    Stop(ctx context.Context, project string) error
    Destroy(ctx context.Context, project string) error
    List(ctx context.Context) ([]container.Info, error)
    Pull(ctx context.Context, image string) error
    ImportBackup(ctx context.Context, p container.ImportParams) error
    Close() error
}
```

### Tier detection at startup

```
backend.New(ctx, logger) Backend
  │
  ├─ Docker daemon reachable?                → DockerBackend      (current Manager, renamed)
  ├─ /run/containerd/containerd.sock exists? → ContainerdBackend  (k8s nodes)
  ├─ runc or crun in $PATH?                  → OCIBackend         (bare metal, post-privesc)
  └─ none                                    → fatal: no supported runtime found
```

- `DockerBackend` = current `container.Manager`, moved to `internal/backend/docker.go`
- `ContainerdBackend` and `OCIBackend` are initially stubbed with `ErrNotImplemented`
- Detection is deterministic and logged at startup (`slog` structured log)

---

## `scrt serve` — HTTP API & Web UI

### Command

```
scrt serve [--addr :8080] [--token <api-key>]
```

Token precedence: `--token` flag → `SCRT_TOKEN` env → auto-generated on first run (printed to stderr, persisted to `~/.scrt.token`).

### API surface (v1)

```
GET  /healthz                              unauthenticated — liveness probe
GET  /                                     serves embedded web UI

GET  /api/v1/containers                    list all SCRT containers + state
POST /api/v1/containers/:name/stop
POST /api/v1/containers/:name/destroy
POST /api/v1/containers/:name/backup
GET  /api/v1/containers/:name/logs         query: ?lines=100
POST /api/v1/images/pull                   body: {"image": "..."}
```

All `/api/v1/` routes require `Authorization: Bearer <token>`.

### Web UI (Phase 1)

- Served from `embed.FS` — no Node.js build step, no external CDN
- Plain HTML + vanilla JS — auto-refreshes container table every 10 seconds
- Read-only by default; stop/backup are the only write actions
- No terminal emulator in Phase 1

---

## Image Build Strategy

### Image 1: `scrt` control plane

```dockerfile
# Stage 1 — build
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY scrt/go.mod scrt/go.sum ./
RUN go mod download
COPY scrt/ .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /scrt ./cmd/scrt

# Stage 2 — runtime
FROM scratch
COPY --from=builder /scrt /scrt
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/scrt"]
```

Target size: ~12–18MB. No shell. No package manager. CA certs for registry TLS.

### Image 2: Security research image (what scrt manages)

Profile-based build via `--build-arg PROFILE`:

| Profile | Base | Use case | Approx size |
|---|---|---|---|
| `minimal` | `debian:slim` | Lightweight, fast pull | ~80MB |
| `standard` | `kalilinux/kali-rolling` | Full pentest tooling | ~2–4GB |
| `custom` | Configurable | Engagement-specific | Varies |

---

## Docker Compose Deployment

### Service topology

```
internet → Caddy (automatic HTTPS) → scrt:8080 (serve mode)
                                         ↕
                                  /var/run/docker.sock
```

### `compose.yaml`

```yaml
services:
  scrt:
    image: ghcr.io/alexrf45/scrt:latest
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - scrt-workspaces:/workspaces
      - scrt-config:/root
    environment:
      SCRT_WORKDIR: /workspaces
      SCRT_TOKEN:   ${SCRT_TOKEN}
    command: ["serve", "--addr", ":8080"]
    expose:
      - "8080"
    networks:
      - internal

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy-data:/data
      - caddy-config:/config
    networks:
      - internal
      - external
    depends_on:
      - scrt

volumes:
  scrt-workspaces:
  scrt-config:
  caddy-data:
  caddy-config:

networks:
  internal:
  external:
```

### `Caddyfile`

```
lab.yourdomain.com {
    reverse_proxy scrt:8080
}
```

### Security notes

- Docker socket mounted read-only — write ops are scoped through the SDK
- `SCRT_TOKEN` injected at runtime via `.env` — never baked into the image
- `scrt` port not exposed to host — only Caddy reaches it
- Caddy handles ACME certificate lifecycle automatically

---

## Kubernetes Deployment

### Security posture — read this first

This manifest requires `privileged: true`. This grants full node access. Scope it to a **dedicated, tainted security research node** — never deploy to shared infrastructure.

### Socket path by distribution

| Distribution | Socket path |
|---|---|
| Standard k8s / EKS / GKE / AKS | `/run/containerd/containerd.sock` |
| k3s | `/run/k3s/containerd/containerd.sock` |
| Docker-based node | `/var/run/docker.sock` |

### Manifests

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: scrt
```

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: scrt-token
  namespace: scrt
type: Opaque
stringData:
  token: ""   # inject via kubectl or external-secrets operator — never hardcode
```

```yaml
# statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: scrt
  namespace: scrt
spec:
  serviceName: scrt
  replicas: 1   # hard limit — concurrent instances corrupt shared state
  selector:
    matchLabels:
      app: scrt
  template:
    metadata:
      labels:
        app: scrt
    spec:
      nodeSelector:
        role: security-research
      tolerations:
        - key: security-research
          operator: Exists
          effect: NoSchedule
      containers:
        - name: scrt
          image: ghcr.io/alexrf45/scrt:latest
          args: ["serve", "--addr", ":8080"]
          env:
            - name: SCRT_WORKDIR
              value: /workspaces
            - name: SCRT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: scrt-token
                  key: token
          ports:
            - containerPort: 8080
          securityContext:
            privileged: true   # required for container runtime operations
          volumeMounts:
            - name: workspaces
              mountPath: /workspaces
            - name: runtime-socket
              mountPath: /run/containerd/containerd.sock
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 15
      volumes:
        - name: runtime-socket
          hostPath:
            path: /run/containerd/containerd.sock   # adjust per distribution
            type: Socket
  volumeClaimTemplates:
    - metadata:
        name: workspaces
      spec:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 20Gi
```

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: scrt
  namespace: scrt
spec:
  selector:
    app: scrt
  ports:
    - port: 8080
      targetPort: 8080
  type: ClusterIP
```

```yaml
# ingress.yaml — requires cert-manager + nginx ingress controller
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: scrt
  namespace: scrt
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts: [lab.yourdomain.com]
      secretName: scrt-tls
  rules:
    - host: lab.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: scrt
                port:
                  number: 8080
```

---

## Decision Log

| # | Decision | Alternatives Considered | Rationale |
|---|---|---|---|
| 1 | Approach A — embedded `serve` mode | Split binary (B), Agent model (C) | Single artifact; drops anywhere; A→B→C is a refactor not a rewrite |
| 2 | `backend.Backend` interface in `internal/backend` | Keeping `container.Manager` as-is | Clean seam for tier detection; isolates Docker dependency; enables daemon-less operation |
| 3 | Tier detection order: Docker → containerd → runc | DinD, always-Docker | Matches realistic environment availability; no daemon overhead |
| 4 | Root / post-privesc privilege assumption | Rootless containers | Realistic for red team use; avoids user namespace complexity |
| 5 | Static binary — `CGO_ENABLED=0`, `scratch` final stage | Alpine, distroless | Zero runtime dependencies; drops clean on any Linux host |
| 6 | Two distinct images — control plane vs. research | Single fat image | Control plane stays minimal; research image size is a separate concern |
| 7 | Research image profiles — minimal / standard / custom | Single image with all tools | Right-size for engagement; avoids 4GB default pull |
| 8 | Caddy for TLS in Compose | nginx, Traefik | Automatic ACME certs; zero config for single-domain |
| 9 | StatefulSet in k8s | Deployment | Workspaces are stateful; `volumeClaimTemplates` enforces per-pod PVC |
| 10 | `privileged: true` in k8s | Specific capability list | Honest — capability lists achieving the same surface are security theater |
| 11 | Node taint/toleration for k8s | Namespace isolation alone | Privileged workloads must not share nodes with other workloads |
| 12 | Single static bearer token for auth | mTLS, OIDC, JWT | Sufficient for solo operator; middleware is the swap point for multi-user |
| 13 | Vanilla JS web UI from `embed.FS` | React/Vue with build toolchain | No Node.js build step; binary stays self-contained |
| 14 | `/healthz` unauthenticated | Auth on all routes | Standard k8s liveness probe pattern; exposes no sensitive data |
