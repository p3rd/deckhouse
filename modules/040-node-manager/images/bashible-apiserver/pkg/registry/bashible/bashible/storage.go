package bashible

import (
	"fmt"
	"io/ioutil"
	"path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"d8.io/bashible/pkg/apis/bashible"
	"d8.io/bashible/pkg/template"
)

const templateName = "bashible.sh.tpl"

// NewStorage returns storage object that will work against API services.
func NewStorage(rootDir string, bashibleContext template.Context) (*Storage, error) {
	templatePath := path.Join(rootDir, "bashible", templateName)

	tplContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read template: %v", err)
	}

	storage := &Storage{
		templateContent: tplContent,
		templateName:    templateName,
		bashibleContext: bashibleContext,
	}

	return storage, nil
}

type Storage struct {
	templateContent []byte
	templateName    string
	bashibleContext template.Context
}

// Render renders single script content by name
func (s Storage) Render(name string) (runtime.Object, error) {
	data, err := s.getContext(name)
	if err != nil {
		return nil, fmt.Errorf("cannot get context: %v", err)
	}
	r, err := template.RenderTemplate(templateName, s.templateContent, data)
	if err != nil {
		return nil, fmt.Errorf("cannot render template: %v", err)
	}

	obj := bashible.Bashible{}
	obj.ObjectMeta.Name = name
	obj.ObjectMeta.CreationTimestamp = metav1.NewTime(time.Now())
	obj.Data = map[string]string{}
	obj.Data[r.FileName] = r.Content.String()

	return &obj, nil
}

func (s Storage) getContext(name string) (map[string]interface{}, error) {
	contextKey, err := template.GetBashibleContextKey(name)
	if err != nil {
		return nil, fmt.Errorf("cannot get context key: %v", err)
	}

	context, err := s.bashibleContext.Get(contextKey)
	if err != nil {
		return nil, fmt.Errorf("cannot get context data: %v", err)
	}

	err = s.bashibleContext.EnrichContext(context)
	if err != nil {
		return nil, err
	}

	return context, nil
}

func (s Storage) New() runtime.Object {
	return &bashible.Bashible{}
}

func (s Storage) NewList() runtime.Object {
	return &bashible.BashibleList{}
}
