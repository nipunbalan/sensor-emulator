package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type cmdChannelConfigType struct {
	protocol string
	topic    string
	broker   string
}

// RunCommandListner listens remote commands
func RunCommandListner(sensorCmdGoChan *chan string) {

	clientid := viper.GetString("deviceid")

	deliveries := make(chan string, 4096)

	// Go routine to initiate and run the MQTT Client
	go InitMQTTClient(clientid, &deliveries)

	for {

		//fmt.Println("InLoop")
		msg := <-deliveries

		evalCommand(msg, sensorCmdGoChan)

	}

}

func evalCommand(cmd string, sensorCmdChannel *chan string) {

	log.Printf("Received a command: %s", cmd)

	cmd = strings.ToLower(cmd)

	if strings.HasPrefix(cmd, "set s_rt") {

		sensorid := ""
		datarate := 0

		_, err := fmt.Sscanf(cmd, "set s_rt %s %d", &sensorid, &datarate)

		failOnError(err, "Error!. Reading command value")

		fmt.Printf("Sensor ID: %s\nData Rate: %d\n", sensorid, datarate)

		*sensorCmdChannel <- fmt.Sprintf("%s %d", sensorid, datarate)

		fmt.Printf("Sent\n")

	} else if strings.HasPrefix(cmd, "get stats") {

		log.Printf("Received 'GET STATS command")

	} else {

		log.Printf("Invalid Command")
	}

}
