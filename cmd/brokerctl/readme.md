# broker-client

## P2P 快速开始

### 创建项目/查询项目(Alice)
```shell
# create project demo in alice
# NOTE: we specify the project-id to simplify the description, generally you should make sure the id is unique or ignore this flag and use the automatically generated one
./brokerctl create project --project-id "demo" --host http://localhost:8081
# check project's information
./brokerctl get project --host http://localhost:8081
[fetch]
+-----------+---------+---------+----------------------------------+
| ProjectId | Creator | Members |               Conf               |
+-----------+---------+---------+----------------------------------+
| demo      | alice   | [alice] | {                                |
|           |         |         |   "protocol":  "SEMI2K",         |
|           |         |         |   "field":  "FM64"               |
|           |         |         | }                                |
+-----------+---------+---------+----------------------------------+
```

### 邀请参与方加入(Alice)
```shell
# alice invite bob to join the project
./brokerctl invite bob --project-id "demo" --host http://localhost:8081
# bob check invitation list
./brokerctl get invitation --host http://localhost:8082
[fetch]
+--------------+---------+---------+-----------+---------+---------+----------------------------------+
| InvitationId | Status  | Inviter | ProjectId | Creator | Members |               Conf               |
+--------------+---------+---------+-----------+---------+---------+----------------------------------+
|            1 | Pending | alice   | demo      | alice   | [alice] | {                                |
|              |         |         |           |         |         |   "protocol":  "SEMI2K",         |
|              |         |         |           |         |         |   "field":  "FM64"               |
|              |         |         |           |         |         | }                                |
+--------------+---------+---------+-----------+---------+---------+----------------------------------+

```

### 参与方接受邀请(Bob)
```shell
# bob decide to join the project with invitation-id 1
./brokerctl process invitation 1 --response "accept" --project-id "demo" --host http://localhost:8082
# check the project, its members should contain alice and bob
./brokerctl get project --host http://localhost:8081
[fetch]
+-----------+---------+-------------+----------------------------------+
| ProjectId | Creator |   Members   |               Conf               |
+-----------+---------+-------------+----------------------------------+
| demo      | alice   | [alice bob] | {                                |
|           |         |             |   "protocol":  "SEMI2K",         |
|           |         |             |   "field":  "FM64"               |
|           |         |             | }                                |
+-----------+---------+-------------+----------------------------------+
```

### 创建数据表
实际上是把数据库中的已存在的表做了一下映射
```shell
# create table for alice
./brokerctl create table ta --project-id "demo" --columns "ID string, credit_rank int, income int, age int" --ref-table alice.user_credit --db-type mysql --host http://localhost:8081
# check the table ta
./brokerctl get table ta --host http://localhost:8081 --project-id "demo"
[fetch]
TableName: ta, Owner: alice, RefTable: alice.user_credit, DBType: mysql
Columns:
+-------------+----------+
| ColumnName  | DataType |
+-------------+----------+
| age         | int      |
| credit_rank | int      |
| ID          | string   |
| income      | int      |
+-------------+----------+

# create table for bob
./brokerctl create table tb --project-id "demo" --columns "ID string, order_amount double, is_active int" --ref-table bob.user_stats --db-type mysql --host http://localhost:8082
# check the table tb
./brokerctl get table tb --host http://localhost:8082 --project-id "demo"
[fetch]
TableName: tb, Owner: bob, RefTable: bob.user_stats, DBType: mysql
Columns:
+--------------+----------+
|  ColumnName  | DataType |
+--------------+----------+
| ID           | string   |
| is_active    | int      |
| order_amount | double   |
+--------------+----------+
```

### 删除表
删除表的映射
```shell
./brokerctl create table ta --project-id "demo" --host http://localhost:8081
```

### 授权 CCL
```shell
# alice set CCL for table ta
./brokerctl grant alice PLAINTEXT --project-id "demo" --table-name ta --column-name ID --host http://localhost:8081
./brokerctl grant alice PLAINTEXT --project-id "demo" --table-name ta --column-name credit_rank --host http://localhost:8081
./brokerctl grant alice PLAINTEXT --project-id "demo" --table-name ta --column-name income --host http://localhost:8081
./brokerctl grant alice PLAINTEXT --project-id "demo" --table-name ta --column-name age --host http://localhost:8081

./brokerctl grant bob PLAINTEXT_AFTER_JOIN --project-id "demo" --table-name ta --column-name ID --host http://localhost:8081
./brokerctl grant bob PLAINTEXT_AFTER_GROUP_BY --project-id "demo" --table-name ta --column-name credit_rank --host http://localhost:8081
./brokerctl grant bob PLAINTEXT_AFTER_AGGREGATE --project-id "demo" --table-name ta --column-name income --host http://localhost:8081
./brokerctl grant bob PLAINTEXT_AFTER_COMPARE --project-id "demo" --table-name ta --column-name age --host http://localhost:8081
# bob set ccl for table tb
./brokerctl grant bob PLAINTEXT --project-id "demo" --table-name tb --column-name ID --host http://localhost:8082
./brokerctl grant bob PLAINTEXT --project-id "demo" --table-name tb --column-name order_amount --host http://localhost:8082
./brokerctl grant bob PLAINTEXT --project-id "demo" --table-name tb --column-name is_active --host http://localhost:8082

./brokerctl grant alice PLAINTEXT_AFTER_JOIN --project-id "demo" --table-name tb --column-name ID --host http://localhost:8082
./brokerctl grant alice PLAINTEXT_AFTER_COMPARE --project-id "demo" --table-name tb --column-name is_active --host http://localhost:8082
./brokerctl grant alice PLAINTEXT_AFTER_AGGREGATE --project-id "demo" --table-name tb --column-name order_amount --host http://localhost:8082

# show grants for alice
# NOTE: you can add flag tables to specify table like: --tables ta
./brokerctl get ccl  --project-id "demo" --parties alice --host http://localhost:8081
[fetch]
+-----------+-----------+--------------+---------------------------+
| PartyCode | TableName |  ColumnName  |        Constraint         |
+-----------+-----------+--------------+---------------------------+
| alice     | ta        | age          | PLAINTEXT                 |
| alice     | ta        | credit_rank  | PLAINTEXT                 |
| alice     | ta        | ID           | PLAINTEXT                 |
| alice     | ta        | income       | PLAINTEXT                 |
| alice     | tb        | ID           | PLAINTEXT_AFTER_JOIN      |
| alice     | tb        | is_active    | PLAINTEXT_AFTER_COMPARE   |
| alice     | tb        | order_amount | PLAINTEXT_AFTER_AGGREGATE |
+-----------+-----------+--------------+---------------------------+
# show grants for bob
./brokerctl get ccl  --project-id "demo" --parties bob --host http://localhost:8081
[fetch]
+-----------+-----------+--------------+---------------------------+
| PartyCode | TableName |  ColumnName  |        Constraint         |
+-----------+-----------+--------------+---------------------------+
| bob       | ta        | age          | PLAINTEXT_AFTER_COMPARE   |
| bob       | ta        | credit_rank  | PLAINTEXT_AFTER_GROUP_BY  |
| bob       | ta        | ID           | PLAINTEXT_AFTER_JOIN      |
| bob       | ta        | income       | PLAINTEXT_AFTER_AGGREGATE |
| bob       | tb        | ID           | PLAINTEXT                 |
| bob       | tb        | is_active    | PLAINTEXT                 |
| bob       | tb        | order_amount | PLAINTEXT                 |
+-----------+-----------+--------------+---------------------------+
```

### 查询
```shell
./brokerctl run "SELECT ta.credit_rank, COUNT(*) as cnt, AVG(ta.income) as avg_income, AVG(tb.order_amount) as avg_amount FROM ta INNER JOIN tb ON ta.ID = tb.ID WHERE ta.age >= 20 AND ta.age <= 30 AND tb.is_active=1 GROUP BY ta.credit_rank;"  --project-id "demo" --host http://localhost:8081 --timeout 3
[fetch]
2 rows in set: (1.221304389s)
+-------------+-----+-------------------+-------------------+
| credit_rank | cnt |    avg_income     |    avg_amount     |
+-------------+-----+-------------------+-------------------+
|           5 |   6 | 18069.77597427368 | 7743.392951965332 |
|           6 |   4 | 336016.8590965271 | 5499.404067993164 |
+-------------+-----+-------------------+-------------------+
```