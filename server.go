package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type SensorReading struct {
	gorm.Model
	Timestamp int64
	Blob      string
	DeviceID  string
}

type server struct {
	db *gorm.DB
	r  *mux.Router
}

func NewServer(db *gorm.DB) *server {
	r := mux.NewRouter()
	s := &server{db, r}
	r.HandleFunc("/devices/{deviceID}/insert", s.handleInsert).Methods("POST")
	r.HandleFunc("/devices/{deviceID}/recent", s.handleGetRecent).Methods("GET")
	return s
}

func (s *server) ListenAndServe(addr string) {
	http.ListenAndServe(addr, s.r)
}

func (s *server) handleInsert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error reading insert body: %s\n", err)
	}

	var readings struct {
		Readings []SensorReading
	}
	if err := json.Unmarshal(bodyBytes, &readings); err != nil {
		fmt.Printf("error unmarshalling input: %q: %s\n", string(bodyBytes), readings)
	}

	tx := s.db.Begin()
	for _, reading := range readings.Readings {
		reading.DeviceID = deviceID
		tx.Create(&reading)
	}
	tx.Commit()
}

func (s *server) handleGetRecent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]

	defer r.Body.Close()
	var readings []SensorReading
	now := time.Now()
	before := now.Add(-4 * time.Hour)
	s.db.Where("\"timestamp\"> ?", before.Unix()).
		Where("\"timestamp\" < ?", now.Unix()).
		Where("device_id = ?", deviceID).
		Find(&readings)
	readingBytes, _ := json.Marshal(map[string]interface{}{"readings": readings})
	w.Write(readingBytes)
}

func main() {
	postgresIP := os.Args[1]
	db, err := gorm.Open("postgres", "postgres://postgres:test@"+postgresIP+"?sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to connect to db: %s", err))
	}

	db.AutoMigrate(&SensorReading{})
	db.Model(&SensorReading{}).AddIndex("idx_device_reading", "timestamp")

	server := NewServer(db)
	fmt.Println("listening...")
	server.ListenAndServe(":8080")
}
