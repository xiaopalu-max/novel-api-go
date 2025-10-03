package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"novel-api/api"
	"novel-api/config"
	"novel-api/logs"

	"gopkg.in/yaml.v2"
)

var cfg config.Config

func main() {
	data, err := ioutil.ReadFile(".env")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Config loaded successfully")
	fmt.Printf("Translation config: Enable=%v, URL=%s, Model=%s\n", cfg.Translation.Enable, cfg.Translation.URL, cfg.Translation.Model)

	// 初始化日志系统
	if err := logs.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	fmt.Println("Logger initialized successfully")
	defer logs.Close()

	// 启动路由 - API路由
	http.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		api.Completions(w, r, &cfg)
	})
	http.HandleFunc("/v1/images/generations", func(w http.ResponseWriter, r *http.Request) {
		api.Generations(w, r, &cfg)
	})

	// 日志管理API路由
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		api.Login(w, r, &cfg)
	})
	http.HandleFunc("/api/logs", api.QueryLogs)
	http.HandleFunc("/api/logs/detail", api.GetLogDetail)

	// 前端页面路由
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/logs.html")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/logs.html")
	})

	log.Println("Starting server on : ", cfg.Server.Addr)
	log.Println("日志查询页面: http://localhost:" + cfg.Server.Addr + "/logs")
	log.Println("默认管理密码: " + cfg.LogsAdmin.Password)

	if err := http.ListenAndServe(":"+cfg.Server.Addr, nil); err != nil {
		log.Fatal(err)
	}

}
