package main

// This is where we create a subscribtion to MQTT & send updates
// to Telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	Mqtt "github.com/eclipse/paho.mqtt.golang"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var currentIP string = "127.0.0.1"

var messagePubHandler Mqtt.MessageHandler = func(client Mqtt.Client, msg Mqtt.Message) {
	fmt.Printf("%s - %s : %s\n", curtime, msg.Topic(), msg.Payload())
	newIP := string(msg.Payload())
	if newIP != currentIP {
		fmt.Printf("%s - New IP '%s' detected, updating ...\n", curtime, newIP)
		updateIP(newIP)
		currentIP = newIP
	}
}

var connectHandler Mqtt.OnConnectHandler = func(client Mqtt.Client) {
	fmt.Printf("%s - Connected to MQTT\n", curtime)
}

var connectLostHandler Mqtt.ConnectionLostHandler = func(client Mqtt.Client, err error) {
	fmt.Printf("%s - Connection to MQTT lost: %v\n", curtime, err)
}

var curtime = time.Now().Format(time.RFC3339)

func mqttsub(client Mqtt.Client) {
	topic := Mqtttopic
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("%s - Subscribed to topic: %s\n", curtime, topic)
}

func updateIP(newIP string) {
	//
	// Setup Telegram bot
	//
	bot, err := tgbotapi.NewBotAPI(Tgbotkey)
	if err != nil {
		log.Fatalf("%s - Error - Bot connection failed", curtime)
	}

	log.Printf("%s - Authorized on account '%s'", curtime, bot.Self.UserName)

	//Send new IP to Telegram
	botmsg := tgbotapi.NewMessage(int64(Tgtarget), "New home IP detected: "+newIP)
	bot.Send(botmsg)
	fmt.Printf("%s - Sent to Telegram\n", curtime)

	//Update Dynu DNS record

	type ApiData struct {
		Name  string `json:"name"`
		Group string `json:"group"`
		Ip    string `json:"ipv4Address"`
		Ttl   int    `json:"ttl"`
		Ip4b  bool   `json:"ipv4"`
	}

	dynuData := ApiData{Name: DomName, Group: DomGrp, Ip: newIP, Ttl: 90, Ip4b: true}

	dynuJson, err := json.Marshal(dynuData)
	if err != nil {
		log.Fatalf("Error %v - Error marshaling JSON", err)
	}

	req, err := http.NewRequest(http.MethodPost, DDNSAPI, bytes.NewBuffer(dynuJson))
	if err != nil {
		log.Fatalf("Error %v - IP API connection failed", err)
	}
	req.Header.Set("API-Key", DDNSKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error %v - IP API JSON post", err)
	}
	defer resp.Body.Close()

	fmt.Printf("%s - API POST response status: %s", curtime, resp.Status)

	botmsg = tgbotapi.NewMessage(int64(Tgtarget), fmt.Sprintf("Updating Dynu host '%s' status: %s", dynuData.Name, resp.Status))
	bot.Send(botmsg)
}

func main() {

	// Run loop (for k8s)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	//Init MQTT broker, connect and listen
	opts := Mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", Mqtthost, Mqtthostport))
	opts.SetClientID(Mqttclid)
	opts.SetUsername(Mqttuser)
	opts.SetPassword(Mqttpwd)
	opts.SetCleanSession(false)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	myclient := Mqtt.NewClient(opts)
	if token := myclient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	mqttsub(myclient)

	//
	<-c

	myclient.Disconnect(15)
}
