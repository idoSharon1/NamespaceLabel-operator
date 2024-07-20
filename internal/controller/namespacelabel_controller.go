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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	// v1 "k8s.io/client-go/applyconfigurations/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// "github.com/idoSharon1/NamespaceLabel-operator/api/v1alpha1"
	"github.com/idoSharon1/NamespaceLabel-operator/api/v1alpha1"
	corev1alpha1 "github.com/idoSharon1/NamespaceLabel-operator/api/v1alpha1"
)

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespaces,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespacelabels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespacelabels/finalizers,verbs=update

func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var namespaceLabel v1alpha1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		logger.Error(err, "unable to load namespace-label")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var wantedNamespace v1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: namespaceLabel.ObjectMeta.Namespace}, &wantedNamespace); err != nil {
		logger.Error(err, "unable to load namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if wantedNamespace.Labels == nil {
		wantedNamespace.Labels = make(map[string]string)
	}

	for key, value := range namespaceLabel.Spec.Labels {
		wantedNamespace.Labels[key] = value
	}

	if err := r.Update(ctx, &wantedNamespace); err != nil {
		logger.Error(err, "unable to update namespace labels")
		return ctrl.Result{}, err
	}

	namespaceLabel.Status.Applied = true
	if err := r.Status().Update(ctx, &namespaceLabel); err != nil {
		logger.Error(err, "unable to update status of namespacelabel")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully updated namespace-", wantedNamespace.Name, "label")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceLabelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.NamespaceLabel{}).
		Complete(r)
}
