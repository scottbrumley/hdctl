package main

import (
	"fmt"
	"time"

	//"time"
)

func main() {
	c := make(chan string)
	user, pass, broker := ReadConfig("config.json")
	client := connect_mqtt(broker, user, pass)
	// Subscribe to a topic
	go SubscribeCtl(client, c)

	timeout := time.After(1 * time.Second)
	for {
		select {
		case s := <-c:
			fmt.Println("Sending: " + s)
			continue
		case <-timeout:
			fmt.Println("timeout")
			break
		}
	}

	// Unscribe
	//if token := client.Unsubscribe("hacmd/#"); token.Wait() && token.Error() != nil {
	//	fmt.Println(token.Error())
	//	os.Exit(1)
	//}

	// Disconnect
	//client.Disconnect(250)
	//time.Sleep(1 * time.Second)

}