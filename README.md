## Telegram bot to subscribe RSS feeds

### Usage
#### Compiling
```
go get
go build -o app

```

#### Initializing the sqlite database
``` 
cat schema.sql | sqlite3 main.db
```

#### Running
```
export TELEGRAM_BOT_TOKEN="SOMETHING_SOMETHINGELSE"
./app main.db
```
