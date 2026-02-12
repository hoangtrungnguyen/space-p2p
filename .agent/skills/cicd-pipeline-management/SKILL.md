---
name: CI/CD Pipeline Management
description: Orchestrate automated build, test, and deployment pipelines for Go LiveKit Edge Nodes.
---

# CI/CD Pipeline Decision Tree

This skill manages the automated lifecycle from commit to production deployment.

1.  **Build Automation**
    *   **Trigger**: Push to `main` branch or manual dispatch.
    *   **Action**: Use GitHub Actions `workflow.yml`.
    *   **Tools**:
        *   `actions/setup-go` (specify Go version 1.25.6).
        *   `docker build` (multi-stage).

2.  **Containerization Strategy**
    *   **Trigger**: Docker Image Creation.
    *   **Action**: Use Multi-Stage Dockerfile.
    *   **Stage 1 (Build)**: `golang:1.25-alpine`.
        *   Set `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`.
        *   Flag: `-ldflags="-w -s"` (Strip DWARF/symbols).
    *   **Stage 2 (Run)**: `scratch` or `alpine:latest`.
        *   Create non-privileged user (UID 10001).
        *   Copy binary from Stage 1.
        *   Expose ports (8080/TCP, 7880/TCP, 7881/TCP, 50000-60000/UDP).

3.  **Testing Gate**
    *   **Trigger**: PR or Push.
    *   **Action**: Run TDD coverage scripts (`check_coverage.sh`).
    *   **Constraint**: FAIL build if coverage < **80%**.

4.  **Deployment Orchestration**
    *   **Trigger**: Successful build on `main`.
    *   **Action**: Execute `scripts/deployment.sh`.
    *   **Steps**:
        1.  **Transfer**: SCP binary + `.env` + `livekit.yaml`.
        2.  **SFU Init**: Check/Install `livekit-server`. Systemd `LimitNOFILE=65535`.
        3.  **Service Def**: Apply `space-p2p.service`. `After=network.target livekit-server.service`.
        4.  **Firewall**: Open UFW/firewalld ports.
        5.  **Health Check**: reload systemd -> curl `localhost:8080/rooms` -> Expect 200 OK.
