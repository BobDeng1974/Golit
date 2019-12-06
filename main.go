package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sirjson/golit/common"
)

func file_exists(name string) bool {
	if _, err := os.Stat(name); err == nil {
		return true

	} else if os.IsNotExist(err) {
		return false

	} else {
		return false
	}
}

func setup_db() {
	script, err := ioutil.ReadFile(common.DatabaseSetup)
	if err != nil {
		log.Fatal("Error while loading database setup")
		log.Fatal(err)
		os.Exit(1)
	}
	db, err := sql.Open("sqlite3", common.DatabaseFile)
	if err != nil {
		log.Fatal("Error while open database")
		log.Fatal(err)
	}
	defer db.Close()
	_, dberr := db.Exec(string(script))
	if dberr != nil {
		log.Fatal("Error while exec database setup")
		log.Fatal(dberr)
	}
}

func setup_config() {
	cfg := common.Config{
		MQTT: common.MqttConfig{
			Host:         string("tcp://localhost:1883"),
			QueryTimeout: 10,
		},
		Hue: common.HueConfig{
			Paired: false,
			IP:     string("127.0.0.1"),
			User:   string("NULL"),
		},
		HostAddress: string(":9090"),
	}

	common.WriteConfig(&cfg)
}

func setup() {
	if !file_exists("golit.db") {
		log.Print("Database setup... ")
		setup_db()
		log.Print("done\n")
	}
	if !file_exists("config.json") {
		log.Print("Config setup... ")
		setup_config()
		log.Print("done\n Exit now!")
		os.Exit(0)
	}
}

func main() {
	log.Print("Golit v0.6")
	setup()
	cfg := common.LoadConfig()
	http.HandleFunc("/", appview_handler)
	http.HandleFunc("/images/", static_handler)
	http.HandleFunc("/view/", static_handler)
	http.HandleFunc("/mqtt/cmd/", mqtt_cmd_handler)
	http.HandleFunc("/mqtt/stat/", mqtt_stat_handler)
	http.HandleFunc("/add/tasmota", tasmota_add_handler)
	http.HandleFunc("/del/tasmota", tasmota_delete_handler)
	http.HandleFunc("/tasmota_add", add_mqtt_view_handler)
	http.HandleFunc("/hue_setup", hue_setup_view_handler)
	http.HandleFunc("/hue/pair", hue_pairing_handler)
	http.HandleFunc("/hue/scene/", hue_scene_handler)
	http.HandleFunc("/hue/light/", hue_light_handler)
	http.HandleFunc("/off", disable_all_handler)
	http.HandleFunc("/on", enable_all_handler)
	http.HandleFunc("/toggle", toggle_handler)
	log.Printf("Starting http server on %s", cfg.HostAddress)
	log.Fatal(http.ListenAndServe(cfg.HostAddress, nil))
}
