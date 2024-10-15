FROM nginx:mainline-alpine3.20-slim

COPY frontend-service/nginx.conf ./etc/nginx/nginx.conf

COPY frontend-service/page/* ./usr/share/nginx/html/

EXPOSE 9090