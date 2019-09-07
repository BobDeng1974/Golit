package hue

import (
	"database/sql"
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
	Config Config
}

type LightUpdate struct {
	Ct  uint16
	Bri uint8
}

type Config struct {
	IP     string
	User   string
	Paired bool
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
	cfg := LoadConfig()
	bridge, _ := huego.Discover()

	user, _ := bridge.CreateUser("Golit_" + name) // Link button needs to be pressed

	cfg.Paired = true
	cfg.User = user
	cfg.IP = bridge.Host

	UpdateConfig(&cfg)
	if len(cfg.IP) > 0 && len(cfg.User) > 0 {
		log.Print("Hue paired! User", cfg.User, " Host: ", cfg.IP)
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

func UpdateConfig(cfg *Config) {
	update_config_entry("hue_ip", cfg.IP)
	update_config_entry("hue_user", cfg.User)
	update_config_entry("hue_registered", bool_str(cfg.Paired))
}

func update_config_entry(key string, value string) {
	db, err := sql.Open("sqlite3", common.DatabaseFile)
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Print(err.Error())
	}
	stmt, err := tx.Prepare("UPDATE config SET Value = ? WHERE Key = ?")
	if err != nil {
		log.Print(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(value, key)
	err = tx.Commit()
	if err != nil {
		log.Print(err.Error())
	}
}

func LoadConfig() Config {
	cfg := Config{}
	db, err := sql.Open("sqlite3", common.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT Key, Value FROM config WHERE Key LIKE('hue_%')")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		var key string

		err = rows.Scan(&key, &value)
		if err != nil {
			log.Fatal(err)
		}
		switch key {
		case "hue_user":
			cfg.User = value
			break
		case "hue_ip":
			cfg.IP = value
			break
		case "hue_registered":
			cfg.Paired = value == "1"
			break
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
