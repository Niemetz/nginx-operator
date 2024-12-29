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

package controller

import (
	"context"
	"errors"

	appv1 "github.com/Niemetz/nginx-operator/api/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NginxDeployReconciler reconciles a NginxDeploy object
type NginxDeployReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.example.com,resources=nginxdeploys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.example.com,resources=nginxdeploys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.example.com,resources=nginxdeploys/finalizers,verbs=update

// Reconcile is part of the main Kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// This function compares the desired state specified by the NginxDeploy object
// against the actual cluster state, and performs operations to align the two.
func (r *NginxDeployReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the NginxDeploy resource
	var nginxDeploy appv1.NginxDeploy
	if err := r.Get(ctx, req.NamespacedName, &nginxDeploy); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("NginxDeploy resource not found. Ignoring since it must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get NginxDeploy")
		return ctrl.Result{}, err
	}

	// Validate Spec
	if nginxDeploy.Spec.Foo == "" {
		err := errors.New("foo field in spec cannot be empty")
		logger.Error(err, "Invalid NginxDeploy spec")
		return ctrl.Result{}, err
	}

	// Define a new Deployment
	nginxDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: req.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "nginx"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: "nginx:latest",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
					}},
				},
			},
		},
	}

	// Check if the Deployment already exists
	found := &v1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: nginxDeployment.Name, Namespace: nginxDeployment.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Info("Creating a new Deployment", "Deployment.Namespace", nginxDeployment.Namespace, "Deployment.Name", nginxDeployment.Name)
		err = r.Create(ctx, nginxDeployment)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", nginxDeployment.Namespace, "Deployment.Name", nginxDeployment.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Deployment already exists - log and finish
	logger.Info("Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

	// Log reconciliation success
	logger.Info("Reconciliation successful", "name", req.Name, "namespace", req.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxDeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.NginxDeploy{}).
		Named("nginxdeploy").
		Complete(r)
}

// int32Ptr is a helper function to get a pointer to an int32 value
func int32Ptr(i int32) *int32 {
	return &i
}
