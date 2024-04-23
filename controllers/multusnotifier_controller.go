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

package controllers

import (
	"context"
	"encoding/json"
	"time"

	netattachdef "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/jsm84/multus-notifier-operator/api/v1alpha1"
)

// MultusNotifierReconciler reconciles a MultusNotifier object
type MultusNotifierReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.f-i.de,resources=multusnotifiers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.f-i.de,resources=multusnotifiers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.f-i.de,resources=multusnotifiers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps.f-i.de,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MultusNotifier object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MultusNotifierReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling MultusNotifier")

	// Fetch the MultusNotifer instance
	instance := &appsv1alpha1.MultusNotifier{}
	err := r.client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Get list of pods using  WatchNamespace from CR instance
	pods, err := listPodsWithLabel(instance)
	if err != nil {
		// Error fetching pods - requeue the request.
		return reconcile.Result{}, err
	}

	// Our soon to be filled out list of Multus Pod IP Addresses
	ipList := []appsv1alpha1.MultusPodIPList{}
	if len(pods) > 0 {
		for idx, pod := range pods {

			netStatus := []netattachdef.NetworkStatus{}
			netStatusData := pod.Annotation["k8s.v1.cni.cncf.io/network-status"]
			netStatusBytes := []byte(netStatusData)

			err := json.Unmarshal(netStatusBytes, &netStatus)
			if err != nil {
				reqLogger.Error("Error unmarshaling network-status from pod", pod.Name)
				return
			}

			ipList[idx] = appsv1alpha1.MultusPodIPList{
				PodName:      pod.Name,
				PodNamespace: pod.Namespace,
				MultusIP:     netStatus[len(netStatus)-1].IPs[0],
			}
			reqLogger.Debug("Found Multus Pod IP", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name, "Pod.MultusIP", ipList[idx].MultusIP)
		}
	}

	if len(ipList) > 0 {
		// Update CR Status once deployed
		instance.Status.PodIPList = ipList
		reqLogger.Info("Updating Object Status", "CR.Kind", instance.Kind, "CR.Name", instance.Name)
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Multus Pod IP status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{RequeueAfter: 300 * time.Second}, nil
}

// listPodsWithLabel fetches a list of pods with a specific label, and optionally also within
// a specific set of namespaces.
func (r *MultusNotifierReconciler) listPodsWithLabel(cr *appsv1alpha1.MultusNotifier) ([]corev1.Pod, error) {
	reqLogger.Info("Getting List of Pods", "labels.Selector", cr.WatchLabel)
	podList := *CoreV1.PodList{}
	err := r.client.List(context.TODO(), labels.Selector{Label: cr.WatchLabel}, podList)
	if err != nil && errors.IsNotFound(err) {
		req.Logger.Error("No pods found with label", "labels.Selector", cr.WatchLabel)
	} else {
		req.Logger.Error("Error occurred while fetching pods")
	}
	return podList.Pods, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MultusNotifierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.MultusNotifier{}).
		Complete(r)
}
