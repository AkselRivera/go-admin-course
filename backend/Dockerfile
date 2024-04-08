FROM golang:1.22


WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download


COPY . .

# RUN go install github.com/cosmtrek/air@latest

# CMD ["air", "-c", "air.toml"]
CMD [ "go", "run", "main.go" ]