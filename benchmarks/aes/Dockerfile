# 1: Builder
FROM vhiveease/golang-builder:latest AS aesGoBuilder
WORKDIR /app/app
RUN apt-get install git
RUN pwd && ls
COPY ./go/go.mod ./ 
COPY ./go/go.sum ./ 
COPY ./go/server.go ./ 

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./server server.go

# 2: Runner
FROM scratch as aesGo
WORKDIR /app/
COPY --from=aesGoBuilder /app/app/server .

ENTRYPOINT [ "/app/server" ]