package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/nipunbalan/sensor-emulator/messaging"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var wg = sync.WaitGroup{}

// RunEmulator runs the sensor emulator
func RunEmulator(sensorCmdGoChan *chan string) {

	rateLimitChannelMap := make(map[string](chan int64))

	sensors := viper.GetStringMap("sensors")

	//Go routine to receive commands and dispatch it to corresponding sensor threads
	go func() {
		for {
			cmd := <-*sensorCmdGoChan
			fmt.Printf("Received '%s'", cmd)
			sensorid := ""
			dataRate := 0
			fmt.Sscanf(cmd, "%s %d", &sensorid, &dataRate)
			fmt.Printf("Sensor:%s|NewDataRate:%d\n", sensorid, dataRate)
			rateLimitChannelMap[sensorid] <- int64(dataRate)
			fmt.Printf("Sent\n")

		}
	}()

	for key := range sensors {

		rateLimitChannelMap[key] = make(chan int64)

		sensorfreq := viper.GetInt64("sensors." + key + ".freq")
		sensorTypeStr := viper.GetString("sensors." + key + ".type")

		wg.Add(1)

		fmt.Println("Key: " + key + " | Frequency: " + strconv.FormatInt(int64(sensorfreq), 10) + " | Type: " + sensorTypeStr)

		go runSensor(key, sensorfreq, sensorTypeStr, rateLimitChannelMap[key])

	}

	wg.Wait()

}

func runSensor(name string, freq int64, sensorType string, rateLimitChannelMap chan int64) {

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

	//Go routine to set new limit
	go func() {
		for {
			newlimit := <-rateLimitChannelMap
			limit := rate.Limit(newlimit)
			fmt.Printf("Setting new data rate '%d' on the sensor '%s'\n", newlimit, name)
			limiter.SetLimit(limit)
		}
	}()

	for {
		file, err := os.Open(fileNameStr)

		failOnError(err, "Error!. Unable to open the file!!")

		reader := bufio.NewReader(file)
		scanner := bufio.NewScanner(reader)

		for i := 0; scanner.Scan(); i++ {
			//Skip header
			if i == 0 {
				i++
				continue
			}
			limiter.Wait(cntx)
			line := scanner.Text()
			body := fmt.Sprintf("%d,%s,%s", time.Now().UnixNano(), name, line)
			//Sending it to MQTTClient through the channel
			deliveries <- body
			//log.Println(name + " : " + body)
			failOnError(err, "Failed to publish a message")
			i++
			//	fmt.Printf("Record Sending Rate: %d records per second\n", counter.Rate())
		}
		file.Close()
	}
}
