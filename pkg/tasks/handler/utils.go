package handler

import (
	"github.com/wulfixyt/anubis-monitors/pkg/modules/eventim"
	"strings"

	"github.com/wulfixyt/anubis-monitors/pkg/modules/fansale"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
)

func startModule(task *structs.Task) {
	if strings.Contains(strings.ToLower(task.Site), "fansale") {
		go fansale.Run(task)
	} else if strings.Contains(strings.ToLower(task.Site), "eventim") {
		go eventim.Run(task)
	} else if strings.Contains(strings.ToLower(task.Site), "ticketmaster") {
		//go ticketmaster.Run(task)
	}
}
