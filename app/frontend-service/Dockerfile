FROM nginx:stable-alpine-slim

COPY frontend-service/nginx.conf ./etc/nginx/conf.d/default.conf

COPY frontend-service/page/* ./usr/share/nginx/html/

EXPOSE 8080