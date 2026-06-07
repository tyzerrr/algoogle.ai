FROM golang:1.22-bookworm AS builder

ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /src/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/algosensei-api ./cmd/api

FROM node:22-bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates python3 \
  && rm -rf /var/lib/apt/lists/* \
  && npm install -g @openai/codex@0.137.0

WORKDIR /app/apps/api

COPY --from=builder /out/algosensei-api /usr/local/bin/algosensei-api
COPY apps/api/seed ./seed

ENV PORT=8000
ENV DATABASE_URL=sqlite:////data/algo_sensei.db
ENV PYTHON_BIN=python3
ENV CODEX_CLI_PATH=codex
ENV CODEX_WORKDIR=/app/apps/api
ENV CODEX_CLI_TIMEOUT_SECONDS=180
ENV CODE_WORKSPACE_DIR=/workspace
ENV CODE_WORKSPACE_PUBLIC_DIR=./workspace

EXPOSE 8000

CMD ["algosensei-api"]
