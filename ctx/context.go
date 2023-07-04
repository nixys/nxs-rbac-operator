package ctx

import (
	appctx "github.com/nixys/nxs-go-appctx/v2"
)

// Ctx defines application custom context
type Ctx struct {
	Conf confOpts
}

// Init initiates application custom context
func (c *Ctx) Init(opts appctx.CustomContextFuncOpts) (appctx.CfgData, error) {

	//a := opts.Args.(*Args)

	// Read config file
	conf, err := confRead(opts.Config)
	if err != nil {
		return appctx.CfgData{}, err
	}

	// Set application context
	c.Conf = conf

	return appctx.CfgData{
		LogFile:  c.Conf.LogFile,
		LogLevel: c.Conf.LogLevel,
		PidFile:  c.Conf.PidFile,
	}, nil
}

// Reload reloads application custom context
func (c *Ctx) Reload(opts appctx.CustomContextFuncOpts) (appctx.CfgData, error) {

	opts.Log.Debug("reloading context")

	return c.Init(opts)
}

// Free frees application custom context
func (c *Ctx) Free(opts appctx.CustomContextFuncOpts) int {

	opts.Log.Debug("freeing context")

	return 0
}
