package testutils

import (
	"log"

	"gopkg.in/ory-am/dockertest.v3"
)

var (
	dockertestPool *dockertest.Pool
)

// DockertestPool returns a *dockertest.Pool. It panics if any error.
func DockertestPool() *dockertest.Pool {

	if dockertestPool == nil {
		var err error
		dockertestPool, err = dockertest.NewPool("")
		if err != nil {
			log.Panic(err)
		}
	}
	return dockertestPool

}
