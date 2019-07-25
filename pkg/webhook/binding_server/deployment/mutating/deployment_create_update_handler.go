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

const (
	BindingTypeAnnotation     = "service.projectriff.io/binding-type"
	BindingProfilesAnnotation = "service.projectriff.io/binding-profiles"
	BindingSecretsAnnotation  = "service.projectriff.io/binding-secrets"
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
	if ann[BindingTypeAnnotation] != "" {
		log.Info("Deployment:", "name", obj.Name, BindingTypeAnnotation, ann[BindingTypeAnnotation])
		boot = true
	}
	profile := ""
	if ann[BindingProfilesAnnotation] != "" {
		log.Info("Deployment:", "name", obj.Name, BindingProfilesAnnotation, ann[BindingProfilesAnnotation])
		profile = ann[BindingProfilesAnnotation]
	}
	if ann[BindingSecretsAnnotation] != "" {
		secretRef, secretPrefix := splitNameAndPrefix(ann[BindingSecretsAnnotation])
		if strings.Contains(secretPrefix, "-") {
			secretPrefix = secretPrefix[0:strings.Index(secretPrefix, "-")] + "_"
		}
		log.Info("Deployment:", "name", obj.Name,
			BindingSecretsAnnotation, secretRef,
			"prefix", secretPrefix)
		if boot {
			//ToDo
			setEnvVarSecret(secretPrefix+"uri", secretRef, "uri", obj)
			setEnvVarSecret("SPRING_DATASOURCE_URL", secretRef, "jdbcUrl", obj)
			setEnvVarSecret("SPRING_DATASOURCE_USERNAME", secretRef, "username", obj)
			setEnvVarSecret("SPRING_DATASOURCE_PASSWORD", secretRef, "password", obj)
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
				setEnvVar("SPRING_PROFILES_ACTIVE", profile, obj)
			}
		}
	}
	return nil
}

func setEnvVarSecret(name string, ref string, key string, obj *appsv1.Deployment) {
	envFound := false
	for i, env := range obj.Spec.Template.Spec.Containers[0].Env {
		log.Info("Env:", "name", env.Name)
		if env.Name == name {
			obj.Spec.Template.Spec.Containers[0].Env[i].ValueFrom = &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					Key: key,
					LocalObjectReference: v1.LocalObjectReference{
						Name: ref,
					},
				},
			}
			envFound = true
			log.Info("Env:", "FOUND", name, "SET-TO", ref+":"+key)
			break
		}
	}
	if !envFound {
		obj.Spec.Template.Spec.Containers[0].Env = append(obj.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name: name,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						Key: key,
						LocalObjectReference: v1.LocalObjectReference{
							Name: ref,
						},
					},
				},
			})
		log.Info("Env:", "ADDED", name, "SET-TO", ref+":"+key)
	}
}

func setEnvVar(name string, value string, obj *appsv1.Deployment) {
	envFound := false
	for i, env := range obj.Spec.Template.Spec.Containers[0].Env {
		log.Info("Env:", "name", env.Name)
		if env.Name == name {
			obj.Spec.Template.Spec.Containers[0].Env[i].Value = value
			envFound = true
			log.Info("Env:", "FOUND", name, "SET-TO", value)
			break
		}
	}
	if !envFound {
		obj.Spec.Template.Spec.Containers[0].Env = append(obj.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name:  name,
				Value: value,
			})
		log.Info("Env:", "ADDED", name, "SET-TO", value)
	}
}

func splitNameAndPrefix(s string) (string, string) {
	var name string
	var prefix string
	if strings.Contains(s, ":") {
		prefix = s[0:strings.Index(s, ":")] + "_"
		name = s[strings.Index(s, ":")+1:]
	} else {
		if strings.Contains(s, "-") {
			prefix = s[0:strings.Index(s, "-")] + "_"
			name = s
		}
	}
	return name, prefix
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
