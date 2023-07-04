package k8sac

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type nsWatcher struct {
	ctx        context.Context
	logger     *logrus.Logger
	k8sClient  *kubernetes.Clientset
	nsInformer cache.SharedIndexInformer
	rules      []rule
}

func (nsw *nsWatcher) run(stopCh <-chan struct{}) {
	go nsw.nsInformer.Run(stopCh)
}

func (nsw *nsWatcher) createRoleBinding(obj any) {

	nsObj, b := obj.(*v1.Namespace)
	if b == false {
		nsw.logger.Errorf("incorrect role binding object")
		return
	}

	nsw.logger.WithFields(logrus.Fields{
		"namespace": nsObj.Name,
	}).Debugf("checking namespace")

	for _, rs := range nsw.rules {

		if rs.ns.MatchString(nsObj.Name) == false {
			continue
		}

		nsw.logger.WithFields(logrus.Fields{
			"namespace": nsObj.Name,
			"regexp":    rs.ns.String(),
		}).Debugf("namespace matched")

		for _, rb := range rs.roleBindings {

			roleBinding := &rbacv1.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("nxs-rbac-%s-%s", rb.role.kind, rb.role.name),
					Namespace: nsObj.Name,
				},
				Subjects: func() []rbacv1.Subject {

					subjs := []rbacv1.Subject{}

					// Users
					for _, u := range rb.subjects.users {
						subjs = append(
							subjs,
							rbacv1.Subject{
								Kind: "User",
								Name: u,
							},
						)
					}

					// Groups
					for _, g := range rb.subjects.groups {
						subjs = append(
							subjs,
							rbacv1.Subject{
								Kind: "Group",
								Name: g,
							},
						)
					}

					// Service accounts
					for _, sa := range rb.subjects.serviceAccounts {
						subjs = append(
							subjs,
							rbacv1.Subject{
								Kind: "ServiceAccount",
								Name: sa,
							},
						)
					}

					return subjs
				}(),

				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     rb.role.kind.String(),
					Name:     rb.role.name,
				},
			}

			_, err := nsw.k8sClient.RbacV1().RoleBindings(nsObj.Name).Create(nsw.ctx, roleBinding, metav1.CreateOptions{})

			// On success
			if err == nil {
				nsw.logger.WithFields(logrus.Fields{
					"namespace":    nsObj.Name,
					"role binding": roleBinding.Name,
				}).Debugf("successfully created role binding")
				continue
			}

			// If error not an "already exists"
			if errors.IsAlreadyExists(err) == false {
				nsw.logger.WithFields(logrus.Fields{
					"namespace":    nsObj.Name,
					"role binding": roleBinding.Name,
					"details":      err,
				}).Warnf("create role binding")
				continue
			}

			// Trying to update existing role binding
			if _, err := nsw.k8sClient.RbacV1().RoleBindings(nsObj.Name).Update(nsw.ctx, roleBinding, metav1.UpdateOptions{}); err != nil {
				nsw.logger.WithFields(logrus.Fields{
					"namespace":    nsObj.Name,
					"role binding": roleBinding.Name,
					"details":      err,
				}).Warnf("update role binding")
				continue
			}

			nsw.logger.WithFields(logrus.Fields{
				"namespace":    nsObj.Name,
				"role binding": roleBinding.Name,
			}).Debugf("successfully updated role binding")
		}
	}
}
