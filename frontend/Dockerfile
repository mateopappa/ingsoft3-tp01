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

COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static SPA — node_modules/coverage excluded via .dockerignore
COPY . /usr/share/nginx/html

EXPOSE 80
