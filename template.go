package press

import (
	"fmt"
	"log"
	"os"
)

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

func CreateYaml(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("初始化 yaml 文件异常：%v\n", err)
	}
	defer f.Close()
	_, err = f.WriteString(TemplateV1)
	if err != nil {
		return fmt.Errorf("初始化 yaml 内容异常：%v\n", err)
	}
	log.Printf("初始化 yaml 成功\n")
	return nil
}
