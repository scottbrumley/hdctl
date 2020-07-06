package main

import (
	"encoding/json"
	"crypto/rand"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

//define a function for the default message handler
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

// Main Functions
func ReadConfig(configstr string)(user string, pass string, broker string){

	plan, _ := ioutil.ReadFile(configstr)
	var data map[string]interface{}
	err := json.Unmarshal(plan, &data)

	if err != nil {
		log.Fatalln(err)
	}
	user = data["user"].(string)
	pass = data["pass"].(string)
	broker = data["broker"].(string)
	return
}
func ProcUUID()(uuid string){
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}
func connect_mqtt(broker string, user string, pass string)(c mqtt.Client) {

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + broker)
	opts.SetClientID(ProcUUID())
	opts.SetDefaultPublishHandler(f)
	opts.SetUsername(user)
	opts.SetPassword(pass)
	//create and start a client using the above ClientOptions
	c = mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return

}
func SubscribeCtl(client mqtt.Client, c chan string){
	if token := client.Subscribe("hacmd/#", 0, func(client mqtt.Client, msg mqtt.Message) {
		time.Sleep(6 * time.Second)
		// Get ProcID from Topic
		topicArray := strings.Split(msg.Topic(),"/")
		procID := topicArray[1]

		// If intiation strings match then return configuration
		if string(msg.Payload()) == "{\"procID\": \"" + procID + "\",\"action\": \"initiate\"}" {
			fmt.Println("hacmd " + procID + " just checked in.")
			brokerconfig := "{\"name\": \"configuration\",\"brokers\": [\"192.168.192.10\"],\"hubs\": [\"192.168.192.185\"]}"
			PublishCtl(client, procID, brokerconfig)
			c <-brokerconfig
			return
		}
	}); token.Wait() && token.Error() != nil {
		return
	}
	return
}

func Unsubscribe(client mqtt.Client){
	// Unscribe
	if token := client.Unsubscribe("hacmd/#"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func PublishCtl(client mqtt.Client, procID string, msg string){
	// Send Configuration
	fmt.Println("Sending configurations to hacmd/" + procID)
	token := client.Publish("hacmd/" + procID, 0, false, msg)
	token.Wait()

	time.Sleep(6 * time.Second)
}