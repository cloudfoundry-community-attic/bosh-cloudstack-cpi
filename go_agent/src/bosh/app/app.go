package app

import (
	boshagent "bosh/agent"
	boshaction "bosh/agent/action"
	boshtask "bosh/agent/task"
	boshblobstore "bosh/blobstore"
	boshboot "bosh/bootstrap"
	boshinf "bosh/infrastructure"
	boshmbus "bosh/mbus"
	boshplatform "bosh/platform"
	"flag"
	"io/ioutil"
)

type app struct {
}

type options struct {
	InfrastructureName string
	PlatformName       string
}

func New() (app app) {
	return
}

func (app app) Run(args []string) (err error) {
	opts, err := parseOptions(args)
	if err != nil {
		return
	}

	infProvider := boshinf.NewProvider()
	infrastructure, err := infProvider.Get(opts.InfrastructureName)
	if err != nil {
		return
	}

	platformProvider := boshplatform.NewProvider()
	platform, err := platformProvider.Get(opts.PlatformName)
	if err != nil {
		return
	}

	boot := boshboot.New(infrastructure, platform)
	settings, err := boot.Run()
	if err != nil {
		return
	}

	mbusHandlerProvider := boshmbus.NewHandlerProvider(settings)
	mbusHandler, err := mbusHandlerProvider.Get()
	if err != nil {
		return
	}

	blobstoreProvider := boshblobstore.NewProvider()
	blobstore, err := blobstoreProvider.Get(settings.Blobstore)
	if err != nil {
		return
	}

	taskService := boshtask.NewAsyncTaskService()
	actionFactory := boshaction.NewFactory(settings, platform, blobstore, taskService)

	agent := boshagent.New(settings, mbusHandler, platform, taskService, actionFactory)
	err = agent.Run()
	return
}

func parseOptions(args []string) (opts options, err error) {
	flagSet := flag.NewFlagSet("bosh-agent-args", flag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard)
	flagSet.StringVar(&opts.InfrastructureName, "I", "", "Set Infrastructure")
	flagSet.StringVar(&opts.PlatformName, "P", "", "Set Platform")

	err = flagSet.Parse(args[1:])
	return
}
