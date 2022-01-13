DockerF: 一个基于云的Docker部署框架
====
##目标：
让应用开发人员只需要关注自己的业务，简单快速的将自己的应用部署到云上的容器中。

##使用方式

### 部署
dockerf deploy --security=$sid -f $confs

``` yaml
cluster-id: xxxxx-platform
cloud-driver: aliyun
pods:
   - feed:
        num: 20
        region: BJ
        containers:
           - front-end:
                num:1
                image: registry.xxxxx.com/feed:latest
                command: bash
           - user-latest:
                num:1
                image: register.xxxxx.com/feed:stable
                command: bash
service-discovery:
   engine: etcd
   nodes: 192.168.1.1,192.168.1.2,192.168.1.3,192.168.1.4,192.168.1.5
```

### 服务发现
dockerf install --type=discovery --nodes=5 --engine=etcd


### 启动停止
dockerf deploy --type=pod --start=true --name=$podname -f $conf

dockerf drop --type=pod --id=$pod-id

### 查看
dockerf list resource | pod | container



### 监控统计
dockerf stats resource | pod | container