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
  "cluster_name": "chia",
  "reserved_space": 100000000000
}
```
2. 异步拉取 **chia-storage-proxy** 的 **plot** 文件

## service 文件
```
cat << EOF > /etc/systemd/system/chia-storage-server.service
[Unit]
Description=Chia Plotter
After=lotus-mount-disk.service

[Service]
ExecStart=/usr/local/bin/chia-storage-server --config /etc/chia-storage-service.conf
Restart=always
RestartSec=10
MemoryAccounting=true
MemoryHigh=infinity
MemoryMax=infinity
Nice=-20
LimitNICE=-20
LimitNOFILE=1048576:1048576
LimitCORE=infinity
LimitNPROC=819200:1048576
IOWeight=9999
CPUWeight=1000
LimitCORE=1024
Delegate=yes
User=root

[Install]
WantedBy=multi-user.target
```

--------

## 部署

部署相关的文件

+ chia-storage-server
+ chia-storage-server.conf
+ chia-storage-server.service

|            文件             |      部署路径       |          说明           |
| :------------------------- | :----------------- | :--------------------- |
|     chia-storage-server     |   /usr/local/bin    | chia 存储服务可执行文件 |
|  chia-storage-server.conf   |        /etc         |        配置文件         |
| chia-storage-server.service | /etc/systemd/system |      service 文件       |

**注**
**chia-storage-server** 服务要设置自启动
```
  systemctl enable chia-storage-server
```
