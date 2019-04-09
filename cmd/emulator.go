package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/nipunbalan/sensorsim/messaging"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var wg = sync.WaitGroup{}

func runEmulator() {
	//fmt.Println("Sensor emulator running")
	sensors := viper.GetStringMap("sensors")

	for key := range sensors {

		fmt.Print(key + ":")

		sensorfreq := viper.GetInt64("sensors." + key + ".freq")
		sensorTypeStr := viper.GetString("sensors." + key + ".type")

		wg.Add(1)
		fmt.Println("Key: " + key + " | Frequency: " + strconv.FormatInt(int64(sensorfreq), 10) + " | Type: " + sensorTypeStr)
		go runSensor(key, sensorfreq, sensorTypeStr)

	}
	wg.Wait()

}

func runSensor(name string, freq int64, sensorType string) {

	defer wg.Done()

	deliveries := make(chan string, 4096)

	dataratereadseconds := viper.GetInt("messaging.dataratereadseconds")

	go messaging.InitMQTTClient(name, &deliveries, dataratereadseconds)

	// Rate limiter limits the number of sensor values read per second
	limit := rate.Limit(freq)
	limiter := rate.NewLimiter(limit, 1)
	cntx := context.Background()

	//Get Filename and open the file
	fileNameStr := viper.GetString("sensors." + name + ".file")

	for {
		file, err := os.Open(fileNameStr)

		failOnError(err, "Error!. Unable to open the file!!")

		reader := bufio.NewReader(file)
		scanner := bufio.NewScanner(reader)

		for i := 0; scanner.Scan(); i++ {
			if i == 0 {
				i++
				continue
			}
			limiter.Wait(cntx)
			line := scanner.Text()
			body := name + " : " + line
			//Sending it to MQTTClient through the channel
			deliveries <- body
			//	log.Println(name + " : " + body)
			failOnError(err, "Failed to publish a message")
			i++
			//	fmt.Printf("Record Sending Rate: %d records per second\n", counter.Rate())
		}
		file.Close()
	}
}
