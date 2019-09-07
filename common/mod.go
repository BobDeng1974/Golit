package common

import (
	"database/sql"
	"log"
)

type ConfigEntry struct {
	Key   string
	Value string
}

type Config struct {
	MQTTHost string
}

const DatabaseFile string = "./golit.db"

func LoadConfig() Config {
	cfg := Config{}
	db, err := sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		log.Fatal(err)
		return cfg
	}
	defer db.Close()
	rows, err := db.Query("SELECT Key, Value FROM config WHERE Key LIKE('common_%')")
	if err != nil {
		log.Fatal(err)
		return cfg
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		var key string

		err = rows.Scan(&key, &value)
		if err != nil {
			log.Fatal(err)
			return cfg
		}
		switch key {
		case "common_mqtt_host":
			cfg.MQTTHost = value
			break
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
		return cfg
	}
	return cfg
}
