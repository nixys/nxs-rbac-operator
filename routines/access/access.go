package access

import (
	"context"

	"github.com/nixys/nxs-rbac-operator/ctx"
	"github.com/nixys/nxs-rbac-operator/modules/k8sac"
	"github.com/sirupsen/logrus"

	appctx "github.com/nixys/nxs-go-appctx/v2"
)

// Runtime executes the routine
func Runtime(cr context.Context, appCtx *appctx.AppContext, crc chan interface{}) {

	cc := appCtx.CustomCtx().(*ctx.Ctx)

	stopCh := make(chan struct{})

	cctx, cf := context.WithCancel(cr)

	k, err := k8sac.Init(
		k8sac.Settings{
			Logger:     appCtx.Log(),
			KubeConfig: cc.Conf.KubeConfigConf,
			Rules: func() []k8sac.RuleSettings {
				rules := []k8sac.RuleSettings{}
				for _, r := range cc.Conf.Rules {
					rules = append(
						rules,
						k8sac.RuleSettings{
							NS: r.NS,
							RoleBindings: func() []k8sac.RoleBindingSettings {
								bs := []k8sac.RoleBindingSettings{}
								for _, b := range r.RoleBindings {
									bs = append(
										bs,
										k8sac.RoleBindingSettings{
											Role: k8sac.RoleSettings{
												Kind: b.Role.Kind,
												Name: b.Role.Name,
											},
											Subjects: k8sac.SubjectsSettings{
												Users:           b.Subjects.Users,
												Groups:          b.Subjects.Groups,
												ServiceAccounts: b.Subjects.ServiceAccounts,
											},
										},
									)
								}
								return bs
							}(),
						},
					)
				}
				return rules
			}(),
		},
	)
	if err != nil {

		appCtx.Log().WithFields(logrus.Fields{
			"details": err,
		}).Errorf("k8s init")

		cf()

		appCtx.RoutineDoneSend(appctx.ExitStatusFailure)
		return
	}

	k.Exec(cctx, stopCh)

	for {
		select {
		case <-cr.Done():
			// Program termination.

			// Call cancel function
			cf()

			// Close stop channel
			close(stopCh)

			appCtx.Log().Info("access routine done")
			return
		case <-crc:
			// Updated context application data.
			// Set the new one in current goroutine.
			appCtx.Log().Info("access routine reload")
		}
	}
}
