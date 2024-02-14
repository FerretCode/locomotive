FROM golang:1.21-alpine
WORKDIR /home/locomotive
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . . 
RUN go build -v -o ./bin/ ./...
CMD [ "./bin/locomotive" ]