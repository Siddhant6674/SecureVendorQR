package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Siddhant6674/vendorQr/cmd/api"
	"github.com/Siddhant6674/vendorQr/config"
	db "github.com/Siddhant6674/vendorQr/database"
	"github.com/Siddhant6674/vendorQr/utils"
	"github.com/go-sql-driver/mysql"
)

func main() {

	utils.InitRedis()
	db, err := db.NewMySQLStorage(mysql.Config{
		User:   config.Envs.DBUser,
		Passwd: config.Envs.DBPassword,
		Addr:   config.Envs.DBAddress,
		DBName: config.Envs.DBName,
		Net:    "tcp",

		AllowNativePasswords: true,
		ParseTime:            true,
	})
	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)
	server := api.NewAPIserver(fmt.Sprintf("%s:%s", config.Envs.PublicHost, config.Envs.Port), db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected")
}
