package main

// This is where we create a subscribtion to MQTT & send updates
// to Telegram

import (
	"fmt"
	"log"
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

	//Send new IP to Telegram and MQTT
	botmsg := tgbotapi.NewMessage(int64(Tgtarget), "New home IP detected: "+newIP)
	bot.Send(botmsg)
	fmt.Printf("%s - Sent to Telegram\n", curtime)
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
