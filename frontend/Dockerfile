FROM nginx:alpine

# Remove default nginx static assets and replace with our configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy static frontend web application
COPY . /usr/share/nginx/html

EXPOSE 80
