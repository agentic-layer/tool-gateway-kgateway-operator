/*
Copyright 2026.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentruntimev1alpha1 "github.com/agentic-layer/agent-runtime-operator/api/v1alpha1"
)

const ToolGatewayKgatewayControllerName = "runtime.agentic-layer.ai/tool-gateway-kgateway-controller"

// Version set at build time using ldflags
var Version = "dev"

// ToolGatewayReconciler reconciles a ToolGateway object
type ToolGatewayReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=runtime.agentic-layer.ai,resources=toolgateways,verbs=get;list;watch
// +kubebuilder:rbac:groups=runtime.agentic-layer.ai,resources=toolgateways/status,verbs=get
// +kubebuilder:rbac:groups=runtime.agentic-layer.ai,resources=toolgatewayclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=runtime.agentic-layer.ai,resources=toolservers,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ToolGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the ToolGateway instance
	var toolGateway agentruntimev1alpha1.ToolGateway
	if err := r.Get(ctx, req.NamespacedName, &toolGateway); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ToolGateway resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ToolGateway")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling ToolGateway",
		"name", toolGateway.Name,
		"namespace", toolGateway.Namespace,
		"toolGatewayClass", toolGateway.Spec.ToolGatewayClassName)

	// Check if this controller should process this ToolGateway
	if !r.shouldProcessToolGateway(ctx, &toolGateway) {
		log.Info("Controller is not responsible for this ToolGateway, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// Get all ToolServer resources
	if _, err := r.getToolServers(ctx); err != nil {
		log.Error(err, "Failed to get ToolServers")
		return ctrl.Result{}, err
	}

	// TODO: Implement resource creation
	if err := r.ensureConfigMap(ctx, &toolGateway); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureDeployment(ctx, &toolGateway); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.ensureService(ctx, &toolGateway); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// shouldProcessToolGateway determines if this controller is responsible for the given ToolGateway
func (r *ToolGatewayReconciler) shouldProcessToolGateway(ctx context.Context, toolGateway *agentruntimev1alpha1.ToolGateway) bool {
	log := logf.FromContext(ctx)

	// List all ToolGatewayClasses
	var toolGatewayClassList agentruntimev1alpha1.ToolGatewayClassList
	if err := r.List(ctx, &toolGatewayClassList); err != nil {
		log.Info("Cannot list ToolGatewayClasses, skipping to avoid errors")
		return false
	}

	// Filter to only classes managed by this controller
	var kgatewayClasses []agentruntimev1alpha1.ToolGatewayClass
	for _, tgc := range toolGatewayClassList.Items {
		if tgc.Spec.Controller == ToolGatewayKgatewayControllerName {
			kgatewayClasses = append(kgatewayClasses, tgc)
		}
	}

	// If className is explicitly set, check if it matches any of our managed classes
	toolGatewayClassName := toolGateway.Spec.ToolGatewayClassName
	if toolGatewayClassName != "" {
		for _, kgc := range kgatewayClasses {
			if kgc.Name == toolGatewayClassName {
				return true
			}
		}
	}

	// Look for ToolGatewayClass with default annotation among filtered classes
	for _, kgc := range kgatewayClasses {
		if kgc.Annotations["toolgatewayclass.kubernetes.io/is-default-class"] == "true" {
			return true
		}
	}

	return false
}

// getToolServers queries all ToolServer resources across all namespaces
func (r *ToolGatewayReconciler) getToolServers(ctx context.Context) ([]*agentruntimev1alpha1.ToolServer, error) {
	var toolServerList agentruntimev1alpha1.ToolServerList
	if err := r.List(ctx, &toolServerList); err != nil {
		return nil, err
	}

	toolServers := make([]*agentruntimev1alpha1.ToolServer, len(toolServerList.Items))
	for i := range toolServerList.Items {
		toolServers[i] = &toolServerList.Items[i]
	}

	return toolServers, nil
}

// ensureConfigMap creates or updates the ConfigMap for the ToolGateway
func (r *ToolGatewayReconciler) ensureConfigMap(_ context.Context, _ *agentruntimev1alpha1.ToolGateway) error {
	// TODO: Implement ConfigMap creation
	return nil
}

// ensureDeployment creates or updates the Deployment for the ToolGateway
func (r *ToolGatewayReconciler) ensureDeployment(_ context.Context, _ *agentruntimev1alpha1.ToolGateway) error {
	// TODO: Implement Deployment creation
	return nil
}

// ensureService creates or updates the Service for the ToolGateway
func (r *ToolGatewayReconciler) ensureService(_ context.Context, _ *agentruntimev1alpha1.ToolGateway) error {
	// TODO: Implement Service creation
	return nil
}

// findToolGatewaysForToolServer returns all ToolGateway resources that need to be reconciled
// when a ToolServer changes. Since all gateways discover ToolServers across all namespaces,
// any ToolServer change affects all ToolGateway resources.
func (r *ToolGatewayReconciler) findToolGatewaysForToolServer(ctx context.Context, _ client.Object) []ctrl.Request {
	var toolGatewayList agentruntimev1alpha1.ToolGatewayList
	if err := r.List(ctx, &toolGatewayList); err != nil {
		return []ctrl.Request{}
	}

	requests := make([]ctrl.Request, len(toolGatewayList.Items))
	for i, gw := range toolGatewayList.Items {
		requests[i] = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      gw.Name,
				Namespace: gw.Namespace,
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *ToolGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentruntimev1alpha1.ToolGateway{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&agentruntimev1alpha1.ToolServer{},
			handler.EnqueueRequestsFromMapFunc(r.findToolGatewaysForToolServer),
		).
		Named(ToolGatewayKgatewayControllerName).
		Complete(r)
}
