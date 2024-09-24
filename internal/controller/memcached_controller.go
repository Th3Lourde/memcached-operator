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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "github.com/example/memcached-operator/api/v1alpha1"
)

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Memcached object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// So one part seems to be receiving the .yaml file and storing it
	// Another part seems to be using what has been stored to update the state

	// log.Info("Reconciler Start")

	// Lookup the Memcached instance for this reconcile request
	memcached := cachev1alpha1.Memcached{}
	if err := r.Get(ctx, req.NamespacedName, &memcached); err != nil {
		log.Error(err, "unable to fetch Memcached")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// log.Info("Found object with name: %s | %s", memcached.Name, memcached.Namespace)

	// log.Info("Creating Deployment")

	// replicas := memcached.Spec.Size
	// image := "memcached:1.6.26-alpine3.19"
	// image := "redis/redis-stack-server:latest"

	// dep := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      memcached.Name,
	// 		Namespace: memcached.Namespace,
	// 	},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Replicas: &replicas,
	// 		Selector: &metav1.LabelSelector{
	// 			MatchLabels: map[string]string{"app.kubernetes.io/name": "project"},
	// 		},
	// 		Template: corev1.PodTemplateSpec{
	// 			ObjectMeta: metav1.ObjectMeta{
	// 				Labels: map[string]string{"app.kubernetes.io/name": "project"},
	// 			},
	// 			Spec: corev1.PodSpec{
	// 				SecurityContext: &corev1.PodSecurityContext{
	// 					RunAsNonRoot: &[]bool{true}[0],
	// 					SeccompProfile: &corev1.SeccompProfile{
	// 						Type: corev1.SeccompProfileTypeRuntimeDefault,
	// 					},
	// 				},
	// 				Containers: []corev1.Container{{
	// 					Image:           image,
	// 					Name:            "memcached",
	// 					ImagePullPolicy: corev1.PullIfNotPresent,
	// 					// Ensure restrictive context for the container
	// 					// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
	// 					SecurityContext: &corev1.SecurityContext{
	// 						RunAsNonRoot:             &[]bool{true}[0],
	// 						RunAsUser:                &[]int64{1001}[0],
	// 						AllowPrivilegeEscalation: &[]bool{false}[0],
	// 						Capabilities: &corev1.Capabilities{
	// 							Drop: []corev1.Capability{
	// 								"ALL",
	// 							},
	// 						},
	// 					},
	// 					Ports: []corev1.ContainerPort{{
	// 						ContainerPort: 11211,
	// 						Name:          "memcached",
	// 					}},
	// 					Command: []string{"memcached", "--memory-limit=64", "-o", "modern", "-v"},
	// 				}},
	// 			},
	// 		},
	// 	},
	// }

	// if err := r.Create(ctx, dep); err != nil {
	// 	log.Error(err, "Failed to create new Deployment",
	// 		"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 	return ctrl.Result{}, err
	// }

	// log.Info("Creation Finished")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Complete(r)
}

// ```
// apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
// kind: Deployment
// metadata:
//   name: redis-master
// spec:
//   selector:
//     matchLabels:
//       app: redis
//       role: master
//       tier: backend
//   replicas: 1
//   template:
//     metadata:
//       labels:
//         app: redis
//         role: master
//         tier: backend
//     spec:
//       containers:
//       - name: master
//         image: registry.k8s.io/redis:e2e  # or just image: redis
//         resources:
//           requests:
//             cpu: 100m
//             memory: 100Mi
//         ports:
//         - containerPort: 6379
// ```

// Command: []string{"memcached", "--memory-limit=64", "-o", "modern", "-v"},
