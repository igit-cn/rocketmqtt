# 本项目修改自 github.com/fhmq/hmq，并进行了大量基于需求的优化，并不保证项目的通用性，请尽量使用原版本。
## Rocketmqtt
一个mqtt消息转发服务器，支持根据topic转发至kafka/rocketmq，并支持订阅rocketmq下发消息。
#### 特性
- 支持转发至kafka
- 支持转发至rocketmq
- 支持订阅rocketmq下发至mqtt客户端
- 基于rocketmq订阅的无状态服务器
- 高性能
- prometheus采集接口
- 简单的管理页面
### 性能评估
阿里云ecs 8核16g
#### 评估内容
- 下发内容 190B tps 1200 订阅自rocketmq
- 上报内容 368B tps 24000 至kafka
- 连接客户端数 60w
#### 评估结果
- cpu 30%
- load5 3
- 内存 10G
- 平均延迟 1ms

### 使用方式
#### 构建
```
go build .
```
#### 运行
```
./rocketmqtt
```
#### 配置文件
- conf/acl.conf
  - 订阅权限设置
- liumqtt.conf
  - 总配置文件，除了配置文件以外的所有配置