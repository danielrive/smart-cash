FROM nginx:stable-alpine-slim

COPY frontend-service/nginx.conf ./etc/nginx/nginx.conf

COPY frontend-service/page/* ./usr/share/nginx/html/

EXPOSE 8080