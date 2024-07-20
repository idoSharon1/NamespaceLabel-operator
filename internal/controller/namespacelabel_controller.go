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
	"fmt"

	"github.com/idoSharon1/NamespaceLabel-operator/api/v1alpha1"
	corev1alpha1 "github.com/idoSharon1/NamespaceLabel-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NamespaceLabelReconciler reconciles a NamespaceLabel object
type NamespaceLabelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespacelabels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespacelabels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.core.namespacelabel.io,resources=namespacelabels/finalizers,verbs=update

func (r *NamespaceLabelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Enter reconcile function")

	// Get the current target namespacelabel
	var namespaceLabel v1alpha1.NamespaceLabel
	if err := r.Get(ctx, req.NamespacedName, &namespaceLabel); err != nil {
		logger.Error(err, "unable to load namespace-label")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get the namespace relevant to the namespacelabel
	var wantedNamespace v1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: namespaceLabel.ObjectMeta.Namespace}, &wantedNamespace); err != nil {
		logger.Error(err, "unable to load namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Define finilaizer name
	finalizerName := "core.namespacelabel.io/finalizer"

	// Check if the target namespacelabel is marked to delete
	if namespaceLabel.GetDeletionTimestamp() != nil {
		logger.Info(fmt.Sprintf("NamespaceLabel %s is marked for deletion", namespaceLabel.Name))

		if controllerutil.ContainsFinalizer(&namespaceLabel, finalizerName) {
			for key := range namespaceLabel.Spec.Labels {
				// Remove all relevant labels from the namespace
				copyCurrentLabels := wantedNamespace.DeepCopy().GetLabels()
				delete(copyCurrentLabels, key)
				wantedNamespace.SetLabels(copyCurrentLabels)
				logger.Info(fmt.Sprintf("Removing label %s", key))

				if err := r.Update(ctx, &wantedNamespace); err != nil {
					logger.Error(err, "Could not preform cleanup on namespace")
					return ctrl.Result{}, err
				}

				// Remove finializer
				controllerutil.RemoveFinalizer(&namespaceLabel, finalizerName)

				if err := r.Update(ctx, &namespaceLabel); err != nil {
					logger.Error(err, "Could not preform cleanup logic")
					return ctrl.Result{}, err
				}
			}

			// Stop reconcile item has been deleted
			return ctrl.Result{}, nil
		}
	} else {
		// Checking if finalizer already exists on namespacelabel and if not try to add it
		if !controllerutil.ContainsFinalizer(&namespaceLabel, finalizerName) {
			controllerutil.AddFinalizer(&namespaceLabel, finalizerName)
			if err := r.Update(ctx, &namespaceLabel); err != nil {
				logger.Error(err, fmt.Sprintf("Unable to add common finializer to namespacelabel %s", namespaceLabel.Name))
			}
		}
	}

	// Init key-value pair for the labels og the namespace in case the namespace doesn't already have labels
	if wantedNamespace.Labels == nil {
		wantedNamespace.Labels = make(map[string]string)
	}

	// Inject all the labels in case the defined labels doesn't conflict with already existing labels on the namespace
	for key, value := range namespaceLabel.Spec.Labels {
		// Checking if the current label already exists on the wanted namespace + the current namespacelabel object is yet to take affect
		if _, exists := wantedNamespace.Labels[key]; exists && !namespaceLabel.Status.Applied {
			return ctrl.Result{}, fmt.Errorf("label %s already registered to this namespace by another namespace label", key)
		} else {
			wantedNamespace.Labels[key] = value
		}
	}

	// Update the namespace object
	if err := r.Update(ctx, &wantedNamespace); err != nil {
		logger.Error(err, "unable to update namespace labels")
		return ctrl.Result{}, err
	}

	// Update the namespacelabel status
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
