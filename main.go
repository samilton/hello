package main

import (
	"github.com/Shopify/sarama"

	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	addr    = flag.String("addr", ":8080", "The address to bind to")
	brokers = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose = flag.Bool("verbose", false, "Turn on Hello logging")
)

type Server struct {
	LogListener sarama.SyncProducer
}

type openWeatherMapData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type weatherData struct {
	Name      string    `json:"name"`
	Temp      float64   `json:"temp"`
	Timestamp time.Time `json:"timestamp"`
}

type logList struct {
	Files []os.FileInfo `json:"files"`
}

func newLogListener(brokerList []string) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10

	producer, err := sarama.NewSyncProducer(brokerList, config)

	if err != nil {
		log.Fatalln("Failed to start Sarama producer: ", err)
	}

	return producer
}

func convertToFahrenheit(kelvin float64) float64 {
	log.Printf("Converting %f to Fahrenheit", kelvin)
	return math.Floor((1.8 * (kelvin - 273)) + 32)
}

func logFileHandler(w http.ResponseWriter, r *http.Request) {
	files, _ := ioutil.ReadDir("./")

	d := logList{}
	d.Files = files

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	log.Printf("files: %s: %.2f", d.Files)
	json.NewEncoder(w).Encode(d)
}

func (s *Server) Run(addr string) error {
	httpServer := &http.Server{
		Addr: addr,
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/logs/", logFileHandler)
	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		city := strings.SplitN(r.URL.Path, "/", 3)[2]

		data, err := query(city)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		d := weatherData{}
		d.Temp = convertToFahrenheit(data.Main.Kelvin)
		d.Name = data.Name
		d.Timestamp = time.Now()

		partition, offset, err := s.LogListener.SendMessage(&sarama.ProducerMessage{
			Topic: "important",
			//Value: sarama.StringEncoder(fmt.Sprintf("City: %s", city)),
			Value: sarama.StringEncoder(r.URL.Path),
		})

		log.Printf("Your data is stored with unique identifier important/%d/%d", partition, offset)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		log.Printf("openWeatherMap: %s: %.2f", d.Name, d.Temp)
		json.NewEncoder(w).Encode(d)
	})

	log.Printf("Listening for requests on %s...\n", addr)
	return httpServer.ListenAndServe()
}

func main() {
	flag.Parse()

	//	Logger := log.New(os.Stdout, "[Hello]", log.LstdFlags)

	brokerList := strings.Split(*brokers, ",")

	server := &Server{
		LogListener: newLogListener(brokerList),
	}

	server.Run(*addr)
}

func query(city string) (openWeatherMapData, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + city)

	if err != nil {
		return openWeatherMapData{}, err
	}

	defer resp.Body.Close()

	var d openWeatherMapData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return openWeatherMapData{}, err
	}

	return d, nil
}
