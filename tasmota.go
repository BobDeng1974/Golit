package main

import (
	"database/sql"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
	"sync"
)

type TasmotaDevice struct {
	ID     int
	Feed   string
	Status TasmotaStatus
	Color  string
	White  string
}

type TasmotaStatusResponse struct {
	Status TasmotaStatus
}

type TasmotaStatus struct {
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

func tasmota_count() int {
	var output int
	db, err := sql.Open("sqlite3", "./huelishous.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT COUNT(*) as count FROM tasmota_device")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&output)
	if err != nil {
		log.Fatal(err)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return output
}

func tasmota_delete(device_feed string) {
	println("tasmota_delete", device_feed)
	db, err := sql.Open("sqlite3", "./huelishous.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		println(err)
	}
	stmt, err := tx.Prepare("DELETE TOP 1 FROM tasmota_device WHERE Feed = ?")
	if err != nil {
		println(err)
	}
	defer stmt.Close()
	res, err2 := stmt.Exec(device_feed)

	err = tx.Commit()
	if err2 == nil {
		rows, err := res.RowsAffected()
		println("Rows affectted", rows)
		if err != nil {
			println(err)
		}
	} else {
		println(err2.Error())
	}
	if err != nil {
		println(err)
	}
}

func tasmota_stat(feed string, prop string) []byte {
	output := make([]byte, 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://moi:1883")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		println(token.Error())
		return output
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if token := client.Subscribe(fmt.Sprintf("stat/%s/%s", feed, strings.ToUpper(prop)), 0, func(client mqtt.Client, msg mqtt.Message) {
		output = msg.Payload()
		client.Disconnect(100)
		wg.Done()
	}); token.Wait() && token.Error() != nil {
		println(token.Error())
		return output
	}

	if token := client.Publish(fmt.Sprintf("cmnd/%s/%s", feed, prop), 0, false, ""); token.Wait() && token.Error() != nil {
		println(token.Error())
		return output
	}
	wg.Wait()
	return output
}

func tasmota_stat_result(feed string, prop string) []byte {
	output := make([]byte, 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://moi:1883")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		println(token.Error())
		return output
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if token := client.Subscribe(fmt.Sprintf("stat/%s/%s", feed, "RESULT"), 0, func(client mqtt.Client, msg mqtt.Message) {
		output = msg.Payload()
		client.Disconnect(100)
		wg.Done()
	}); token.Wait() && token.Error() != nil {
		println(token.Error())
		return output
	}

	if token := client.Publish(fmt.Sprintf("cmnd/%s/%s", feed, prop), 0, false, ""); token.Wait() && token.Error() != nil {
		println(token.Error())
		return output
	}
	wg.Wait()
	return output
}

func tasmota_add(device TasmotaDevice) {
	println("tasmota_add", device.Feed)
	db, err := sql.Open("sqlite3", "./huelishous.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		println(err)
	}
	stmt, err := tx.Prepare("INSERT INTO tasmota_device (Feed) VALUES(?)")
	if err != nil {
		println(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(device.Feed)
	err = tx.Commit()
	if err != nil {
		println(err)
	}
}

func tasmota_fetch() []TasmotaDevice {
	var devlist []TasmotaDevice
	db, err := sql.Open("sqlite3", "./huelishous.db")
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
		device := TasmotaDevice{}

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
