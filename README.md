# Chia Storage Server
## TODO

1. 数据库设计
2. 生成 API 文档
3. shutdown
4. logrotate

### 数据库设计

```sql
create table chia_storage(
    id int(11) UNSIGNED AUTO_INCREMENT,
    url varchar(255) not null default '',
    state tinyint not null default 0,
    create_at int(11) not null default 0,
    update_at int(11) not null default 0,
    delete_at int(11) not null default 0,
    PRIMARY KEY ( `runoob_id` )
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

## 最新修改

1. 配置文件

```json
{
  "port": 18080,
  "db_path": "/etc/chia-storage-server.db",
  "cluster_name": "chia"
}
```
2. 异步拉取 **chia-storage-proxy** 的 **plot** 文件 