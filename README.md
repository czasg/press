# press

基于 go 实现的压力测试工具。     
安装方法
```shell script
go install github.com/czasg/press/cmd/press@latest
```

优点：     
1、提供了较友好的 yaml 配置模板，可读与可配置性更强。     
2、简单易上手。  

## 快速开始
1、初始化 yaml 模板
```shell script
press init
```

2、启动压测  
```shell script
press start -f press.yaml
```

除了 yaml，还支持快速启动一个简单压测：
```shell script
press test www.baidu.com
```

## yaml 解读
```yaml
---
version: "1"
metadata:
  name: "press"
steps:
  - name: "demo"
    logInterval: 1                         # 日志输出间隔
    threadGroup:                           # 线程组
      thread: 1                            # 并发线程数
      threadRampUp: 1                      # 线程唤醒时间
      duration: 5                          # 持续时间
    http:
      url: "http://www.baidu.com"          # 请求入口
      method: "GET"                        # 请求方法
      timeout: 10                          # 请求超时
#      headers:                            # 请求头 
#        content-type: "application/json"
#      body: |                             # 请求体
#        {
#          "hello":"press"
#        }
    assert:                                # 断言
      statusCode: 200                      # 响应状态码
#      headers:                            # 响应头
#        - content-type: "application/json"
#      body: ""                            # 响应体
#      jsonMap:                            # 仅支持第一层json解析
#        - errCode: 0
#    output:                               # 保存路径                       
#      path: "."
```
