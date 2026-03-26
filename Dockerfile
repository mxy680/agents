# ============================================================
# Multi-stage build: Go binary + Node.js portal + Agent SDK
# ============================================================

# Stage 1: Build Go integrations binary
FROM golang:1.25-bookworm AS go-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/integrations ./cmd/integrations/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/sync-templates ./cmd/sync-templates/

# Stage 2: Build Next.js portal
FROM node:22-bookworm-slim AS portal-builder
WORKDIR /build/portal
COPY portal/package.json portal/package-lock.json ./
RUN npm ci --production=false
COPY portal/ ./
# Build requires env vars at build time for Next.js
ARG NEXT_PUBLIC_SUPABASE_URL
ARG NEXT_PUBLIC_SUPABASE_ANON_KEY
RUN npm run build

# Stage 3: Runtime
FROM node:22-bookworm-slim

# Install system deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    python3 \
    python3-pip \
    texlive-base \
    texlive-latex-extra \
    texlive-fonts-recommended \
    texlive-latex-recommended \
    && rm -rf /var/lib/apt/lists/*

# Install Python deps for pipeline scripts
RUN pip3 install --break-system-packages openpyxl

# Install Claude Agent SDK globally
RUN npm install -g @anthropic-ai/claude-agent-sdk

# Install Doppler CLI
RUN curl -sL https://cli.doppler.com/install.sh | sh

# Copy Go binaries
COPY --from=go-builder /bin/integrations /usr/local/bin/integrations
COPY --from=go-builder /bin/sync-templates /usr/local/bin/sync-templates

# Copy portal
WORKDIR /app
COPY --from=portal-builder /build/portal/.next ./.next
COPY --from=portal-builder /build/portal/node_modules ./node_modules
COPY --from=portal-builder /build/portal/package.json ./package.json
COPY --from=portal-builder /build/portal/public ./public

# Copy agent files
COPY agents/ /app/agents/

# Expose port
ENV PORT=3000
EXPOSE 3000

# Doppler injects all secrets at runtime via DOPPLER_TOKEN env var
CMD ["doppler", "run", "--", "npm", "start"]
