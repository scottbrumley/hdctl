package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func main() {
	ctrlch1 := make(chan string)
	//cmdch1 := make(chan string)
	//configuredHubs := make(map[string]bool)
	user, pass, broker, mongoDB := ReadConfig("config.json")
	client := connect_mqtt(broker, user, pass)
	mongoClient := connect_mongoDB(mongoDB)

	//Checking the connection
	mongoClient.Ping(context.TODO(), nil)
	fmt.Println("Database connected")

	// Subscribe to a topic
	go SubscribeTo("hacmd/ctrl", client, ctrlch1)

	for {
		//fmt.Println("hacmd: " + <-procID + " msg: " + <-ctrlch1)
		select {
		case msgStr := <-ctrlch1:
			procID, action := readCtrl(msgStr)

			//fmt.Println(msgStr + " message. ")
			// If intiation strings match then return configuration
			if action == "initiate" {
				fmt.Println(procID + " Initiated")
				//configuredHubs[<-procID] = true
				fmt.Println(time.Now().Format(time.RFC850) + " hacmd/ctrl just checked in.")
				pingCMD := "{\"commands\": [\"https://192.168.192.185/api/config\",\"https://192.168.192.56/api/config\",\"https://192.168.192.58/api/config\"]}"
				go PublishTo("hacmd/cmd", client, pingCMD)
			}

			if strings.Contains(msgStr, "sensors") {
				fmt.Println("Received Sensors Results")
			}

		//Issue Sensor Commands
		case <-time.After(60 * time.Second):
			fmt.Println("Timmer Up.  Issue Commands.")
			//if (configuredHubs[<-procID] == true) {
			pingCMD := "{\"commands\": [\"https://192.168.192.185/api/vw6OE0D7kDffGeMpA5JXhHuQZaXMtX8Jh8zcEyyb/sensors\",\"https://192.168.192.56/api/OdZrhUY-514oY5iuhkg4lFgm0iL6qRlCIAAqvA3y/sensors\",\"https://192.168.192.58/api/Yj3knZa5VWYGLYo6n7TAOVWrRW-3VK9Un1UALd9t/sensors\"]}"
			go PublishTo("hacmd/cmd", client, pingCMD)
			//configuredHubs[<-procID] = false
			//}

		}
	}
}
