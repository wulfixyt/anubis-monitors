package socket

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/wulfixyt/anubis-monitors/pkg/config"
	"github.com/wulfixyt/anubis-monitors/pkg/log"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/handler"
	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
)

func Connect() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			time.Sleep(20 * time.Second)
			return
		}
	}()

	// Initialize the logger
	log.Init()

	// Load the settings from settings.json
	err := loadSettings()
	if err != nil {
		return
	}

	// DE

	task := &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.de",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.de"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// AT

	task = &structs.Task{
		Site:      "www.fansale.at",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.at"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.at",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.at"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.at",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.at"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// CH

	task = &structs.Task{
		Site:      "www.fansale.ch",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.ch"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.ch",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.ch"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.ch",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.ch"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// IT

	task = &structs.Task{
		Site:      "www.fansale.it",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.it"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.it",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.it"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// PL

	task = &structs.Task{
		Site:      "www.fansale.pl",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.pl"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.pl",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.pl"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// ES

	task = &structs.Task{
		Site:      "www.fansale.es",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.es"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.es",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.es"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// SE

	task = &structs.Task{
		Site:      "www.fansale.se",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.se"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.se",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.se"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// NO

	task = &structs.Task{
		Site:      "www.fansale.no",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.no"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.no",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.no"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	// FI

	task = &structs.Task{
		Site:      "www.fansale.fi",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.fi"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	task = &structs.Task{
		Site:      "www.fansale.fi",
		Delay:     100,
		ProxyFile: formatWebsite("www.fansale.fi"),
	}

	handler.Create(task)
	handler.Start(task.Id)

	for {
		time.Sleep(100 * time.Second)
	}

	for {
		// Check for what monitors should be active
		err = getMonitors()
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(5 * time.Second)
	}
}

func getMonitors() error {
	resp, err := http.Get("https://api.anubisio.com/api/monitor?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzM1Mjk4NjN9.2flxemBNLZ3BqBDsk0vX3hrBPHhaIJWGj4vfj2BGRWw")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response monitorResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	var productList []string

	// Go through all active websites
	for _, website := range response.Message {
		// Go through all products that should being monitored
		for _, product := range website.Products {
			// Create a hash and use it as the groupId
			hash := md5.Sum([]byte(strings.ToLower(website.Website + "_" + product)))
			groupId := hex.EncodeToString(hash[:])

			// Check if tasks with that groupId do exist
			if len(handler.Filter(groupId)) == 0 {
				delay, ok := config.Config.Config.Delays[formatWebsite(website.Website)]
				if !ok {
					// Default delay
					delay = 2500
				}

				task := &structs.Task{
					Site:      website.Website,
					Input:     product,
					Delay:     delay,
					GroupId:   groupId,
					ProxyFile: formatWebsite(website.Website),
				}

				handler.Create(task)
				handler.Start(task.Id)
			}

			productList = append(productList, groupId)
		}
	}

	tasks := handler.GetTasks()
	// Go through all active tasks
	for _, task := range tasks {
		// Check if each task group should be active
		if !contains(productList, task.GroupId) {
			// Stop and delete that task
			handler.Delete(task.Id)
		}
	}

	return nil
}

func contains(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}

	return false
}
