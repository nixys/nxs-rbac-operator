package ctx

import (
	conf "github.com/nixys/nxs-go-conf"
)

type confOpts struct {
	LogFile        string     `conf:"logfile" conf_extraopts:"default=stdout"`
	LogLevel       string     `conf:"loglevel" conf_extraopts:"default=info"`
	PidFile        string     `conf:"pidfile"`
	KubeConfigConf string     `conf:"kubeConfig"`
	Rules          []ruleConf `conf:"rules" conf_extraopts:"required"`
}

type ruleConf struct {
	NS           string            `conf:"ns" conf_extraopts:"required"`
	RoleBindings []roleBindingConf `conf:"roleBindings" conf_extraopts:"required"`
}

type roleBindingConf struct {
	Role     roleConf     `conf:"role" conf_extraopts:"required"`
	Subjects subjectsConf `conf:"subjects" conf_extraopts:"required"`
}

type roleConf struct {
	Kind string `conf:"kind" conf_extraopts:"required"`
	Name string `conf:"name" conf_extraopts:"required"`
}

type subjectsConf struct {
	Users           []string `conf:"users"`
	Groups          []string `conf:"groups"`
	ServiceAccounts []string `conf:"serviceAccounts"`
}

func confRead(confPath string) (confOpts, error) {

	var c confOpts

	err := conf.Load(&c, conf.Settings{
		ConfPath:    confPath,
		ConfType:    conf.ConfigTypeYAML,
		UnknownDeny: true,
	})
	if err != nil {
		return c, err
	}

	return c, err
}
