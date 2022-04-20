Generic database webservice layer
=================================
This code publishes automatically CRUD REST web services for all tables available in a SQL database.  
Some special tables can be used for defining database access restrictions based on an RBAC model.

Running in debug mode
---------------------
> bee run -downdoc=true -gendoc=true

Testing
-------
 http://localhost:8080/swagger