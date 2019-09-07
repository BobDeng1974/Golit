package main

import (
	"fmt"
	"github.com/amimof/huego"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"net/http"
	"sirjson/golit/hue"
	"sirjson/golit/tasmota"
	"strings"
)

type AppViewState struct {
	Tasmota []tasmota.Device
	Hue     hue.Device
}

func appview_handler(w http.ResponseWriter, r *http.Request) {
	state := &AppViewState{Tasmota: []tasmota.Device{}, Hue: hue.Device{}}
	state.Tasmota = tasmota.Fetch()
	state.Hue.Config = hue.LoadConfig()
	for i, _ := range state.Tasmota {
		responseData := tasmota.GetInfo(state.Tasmota[i].Feed, "Status", false)
		status, jerr := tasmota.UnmarshalStatus(responseData)
		if jerr != nil {
			log.Print("Unmarshal error", jerr.Error())
			WriteResponse(w, ErrResult(jerr.Error()))
			return
		}
		state.Tasmota[i].Status = status.Status

		responseData = tasmota.GetInfo(state.Tasmota[i].Feed, "Color", true)
		colorState, jerr := tasmota.UnmarshalColor(responseData)
		if jerr != nil {
			log.Print("Unmarshal error", jerr.Error())
			WriteResponse(w, ErrResult(jerr.Error()))
			return
		}
		if len(colorState.Color) > 0 {
			state.Tasmota[i].Color = colorState.Color[:len(colorState.Color)-2]
			state.Tasmota[i].White = colorState.Color[len(colorState.Color)-2:]
		}
	}

	if state.Hue.Config.Paired {
		bridge, _ := huego.Discover()
		bridge = bridge.Login(state.Hue.Config.User)
		lights, err := bridge.GetLights()
		if err != nil {
			log.Print("Hue error", err.Error())
			WriteResponse(w, ErrResult(err.Error()))
			return
		}
		state.Hue.Lights = lights

		scenes, err := bridge.GetScenes()
		if err != nil {
			log.Print("Hue error", err.Error())
			WriteResponse(w, ErrResult(err.Error()))
			return
		}
		state.Hue.Scenes = scenes
	}
	Template(w, "view/app.html", state)
}

func add_mqtt_view_handler(w http.ResponseWriter, r *http.Request) {
	state := &AppViewState{Tasmota: []tasmota.Device{}}
	state.Tasmota = tasmota.Fetch()
	Template(w, "view/addmqtt.html", state)
}

func hue_setup_view_handler(w http.ResponseWriter, r *http.Request) {
	state := &AppViewState{}
	state.Hue.Config = hue.LoadConfig()
	Template(w, "view/huesetup.html", state)
}

func hue_pairing_handler(w http.ResponseWriter, r *http.Request) {
	hueErr := hue.Pair()
	if hueErr != nil {
		log.Print("Hue error", hueErr.Error())
		WriteResponse(w, ErrResult(hueErr.Error()))
		return
	}
	WriteResponse(w, OKResult)
}

func hue_scene_handler(w http.ResponseWriter, r *http.Request) {
	sceneReq := r.URL.Path[len("/hue/scene/"):]
	config := hue.LoadConfig()
	if config.Paired {
		bridge, _ := huego.Discover()
		bridge = bridge.Login(config.User)
		hue.SetScene(sceneReq, bridge)
		WriteResponse(w, OKResult)
	}
}

func hue_light_handler(w http.ResponseWriter, r *http.Request) {
	lightReq := r.URL.Path[len("/hue/light/"):]
	lightParam := strings.Split(lightReq, "/")
	config := hue.LoadConfig()
	if config.Paired {
		bridge, _ := huego.Discover()
		bridge = bridge.Login(config.User)
		if lightParam[1] == "on" {
			hue.LightOn(lightParam[0], bridge)
		} else {
			hue.LightOff(lightParam[0], bridge)
		}
		WriteResponse(w, OKResult)
	}
}

func static_handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	log.Print("Serving static file:", path)
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
		log.Print("IO error", err.Error())
		WriteResponse(w, ErrResult(err.Error()))
		return
	}
}

func tasmota_delete_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print("IO error", err.Error())
		WriteResponse(w, ErrResult(err.Error()))
		return
	}

	t, jerr := tasmota.UnmarshalDevice(body)
	if jerr != nil {
		log.Print("Unmarshal error", jerr.Error())
		WriteResponse(w, ErrResult(jerr.Error()))
		return
	}

	tasmota.Delete(t.Feed)
	WriteResponse(w, OKResult)
}

func tasmota_add_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print("Tasmota Error", err.Error())
		WriteResponse(w, ErrResult(err.Error()))
		return
	}
	t, jerr := tasmota.UnmarshalDevice(body)
	if jerr != nil {
		log.Print("Unmarshal error", jerr.Error())
		WriteResponse(w, ErrResult(jerr.Error()))
		return
	}
	tasmota.Add(t)
	WriteResponse(w, OKResult)
}

func mqtt_cmd_handler(w http.ResponseWriter, r *http.Request) {
	cmdRequest := r.URL.Path[len("/mqtt/cmd/"):]
	cmdRequest = strings.ReplaceAll(cmdRequest, "*", "#")
	log.Print("Tasmota:", cmdRequest)

	cmds := strings.Split(cmdRequest, "/")
	if len(cmds) < 3 {
		return
	}

	for _, element := range cmds {
		log.Print("\t", element)
	}
	opts := mqtt.NewClientOptions().AddBroker("tcp://192.168.178.37:1883")
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

func mqtt_stat_handler(w http.ResponseWriter, r *http.Request) {
	cmdRequest := r.URL.Path[len("/mqtt/stat/"):]
	log.Print("Tasmota:", cmdRequest)
	cmds := strings.Split(cmdRequest, "/")
	if len(cmds) < 2 {
		log.Print("not enough args")
		return
	}
	WriteByteResponse(w, tasmota.GetInfo(cmds[0], cmds[1], false))
}
