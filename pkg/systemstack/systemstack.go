package systemstack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"sigs.k8s.io/yaml"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/rio/stacks"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

type SystemStack struct {
	namespace string
	apply     apply.Apply
	name      string
	contents  []byte
}

func NewStack(apply apply.Apply, systemNamespace string, name string, system bool) *SystemStack {
	setID := "stack-" + name
	if system {
		setID = "system-" + setID
	}
	return &SystemStack{
		namespace: systemNamespace,
		apply:     apply.WithSetID(setID).WithDefaultNamespace(systemNamespace),
		name:      name,
	}
}

func (s *SystemStack) WithContent(contents []byte) {
	s.contents = contents
}

func (s *SystemStack) Questions() ([]v1.Question, error) {
	content, err := s.content()
	if err != nil {
		return nil, err
	}

	t := template.Template{
		Content: content,
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t.Questions()
}

func (s *SystemStack) content() ([]byte, error) {
	if len(s.contents) > 0 {
		return s.contents, nil
	}
	if os.Getenv("RIO_DEV") != "" {
		return ioutil.ReadFile("stacks/" + s.name + "-stack.yaml")
	}
	return stacks.Asset("stacks/" + s.name + "-stack.yaml")
}

func (s *SystemStack) Deploy(answers map[string]string) error {
	content, err := s.content()
	if err != nil {
		return err
	}

	rf, err := riofile.Parse(bytes.NewBuffer(content), template.AnswersFromMap(answers))
	if err != nil {
		return err
	}

	os := objectset.NewObjectSet()
	os.Add(rf.Objects()...)
	return s.apply.Apply(os)
}

func (s *SystemStack) Yaml(answers map[string]string) (string, error) {
	content, err := s.content()
	if err != nil {
		return "", err
	}

	rf, err := riofile.Parse(bytes.NewBuffer(content), template.AnswersFromMap(answers))
	if err != nil {
		return "", err
	}

	output := strings.Builder{}
	objs := rf.Objects()
	for _, obj := range objs {
		data, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		r, err := yaml.JSONToYAML(data)
		if err != nil {

			return "", err
		}
		output.Write(r)
		output.WriteString("---\n")
	}
	return output.String(), nil
}

func (s *SystemStack) Objects(answers map[string]string) ([]runtime.Object, error) {
	content, err := s.content()
	if err != nil {
		return nil, err
	}

	rf, err := riofile.Parse(bytes.NewBuffer(content), template.AnswersFromMap(answers))
	if err != nil {
		return nil, err
	}
	return rf.Objects(), nil
}

func (s *SystemStack) Remove() error {
	return s.apply.Apply(nil)
}
