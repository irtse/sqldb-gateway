FROM golang:alpine as builder
 
WORKDIR /app
 
COPY . . 

RUN go run main.go

RUN cd plugins/cegid
RUN go build -buildmode=plugin -o plugin.so plugin.go

RUN cd ../autoload_cegid
RUN go build -buildmode=plugin -o plugin.so plugin.go

RUN cd ../..

RUN go run main.go
 
RUN sed -i 's/http:\/\/127.0.0.1:8080\/swagger\/swagger.json/swagger.json/g' swagger/index.html

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .
  
FROM scratch
 
WORKDIR /app

COPY --from=builder /app/project_test.csv /app/project_test.csv 
COPY --from=builder /app/user_test.csv /app/user_test.csv 

COPY ./plugins /app/plugins
COPY --from=builder /app/conf /app/conf
COPY --from=builder /app/sqldb-ws /usr/bin/
COPY --from=builder /app/swagger /app/swagger
COPY --from=builder /app/web /app/web

EXPOSE 8080
 
ENTRYPOINT ["sqldb-ws"]