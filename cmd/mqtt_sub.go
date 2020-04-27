package cmd

import (
	"fmt"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

//InitMQTTClient Initiates the MQTT client and connects to the broker
func InitMQTTClient(clientid string, deliveries *chan string) {

	topic := viper.GetString("messaging.command_topic")
	broker := viper.GetString("messaging.broker")
	//	password := viper.GetString("messaging.password")
	//	user := viper.GetString("messaging.user")
	id := clientid
	cleansess := false
	qos := 2
	store := ":memory:"

	//fmt.Println("Topic:" + broker)

	//topic = strings.Replace(topic, "{sensor_name}", clientid, -1)

	if topic == "" {
		fmt.Println("Invalid topic, must not be empty")
		return
	}

	if broker == "" {
		fmt.Println("Invalid broker URL, must not be empty")
		return
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(id)
	//	opts.SetUsername(user)
	//	opts.SetPassword(password)
	opts.SetCleanSession(cleansess)
	if store != ":memory:" {
		opts.SetStore(MQTT.NewFileStore(store))
	}

	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := MQTT.NewClient(opts)
	defer client.Disconnect(250)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	for {
		incoming := <-choke
		*deliveries <- incoming[1]
		//	fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming[0], incoming[1])
	}

}
