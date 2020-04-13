package controller

import (
	"github.com/luis5tb/worker-osp-operator/pkg/controller/computeopenstack"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, computeopenstack.Add)
}
