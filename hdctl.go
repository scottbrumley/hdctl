package main

import (
	"encoding/json"
	"fmt"
	"hdctl"
	"os"
	"strings"
	"time"
)

func main() {

	configStr := "config.json"

	if len(os.Args[1:]) > 0 {
		if os.Args[1] != "" {
			configStr = os.Args[1]
		}
	}
	controlCenter := hdctl.New(configStr)

	for {
		select {
		case msgStr := <-controlCenter.CtrlMessages:
			if json.Valid([]byte(msgStr)) {
				procID, hubid, action, command, retval := hdctl.ReadCtrl(msgStr)
				// If intiation strings match then return configuration
				if action == "initiate" {
					fmt.Println(procID + " Initiated")
					fmt.Println(time.Now().Format(time.RFC850) + " hacmd/ctrl just checked in.")
					results := controlCenter.FindJobs(procID, "config")
					for _, job := range results {
						go controlCenter.SendCommands(job)
					}
				}
				if action == "result" {
					commandStr := strings.Split(command, "/")
					switch len(commandStr) {
					//Priviledged Commands
					case 6:
						switch commandStr[5] {
						case "sensors":
							// Update the Sensor Details
							controlCenter.UpdateSensors(hubid, retval)
						case "config":
							// Update Configuration Database

						}
					//Base Commands
					case 5:
						switch commandStr[4] {
						case "config":
							// Update Configuration Database

						}
					}

				}
			} else {
				fmt.Println("Not a JSON String")
				continue
			}

		//Check Jobs in Mongo DB every 60 seconds
		case <-time.After(60 * time.Second):
			fmt.Println("Check Jobs")
			results := controlCenter.FindJobs(controlCenter.ProcID, "command")

			if results != nil {
				for _, job := range results {
					/*
						myJobID, myTrigger, myProcID, myAction, myActionType, myCommands := controlCenter.ReadJob(job)
						fmt.Println(myJobID)
						fmt.Println(myTrigger)
						fmt.Println(myProcID)
						fmt.Println(myAction)
						fmt.Println(myActionType)
						fmt.Println(myCommands)

					*/
					go controlCenter.SendCommands(job)
				}
			}
		}
	}
}
