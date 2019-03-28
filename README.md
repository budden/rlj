# rlj = Redis and Left Join

## Concept

Redis is cool, but sometimes we need LEFT JOINSs like those we find in relational databases. 
Here we did an excercise and implemented something like the following SQL code snippet

```
create table client (id, name)

create table [order] (id, clientid, name, total)

select client.*, [order].* from client left join [order] on client.id = [order].clientid
```
To make this query efficient, usually we need an index. To make things more meaningful, we use this one: 
```
create index order_by_clientid by [order] (clientid) 
```

## Implementation

The main idea comes from [Secondary indexing with Redis](https://redis.io/topics/indexes), which 
is a part of official redis documentation. 

We implement a primary key index for each table by two hashes named `client` and `order`. Each hash maps
ID to the json serialized data. 

To imitate the order_by_clientid "index", we use a series of hashes in the Redis database, where each hash 
represents a bucket of orders of a given client, and keys for those buckets are of the form
`order-by-clientid.$clientid`. 

Actually it is not just an "index" (btree), but a hash clustered index, so we have two copies of each order, 
one in the primary key index and another one in the order_by_clientid index.

To implement left join, we SCAN `client` hash and join orders to them via `order-by-clientid`

We have two indexes for client which must be coherent for any reader and writer. So we had to introduce 
a lock at the orders. We used a simplest form of global lock, which is based on SETNX idiom, see
[](pkg/leftjoin/lock.go). Every read and write operation obtains and holds this lock. 

## Installation & Run

### OS
We only tested on Debian 9.4 (stretch)

### Redis
```
>sudo apt-get install redis-server
>redis-cli --version
redis-cli 3.2.6
```

### Golang
And also golang (`go version` shows 1.11.6). See the [installation manual](https://golang.org/doc/install)

### Rlj
Be careful, TESTS AND MAIN PROGRAM CLEAR THE DEFAULT REDIS DATABASE. 
```
>go get github.com/budden/rlj
>cd $GOPATH/src/github.com/budden/rlj
>go get ./...
>go test ./...
>go run main.go
&{1 Vasya} <=> &{1 1 Car {{false [100]} {false []}}}
&{2 Маша} <=> &{2 2 Dress {{false [50]} {false []}}}
&{2 Маша} <=> &{3 2 Туфельки {{false [50]} {false []}}}
```
Main application creates about 5 records, joins them and prints the result. 

## Docs & Tests

Packages documentation is scant. It can be found at [godoc](https://godoc.org/github.com/budden/rlj)

There are some tests. We call them integration tests because we use a real database connection. 
We have 100% coverage on one package only, and only some coverage for the whole program.

Development was Google driven, there are references to knowledge sources around the code.

## TODOs and Flaws

- no full CRUD implementation. We only implemented insert.
- joined query is not atomic. For each client, bucket of client's orders is obtained atomically.
However, scan of clients is not atomic, see [the consequences](https://redis.io/commands/scan#scan-guarantees)
Also orders can be updated in between of two bucket's reads. As a result, the overall result is inconsistent
if other connections are updating either clients or orders. This could be remedied by adding more locking, or 
by implementing a multi-version architecture like that found in Oracle and Postgresql. As an apology, Firebird is an 
example of SQL DBMS with non-atomic selects. 
- performance sucks. We expect that runtime complexity is a sort of `O(#clients * #orders-per-client * Log(#clients * #orders-per-client * #keys in the database))`, but there is a network roundtrip for each client, so those `«O»s` are misleading. I think to make things really fast one should use Lua to minimize network I/O. Another approach might be using scan and/or multi.
- test coverage is low
- if the app crashes while the orders lock is held, lock is never released and all subsequent `order` queries would fail with timeout. To fix that, we could implement a sort of "recover broken database" tool and procedure. All SQL DBMSs have this kind of tool, which requires an exclusive access to the database. In a simplest form, we could recreate a secondary index using the data from the primary one and then release a lock. This would make the DB operational again, but the status of last transaction is left unclear. From the connections's POV, last transaction failed. But we don't know if the last transaction was recorded in the DB or not. So from the DB's POV, last transaction could succeed. 

## Conclusion

We just demonstrated a pattern of making a sort of LEFT JOIN in a Redis
