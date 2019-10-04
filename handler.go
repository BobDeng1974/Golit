package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sirjson/golit/common"
	"sirjson/golit/hue"
	"sirjson/golit/tasmota"
	"strings"

	"github.com/amimof/huego"
)

type AppViewState struct {
	Tasmota []tasmota.Device
	Hue     hue.Device
}

func disable_all() {
	cfg := common.LoadConfig()
	state := fetch_state(&cfg)
	for _, d := range state.Tasmota {
		var cmd []string
		cmd = append(cmd, d.Feed)
		cmd = append(cmd, "power")
		cmd = append(cmd, "off")
		tasmota.Command(cfg.MQTTHost, cmd)
	}
	huecfg := hue.LoadConfig()
	bridge := huego.New(huecfg.IP, huecfg.User)
	for _, d := range state.Hue.Lights {
		hue.LightOff(d.UniqueID, bridge)
	}
}

func enable_all() {
	cfg := common.LoadConfig()
	state := fetch_state(&cfg)
	for _, d := range state.Tasmota {
		var cmd []string
		cmd = append(cmd, d.Feed)
		cmd = append(cmd, "power")
		cmd = append(cmd, "on")
		tasmota.Command(cfg.MQTTHost, cmd)
	}
	huecfg := hue.LoadConfig()
	bridge := huego.New(huecfg.IP, huecfg.User)
	for _, d := range state.Hue.Lights {
		hue.LightOn(d.UniqueID, bridge)
	}
}

func toggle_all() {
	cfg := common.LoadConfig()
	state := fetch_state(&cfg)
	is_active := false
	for _, d := range state.Tasmota {
		if d.Status.Power == 1 {
			is_active = true
			break
		}
	}
	if is_active == false {
		for _, d := range state.Hue.Lights {
			if d.State.On {
				is_active = true
				break
			}
		}
	}
	if is_active {
		disable_all()
	} else {
		enable_all()
	}
}

func disable_all_handler(w http.ResponseWriter, req *http.Request) {
	disable_all()
	fmt.Fprintf(w, JSONResult("OK"))
}

func enable_all_handler(w http.ResponseWriter, req *http.Request) {
	enable_all()
	fmt.Fprintf(w, JSONResult("OK"))
}

func toggle_handler(w http.ResponseWriter, req *http.Request) {
	toggle_all()
	fmt.Fprintf(w, JSONResult("OK"))
}

func appview_handler(w http.ResponseWriter, r *http.Request) {
	cfg := common.LoadConfig()
	state := fetch_state(&cfg)
	Template(w, "view/app.html", state)
}

func fetch_state(cfg *common.Config) *AppViewState {
	state := &AppViewState{Tasmota: []tasmota.Device{}, Hue: hue.Device{}}
	state.Tasmota = tasmota.Fetch()
	state.Hue.Config = hue.LoadConfig()

	for i, _ := range state.Tasmota {
		responseData := tasmota.GetInfo(cfg.MQTTHost, state.Tasmota[i].Feed, "Status", false)
		status, jerr := tasmota.UnmarshalStatus(responseData)
		if jerr != nil {
			log.Print("Unmarshal error", jerr.Error())
			return state
		}
		state.Tasmota[i].Status = status.Status

		responseData = tasmota.GetInfo(cfg.MQTTHost, state.Tasmota[i].Feed, "Color", true)
		colorState, jerr := tasmota.UnmarshalColor(responseData)
		if jerr != nil {
			log.Print("Unmarshal error", jerr.Error())
			return state
		}
		if len(colorState.Color) > 0 {
			state.Tasmota[i].Color = colorState.Color[:len(colorState.Color)-2]
			state.Tasmota[i].White = colorState.Color[len(colorState.Color)-2:]
		}
	}

	if state.Hue.Config.Paired {
		bridge := huego.New(state.Hue.Config.IP, state.Hue.Config.User)
		lights, err := bridge.GetLights()
		if err != nil {
			log.Print("Hue error", err.Error())
			state.Hue.Config.Paired = false
			hue.UpdateConfig(&state.Hue.Config)
			return state
		}
		state.Hue.Lights = lights

		scenes, err := bridge.GetScenes()
		if err != nil {
			log.Print("Hue error", err.Error())
			state.Hue.Config.Paired = false
			hue.UpdateConfig(&state.Hue.Config)
			return state
		}
		state.Hue.Scenes = scenes

		groups, err := bridge.GetGroups()
		if err != nil {
			log.Print("Hue error", err.Error())
			state.Hue.Config.Paired = false
			hue.UpdateConfig(&state.Hue.Config)
			return state
		}
		state.Hue.Groups = groups
	}

	return state
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
		log.Print("Hue pairing error ", hueErr.Error())
		fmt.Fprintf(w, JSONResult(hueErr.Error()))
		return
	}
	fmt.Fprintf(w, JSONResult("OK"))
}

func hue_scene_handler(w http.ResponseWriter, r *http.Request) {
	sceneReq := r.URL.Path[len("/hue/scene/"):]
	config := hue.LoadConfig()
	if config.Paired {
		bridge := huego.New(config.IP, config.User)
		hue.SetScene(sceneReq, bridge)
		fmt.Fprintf(w, JSONResult("OK"))
	}
}

func hue_light_handler(w http.ResponseWriter, r *http.Request) {
	lightReq := r.URL.Path[len("/hue/light/"):]
	lightParam := strings.Split(lightReq, "/")
	config := hue.LoadConfig()
	if config.Paired {
		bridge := huego.New(config.IP, config.User)
		switch lightParam[1] {
		case "on":
			hue.LightOn(lightParam[0], bridge)
			break
		case "off":
			hue.LightOff(lightParam[0], bridge)
			break
		case "update":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Print("IO error: ", err.Error())
				fmt.Fprintf(w, JSONResult(err.Error()))
				return
			}

			t, jerr := hue.UnmarshalLightUpdate(body)
			if jerr != nil {
				log.Print("Unmarshal error: ", jerr.Error())
				fmt.Fprintf(w, JSONResult(jerr.Error()))
				return
			}
			uerr := hue.UpdateLight(bridge, lightParam[0], t)
			if uerr != nil {
				log.Print("UpdateLight error: ", uerr.Error())
				fmt.Fprintf(w, JSONResult(uerr.Error()))
				return
			}
			break
		}
		fmt.Fprintf(w, JSONResult("OK"))
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
		fmt.Fprintf(w, JSONResult(err.Error()))
		return
	}
}

func tasmota_delete_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print("IO error", err.Error())
		fmt.Fprintf(w, JSONResult(err.Error()))
		return
	}

	t, jerr := tasmota.UnmarshalDevice(body)
	if jerr != nil {
		log.Print("Unmarshal error", jerr.Error())
		fmt.Fprintf(w, JSONResult(jerr.Error()))
		return
	}

	tasmota.Delete(t.Feed)
	fmt.Fprintf(w, JSONResult("OK"))
}

func tasmota_add_handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Print("Tasmota Error", err.Error())
		fmt.Fprintf(w, JSONResult(err.Error()))
		return
	}
	t, jerr := tasmota.UnmarshalDevice(body)
	if jerr != nil {
		log.Print("Unmarshal error", jerr.Error())
		fmt.Fprintf(w, JSONResult(jerr.Error()))
		return
	}
	tasmota.Add(t)
	fmt.Fprintf(w, JSONResult("OK"))
}

func mqtt_cmd_handler(w http.ResponseWriter, r *http.Request) {
	cmdRequest := r.URL.Path[len("/mqtt/cmd/"):]
	cmdRequest = strings.ReplaceAll(cmdRequest, "*", "#")
	log.Print("Tasmota:", cmdRequest)
	cfg := common.LoadConfig()

	cmds := strings.Split(cmdRequest, "/")
	if len(cmds) < 3 {
		return
	}

	tasmota.Command(cfg.MQTTHost, cmds)
}

func mqtt_stat_handler(w http.ResponseWriter, r *http.Request) {
	cmdRequest := r.URL.Path[len("/mqtt/stat/"):]
	log.Print("Tasmota:", cmdRequest)
	cmds := strings.Split(cmdRequest, "/")
	if len(cmds) < 2 {
		log.Print("not enough args")
		return
	}
	cfg := common.LoadConfig()
	WriteByteResponse(w, tasmota.GetInfo(cfg.MQTTHost, cmds[0], cmds[1], false))
}
