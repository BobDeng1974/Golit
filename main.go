package main

import (
	"encoding/json"
	"fmt"
	"github.com/amimof/huego"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type StaticFile struct {
	Data []byte
}

type AppViewState struct {
	Tasmota []TasmotaDevice
	Hue     HueDevice
}

type ConfigEntry struct {
	Key   string
	Value string
}

func load_static(path string) (*StaticFile, error) {
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &StaticFile{Data: buffer}, nil
}

func appview_handler(w http.ResponseWriter, r *http.Request) {
	state := &AppViewState{Tasmota: []TasmotaDevice{}, Hue: HueDevice{}}
	state.Tasmota = tasmota_fetch()
	state.Hue.Config = hue_load_config()
	for i, _ := range state.Tasmota {
		responseData := tasmota_stat(state.Tasmota[i].Feed, "Status")
		var status TasmotaStatusResponse
		err := json.Unmarshal(responseData, &status)
		if err != nil {
			println(err.Error())
			_, err := fmt.Fprintf(w, "{\"result\": \"%s\"}", err.Error())
			if err != nil {
				println("Error", err)
			}
			return
		}
		state.Tasmota[i].Status = status.Status

		responseData = tasmota_stat_result(state.Tasmota[i].Feed, "Color")
		var colorState ColorState
		err = json.Unmarshal(responseData, &colorState)
		if err != nil {
			println(err.Error())
			_, err := fmt.Fprintf(w, "{\"result\": \"%s\"}", err.Error())
			if err != nil {
				println("Error", err)
			}
			return
		}
		state.Tasmota[i].Color = colorState.Color[:len(colorState.Color)-2]
		state.Tasmota[i].White = colorState.Color[len(colorState.Color)-2:]
	}

	if state.Hue.Config.Paired {
		bridge, _ := huego.Discover()
		bridge = bridge.Login(state.Hue.Config.User)
		lights, err := bridge.GetLights()
		if err != nil {
			println("Hue error", err.Error())
		}
		state.Hue.Lights = lights

		scenes, err := bridge.GetScenes()
		if err != nil {
			println("Hue error", err.Error())
		}
		state.Hue.Scenes = scenes
	}

	page, _ := template.ParseFiles("view/app.html")
	page.Execute(w, state)
}

func add_mqtt_view_handler(w http.ResponseWriter, r *http.Request) {
	state := &AppViewState{Tasmota: []TasmotaDevice{}}
	state.Tasmota = tasmota_fetch()

	page, _ := template.ParseFiles("view/addmqtt.html")
	page.Execute(w, state)
}

func hue_setup_view_handler(w http.ResponseWriter, r *http.Request) {
	state := &AppViewState{}
	state.Hue.Config = hue_load_config()
	page, _ := template.ParseFiles("view/huesetup.html")
	page.Execute(w, state)
}

func hue_pairing_handler(w http.ResponseWriter, r *http.Request) {
	hue_pair()
	_, err := fmt.Fprintf(w, "{\"result\": \"OK\"}")
	if err != nil {
		println("Error", err.Error())
	}
}

func hue_scene_handler(w http.ResponseWriter, r *http.Request) {
	sceneReq := r.URL.Path[len("/hue/scene/"):]
	config := hue_load_config()
	if config.Paired {
		bridge, _ := huego.Discover()
		bridge = bridge.Login(config.User)
		hue_set_scene(sceneReq, bridge)
		_, err := fmt.Fprintf(w, "{\"result\": \"OK\"}")
		if err != nil {
			println("Error", err.Error())
		}
	}
}

func hue_light_handler(w http.ResponseWriter, r *http.Request) {
	lightReq := r.URL.Path[len("/hue/light/"):]
	lightParam := strings.Split(lightReq, "/")
	config := hue_load_config()
	if config.Paired {
		bridge, _ := huego.Discover()
		bridge = bridge.Login(config.User)
		if lightParam[1] == "on" {
			hue_light_on(lightParam[0], bridge)
		} else {
			hue_light_off(lightParam[0], bridge)
		}
		_, err := fmt.Fprintf(w, "{\"result\": \"OK\"}")
		if err != nil {
			println("Error", err.Error())
		}
	}
}

func static_handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	println("Serving static file:", path)
	ext := strings.Split(path, ".")[1]

	file, _ := load_static(path)

	switch ext {
	case "css":
		w.Header().Set("Content-Type", "text/css")
		break
	case "js":
		w.Header().Set("Content-Type", "application/javascript")
		break
	case "png":
		w.Header().Set("Content-Type", "image/png")
		break
	case "jpg":
	case "jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
		break
	}
	_, err := w.Write(file.Data)
	if err != nil {
		println("Error", err)
	}
}

func tasmota_delete_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		println(err.Error())
		_, err := fmt.Fprintf(w, "{\"result\": \"%s\"}", err.Error())
		if err != nil {
			println("Error", err)
		}
		return
	}
	println(string(body))
	var t TasmotaDevice
	err = json.Unmarshal(body, &t)
	if err != nil {
		println(err.Error())
		_, err := fmt.Fprintf(w, "{\"result\": \"%s\"}", err.Error())
		if err != nil {
			println("Error", err)
		}
		return
	}
	tasmota_delete(t.Feed)
	_, err = fmt.Fprintf(w, "{\"result\": \"OK\"}")
	if err != nil {
		println("Error", err)
	}
}

func tasmota_add_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		println(err.Error())
		_, err := fmt.Fprintf(w, "{\"result\": \"%s\"}", err.Error())
		if err != nil {
			println("Error", err)
		}
		return
	}
	println(string(body))
	var t TasmotaDevice
	err = json.Unmarshal(body, &t)
	if err != nil {
		println(err.Error())
		_, err := fmt.Fprintf(w, "{\"result\": \"%s\"}", err.Error())
		if err != nil {
			println("Error", err)
		}
		return
	}
	tasmota_add(t)
	_, err = fmt.Fprintf(w, "{\"result\": \"OK\"}")
	if err != nil {
		println("Error", err.Error())
	}
}

func mqtt_cmd_handler(w http.ResponseWriter, r *http.Request) {
	cmdRequest := r.URL.Path[len("/mqtt/cmd/"):]
	cmdRequest = strings.ReplaceAll(cmdRequest, "*", "#")
	println("Tasmota:", cmdRequest)

	cmds := strings.Split(cmdRequest, "/")
	if len(cmds) < 3 {
		return
	}

	for _, element := range cmds {
		println("\t", element)
	}
	opts := mqtt.NewClientOptions().AddBroker("tcp://moi:1883")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		println(token.Error())
		return
	}

	if token := client.Publish(fmt.Sprintf("cmnd/%s/%s", cmds[0], cmds[1]), 0, false, cmds[2]); token.Wait() && token.Error() != nil {
		println(token.Error())
		return
	}
	client.Disconnect(100)
}

func mqtt_stat_handler(w http.ResponseWriter, r *http.Request) {
	cmdRequest := r.URL.Path[len("/mqtt/stat/"):]
	println("Tasmota:", cmdRequest)
	cmds := strings.Split(cmdRequest, "/")
	if len(cmds) < 2 {
		println("not enough args")
		return
	}
	_, err := w.Write(tasmota_stat(cmds[0], cmds[1]))
	if err != nil {
		println("Error", err.Error())
	}
}

func main() {
	println("Huelishous v0.1")
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
