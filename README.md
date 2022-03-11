# go-tagdog
[![Go](https://github.com/kcat-io/go-tagdog/actions/workflows/go-tagdog.yml/badge.svg?branch=master)](https://github.com/kcat-io/go-tagdog/actions/workflows/go-tagdog.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kcat-io/go-tagdog)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://copyfree.org)

一个基于k8s clinet-go实现的容器镜像版本检查工具

业务中有使用k8s部署和管理微服务，微服务的数量多了之后偶尔会出现因人工操作疏漏而导致版本错乱的现象。比如某个微服务临时要切到灰度版本，结果灰度执行完成后忘记切换回正式版本；再比如一些cronjob所使用的镜像版本可能落后于deployment中运行的版本（cronjob配置忘记更新镜像版本）。

这些状况都可能给业务带来损害，为此我编写了一个检查规则可配置的工具来兜底提醒，工具通过k8s的client-go客户端获取微服务的信息跟配置中的信息做比对，对于异常信息做出预警。

可以把go-tagdog编译成镜像配置到k8s集群的cronjob中定期检查，提供预警。

### 配置内容
```yaml
# 配置kube config 路径
kube_config: "/Users/kcat/.kube/config"
# 配置目标集群中的namespace
namespace: "default"
#配置调用k8s接口的超时时间，单位：秒
timeout: 600

# 配置需要启用的规则插件，alarm建议配置在最前，ignore建议配置在最后
# 每个插件都可以定制（前序逻辑）和（后续逻辑），系统将按照插件顺序执行（前序逻辑）再逆序执行（后续逻辑）
# 其中alarm在（后续逻辑）中实现了告警输出
plugins:
  - alarm
  - tagcheck
  - ignore

# 以下是各插件的自主配置，由各插件自主解析和使用
alarm:
  enable: true

ignore:
  enable: true
  list: # 类型::pod名称.container名称:屏蔽告警截止时间（在此时间之前 将忽略该container的告警）
    deployment::testpod.testcontainer0: "2022-03-10 00:00:00"
    deployment::testpod.testcontainer3: "2022-03-10 00:00:00"
    deployment::testpod.testcontainer4: ""
    deployment::testpod.testcontainer5: "2022-03-10 00:00:00"
    deployment::testpod.testcontainer7: "2022-03-07 00:00:00"
    deployment::testpod.testcontainer9: "2022-03-10 00:00:00"
    crontab::testpod.testcontainer: "" # 也可以不设置时间 将无限期忽略

tagcheck:
  enable: true
  list: # 检查使用以下镜像的容器tag是否正确
    ccr.ccs.tencentyun.com/kcat/nginx: "latest"
    ccr.ccs.tencentyun.com/kcat/typecho: "latest"
  cronjob: true #检查cronjob所采用的镜像tag是否与deployment一致
```

### 目录结构
```
.
├── config.yaml # 主配置文件，可在配置中启用和关闭插件
├── go.mod
├── go.sum
├── main.go # 入口文件
└── plugins # 插件目录
    ├── alarm.go # 告警插件，可按需修改为邮件、企业微信等通知方式
    ├── entity.go # 实体定义
    ├── ignore.go # 忽略插件，可配置免告警策略
    └── tagcheck.go # Tag校验插件，校验目标镜像的Tag是否符合预期


```

### 编译执行
```
go build -v
./go-tagdog
```
