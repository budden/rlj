# rlj = Redis and Left Join

## Concept

Redis is cool, but sometimes we need a left joins like those we find in relational RDBMS. 
Here we did an excercise and implemented something like

```
create table client (id, name)

create table order (id, clientid, name, total)

select client.*, [order].* from client left join [order] on client.id = [order].clientid
```

To make this query efficient, usually we need an index. For instance, this one can help: 
```
create index order_by_clientid by [order] (clientid) 
```

## Implementation

The main idea is from [Secondary indexing with Redis](https://redis.io/topics/indexes) which 
is a part of official redis documentation. 

To imitate an "index", which we implemented as a series of hashes with the keys of the form
`order-by-clientid.$clientid`. 

Actually it is not just an "index" (btree), but a hash clustered index, but that does not matter much.
So we have two copies of orders data - one indexed by primary key (id) and another one indexed by clientid.

To implement left join, we scan all clients and join orders to them via this 'index'

To add and access orders safely, we had to introduce a lock at the order table to make sure that no more 
that one client has an access to the table's data. So, `insert into order` is "thread safe" as well as
`select from order where ...`. OTOH, retrieveing our join is not thread safe. Other clients can write
to either table while we're doing the query, so it won't represent the state of the database at
some specific moment. We only lock the time when we read one bucket in our index, so it is only locally
coherent, but not the entire dataset wise. 

Strictly speaking it is a flaw but some SQL RDBMS like Firebird behave the same way. 

## Docs & Tests

Documentation can be found at [godoc](https://godoc.org/github.com/budden/rlj)

There are some tests. We call them integration tests because we use a real database connection. 

Be careful, TESTS AND MAIN PROGRAM FLUSH THE DEFAULT REDIS DATABASE. 

We have 100% coverage on one package only, and only some coverage for the whole program.

Main application creates about 5 records, joins them and prints the result. 

Development was Google driven, there are references to knowledge sources around the code.

## Flaws

As we mentioned before, our joined query is not atomic. 

Obviously, performance sucks. We expect it to be a sort of `O(#clients * #orders-per-client * Log(#clients * #orders-per-client * #keys in the database))`, 
but there is a network roundtrip for each client, so those `Os` are misleading. I think to make things really fast one should use Lua to 
minimize network I/O. Another approach might be using scan and/or multi.

Tests are scant. 

If the app crashes while lock is held, it never recovers. OTOH it will never show an incoherent view of order's 
data, so is it not so bad. 


WBR, Budden


