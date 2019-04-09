package messaging

import (
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	RATECOUNTER "github.com/paulbellamy/ratecounter"
	"github.com/spf13/viper"
)

//InitMQTTClient Initiates the MQTT client and connects to the broker
func InitMQTTClient(clientid string, deliveries *chan string, dataRateReadSeconds int) {

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
	counter := RATECOUNTER.NewRateCounter(1 * time.Second)

	//Go routine to print out data sending rate
	go func() {
		for {
			fmt.Printf("%s | Sending rate of '%s' : %d \t records/sec\n", time.Now().Format(time.RFC3339), clientid, counter.Rate())
			time.Sleep(time.Second * time.Duration(dataRateReadSeconds))
		}
	}()

	for {
		payload := <-*deliveries
		token := client.Publish(topic, byte(qos), false, payload)
		token.Wait()
		counter.Incr(1)
		//fmt.Printf("Rate of %s:%d records/sec\n", clientid, counter.Rate())
	}
}
