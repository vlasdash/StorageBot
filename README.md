# StorageBot

1. Получить url с помощью ngrok

```
ngrok http 8080
```

2. Добавить url в `config.yaml` по ключу `webhook_url` 
3. Добавить в `env.example` `BOT_TOKEN`
4. Запустить
```
docker-compose up
```
