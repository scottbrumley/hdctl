package main

import (
	"encoding/json"
	"fmt"
	"hdctl"
	"strings"
	"time"
)

func main() {
	controlCenter := hdctl.New("config.json")

	for {
		select {
		case msgStr := <-controlCenter.CtrlMessages:
			if json.Valid([]byte(msgStr)) {
				procID, hubid, action, command, retval := hdctl.ReadCtrl(msgStr)
				//procID, action, command, result := readCtrl(msgStr)
				// If intiation strings match then return configuration
				if action == "initiate" {
					fmt.Println(procID + " Initiated")
					fmt.Println(time.Now().Format(time.RFC850) + " hacmd/ctrl just checked in.")
					results := controlCenter.FindJobs(controlCenter.ProcID, "config")
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

		//Issue Sensor Commands
		case <-time.After(60 * time.Second):
			fmt.Println("Check Jobs")
			//results := hdctl.find_mongoDB(mongoClient, "homeSysDB", "jobs", brokerID)
			results := controlCenter.FindJobs(controlCenter.ProcID, "command")
			for _, job := range results {
				go controlCenter.SendCommands(job)
			}
		}
	}
}
