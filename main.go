package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	c := make(chan string)
	user, pass, broker, mongoDB := ReadConfig("config.json")
	client := connect_mqtt(broker, user, pass)
	mongoClient := connect_mongoDB(mongoDB)

	//Checking the connection
	mongoClient.Ping(context.TODO(), nil)
	fmt.Println("Database connected")

	// Subscribe to a topic
	go SubscribeCtl(client, c)
	procID := <-c
	timeout := time.After(1 * time.Second)
	for {
		select {
		case procID = <-c:
			fmt.Println(time.Now().Format(time.RFC850) + " Sending: " + procID)
			continue
		case <-timeout:
			fmt.Println("timeout")
			break
		}
		pingCMD := "{\"commands\": [\"https://192.168.192.185/api/vw6OE0D7kDffGeMpA5JXhHuQZaXMtX8Jh8zcEyyb/sensors\",\"https://192.168.192.56/api/OdZrhUY-514oY5iuhkg4lFgm0iL6qRlCIAAqvA3y/sensors\",\"https://192.168.192.58/api/Yj3knZa5VWYGLYo6n7TAOVWrRW-3VK9Un1UALd9t/sensors\"]}"
		PublishCtl(client, procID, pingCMD)
		timeout = time.After(60 * time.Second)
	}
}
