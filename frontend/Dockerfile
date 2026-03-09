# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies (ci = exact lockfile, no scripts)
COPY package*.json ./
RUN npm ci --prefer-offline

# Copy source and build
COPY . .
RUN npm run build

# ── Stage 2: Serve ────────────────────────────────────────────────────────────
FROM nginx:1.25-alpine

# envsubst is included in nginx:alpine (part of gettext)
RUN apk --no-cache add gettext

# Clear default site config
RUN rm -f /etc/nginx/conf.d/default.conf

# Static files
COPY --from=builder /app/dist /usr/share/nginx/html

# Nginx config template (processed at container start by entrypoint)
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Custom entrypoint: substitutes ${BACKEND_URL}, then starts nginx
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
