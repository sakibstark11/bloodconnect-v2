FROM golang:latest

WORKDIR /app

# Install Air for hot reloading
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Start Air
CMD ["air", "-c", ".air.toml"]
