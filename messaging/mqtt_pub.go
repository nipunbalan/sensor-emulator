package messaging

import (
	"fmt"

	"github.com/spf13/viper"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

//InitMQTTClient Initiates the MQTT client and connects to the broker
func InitMQTTClient(clientid string, deliveries *chan string) {

	topic := viper.GetString("messaging.topic")
	broker := viper.GetString("messaging.broker")
	//	password := viper.GetString("messaging.password")
	//	user := viper.GetString("messaging.user")
	id := clientid
	cleansess := false
	qos := 0
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

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer client.Disconnect(250)
	fmt.Printf("Sensor: %s data started publsihing to topic: %s \n", clientid, topic)
	for {
		payload := <-*deliveries
		token := client.Publish(topic, byte(qos), false, payload)
		token.Wait()
	}
}
