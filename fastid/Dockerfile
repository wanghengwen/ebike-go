FROM alpine:latest
WORKDIR /app
COPY build/luopingtech-fastid-register .
COPY conf/application.yml /app/conf/application.yml

EXPOSE 8080
ENTRYPOINT ["./luopingtech-fastid-register"]
