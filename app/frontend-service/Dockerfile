FROM nginx:mainline-alpine3.20-slim

COPY ./nginx.conf ./etc/nginx/nginx.conf

COPY ./page/* ./usr/share/nginx/html/

EXPOSE 9090