FROM nginxinc/nginx-unprivileged:alpine

COPY frontend-service/nginx.conf ./etc/nginx/nginx.conf

COPY frontend-service/page/* ./usr/share/nginx/html/

EXPOSE 9090