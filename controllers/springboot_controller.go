/*
Copyright 2023.

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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	springv1 "yujiangjun/spring-boot-controller/api/v1"
)

// SpringBootReconciler reconciles a SpringBoot object
type SpringBootReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=spring.yujiangjun.github.com,resources=springboots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=spring.yujiangjun.github.com,resources=springboots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=spring.yujiangjun.github.com,resources=springboots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SpringBoot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SpringBootReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	c := r.Client

	boot := &springv1.SpringBoot{}

	err := r.Get(ctx, req.NamespacedName, boot)
	if err != nil {
		logger.Info("" + req.NamespacedName.Name + "is Delete")
		return reconcile.Result{}, nil
	}

	logger.Info("image================:" + boot.Spec.Image)
	name := boot.GetObjectMeta().GetName()

	labels := map[string]string{
		"app": name,
	}

	podObjMeta := metav1.ObjectMeta{
		Name:      name + "-pod",
		Namespace: req.Namespace,
		Labels:    labels,
	}

	serviceTypeMeta := metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
	serviceObjMeta := metav1.ObjectMeta{
		Name:      boot.Name + "-svc",
		Namespace: boot.Namespace,
		Labels:    labels,
	}
	service := &v1.Service{
		TypeMeta:   serviceTypeMeta,
		ObjectMeta: serviceObjMeta,
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{{
				Port: 80,
			}},
			Selector: labels,
		},
	}

	err = controllerutil.SetControllerReference(boot, service, r.Scheme)
	if err != nil {
		logger.Info("service set reference occur error", service.ObjectMeta.Name)
		return ctrl.Result{}, err
	}

	// 创建service
	_, err = controllerutil.CreateOrUpdate(ctx, c, service, func() error {
		return nil
	})
	if err != nil {
		logger.Info("service 发生了错误")
		return ctrl.Result{}, nil
	}

	typeMeta := metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	pod := &v1.Pod{
		TypeMeta:   typeMeta,
		ObjectMeta: podObjMeta,
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  boot.Name + "-c",
					Image: boot.Spec.Image,
					//Ports: []v1.ContainerPort{{
					//	ContainerPort: boot.Spec.Port,
					//}},
				},
			},
		},
	}
	err = controllerutil.SetControllerReference(boot, pod, r.Scheme)
	if err != nil {
		logger.Info("绑定owner发生了错误")
		return ctrl.Result{}, err
	}
	// 创建pod
	err = c.Create(ctx, pod)
	if err != nil {
		logger.Info("pod create err")
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpringBootReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&springv1.SpringBoot{}).
		Complete(r)
}
