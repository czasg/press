package yamltemplate

import "fmt"

func GetTemplate(version string) (string, error) {
	switch version {
	case "1":
		return NewTemplateV1(), nil
	case "2":
		return NewTemplateV2(), nil
	default:
		return "", fmt.Errorf("unsupport version[%s]", version)
	}
}
