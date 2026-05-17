#!/bin/sh
set -e

: "${PORT:=80}"

envsubst '${BACKEND_URL}' < /usr/share/nginx/html/index.html.template > /usr/share/nginx/html/index.html

cat > /etc/nginx/conf.d/default.conf <<EOF
server {
    listen ${PORT};
    root /usr/share/nginx/html;
    index index.html;
}
EOF

exec nginx -g 'daemon off;'
