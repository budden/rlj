# rlj = Redis and Left Join

Redis is cool, but sometimes we need a left joins like those we find in relational RDBMS. 
Here we did an excercise and implemented something like

```
create table client (id, name)

create table order (id, clientid, name, total)

select client.*, [order].* from client left join [order] on client.id = [order].clientid
```

For that, we created an "index", which we implemented as a bunch of hashes of the form
`order-by-clientid.$clientid`. Thus we imitated 

```
create index order_by_clientid by [order] 
```

Actually it is not just an "index" (btree), but a hash clustered index, but that does not matter much.
So we have two copies of orders data - one indexed by primary key (id) and another one indexed by clientid.

To implement left join, we scal all clients and join orders to them via order_by_clientid index.

To add records safely, we had to introduce a lock at the order table to make sure that no more that one
client has an access to the table's data. So, `insert into order` is "thread safe". OTOH, 
retrieveing data is not thread safe. Other clients can write to either table while we're doing the query. 
We only lock the time when we read one bucket in our index, but we don't lock the entire table or
entire database during the time of the query. But that's only an excersize :) 

Documentation can be found at [godoc](https://godoc.org/github.com/budden/rlj)

There are some tests. We call them integration test because we use a real database connection. 
Be careful, tests flush the default redis database. We have 100% coverage on one package only, 
and only some coverage for the whole program.

Main application creates about 5 records, joins them and prints the result. 

Development was Google driven, there are references to knowledge sources around the code.
The main idea is from [Secondary indexing with Redis](https://redis.io/topics/indexes) which 
is a part of official redis documentation. 

Obviously, performance sucks. We expect it to be a sort of `O(#clients * #orders-per-client * Log(#clients * #orders-per-client * #keys in the database))`, 
but there is a network roundtrip for each client, so those `Os` are misleading. I think to make things really fast one should use Lua to 
minimize network I/O. Another approach might be using scan and/or multi.


WBR, Budden


