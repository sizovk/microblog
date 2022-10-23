FROM golang:1.19

WORKDIR /microblog

COPY . .

RUN go mod tidy
RUN go build -o microblog

CMD ["/microblog/microblog"]