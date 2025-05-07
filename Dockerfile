FROM golang:alpine as builder
 
WORKDIR /app
 
COPY . .
 
RUN apk add git
 
RUN go get github.com/beego/bee/v2 && go install github.com/beego/bee/v2@master


RUN timeout 15 bee run -gendoc=true -downdoc=true -runmode=dev || :
 
RUN sed -i 's/http:\/\/127.0.0.1:8080\/swagger\/swagger.json/swagger.json/g' swagger/index.html
 
RUN cd plugins/cegid && go build -buildmode=plugin -o plugin.so plugin.go
RUN ../autoload_cegid && go build -buildmode=plugin -o plugin.so plugin.go
RUN ../..

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .
  
FROM scratch
 
WORKDIR /app

COPY --from=builder /app/project_test.csv /app/project_test.csv 
COPY --from=builder /app/user_test.csv /app/user_test.csv 

COPY --from=builder /app/plugins /app/plugins
COPY --from=builder /app/conf /app/conf
COPY --from=builder /app/sqldb-ws /usr/bin/
COPY --from=builder /app/swagger /app/swagger
COPY --from=builder /app/web /app/web

EXPOSE 8080
 
ENTRYPOINT ["sqldb-ws"]