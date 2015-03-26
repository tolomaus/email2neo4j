#### Summary:

These two little go command line tool allows you to import your emails from either a mail server via IMAP or an mbox file into a neo4j graph database. I have used the model from the email example project that is explained in Chapter 3 of the book [Graph Databases](http://graphdatabases.com/) by Ian Robinson, Jim Webber and Emil Eifrem (which is a truly great introduction to the graph btw)

#### Usage:
```shell
imap2neo4j imap.myserver.com Inbox user@myserver.com password [paging, eg import by batches of 1000] [specific range of messages, eg 5000:*]
```
```shell
mbox2neo4j /path/to/mbox/file
```

It assumes that you have neo4j installed on http://localhost:7474/db/data

#### Note:
It *shouldn't* do any harm to your emails as all it does is read it's headers but you may test it first on a folder with less important mails, just to be sure.

#### mbox
On a Mac you can export your mails to an mbox file with Mac Mail as well as Outlook for Mac. On a Windows you may have to install a "whatever format"-to-mbox converter.

#### Some useful neo4j/Cypher queries:

```sql
// Create meta graph  
  MATCH (a)-[r]->(b)   
  WITH labels(a) AS a_labels,type(r) AS rel_type,labels(b) AS b_labels   
  UNWIND a_labels as l   
  UNWIND b_labels as l2   
  MERGE (a:Meta_Node {name:l})   
```
and
```sql
// Show meta graph 
MATCH (n:Meta_Node) RETURN n
```
![alt text](https://github.com/tolomaus/email2neo4j/blob/master/images/model.png "model")

```sql
//Biggest email threads
MATCH p=(e:Email)<-[:REPLY*]-(r:Email)<-[]-(sender:Account)
WHERE NOT (e)-[:REPLY]->()
RETURN sender.name, e.subject, Id(e), length(p) - 1 AS depth
ORDER BY depth DESC
LIMIT 100
```
This will return a table with the id's of the starting emails which can be used in the queries below

```sql
// Email thread by email id
MATCH p=(n:Email)<-[:REPLY*]-(:Email)
WHERE id(n)=123
RETURN p
```
![alt text](https://github.com/tolomaus/email2neo4j/blob/master/images/emailthread.png "Email thread")

```sql
// Email thread by email id now with all accounts
MATCH p=(:Account)-[]-(n:Email)<-[:REPLY*]-(:Email)-[]-(:Account)
WHERE id(n)=135256
RETURN p
```
![alt text](https://github.com/tolomaus/email2neo4j/blob/master/images/emailthreadwithaccounts.png "Email thread with accounts")
