FROM golang:1.24-alpine
WORKDIR /app
RUN apk add --no-cache git
COPY . .
EXPOSE 8080
CMD ["go", "run", "main.go"]