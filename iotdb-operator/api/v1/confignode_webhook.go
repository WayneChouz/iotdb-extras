/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var confignodelog = logf.Log.WithName("confignode-resource")

var configNodeMgrClient client.Client

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *ConfigNode) SetupWebhookWithManager(mgr ctrl.Manager) error {
	configNodeMgrClient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-iotdb-apache-org-v1-confignode,mutating=true,failurePolicy=fail,sideEffects=None,groups=iotdb.apache.org,resources=confignodes,verbs=create;update,versions=v1,name=mconfignode.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ConfigNode{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ConfigNode) Default() {
	confignodelog.Info("default", "name", r.Name)
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
//+kubebuilder:webhook:path=/validate-iotdb-apache-org-v1-confignode,mutating=false,failurePolicy=fail,sideEffects=None,groups=iotdb.apache.org,resources=confignodes,verbs=create;update,versions=v1,name=vconfignode.kb.io,admissionReviewVersions=v1
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

var _ webhook.Validator = &ConfigNode{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ConfigNode) ValidateCreate() (admission.Warnings, error) {
	confignodelog.Info("validate create", "name", r.Name)
	return r.validateReplicas()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ConfigNode) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	confignodelog.Info("validate update", "name", r.Name)
	return r.validateReplicas()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ConfigNode) ValidateDelete() (admission.Warnings, error) {
	confignodelog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *ConfigNode) validateReplicas() (admission.Warnings, error) {
	nodeList := &corev1.NodeList{}
	if err := configNodeMgrClient.List(context.Background(), nodeList); err != nil {
		return nil, err
	}

	workerNodeCount := 0
	for _, node := range nodeList.Items {
		hasNoSchedule := false
		for _, taint := range node.Spec.Taints {
			if taint.Effect == corev1.TaintEffectNoSchedule || taint.Effect == corev1.TaintEffectNoExecute {
				hasNoSchedule = true
				break
			}
		}
		if !hasNoSchedule {
			workerNodeCount++
		}
	}

	confignodelog.Info("validate ConfigNode", "replicas count", r.Spec.Replicas, "NoSchedule node count", workerNodeCount)

	if r.Spec.Replicas > workerNodeCount {
		return nil, errors.New("ConfigNode replicas cannot exceed the number of available worker nodes in the cluster")
	}

	return nil, nil
}