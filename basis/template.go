package basis

var TemplateV1 = `
version: "1"
metadata:
  name: "press"
  gRPCServer:
	- localhost:8000
	- localhost:8001
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
      headers:
        - cookie: ""
      body: ""
#      jsonMap:
#        - key1: 1
#        - key2: 2
#    output:
#      path: "."
`
