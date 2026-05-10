package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type MessagePusherConfig struct {
	Address string
	Token   string
}

type request struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	Channel     string `json:"channel"`
	Token       string `json:"token"`
}

type response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MessageContent struct {
	Title       string
	Description string
	Content     string
}

func SendMessagePusher(cfg MessagePusherConfig, msg MessageContent) error {
	if cfg.Address == "" {
		return errors.New("message pusher address is not set")
	}
	req := request{
		Title:       msg.Title,
		Description: msg.Description,
		Content:     msg.Content,
		Token:       cfg.Token,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := http.Post(cfg.Address,
		"application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	var res response
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.Message)
	}
	return nil
}
