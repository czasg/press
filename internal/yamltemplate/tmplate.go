package yamltemplate

import "fmt"

func GetTemplate(version string) (string, error) {
	switch version {
	case "1":
		return NewTemplateV1(), nil
	case "latest":
		return NewTemplateV1(), nil
	default:
		return "", fmt.Errorf("暂不支持版本[%s]", version)
	}
}
