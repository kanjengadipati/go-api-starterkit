FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache make curl \
	&& curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
	| tar xvz \
	&& mv migrate /usr/local/bin/migrate

CMD ["make", "db-setup"]
