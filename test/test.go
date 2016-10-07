package main

import (
	"log"
	"net/http"

	"github.com/insionng/vodka"
	"github.com/insionng/vodka/engine/fasthttp"
	"github.com/vodka-contrib/session"
	//_ "github.com/vodka-contrib/session/redis"
	m "github.com/insionng/vodka/middleware"
)

func main() {
	if err := session.InitSession(session.Options{"file", `{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"./data/session"}`}); err != nil {
		log.Fatalln("session errors:", err)
	}

	v := vodka.New()
	v.Use(m.Recover())
	v.Use(m.Gzip())
	v.Use(session.Sessioner())

	v.GET("/get", func(self vodka.Context) error {
		sess := session.GetStore(self)

		value := "nil"
		valueIf := sess.Get("key")
		if valueIf != nil {
			value = valueIf.(string)
		}

		return self.String(http.StatusOK, value)

	})

	v.GET("/set", func(self vodka.Context) error {
		sess := session.GetStore(self)

		val := self.QueryParam("v")
		if len(val) == 0 {
			val = "value"
		}

		err := sess.Set("key", val)
		if err != nil {
			log.Printf("sess.set %v \n", err)
		}
		return self.String(http.StatusOK, "ok")
	})

	v.Run(fasthttp.New(":8080"))
}
