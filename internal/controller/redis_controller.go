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
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "github.com/example/memcached-operator/api/v1alpha1"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	length  = 16
)

func GenerateSecurePassword() []byte {

	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))
	for i := range password {
		index, _ := rand.Int(rand.Reader, charsetLength)
		password[i] = charset[index.Int64()]
	}
	return password
}

// RedisReconciler reconciles a Redis object
type RedisReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func genRedisService(name, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-service", name),
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     "redis",
				Protocol: "TCP",
				Port:     6379,
			}},
			Selector: map[string]string{"app": "redis"},
		},
	}

}

func genRedisDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
	image := "redis:alpine"

	q := &resource.Quantity{}
	q.Set(2)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": "project", "app": "redis"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": "project", "app": "redis"},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{{
						Image:           image,
						Name:            "redis",
						ImagePullPolicy: corev1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             &[]bool{true}[0],
							RunAsUser:                &[]int64{1001}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 6379,
							Name:          "redis",
						}},
						Command: []string{"redis-server"},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "redis-volume",
							MountPath: "/data",
						}},
					}},
					// https://github.com/EngineerBetter/k8s-course-materials/blob/master/stateful-apps/redis-with-volume.yaml
					Volumes: []corev1.Volume{{
						Name: "redis-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "redis-volume-claim",
							},
						},
					}},
				},
			},
		},
	}
}

//+kubebuilder:rbac:groups=cache.example.com,resources=redis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=redis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=redis/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Redis object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	nextLoopTime := ctrl.Result{
		Requeue:      true,
		RequeueAfter: 5 * time.Second,
	}

	log.Info("Reconciler start - redis")

	// Lookup the redis instance for this reconcile request
	redis := cachev1alpha1.Redis{}
	if err := r.Get(ctx, req.NamespacedName, &redis); err != nil {
		log.Error(err, "unable to fetch Redis")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return nextLoopTime, client.IgnoreNotFound(err)
	}

	if redis.Spec.Password == nil {

		log.Info("starting to create secret")

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-secret", "redis"),
				Namespace: "memcached-operator-system",
			},
			Type: corev1.SecretTypeBasicAuth,
			Data: map[string][]byte{
				corev1.BasicAuthUsernameKey: []byte("admin"),
				corev1.BasicAuthPasswordKey: GenerateSecurePassword(),
			},
		}

		log.Info("applying to cluster")

		if err := r.Create(ctx, secret); err != nil {
			log.Error(err, "Failed to create secret")
		} else {
			optional := false

			log.Info("updating cluster spec")

			redis.Spec.Password = &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
				Optional: &optional,
			}
			log.Info("applying update of redis")

			err = r.Update(ctx, &redis)

			log.Info("finished updating")

			if err != nil {
				log.Error(err, "Failed to update secret")
			}
		}

	}

	// Just create the service for now - permissions issue, fix later
	// svc := genRedisService(redis.Name, redis.Namespace)

	// if err := r.Create(ctx, svc); err != nil {
	// 	log.Error(err, "Failed to create new Service",
	// 		"Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
	// 	return nextLoopTime, err
	// }

	log.Info("Checking if deployment exists")

	// TODO - gen deployment based upon the .yaml file in the api server
	dep := genRedisDeployment(redis.Name, redis.Namespace, int32(1))

	cluster_deployment := &appsv1.Deployment{}

	// Fetch the deployment
	err := r.Get(ctx, req.NamespacedName, cluster_deployment)

	if err == nil {
		log.Info("Deployment exists, exiting")
		return nextLoopTime, nil
	}

	if !apierrors.IsNotFound(err) {
		log.Error(err, "Error while fetching deployment")
	}

	log.Info("Deployment dne, creating")

	if err := r.Create(ctx, dep); err != nil {
		log.Error(err, "Failed to create new Deployment",
			"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		return nextLoopTime, err
	}

	return nextLoopTime, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Redis{}).
		Complete(r)
}
