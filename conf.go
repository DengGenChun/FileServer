package main

import (
	"encoding/json"
	"os"
)

var (
	// main.go
	Address           = ":4000"
	TemplateDir       = "./template"
	SuperAdminAccount = "ddd"
	SuperAdminPwd     = "4400"
	AdminPwd          = "4399"

	// dir.go
	FileMainDir = "./file"
	LogMainDir  = "./log"

	// user.go
	AccessCountLimit            int   = 5
	TimeLimitInSecond           int64 = 3
	TimeBoforeRestoreInSecond   int64 = 60
	TimeForUserToExistInSecond  int64 = 180
	TimeForExpiredCheckInSecond int64 = 60
)

func init() {
	// Open configuration file
	file, err := os.Open("./conf.json")
	if err != nil {
		panic(err.Error())
	}
	// Create a new decoder for configuration file
	decoder := json.NewDecoder(file)
	// Struct to store configuration parameter
	cfg := Config{}
	// Decode configuration file
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err.Error())
	}

	Address = cfg.Address
	TemplateDir = cfg.TemplateDir
	AdminPwd = cfg.AdminPwd
	SuperAdminAccount = cfg.SuperAdminAccount
	SuperAdminPwd = cfg.SuperAdminPwd
	FileMainDir = cfg.FileMainDir
	LogMainDir = cfg.LogMainDir
	AccessCountLimit = cfg.AccessCountLimit
	TimeLimitInSecond = int64(cfg.TimeLimitInSecond)
	TimeBoforeRestoreInSecond = int64(cfg.TimeBoforeRestoreInSecond)
	TimeForUserToExistInSecond = int64(cfg.TimeForUserToExistInSecond)
	TimeForExpiredCheckInSecond = int64(cfg.TimeForExpiredCheckInSecond)
}

// Config struct used for decode configuration file
type Config struct {
	Address           string
	TemplateDir       string
	AdminPwd          string
	SuperAdminAccount string
	SuperAdminPwd     string

	FileMainDir string
	LogMainDir  string

	TimeLimitInSecond           int
	AccessCountLimit            int
	TimeBoforeRestoreInSecond   int
	TimeForUserToExistInSecond  int
	TimeForExpiredCheckInSecond int
}
