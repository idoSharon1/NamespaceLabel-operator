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

package controller_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	corev1alpha1 "github.com/idoSharon1/NamespaceLabel-operator/api/v1alpha1"
)

const (
	NamespaceLabelName      = "test-namespacelabel"
	NamespaceLabelNamespace = "default"
)

var _ = Describe("NamespaceLabel Controller", func() {

	Context("When creating namespacelabel object", func() {
		ctx := context.Background()
		It("Should create new regular namespacelabel crd", func() {
			namespacelabel := &corev1alpha1.NamespaceLabel{
				TypeMeta: metav1.TypeMeta{
					Kind:       "NamespaceLabel",
					APIVersion: "core.core.namespacelabel.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      NamespaceLabelName,
					Namespace: NamespaceLabelNamespace,
				},
				Spec: corev1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"a": "1",
						"b": "2",
						"c": "3",
					},
				},
			}
			Expect(k8sClient.Create(ctx, namespacelabel)).Should(Succeed())

			namespaceLookupKey := types.NamespacedName{Name: "default"}
			affectedNamespace := corev1.Namespace{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespaceLookupKey, &affectedNamespace)
				return err == nil
			}).Should(BeTrue())
			Expect(affectedNamespace.ObjectMeta.GetLabels()).Should(Equal(map[string]string{"a": "1", "b": "2", "c": "3"}))
		})

		It("Should not allow to create new namespacelabel crd with managment labels", func() {
			namespacelabel := &corev1alpha1.NamespaceLabel{
				TypeMeta: metav1.TypeMeta{
					Kind:       "NamespaceLabel",
					APIVersion: "core.core.namespacelabel.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      NamespaceLabelName,
					Namespace: NamespaceLabelNamespace,
				},
				Spec: corev1alpha1.NamespaceLabelSpec{
					Labels: map[string]string{
						"app.kubernetes.io/test": "1",
					},
				},
			}
			err := (k8sClient.Create(ctx, namespacelabel))
			Expect(err).To(HaveOccurred())
		})
	})
})
