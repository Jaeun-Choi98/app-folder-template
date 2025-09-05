FROM golang:1.24.6 as go-build
WORKDIR /app
COPY ./ ./
WORKDIR ./cmd/server
RUN go build -o ./main main.go

FROM golang:1.24.6
WORKDIR /home/app
COPY --from=go-build /app/cmd/server/ /home/app/
CMD [ "./main" ]