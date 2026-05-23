FROM node:22-alpine AS builder

WORKDIR /src/apps/h5

COPY apps/h5/package*.json ./
RUN npm ci

COPY apps/h5 ./
ARG VITE_API_BASE
ARG VITE_ICE_SERVERS
ARG VITE_FORCE_TURN
ARG VITE_VAPID_PUBLIC_KEY
ENV VITE_API_BASE=$VITE_API_BASE
ENV VITE_ICE_SERVERS=$VITE_ICE_SERVERS
ENV VITE_FORCE_TURN=$VITE_FORCE_TURN
ENV VITE_VAPID_PUBLIC_KEY=$VITE_VAPID_PUBLIC_KEY
RUN npm run build

FROM nginx:1.27-alpine

COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /src/apps/h5/dist /usr/share/nginx/html

EXPOSE 80
