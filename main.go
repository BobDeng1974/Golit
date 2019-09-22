package main

import (
	"log"
	"net/http"
)

func main() {
	log.Print("Golit v0.3")
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
