package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"novel-api/api"
	"novel-api/config"

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
	// 示例使用全局变量
	// fmt.Println("Translation URL:", config.Translation.URL)
	// fmt.Println("Nkey Path:", config.Nkey.Path)
	// fmt.Println("Parameters Width:", config.Parameters.Width)

	// 启动路由

	http.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		api.Completions(w, r, &cfg)
	}) // 修改了路由

	// http.HandleFunc("/v1/images/generations", api.Completions) // 修改了路由
	// http.HandleFunc("/tokens/upload", api.HandleUploadTokens)
	// http.HandleFunc("/tokens/count", api.HandleGetAvailableTokensCount)
	// http.HandleFunc("/tokens", api.HandleClearTokens)           // 使用 DELETE 方法清空
	// http.HandleFunc("/tokens/errors", api.HandleGetErrorTokens) // 如果你实现了这个接口
	// http.HandleFunc("/web/", api.WebCheck)                      // 前端页面

	log.Println("Starting server on : ", cfg.Server.Addr)

	if err := http.ListenAndServe(":"+cfg.Server.Addr, nil); err != nil {
		log.Fatal(err)
	}

}
