# Start from the latest golang base image
FROM golang:1.20.6

# Add Maintainer Info
LABEL maintainer="Daniel Lizarazo <daniel7lizarazo@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

ENV GO111MODULE=on

ENV DBADRESS=weeding-database.c0mkdcum6dbh.us-east-2.rds.amazonaws.com:3306
ENV DBNAME=WeddingDB
ENV DBPASS=daniel7ayde
ENV DBUSER=danielAdmin
ENV LOCALPORT=:8080

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]