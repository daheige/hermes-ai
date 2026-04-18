# curl请求方式
```shell
curl http://127.0.0.1:1337/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxx" \
  -d '{
        "model": "deepseek-chat",
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "go语言是什么"}
        ],
        "stream": false
      }'
```

# go sdk
https://github.com/voocel/litellm
