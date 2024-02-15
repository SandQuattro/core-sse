package main

import (
	"flag"
	"github.com/pkoukk/tiktoken-go"
	"log"
	"os"
	"sse-demo-core/internal/app/structs"
	"sse-demo-core/internal/config"
	"sse-demo-core/internal/db"
	"strings"
)

func main() {
	confFile := flag.String("config", "conf/application.conf", "-config=<config file name>")
	flag.Parse()

	// и подгрузим конфиг
	config.MustConfig(confFile)
	conf := config.GetConfig()

	// Коннектимся к базе
	dbPass := os.Getenv("PGPASS")
	if dbPass == "" {
		log.Fatal("db password is empty")
	}

	d := db.Connect(conf, dbPass)
	defer d.Close()
	log.Println(">> DATABASE CONNECTION SUCCESSFUL")

	var messages []structs.ThreadMessage
	err := d.Select(&messages, `select tm.id,
										 tm.object,
										 EXTRACT(epoch FROM tm.created_at)::bigint as created_at,
										 tm.thread_id,
										 tm.role,
										 tm.prompt,
										 tm.assistant_id,
										 tm.run_id,
										 tm.content::json,
									     tm.file_ids::json,
										 tm.hidden,
										 tm.metadata
								from messages tm`)
	if err != nil {
		log.Panic(">> Ошибка поиска сообщений")
	}

	codec, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		log.Panic(err)
	}
	for _, msg := range messages {
		p := sanitize(msg.Prompt)

		tokens := codec.Encode(p, nil, nil)
		_, err = d.Exec(`update messages set tokens = $1 where id = $2`,
			len(tokens),
			msg.ID)
		if err != nil {
			log.Panic("Ошибка обновления токенов сообщения, ", err.Error())
		}
	}
}

func sanitize(s string) string {
	res := strings.ReplaceAll(s, "```python\n", "")
	res = strings.ReplaceAll(res, "```", "")
	return res
}
