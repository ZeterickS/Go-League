FROM golang:1.23.4-alpine

WORKDIR /app

# Copy the entire project into the container
COPY . .

RUN go mod download

# Build the Go application
RUN go build -o main .

# Ensure migration files, helper scripts, and assets are copied
COPY ./migrations/db /app/migrations/db
COPY ./assets /app/assets

# Start the application
CMD ["/app/main"]