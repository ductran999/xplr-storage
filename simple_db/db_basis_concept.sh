#!/bin/bash
db_set () {
 echo "$1,$2" >> database
}

db_get () {
 grep "^$1," database | sed -e "s/^$1,//" | tail -n 1
}

db_set 42 {"name":"xplr","age":1}
db_set 42 {"name":"jovian","age":2}

db_get 42