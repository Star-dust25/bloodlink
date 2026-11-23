# ==============================================================================
# Stage 1: Build Angular Production SPA
# ==============================================================================
FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# ==============================================================================
# Stage 2: Serve with Nginx 1.27 Alpine + Reverse Proxy
# ==============================================================================
FROM nginx:1.27-alpine

RUN apk add --no-cache wget

ENV BACKEND_URL=http://backend:8080

COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=builder /app/dist/frontend/browser /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1/ || exit 1
