package main

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/robotalks/robo.go/pkg/l1/comm/mqtt"
	"github.com/robotalks/robo.go/pkg/l1/msgs"

	_ "github.com/robotalks/robo.go/pkg/joystick/msgs"
)

var (
	mqttURL = "mqtt://localhost:1883/robo/"
)

func init() {
	if val := os.Getenv("ROBO_MQTT_URL"); val != "" {
		mqttURL = val
	}
	flag.StringVar(&mqttURL, "mqtt", mqttURL, "MQTT broker URL.")
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lmicroseconds)

	q, err := mqtt.NewQueueFromURL(mqttURL)
	if err != nil {
		log.Fatalln(err)
	}

	q.Sub("#", mqtt.Handler(func(topic string, payload []byte) {
		if strings.HasSuffix(topic, "/meta") {
			log.Printf("%s: %s", topic, string(payload))
			return
		}
		typed, err := msgs.DecodeTyped(payload)
		if err != nil {
			log.Printf("%s: bad message: %v", topic, err)
			return
		}
		msg, err := typed.Decode()
		if err != nil {
			log.Printf("%s: decode error: (type_id=%x) %v", topic, typed.TypeId, err)
			return
		}
		log.Printf("%s: [%s] %s", topic,
			reflect.Indirect(reflect.ValueOf(msg)).Type().Name(),
			msg.(msgs.SerializableMessage).Serializable().String())
	}))
	<-(chan struct{})(nil)
}
