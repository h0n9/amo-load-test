package main

import (
	"github.com/amolabs/amo-load-test/pkg/amoabci"
	"github.com/interchainio/tm-load-test/pkg/loadtest"
)

func main() {
	f := amoabci.NewAMOABCIClientFactory()
	err := loadtest.RegisterClientFactory("amoabci", f)
	if err != nil {
		panic(err)
	}

	loadtest.Run(&loadtest.CLIConfig{
		AppName:              "amo-load-test",
		AppShortDesc:         "Load testing application for amoabci",
		DefaultClientFactory: "amoabci",
	})
}
