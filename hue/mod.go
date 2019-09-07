package hue

import (
	"database/sql"
	"errors"
	"github.com/amimof/huego"
	"log"
)

type Device struct {
	Lights []huego.Light
	Scenes []huego.Scene
	Config Config
}

type Config struct {
	IP     string
	User   string
	Paired bool
}

func Pair() error {
	cfg := LoadConfig()
	bridge, _ := huego.Discover()
	user, _ := bridge.CreateUser("Golit") // Link button needs to be pressed

	cfg.Paired = true
	cfg.User = user
	cfg.IP = bridge.Host

	UpdateConfig(&cfg)
	if len(cfg.IP) > 0 && len(cfg.User) > 0 {
		log.Print("Hue paired!", cfg.User, cfg.IP)
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
	db, err := sql.Open("sqlite3", "./huelishous.db")
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
	db, err := sql.Open("sqlite3", "./huelishous.db")
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
