package acceptance

import "github.com/dshills/vellum/app"

func currentFixtureRef(a *app.Application, name string) app.WindowRef {
	ref, err := a.WindowRef(name)
	if err != nil {
		panic(err)
	}
	return ref
}
