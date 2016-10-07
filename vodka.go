package session

import (
	"encoding/gob"
	"errors"
	"log"
	"net/url"

	"github.com/insionng/vodka"
)

var GlobalSessions *Manager

var defaultOtions = Options{"memory", `{"cookieName":"gosessionid","gclifetime":3600}`}

//var defaultOtions = Options{"file", `{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"./data/session"}`}

//var defaultOtions = Options{"redis", `{"cookieName":"gosessionid","gclifetime":3600,"ProviderConfig":"127.0.0.1:6379"}`}

const (
	CONTEXT_SESSION_KEY = "_SESSION_STORE"
	COOKIE_FLASH_KEY    = "_COOKIE_FLASH"
	CONTEXT_FLASH_KEY   = "_FLASH_VALUE"
	SESSION_FLASH_KEY   = "_SESSION_FLASH"
	SESSION_INPUT_KEY   = "_SESSION_INPUT"
)

type Options struct {
	Provider string
	Config   string
}

/*
type managerConfig struct {
	CookieName      string `json:"cookieName"`
	EnableSetCookie bool   `json:"enableSetCookie,omitempty"`
	Gclifetime      int64  `json:"gclifetime"`
	Maxlifetime     int64  `json:"maxLifetime"`
	Secure          bool   `json:"secure"`
	CookieLifeTime  int    `json:"cookieLifeTime"`
	ProviderConfig  string `json:"providerConfig"`
	Domain          string `json:"domain"`
	SessionIDLength int64  `json:"sessionIDLength"`
}
*/

func init() {
	gob.Register(url.Values{})
}

func Setup(op ...Options) error {
	option := defaultOtions
	if len(op) > 0 {
		option = op[0]
	}

	if len(option.Provider) == 0 {
		option.Provider = defaultOtions.Provider
		option.Config = defaultOtions.Config
	}

	// if len(option.Config) == 0 {
	// 	option.Config = defaultOtions.Config
	// }

	log.Println("session config ", option)

	var err error
	GlobalSessions, err = NewManager(option.Provider, option.Config)
	if err != nil {
		return err
	}
	go GlobalSessions.GC()

	return nil
}

func Sessioner() vodka.MiddlewareFunc {
	return func(next vodka.HandlerFunc) vodka.HandlerFunc {
		return func(c vodka.Context) error {
			if GlobalSessions == nil {

				return errors.New("session manager not found, use session middleware but not init ?")
			}

			sess, err := GlobalSessions.SessionStart(c.Response(), c.Request())
			if err != nil {
				return err
			}

			c.Set(CONTEXT_FLASH_KEY, Flash{})

			flashVals := url.Values{}

			flashIf := sess.Get(SESSION_FLASH_KEY)
			if flashIf != nil {
				vals, _ := url.QueryUnescape(flashIf.(string))
				flashVals, _ = url.ParseQuery(vals)
				if len(flashVals) > 0 {
					flash := Flash{}
					flash.ErrorMsg = flashVals.Get("error")
					flash.WarningMsg = flashVals.Get("warning")
					flash.InfoMsg = flashVals.Get("info")
					flash.SuccessMsg = flashVals.Get("success")
					// c.SetData("FLASH", flash)
					// vodka.v2没有直接分配变量到模板这个方法，flash先存到context里
					c.Set(CONTEXT_FLASH_KEY, flash)

				}
			}

			f := NewFlash()

			sess.Set(SESSION_FLASH_KEY, f)

			c.Set(CONTEXT_SESSION_KEY, sess)

			defer func() {
				log.Println("save session", sess)
				sess.Set(SESSION_FLASH_KEY, url.QueryEscape(f.Encode()))
				sess.SessionRelease(c.Response())
			}()

			return next(c)
		}
	}
}

func GetStore(c vodka.Context) Store {
	store := c.Get(CONTEXT_SESSION_KEY)
	if store != nil {
		return store.(Store)
	}

	return nil
}

func GetFlash(c vodka.Context) *Flash {
	return GetStore(c).Get(SESSION_FLASH_KEY).(*Flash)
}

func FlashValue(c vodka.Context) Flash {
	return c.Get(CONTEXT_FLASH_KEY).(Flash)
}

func SaveInput(c vodka.Context) {
	GetStore(c).Set(SESSION_INPUT_KEY, url.Values(c.FormParams()))
}

func GetInput(c vodka.Context) url.Values {
	input := GetStore(c).Get(SESSION_INPUT_KEY)
	if input != nil {
		return input.(url.Values)
	}

	return url.Values{}
}

func CleanInput(c vodka.Context) {
	GetStore(c).Set(SESSION_INPUT_KEY, url.Values{})
}

func NewFlash() *Flash {
	return &Flash{url.Values{}, "", "", "", ""}
}

type Flash struct {
	url.Values
	ErrorMsg, WarningMsg, InfoMsg, SuccessMsg string
}

func (f *Flash) set(name, msg string) {
	f.Set(name, msg)
}

func (f *Flash) Error(msg string) {
	f.ErrorMsg = msg
	f.set("error", msg)
}

func (f *Flash) Warning(msg string) {
	f.WarningMsg = msg
	f.set("warning", msg)
}

func (f *Flash) Info(msg string) {
	f.InfoMsg = msg
	f.set("info", msg)
}

func (f *Flash) Success(msg string) {
	f.SuccessMsg = msg
	f.set("success", msg)
}
