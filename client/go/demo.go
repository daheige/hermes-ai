package main

import (
	"context"
	"fmt"
	"log"

	"github.com/voocel/litellm"
)

func main() {
	client, err := litellm.NewWithProvider("qwen", litellm.ProviderConfig{
		// APIKey: os.Getenv("OPENAI_API_KEY"),
		APIKey:  "sk-xxx",                   // 虚拟apikey或真实的apikey
		BaseURL: "http://localhost:1337/v1", // 网关地址或真实的大模型provider base_url地址
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Chat(context.Background(), &litellm.Request{
		Model:    "deepseek-chat",
		Messages: []litellm.Message{litellm.UserMessage("go语言是什么")},
	})
	if err != nil {
		log.Fatalln("request err:", err)
	}

	fmt.Println(resp.Content)
}
