package hdctl

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
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

type jobsStruct struct {
	JobID      int        `json:"jobid"`
	Trigger    string     `json:"trigger"`
	ProcID     string     `json:"procid"`
	Action     string     `json:"action"`
	ActionType string     `json:"actiontype"`
	Commands   []Commands `json:"commands"`
}
type Commands struct {
	URL       string `json:"url"`
	Hubid     string `json:"hubid"`
	VendorTag string `json:"vendortag"`
}

type LutronCommands struct {
	Id        string `json:"id"`
	Ipadd     string `json:"ipadd"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Face      string `json:"face"`
	Type      string `json:"type"`
	Action    string `json:"action"`
	VendorTag string `json:"vendortag"`
}

type hacmdInit struct {
	ProcID     string                 `json:"procid"`
	HubID      string                 `json:"hubid"`
	Action     string                 `json:"action"`
	ActionType string                 `json:"actiontype"`
	Command    string                 `json:"command"`
	Results    map[string]interface{} `json:"results"`
}

type hactl struct {
	ProcID       string
	CtrlMessages chan string
	mqttClient   mqtt.Client
	mongoClient  *mongo.Client
}

// Main Functions
func New(configStr string) hactl {
	ch1 := make(chan string)
	user, pass, broker, procID, mongoDB := ReadConfig("config.json")
	mqttClient := connect_mqtt(broker, user, pass)
	mongoClient := connect_mongoDB(mongoDB)

	//Checking the connection
	mongoClient.Ping(context.TODO(), nil)
	fmt.Println("Database connected")

	// Subscribe to a topic
	go SubscribeTo("hacmd/ctrl", mqttClient, ch1)
	ctrlProc := hactl{procID, ch1, mqttClient, mongoClient}
	go handleRequests()
	return ctrlProc
}
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}
func pingPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Ping Page!")
	fmt.Println("Endpoint Hit: Ping")
}
func lutronPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Lutron Page!")
	fmt.Println("Endpoint Hit: Lutron")
}
func huePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Hue Page!")
	fmt.Println("Endpoint Hit: Hue")
}
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/ping", pingPage)
	myRouter.HandleFunc("/lutron", lutronPage)
	myRouter.HandleFunc("/hue", huePage)
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}
func (controlCenter hactl) SendCommands(msg string) {
	topic := "hacmd/cmd"
	fmt.Println(time.Now().Format(time.RFC850) + " Sending command " + msg + " to " + topic)
	publishTo(topic, controlCenter.mqttClient, msg)
}
func (controlCenter hactl) UpdateSensors(hubid string, sensorVal map[string]interface{}) {
	updateOne_mongoDB(controlCenter.mongoClient, "homeSysDB", "sensors", hubid, sensorVal)
}
func (controlCenter hactl) FindJobs(brokerId string, actiontype string) (results []string) {
	results = find_mongoDB(controlCenter.mongoClient, "homeSysDB", "jobs", brokerId, actiontype)
	return results
}
func ReadCtrl(configstr string) (procID string, hubid string, action string, command string, results map[string]interface{}) {
	res := hacmdInit{}
	err := json.Unmarshal([]byte(configstr), &res)
	if err != nil {
		log.Fatalln(err)
	}
	procID = res.ProcID
	hubid = res.HubID
	action = res.Action
	command = res.Command
	results = res.Results
	return
}
func (controlCenter hactl) ReadJob(jobStr string) (jobid int, trigger string, procid string, action string, actiontype string, commands []Commands) {
	data := jobsStruct{}
	err := json.Unmarshal([]byte(jobStr), &data)
	if err != nil {
		log.Fatalln(err)
	}
	jobid = data.JobID
	trigger = data.Trigger
	procid = data.ProcID
	action = data.Action
	actiontype = data.ActionType
	commands = data.Commands
	return
}
func ReadConfig(configstr string) (user string, pass string, broker string, procID, mongoDB string) {
	plan, _ := ioutil.ReadFile(configstr)
	var data map[string]interface{}
	err := json.Unmarshal(plan, &data)

	if err != nil {
		log.Fatalln(err)
	}
	user = data["user"].(string)
	pass = data["pass"].(string)
	broker = data["broker"].(string)
	procID = data["procID"].(string)
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
func updateOne_mongoDB(mongoClient *mongo.Client, dbStr string, collectionStr string, hubid string, retval map[string]interface{}) {
	//Set up a context required by mongo.Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//To close the connection at the end
	defer cancel()
	// Update Sensor Database
	SensorsCollection := mongoClient.Database(dbStr).Collection(collectionStr)
	filter := bson.D{{"hubid", hubid}}
	// Need to specify the mongodb output operator too
	newName := bson.D{
		{"$set", bson.D{
			{collectionStr, retval},
		}},
	}
	res, err := SensorsCollection.UpdateOne(ctx, filter, newName)
	if err != nil {
		log.Fatal(err)
	}
	updatedObject := *res
	fmt.Printf("Modified count is : %d", updatedObject.ModifiedCount)
	//fmt.Printf("The matched count is : %d, the modified count is : %d", updatedObject.MatchedCount, updatedObject.ModifiedCount)

}
func find_mongoDB(mongoClient *mongo.Client, dbStr string, collectionStr string, procid string, actionStr string) (results []string) {
	//Set up a context required by mongo.Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//To close the connection at the end
	defer cancel()

	// Collection to retrieve from
	Collection := mongoClient.Database(dbStr).Collection(collectionStr)

	searchStr := bson.D{{"procid", procid}}
	if actionStr != "" {
		searchStr = bson.D{{"procid", procid}, {"actiontype", actionStr}}
	}

	cur, error := Collection.Find(ctx, searchStr)

	var alljobs []*jobsStruct

	//Loops over the cursor stream and appends to result array
	for cur.Next(context.TODO()) {
		var jobResult jobsStruct
		err := cur.Decode(&jobResult)
		if err != nil {
			log.Fatal(error)
		}
		alljobs = append(alljobs, &jobResult)
	}
	//dont forget to close the cursor
	defer cur.Close(context.TODO())

	for _, element := range alljobs {
		res2B, _ := json.Marshal(element)
		results = append(results, string(res2B))
	}

	return
}
func SubscribeTo(topic string, client mqtt.Client, c chan string) {
	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payloadStr := string(msg.Payload())
		//time.Sleep(6 * time.Second)
		// Get ProcID from Topic
		c <- payloadStr

		/*
			if len(payloadStr) > 0 {
				c <- payloadStr
			}
		*/

		return
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
func publishTo(topic string, client mqtt.Client, msg string) {
	// Publish Response
	token := client.Publish(topic, 0, false, msg)
	token.Wait()
}
