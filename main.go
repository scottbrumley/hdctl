package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func main() {
	ctrlch1 := make(chan string)
	//cmdch1 := make(chan string)
	user, pass, broker, brokerID, mongoDB := ReadConfig("config.json")
	client := connect_mqtt(broker, user, pass)
	mongoClient := connect_mongoDB(mongoDB)

	//Checking the connection
	mongoClient.Ping(context.TODO(), nil)
	fmt.Println("Database connected")

	// Subscribe to a topic
	go SubscribeTo("hacmd/ctrl", client, ctrlch1)

	for {
		select {
		case msgStr := <-ctrlch1:
			if json.Valid([]byte(msgStr)) {
				procID, hubid, action, command, retval := readCtrl(msgStr)
				//procID, action, command, result := readCtrl(msgStr)
				// If intiation strings match then return configuration
				if action == "initiate" {
					fmt.Println(procID + " Initiated")
					fmt.Println(time.Now().Format(time.RFC850) + " hacmd/ctrl just checked in.")
					pingCMD := "{\"procid\":\"" + procID + "\",\"commands\": [{\"url\": \"https://192.168.192.55/api/config\",\"hubid\": \"001788FFFE277094\"},{\"url\": \"https://192.168.192.56/api/config\",\"hubid\": \"ECB5FAFFFE10C52F\"},{\"url\": \"https://192.168.192.58/api/config\",\"hubid\": \"ECB5FAFFFE0DA7C7\"}]}"
					go PublishTo("hacmd/cmd", client, pingCMD)
				}
				if action == "result" {
					commandStr := strings.Split(command, "/")
					switch len(commandStr) {
					//Priviledged Commands
					case 6:
						switch commandStr[5] {
						case "sensors":
							// Update the Sensor Details
							updateOne_mongoDB(mongoClient, "homeSysDB", "sensors", hubid, retval)
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
			results := find_mongoDB(mongoClient, "homeSysDB", "jobs", brokerID)
			for _, result := range results {
				fmt.Println(result)
			}

			//pingCMD := "{\"procid\":\"3edb99fa-5be6-7cc3-025b-b834efe62fd9\",\"commands\": [{\"url\": \"https://192.168.192.55/api/vw6OE0D7kDffGeMpA5JXhHuQZaXMtX8Jh8zcEyyb/sensors\",\"hubid\": \"001788FFFE277094\"},{\"url\": \"https://192.168.192.56/api/OdZrhUY-514oY5iuhkg4lFgm0iL6qRlCIAAqvA3y/sensors\",\"hubid\": \"ECB5FAFFFE10C52F\"},{\"url\": \"https://192.168.192.58/api/Yj3knZa5VWYGLYo6n7TAOVWrRW-3VK9Un1UALd9t/sensors\",\"hubid\": \"ECB5FAFFFE0DA7C7\"}]}"
			//go PublishTo("hacmd/cmd", client, pingCMD)
		}
	}
}
