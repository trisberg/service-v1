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

package binding

import (
	"context"
	"encoding/json"
	"fmt"
	servicev1alpha1 "github.com/trisberg/service/pkg/apis/service/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Binding Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBinding{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("binding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Binding
	err = c.Watch(&source.Kind{Type: &servicev1alpha1.Binding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Secret created by Binding - change this for objects you create
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &servicev1alpha1.Binding{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBinding{}

// ReconcileBinding reconciles a Binding object
type ReconcileBinding struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Binding object and makes changes based on the state read
// and what is in the Binding.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=service.projectriff.io,resources=bindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.projectriff.io,resources=bindings/status,verbs=get;update;patch
func (r *ReconcileBinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Binding instance
	instance := &servicev1alpha1.Binding{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	bindingSecret := instance.Name + "-binding"
	if instance.Spec.BindingSecret != "" {
		bindingSecret = instance.Spec.BindingSecret
	}
	credSecret := &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingSecret,
			Namespace: instance.Namespace,
		},
		StringData: map[string]string{
			"name": instance.Name,
		},
	}
	creds := getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, "credentials")
	var credentials string
	if len(creds) > 0 {
		credentials = string(getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, "credentials"))
		var credMap map[string]interface{}
		if err := json.Unmarshal(creds, &credMap); err != nil {
			log.Info("Oops, error parsing creds", "err", err.Error())
		}
		credSecret.StringData["uri"] = fmt.Sprintf("%v", credMap["uri"])
		credSecret.StringData["jdbcUrl"] = fmt.Sprintf("jdbc:%v", credMap["uri"])
		credSecret.StringData["username"] = fmt.Sprintf("%v", credMap["username"])
		credSecret.StringData["password"] = fmt.Sprintf("%v", credMap["password"])
		credSecret.StringData["host"] = fmt.Sprintf("%v", credMap["host"])
	} else {
		credentials = "{\"name\": \"" + instance.Name + "\""
		if instance.Spec.URI != "" || instance.Spec.URIKey != "" {
			uri := instance.Spec.URI
			if instance.Spec.URIKey != "" {
				value := getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, instance.Spec.URIKey)
				if len(value) > 0 {
					uri = string(value)
				}
			}
			credSecret.StringData["uri"] = uri
			credSecret.StringData["jdbcUrl"] = "jdbc:" + uri
			credentials += ", \"uri\": \"" + instance.Spec.URI + "\""
			credentials += ", \"jdbcUrl\": \"jdbc:" + instance.Spec.URI + "\""
		}
		if instance.Spec.Host != "" || instance.Spec.HostKey != "" {
			host := instance.Spec.Host
			if instance.Spec.HostKey != "" {
				value := getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, instance.Spec.HostKey)
				if len(value) > 0 {
					host = string(value)
				}
			}
			credSecret.StringData["host"] = host
			credentials += ", \"host\": \"" + host + "\""
		} else {
			credSecret.StringData["host"] = ""
		}
		if instance.Spec.Port != "" || instance.Spec.PortKey != "" {
			port := instance.Spec.Port
			if instance.Spec.PortKey != "" {
				value := getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, instance.Spec.PortKey)
				if len(value) > 0 {
					port = string(value)
				}
			}
			credSecret.StringData["port"] = port
			credentials += ", \"port\": " + port
		}
		if instance.Spec.Username != "" || instance.Spec.UsernameKey != "" {
			username := instance.Spec.Username
			if instance.Spec.UsernameKey != "" {
				value := getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, instance.Spec.UsernameKey)
				if len(value) > 0 {
					username = string(value)
				}
			}
			credSecret.StringData["username"] = username
			credentials += ", \"username\": \"" + username + "\""
		}
		if instance.Spec.SecretRef != "" {
			passwordKey := "password"
			if instance.Spec.PasswordKey != "" {
				passwordKey = instance.Spec.PasswordKey
			}
			password := getSecretValue(r.Client, context.TODO(), instance.Namespace, instance.Spec.SecretRef, passwordKey)
			credSecret.StringData["password"] = string(password)
			credentials += ", \"password\": \"" + string(password) + "\""
		}
		credentials += "}"
	}
	credSecret.StringData["credentials"] = credentials
	//credSecret.StringData["vcap.services"] =
	//	"{\"mysql\":[{\"name\": \"" + instance.Name + "\", \"credentials\": " + credentials + "}]}"
	if err := controllerutil.SetControllerReference(instance, credSecret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Secret already exists
	found := &corev1.Secret{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: credSecret.Name, Namespace: credSecret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Secret", "namespace", credSecret.Namespace, "name", credSecret.Name)
		err = r.Create(context.TODO(), credSecret)
		return reconcile.Result{}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	//if !reflect.DeepEqual(vcap.Data, found.Data) {
	//	found.Data = vcap.Data
	//	log.Info("Updating Secret", "namespace", vcap.Namespace, "name", vcap.Name)
	//	err = r.Update(context.TODO(), found)
	//	if err != nil {
	//		return reconcile.Result{}, err
	//	}
	//}
	return reconcile.Result{}, nil
}

func getSecretValue(c client.Client, ctx context.Context, namespace string, name string, key string) []byte {
	log.Info("Reading Secret", "namespace", namespace, "name", name)
	secret := &corev1.Secret{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, secret)
	if errors.IsNotFound(err) {
		log.Info("Reading Secret", "namespace", "default", "name", name)
		err = c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: "default"}, secret)
	}
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Oops, not found")
		} else {
			log.Info("Oops, ERROR", "err", err.Error())
		}
	}
	return secret.Data[key]
}
