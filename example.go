package main

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/basicauth"
	"os"
	"regexp"
	"time"
)

func newApp() *iris.Application {

	app := iris.New()
	app.StaticEmbedded("/public", "./public", Asset, AssetNames)

	authConfig := basicauth.Config{
		Users:   map[string]string{"wtq": "wtq", "www": "www"},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(30) * time.Minute,
	}

	authentication := basicauth.New(authConfig)
	app.RegisterView(iris.HTML("./templates", ".html").Binary(Asset, AssetNames))
	app.Get("/", func(ctx iris.Context) { ctx.Redirect("/admin") })

	// to party
	needAuth := app.Party("/admin", authentication)
	{
		//http://localhost:8080/admin
		needAuth.Get("/", func(ctx iris.Context) {
			//ctx.ViewData("Name", "iris") // the .Name inside the ./templates/hi.html
			ctx.Gzip(true) // enable gzip for big files
			ctx.View("onepage.html")
		})
		needAuth.Get("/updatetime", func(ctx iris.Context) {
			c, err := redis.Dial("tcp", "localhost:6379")
			errCheck(err)
			defer c.Close()
			val, err := redis.String(c.Do("get", "time"))
			errCheck(err)
			ctx.Writef(string(val))
		})
		needAuth.Get("/data", func(ctx iris.Context) {
			c, err := redis.Dial("tcp", "localhost:6379")
			errCheck(err)
			defer c.Close()
			var eq_slice []string
			var eq_map map[string]map[string]string
			eq_map = make(map[string]map[string]string)

			key_ls, getErr := redis.Strings(c.Do("KEYS", "*"))
			//fmt.Println(key_ls)
			errCheck(getErr)
			re_valid_keys, _ := regexp.Compile("^[a-zA-Z]+\\d+[a-zA-Z]+$")
			re_valid_eqs, _ := regexp.Compile("^[a-zA-Z]+\\d+")
			re_item, _ := regexp.Compile("[a-zA-Z]+$")
			for _, element := range key_ls {
				if re_valid_keys.MatchString(element) {
					// 处理设备列表
					eq_slice = append(eq_slice, element)
					// 处理设备映射
					tmpkey := re_valid_eqs.FindString(element)
					tmpitem := re_item.FindString(element)
					//fmt.Println("aaa",tmpkey)
					if _, exist := eq_map[tmpkey]; exist {
						val, _ := redis.String(c.Do("get", element))
						eq_map[tmpkey][tmpitem] = val
					} else {
						val, _ := redis.String(c.Do("get", element))
						eq_map[tmpkey] = map[string]string{tmpitem: val}
					}
				}
				//fmt.Println(index, element)
			}
			mapB, _ := json.Marshal(eq_map)
			ctx.Writef(string(mapB))
		})

		needAuth.Get("/set/{key}/{value}", func(ctx iris.Context) {
			key, value := ctx.Params().Get("key"), ctx.Params().Get("value")
			c, err := redis.Dial("tcp", "localhost:6379")
			errCheck(err)
			defer c.Close()

			val, err := redis.String(c.Do("SET", key, value))
			errCheck(err)
			// test if setted here
			ctx.Writef("All ok with the key: '%s' and val is: '%s'", key, val)
		})

		needAuth.Get("/get/{key}", func(ctx iris.Context) {
			key := ctx.Params().Get("key")
			c, err := redis.Dial("tcp", "localhost:6379")
			errCheck(err)
			defer c.Close()

			val, err := redis.String(c.Do("GET", key))
			errCheck(err)
			// test if setted here
			ctx.Writef("The '%s' on the /get was: %v", key, val)

		})

		needAuth.Get("/delete/{key}", func(ctx iris.Context) {
			// delete a specific key
			key := ctx.Params().Get("key")
			c, err := redis.Dial("tcp", "localhost:6379")
			errCheck(err)
			defer c.Close()

			_, err = c.Do("DEL", key)
			errCheck(err)
			// test if setted here
			ctx.Writef("The '%s' on the /delete was deleted: ", key)
		})

		needAuth.Get("/clear", func(ctx iris.Context) {
			// removes all entries
			c, err := redis.Dial("tcp", "localhost:6379")
			errCheck(err)
			defer c.Close()
			val, err := c.Do("FLUSHDB")

			errCheck(err)
			// test if setted here
			ctx.Writef("all keys were deleted!  %v:", val)
		})
	}
	return app
}

func main() {

	app := newApp()
	//app.StaticWeb("/public", "./public")
	app.Run(iris.Addr(":80"))
}

func errCheck(err error) {
	if err != nil {
		fmt.Println("sorry,has some error:", err)
		os.Exit(-1)
	}
}
