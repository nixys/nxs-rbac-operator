package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/nixys/nxs-rbac-operator/ctx"
	"github.com/nixys/nxs-rbac-operator/routines/access"

	appctx "github.com/nixys/nxs-go-appctx/v2"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func main() {

	// Read command line arguments
	args := ctx.ArgsRead()

	// Init appctx
	appCtx, err := appctx.ContextInit(appctx.Settings{
		CustomContext:    &ctx.Ctx{},
		Args:             &args,
		CfgPath:          args.ConfigPath,
		TermSignals:      []os.Signal{syscall.SIGTERM, syscall.SIGINT},
		ReloadSignals:    []os.Signal{syscall.SIGHUP},
		LogrotateSignals: []os.Signal{syscall.SIGUSR1},
		LogFormatter:     &logrus.JSONFormatter{},
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	appCtx.Log().Info("program started")

	// main() body function
	defer appCtx.MainBodyGeneric()

	// Create main context
	c := context.Background()

	// Create access
	appCtx.RoutineCreate(c, access.Runtime)
}
