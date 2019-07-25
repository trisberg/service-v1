/*

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

package mutating

import (
	"context"
	"github.com/knative/pkg/ptr"
	"k8s.io/api/core/v1"
	//extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
	"strings"
)

var log = logf.Log.WithName("service.binding.deployment.webhook")

func init() {
	webhookName := "mutating-create-update-deployment"
	if HandlerMap[webhookName] == nil {
		HandlerMap[webhookName] = []admission.Handler{}
	}
	HandlerMap[webhookName] = append(HandlerMap[webhookName], &DeploymentCreateUpdateHandler{})
}

// DeploymentCreateUpdateHandler handles Deployment
type DeploymentCreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	Client client.Client

	// Decoder decodes objects
	Decoder types.Decoder
}

func (h *DeploymentCreateUpdateHandler) mutatingDeploymentFn(ctx context.Context, obj *appsv1.Deployment) error {
	// TODO(user): implement your admission logic
	ann := obj.Annotations
	boot := false
	if ann["service.projectriff.io/binding-type"] != "" {
		log.Info("Deployment:", "name", obj.Name, "service.projectriff.io/binding-type", ann["service.projectriff.io/binding-type"])
		boot = true
	}
	profile := ""
	if ann["service.projectriff.io/binding-profiles"] != "" {
		log.Info("Deployment:", "name", obj.Name, "service.projectriff.io/binding-profiles", ann["service.projectriff.io/binding-profiles"])
		profile = ann["service.projectriff.io/binding-profiles"]
	}
	if ann["service.projectriff.io/binding-secrets"] != "" {
		secretRef := ann["service.projectriff.io/binding-secrets"]
		secretPrefix := secretRef + "_"
		if strings.Contains(secretPrefix, "-") {
			secretPrefix = secretPrefix[0:strings.Index(secretPrefix, "-")] + "_"
		}
		log.Info("Deployment:", "name", obj.Name,
			"service.projectriff.io/binding-secrets", ann["service.projectriff.io/binding-secrets"],
			"prefix", secretPrefix)
		if boot {
			//ToDo
		} else {
			envFromFound := false
			for _, envFrom := range obj.Spec.Template.Spec.Containers[0].EnvFrom {
				if envFrom.SecretRef != nil && envFrom.SecretRef.Name == secretRef {
					envFromFound = true
					log.Info("EnvFrom:", "FOUND", secretRef)
					break
				}
			}
			if !envFromFound {
				obj.Spec.Template.Spec.Containers[0].EnvFrom = append(obj.Spec.Template.Spec.Containers[0].EnvFrom,
					v1.EnvFromSource{
						SecretRef: &v1.SecretEnvSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: secretRef,
							},
							Optional: ptr.Bool(true),
						},
						Prefix: secretPrefix,
					})
			}
			if profile != "" {
				envVar := "SPRING_PROFILES_ACTIVE"
				envFound := false
				for i, env := range obj.Spec.Template.Spec.Containers[0].Env {
					log.Info("Env:", "name", env.Name)
					if env.Name == envVar {
						obj.Spec.Template.Spec.Containers[0].Env[i].Value = profile
						envFound = true
						log.Info("Env:", "FOUND", envVar)
						break
					}
				}
				if !envFound {
					obj.Spec.Template.Spec.Containers[0].Env = append(obj.Spec.Template.Spec.Containers[0].Env,
						v1.EnvVar{
							Name:  envVar,
							Value: profile,
						})
				}
			}
		}
	}
	return nil
}

var _ admission.Handler = &DeploymentCreateUpdateHandler{}

// Handle handles admission requests.
func (h *DeploymentCreateUpdateHandler) Handle(ctx context.Context, req types.Request) types.Response {
	obj := &appsv1.Deployment{}

	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := obj.DeepCopy()

	log.Info("Handle:", "deployment", obj.Name)

	err = h.mutatingDeploymentFn(ctx, copy)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return admission.PatchResponse(obj, copy)
}

var _ inject.Client = &DeploymentCreateUpdateHandler{}

// InjectClient injects the client into the DeploymentCreateUpdateHandler
func (h *DeploymentCreateUpdateHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}

var _ inject.Decoder = &DeploymentCreateUpdateHandler{}

// InjectDecoder injects the decoder into the DeploymentCreateUpdateHandler
func (h *DeploymentCreateUpdateHandler) InjectDecoder(d types.Decoder) error {
	h.Decoder = d
	return nil
}
