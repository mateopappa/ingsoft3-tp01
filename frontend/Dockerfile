# ---- Stage 1: install deps + sources (shared by test) ----
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
# DevDependencies (vitest, coverage) are required by the test stage —
# do NOT use --omit=dev here.
RUN npm ci

COPY . .

# ---- Stage 2: unit tests + coverage threshold ----
FROM build AS test
ENV OUT_DIR=/out
ENTRYPOINT ["/bin/sh", "/app/scripts/run-tests.sh"]

# ---- Stage 3: production nginx image (no node, no tests) ----
FROM nginx:alpine AS final

# Template is processed by the official nginx entrypoint: ${VAR} → env value.
COPY default.conf.template /etc/nginx/templates/default.conf.template

# Static SPA — node_modules/coverage excluded via .dockerignore
COPY index.html /usr/share/nginx/html/index.html
COPY css /usr/share/nginx/html/css
COPY static /usr/share/nginx/html/static
COPY js /usr/share/nginx/html/js

# Defaults match docker-compose (service name + Docker DNS).
# Render overrides these with the API URL of THAT environment.
ENV BACKEND_URL=http://backend:8080 \
    DNS_RESOLVER=127.0.0.11

EXPOSE 80
