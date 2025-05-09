Generic database webservice layer
=================================

WARNING : THIS IS A DEVELOPMENT VERSION OF THE SERVICE "LIB" MUST BE EXTERNALIZE OR PARSED INTO MULTIPLE LIB ACTUALLY EMBEDDED.
THIS VERSION IS NOT AT ITS FINEST (only full generic accessor revised to give access to table, rows, columns level in DB)
Actually those are the features implemented (but not confirmed) :
- Generic accessor on DB table at level : Schema (Table), Column & Rows
    Those level are invoke by presence of query params such as : "rows" : to invoke the overhidde rows Manager , "columns" : to manage columns. Basic param to those is "all" 
    exemples : 
        - table params will focus on a particular table. 
        - rows="all" (POST.PUT.DELETE or GET ALL), rows=1 (GET row id=1). Row alway overhide columns
        - columns="all" (POST.PUT.DELETE or GET ALL), columns=name (Manage name column). With rows, columns will refine the results.
- High Query filter : any field of a table could be use in query params to filter results. 
- Alternative Token Authentication.
- DB integrity protection.
- Auto generated docs depending on schemas...
- Basic permission on table. 
- DBschema (elder DBform) will automatically generate a new table, DBschema_field (elder DBformfield) will automatically generate a new row in table. 
- And so on but i'm bored to continue... 
"COMING SOON FOR NEXT"

This code publishes automatically CRUD REST web services for all tables available in a SQL database.  
Some special tables can be used for defining database access restrictions based on an RBAC model.

> export GOPRIVATE=forge.redroom.link
before doing go mod tidy.

To build :

    bee generate routers
    bee run -gendoc=true -downdoc=true

RUN GO PROJECT 
`go run main.go` OR `authmode=token go run main.go`

RUN DB
Prerequisite : Get docker on your local machine. 

- `cd ./db`
(Optionnal) - `cp <my file path> ./db/autoload` 
- `<sudo> docker-compose up`

You can add any sql file in ./db/autoload and db will start with your SQL at run. 

Raw DB config.
    Type : PostgresSQL
    Host: db_pg_1
    Database: test
    User : test
    Password : test

Adminer : localhost:8888
SQLDB-WS SWAGGER : [localhost:8080/swagger](http://localhost:8080/swagger)

Super Admin SQLDB-WS
username : root
password : admin

{
  "login": "root",
  "password": "admin"
}