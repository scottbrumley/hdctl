package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Helper Functions
func ValidIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

	// Is a valid IP Address
	if re.MatchString(ipAddress) {
		// Test the octet ranges
		octets := strings.Split(ipAddress, ".")

		for i := range octets {
			octet, err := strconv.Atoi(octets[i])
			if err != nil {
				log.Fatalln(err)
			}

			if octet > 255 {
				return false
			}
			if octet < 0 {
				return false
			}
		}
		return true
	}

	return false
}
func validHost(host string) bool {
	host = strings.Trim(host, " ")
	re, _ := regexp.Compile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	if re.MatchString(host) {
		return true
	}
	return false
}

//define a function for the default message handler
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

// Main Functions
func ReadConfig(configstr string) (user string, pass string, broker string, mongoDB string) {
	plan, _ := ioutil.ReadFile(configstr)
	var data map[string]interface{}
	err := json.Unmarshal(plan, &data)

	if err != nil {
		log.Fatalln(err)
	}
	user = data["user"].(string)
	pass = data["pass"].(string)
	broker = data["broker"].(string)
	mongoDB = data["mongoDB"].(string)
	return
}
func ProcUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}
func connect_mqtt(broker string, user string, pass string) (client mqtt.Client) {

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + broker)
	opts.SetClientID(ProcUUID())
	opts.SetDefaultPublishHandler(f)
	opts.SetUsername(user)
	opts.SetPassword(pass)
	//create and start a client using the above ClientOptions
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return

}
func connect_mongoDB(mongoDB string) (client *mongo.Client) {
	//Set up a context required by mongo.Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//To close the connection at the end
	defer cancel()
	//We need to set up a client first
	//It takes the URI of your database
	client, error := mongo.NewClient(options.Client().ApplyURI(mongoDB))
	if error != nil {
		log.Fatal(error)
	}
	//Call the connect function of client
	error = client.Connect(ctx)
	//Checking the connection
	error = client.Ping(context.TODO(), nil)
	fmt.Println("Database connected")
	return
}
func SubscribeCtl(client mqtt.Client, c chan string) {
	if token := client.Subscribe("hacmd/#", 0, func(client mqtt.Client, msg mqtt.Message) {
		time.Sleep(6 * time.Second)
		// Get ProcID from Topic
		topicArray := strings.Split(msg.Topic(), "/")
		procID := topicArray[1]

		// If intiation strings match then return configuration
		if string(msg.Payload()) == "{\"procID\": \""+procID+"\",\"action\": \"initiate\"}" {
			fmt.Println("hacmd " + procID + " just checked in.")
			//brokerconfig := "{\"name\": \"configuration\",\"brokers\": [\"192.168.192.10\"],\"hubs\": [\"192.168.192.185\"]}"
			pingCMD := "{\"commands\": [\"https://192.168.192.185/api/config\",\"https://192.168.192.56/api/config\",\"https://192.168.192.58/api/config\"]}"
			PublishCtl(client, procID, pingCMD)
			c <- procID
			return
		}
	}); token.Wait() && token.Error() != nil {
		return
	}
	return
}
func Unsubscribe(client mqtt.Client) {
	// Unscribe
	if token := client.Unsubscribe("hacmd/#"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}
func Disconnect(client mqtt.Client) {
	// Disconnect
	client.Disconnect(250)
	time.Sleep(1 * time.Second)
}
func PublishCtl(client mqtt.Client, procID string, msg string) {
	// Send Configuration
	fmt.Println("Sending configurations to hacmd/" + procID)
	token := client.Publish("hacmd/"+procID, 0, false, msg)
	token.Wait()

	time.Sleep(6 * time.Second)
}
