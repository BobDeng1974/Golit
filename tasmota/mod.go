package tasmota

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sirjson/golit/common"
	"strings"
	"sync"
)

type Device struct {
	ID     int
	Feed   string
	Status Status
	Color  string
	White  string
}

type StatusResponse struct {
	Status Status
}

type Status struct {
	Module       int
	FriendlyName []string
	Topic        string
	ButtonTopic  string
	Power        int
	PowerOnState int
	LedState     int
	SaveData     int
	SaveState    int
	SwitchTopic  string
	SwitchMode   []int
	ButtonRetain int
	SwitchRetain int
	SensorRetain int
	PowerRetain  int
}

type ColorState struct {
	Color string
}

func Delete(device_feed string) {
	log.Print("tasmota_delete", device_feed)
	db, err := sql.Open("sqlite3", common.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Print(err)
	}
	stmt, err := tx.Prepare("DELETE TOP 1 FROM tasmota_device WHERE Feed = ?")
	if err != nil {
		log.Print(err)
	}
	defer stmt.Close()
	res, err2 := stmt.Exec(device_feed)

	err = tx.Commit()
	if err2 == nil {
		rows, err := res.RowsAffected()
		log.Print("Rows affectted", rows)
		if err != nil {
			log.Print(err)
		}
	} else {
		log.Print(err2.Error())
	}
	if err != nil {
		log.Print(err)
	}
}

func Command(mqtt_host string, cmds []string) {
	opts := mqtt.NewClientOptions().AddBroker(mqtt_host)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Print(token.Error())
		return
	}

	if token := client.Publish(fmt.Sprintf("cmnd/%s/%s", cmds[0], cmds[1]), 0, false, cmds[2]); token.Wait() && token.Error() != nil {
		log.Print(token.Error())
		return
	}
	client.Disconnect(100)
}

func GetInfo(mqtt_host string, feed string, prop string, subscriptionIsResult bool) []byte {
	output := make([]byte, 0)
	opts := mqtt.NewClientOptions().AddBroker(mqtt_host)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Print(token.Error())
		return output
	}

	var wg sync.WaitGroup
	wg.Add(1)
	var sub string
	if subscriptionIsResult {
		sub = "RESULT"
	} else {
		sub = strings.ToUpper(prop)
	}
	if token := client.Subscribe(fmt.Sprintf("stat/%s/%s", feed, sub), 0, func(client mqtt.Client, msg mqtt.Message) {
		output = msg.Payload()
		client.Disconnect(100)
		wg.Done()
	}); token.Wait() && token.Error() != nil {
		log.Print(token.Error())
		return output
	}

	if token := client.Publish(fmt.Sprintf("cmnd/%s/%s", feed, prop), 0, false, ""); token.Wait() && token.Error() != nil {
		log.Print(token.Error())
		return output
	}
	wg.Wait()
	return output
}

func UnmarshalStatus(data []byte) (*StatusResponse, error) {
	var t StatusResponse
	if len(data) == 0 {
		log.Print("TasmotaStatusResponse unmarshal failed: no data")
		return nil, errors.New("TasmotaStatusResponse unmarshal failed: no data")
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		log.Print("JSON", err.Error())
		return nil, errors.New("JSON failure")
	}
	return &t, nil
}

func UnmarshalColor(data []byte) (*ColorState, error) {
	var t ColorState
	if len(data) == 0 {
		log.Print("TasmotaStatusResponse unmarshal failed: no data")
		return nil, errors.New("TasmotaStatusResponse unmarshal failed: no data")
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		log.Print("JSON", err.Error())
		return nil, errors.New(err.Error())
	}
	return &t, nil
}

func UnmarshalDevice(data []byte) (*Device, error) {
	var t Device
	if len(data) == 0 {
		log.Print("TasmotaDevice unmarshal failed: no data")
		return nil, errors.New("TasmotaDevice unmarshal failed: no data")
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		log.Print("JSON", err.Error())
		return nil, errors.New(err.Error())
	}
	return &t, nil
}

func Add(device *Device) {
	log.Print("tasmota_add", device.Feed)
	db, err := sql.Open("sqlite3", common.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Print(err)
	}
	stmt, err := tx.Prepare("INSERT INTO tasmota_device (Feed) VALUES(?)")
	if err != nil {
		log.Print(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(device.Feed)
	err = tx.Commit()
	if err != nil {
		log.Print(err)
	}
}

func Fetch() []Device {
	var devlist []Device
	db, err := sql.Open("sqlite3", common.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT Feed as count FROM tasmota_device")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		device := Device{}

		err = rows.Scan(&device.Feed)
		if err != nil {
			log.Fatal(err)
		}

		devlist = append(devlist, device)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return devlist
}
