package controller

import (
	"github.com/silveryfu/ray-operator/pkg/controller/rayhead"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, rayhead.Add)
}
