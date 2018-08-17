package kubernetes

import (
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
)

func flattenDeploymentSpec(in v1beta1.DeploymentSpec) ([]interface{}, error) {
	att := make(map[string]interface{})
	att["min_ready_seconds"] = in.MinReadySeconds
	if in.Replicas != nil {
		att["replicas"] = *in.Replicas
	}
	att["selector"] = in.Selector.MatchLabels
	att["strategy"] = flattenDeploymentStrategy(in.Strategy)
	podSpec, err := flattenPodSpec(in.Template.Spec)
	if err != nil {
		return nil, err
	}
	att["template"] = []interface{}{
		map[string]interface{}{
			"metadata": flattenMetadata(in.Template.ObjectMeta),
			"spec":     podSpec,
		},
	}
	return []interface{}{att}, nil
}

func flattenDeploymentStrategy(in v1beta1.DeploymentStrategy) []interface{} {
	att := make(map[string]interface{})
	if in.Type != "" {
		att["type"] = in.Type
	}
	if in.RollingUpdate != nil {
		att["rolling_update"] = flattenDeploymentStrategyRollingUpdate(in.RollingUpdate)
	}
	return []interface{}{att}
}

func flattenDeploymentStrategyRollingUpdate(in *v1beta1.RollingUpdateDeployment) []interface{} {
	att := make(map[string]interface{})
	if in.MaxSurge != nil {
		att["max_surge"] = in.MaxSurge.String()
	}
	if in.MaxUnavailable != nil {
		att["max_unavailable"] = in.MaxUnavailable.String()
	}
	return []interface{}{att}
}

func expandDeploymentTemplatePodSpec(template []interface{}) (v1.PodSpec, error) {
	if len(template) == 0 || template[0] == nil {
		return v1.PodSpec{}, nil
	}
	in := template[0].(map[string]interface{})
	return expandPodSpec(in["spec"].([]interface{}))
}

func expandDeploymentSpec(deployment []interface{}) (v1beta1.DeploymentSpec, error) {
	obj := v1beta1.DeploymentSpec{}
	if len(deployment) == 0 || deployment[0] == nil {
		return obj, nil
	}
	in := deployment[0].(map[string]interface{})
	obj.MinReadySeconds = int32(in["min_ready_seconds"].(int))
	obj.Replicas = ptrToInt32(int32(in["replicas"].(int)))
	obj.Selector = &metav1.LabelSelector{
		MatchLabels: expandStringMap(in["selector"].(map[string]interface{})),
	}
	obj.Strategy = expandDeploymentStrategy(in["strategy"].([]interface{}))
	podSpec, err := expandDeploymentTemplatePodSpec(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: obj.Selector.MatchLabels,
		},
		Spec: podSpec,
	}

	return obj, nil
}

func expandDeploymentStrategy(p []interface{}) v1beta1.DeploymentStrategy {
	obj := v1beta1.DeploymentStrategy{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["type"]; ok {
		obj.Type = v1beta1.DeploymentStrategyType(v.(string))
	}
	if v, ok := in["rolling_update"]; ok {
		obj.RollingUpdate = expandRollingUpdateDeployment(v.([]interface{}))
	}
	return obj
}

func expandRollingUpdateDeployment(p []interface{}) *v1beta1.RollingUpdateDeployment {
	obj := v1beta1.RollingUpdateDeployment{}
	if len(p) == 0 || p[0] == nil {
		return &obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["max_surge"]; ok {
		obj.MaxSurge = expandRollingUpdateDeploymentIntOrString(v.(string))
	}
	if v, ok := in["max_unavailable"]; ok {
		obj.MaxUnavailable = expandRollingUpdateDeploymentIntOrString(v.(string))
	}
	return &obj
}

func expandRollingUpdateDeploymentIntOrString(v string) *intstr.IntOrString {
	i, err := strconv.Atoi(v)
	if err != nil {
		return &intstr.IntOrString{
			Type:   intstr.String,
			StrVal: v,
		}
	}
	return &intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(i),
	}
}
