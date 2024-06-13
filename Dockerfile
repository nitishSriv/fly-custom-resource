FROM golang:1.19 AS builder

WORKDIR /
COPY . .

RUN go build -o /opt/resource/check ./main.go
RUN go build -o /opt/resource/in ./main.go
RUN go build -o /opt/resource/out ./main.go

FROM alpine:3.15

COPY --from=builder /opt/resource/check /opt/resource/check
COPY --from=builder /opt/resource/in /opt/resource/in
COPY --from=builder /opt/resource/out /opt/resource/out

RUN chmod +x /opt/resource/check /opt/resource/in /opt/resource/out

CMD ["/opt/resource/check"]
