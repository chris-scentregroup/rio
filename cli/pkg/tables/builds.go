package tables

import (
	"fmt"

	"github.com/knative/pkg/apis"

	"github.com/rancher/rio/cli/pkg/table"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func NewBuild(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"Name", "{{stackScopedName .Obj.Namespace .Obj.Name ``}}"},
		{"SERVICE", "{{.Obj | findService}}"},
		{"REVISION", "{{.Obj | findRevision}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SUCCEED", "{{ .Obj | succeed }}"},
		{"REASON", "{{ .Obj | reason }}"},
	}, cfg)

	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetSetNamespace()))
	writer.AddFormatFunc("findService", findService)
	writer.AddFormatFunc("findRevision", findRevision)
	writer.AddFormatFunc("succeed", findSucceed)
	writer.AddFormatFunc("reason", findReason)
	return &tableWriter{
		writer: writer,
	}
}

func findSucceed(data interface{}) (string, error) {
	b, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	cond := b.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		return "", nil
	}
	return string(cond.Status), nil
}

func findReason(data interface{}) (string, error) {
	b, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	cond := b.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		return "", nil
	}
	return string(cond.Reason), nil
}

func findService(data interface{}) (string, error) {
	m, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	name := m.Labels["service-name"]
	namespace := m.Labels["service-namespace"]
	return fmt.Sprintf("%s/%s", namespace, name), nil
}

func findRevision(data interface{}) (string, error) {
	m, ok := data.(*tektonv1alpha1.TaskRun)
	if !ok {
		return "", nil
	}
	for _, param := range m.Spec.Inputs.Resources[0].ResourceSpec.Params {
		if param.Name == "revision" {
			return param.Value, nil
		}
	}
	return "", nil
}
