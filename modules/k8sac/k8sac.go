package k8sac

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sAC struct {
	k8sClient *kubernetes.Clientset
	rules     []rule
	logger    *logrus.Logger
}

type Settings struct {
	KubeConfig string
	Rules      []RuleSettings
	Logger     *logrus.Logger
}

type RuleSettings struct {
	NS           string
	RoleBindings []RoleBindingSettings
}

type RoleBindingSettings struct {
	Role     RoleSettings
	Subjects SubjectsSettings
}

type RoleSettings struct {
	Kind string
	Name string
}

type SubjectsSettings struct {
	Users           []string
	Groups          []string
	ServiceAccounts []string
}

type rule struct {
	ns           *regexp.Regexp
	roleBindings []roleBinding
}

type roleBinding struct {
	role     role
	subjects subjects
}

type role struct {
	kind roleKind
	name string
}

type subjects struct {
	users           []string
	groups          []string
	serviceAccounts []string
}

type roleKind string

const (
	roleKindClusterRole roleKind = "ClusterRole"
	roleKindRole        roleKind = "Role"
)

func (k roleKind) String() string {
	return string(k)
}

func Init(s Settings) (K8sAC, error) {

	var (
		k      K8sAC
		config *rest.Config
		err    error
	)

	if len(s.KubeConfig) == 0 {
		config, err = rest.InClusterConfig()
		if err != nil {
			return K8sAC{}, fmt.Errorf("k8s access controller: %w", err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", s.KubeConfig)
		if err != nil {
			return K8sAC{}, fmt.Errorf("k8s access controller: %w", err)
		}
	}

	k.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return K8sAC{}, fmt.Errorf("k8s access controller: %w", err)
	}

	k.logger = s.Logger

	for _, rs := range s.Rules {

		ns, err := regexp.Compile(rs.NS)
		if err != nil {
			return K8sAC{}, fmt.Errorf("k8s access controller: %w", err)
		}

		rbs := []roleBinding{}
		for _, rb := range rs.RoleBindings {

			var rk roleKind
			switch rb.Role.Kind {
			case roleKindClusterRole.String():
				rk = roleKindClusterRole
			case roleKindRole.String():
				rk = roleKindRole
			default:
				return K8sAC{}, fmt.Errorf("unknown role kind '%s'", rb.Role.Kind)
			}

			rbs = append(
				rbs,
				roleBinding{
					role: role{
						kind: rk,
						name: rb.Role.Name,
					},
					subjects: subjects{
						users:           rb.Subjects.Users,
						groups:          rb.Subjects.Groups,
						serviceAccounts: rb.Subjects.ServiceAccounts,
					},
				},
			)
		}

		k.rules = append(
			k.rules,
			rule{
				ns:           ns,
				roleBindings: rbs,
			},
		)
	}

	return k, nil
}

func (k K8sAC) Exec(ctx context.Context, stopCh <-chan struct{}) {

	nsw := &nsWatcher{
		ctx:       ctx,
		logger:    k.logger,
		k8sClient: k.k8sClient,
		nsInformer: cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return k.k8sClient.CoreV1().Namespaces().List(ctx, options)
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return k.k8sClient.CoreV1().Namespaces().Watch(ctx, options)
				},
			},
			&v1.Namespace{},
			1*time.Minute,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		),
		rules: k.rules,
	}

	nsw.nsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: nsw.createRoleBinding,
	})

	nsw.run(stopCh)
}
