package press

var TemplateV1 = `---
version: "1"
metadata:
  name: "press"
steps:
  - name: "压力测试"
    logInterval: 1
    threadGroup:
      thread: 2
      threadRampUp: 1
      duration: 10
    http:
      url: "http://www.baidu.com"
      method: "POST"
      timeout: 10
      headers:
        content-type: "application/json"
      body: |
        {
          "hello":"press"
        }
    assert:
      statusCode: 200
#      headers:
#        - auth: "authKey"
#      body: ""
#      jsonMap:  # 仅支持获取 Map 第一层
#        - errCode: 0
#        - status: true
#    output:
#      path: "."
`
