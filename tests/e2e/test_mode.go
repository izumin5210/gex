package main

import "golang.org/x/tools/go/packages/packagestest"

type TestMode int

const (
	_ TestMode = iota
	TestModeMod
	TestModeDep
)

func (tm TestMode) String() string {
	switch tm {
	case TestModeMod:
		return "mod"
	case TestModeDep:
		return "dep"
	default:
		panic("unreachable")
	}
}

func (tm TestMode) Exporter() packagestest.Exporter {
	switch tm {
	case TestModeMod:
		return packagestest.Modules
	case TestModeDep:
		return packagestest.GOPATH
	default:
		panic("unreachable")
	}
}
