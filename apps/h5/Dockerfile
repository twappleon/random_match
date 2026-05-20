FROM node:22-alpine AS builder

WORKDIR /src/apps/h5

COPY apps/h5/package*.json ./
RUN npm ci

COPY apps/h5 ./
ARG VITE_API_BASE
ENV VITE_API_BASE=$VITE_API_BASE
RUN npm run build

FROM nginx:1.27-alpine

COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /src/apps/h5/dist /usr/share/nginx/html

EXPOSE 80
