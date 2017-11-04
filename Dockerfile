FROM golang

RUN go get -u gopkg.in/russross/blackfriday.v2

COPY ./src/ /app/

RUN go build -o /app/app /app/app.go

RUN chmod u+x /app/app

EXPOSE 80

CMD ["/app/app"]
