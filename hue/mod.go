package hue

import (
	"encoding/json"
	"errors"
	"github.com/amimof/huego"
	"log"
	"os"
	"sirjson/golit/common"
)

type Device struct {
	Lights []huego.Light
	Scenes []huego.Scene
	Groups []huego.Group
}

type LightUpdate struct {
	Ct  uint16
	Bri uint8
}

func UnmarshalLightUpdate(data []byte) (*LightUpdate, error) {
	var t LightUpdate
	if len(data) == 0 {
		log.Print("LightUpdate unmarshal failed: no data")
		return nil, errors.New("LightUpdate unmarshal failed: no data")
	}
	err := json.Unmarshal(data, &t)
	if err != nil {
		log.Print("JSON", err.Error())
		return nil, errors.New("JSON failure")
	}
	return &t, nil
}

func UpdateLight(bridge *huego.Bridge, uid string, update *LightUpdate) error {
	lights, err := bridge.GetLights()
	if err != nil {
		return err
	}
	for _, e := range lights {
		if e.UniqueID == uid {
			cterr := e.Ct(update.Ct)
			if cterr != nil {
				return cterr
			}
			brierr := e.Bri(update.Bri)
			if cterr != nil {
				return brierr
			}
			break
		}
	}
	return nil
}

func Pair() error {
	name, err := os.Hostname()
	if err != nil {
		return err
	}
	cfg := common.LoadConfig()
	bridge, _ := huego.Discover()

	user, _ := bridge.CreateUser("Golit_" + name) // Link button needs to be pressed

	cfg.Hue.Paired = true
	cfg.Hue.User = user
	cfg.Hue.IP = bridge.Host

	common.WriteConfig(&cfg)
	if len(cfg.Hue.IP) > 0 && len(cfg.Hue.User) > 0 {
		log.Print("Hue paired! User", cfg.Hue.User, " Host: ", cfg.Hue.IP)
		return nil
	} else {
		return errors.New("Pairing failed")
	}
}

func SetScene(scene string, bridge *huego.Bridge) {
	_, err := bridge.RecallScene(scene, 0)
	if err != nil {
		log.Print("Hue error", err.Error())
	}
}

func LightOn(luid string, bridge *huego.Bridge) {
	lights, err := bridge.GetLights()
	for _, elem := range lights {
		if elem.UniqueID == luid {
			err := elem.On()
			if err != nil {
				log.Print("Hue error", err.Error())
			}
			return
		}
	}
	if err != nil {
		log.Print("Hue error", err.Error())
	}
}

func LightOff(luid string, bridge *huego.Bridge) {
	lights, err := bridge.GetLights()
	for _, elem := range lights {
		if elem.UniqueID == luid {
			err := elem.Off()
			if err != nil {
				log.Print("Hue error", err.Error())
			}
			return
		}
	}
	if err != nil {
		log.Print("Hue error", err.Error())
	}
}

func bool_str(a bool) string {
	if a {
		return "1"
	} else {
		return "0"
	}
}
