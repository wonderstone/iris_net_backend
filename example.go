package main

import (
	"database/sql" //http://go-database-sql.org/accessing.html
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //install with $ go get -u github.com/go-sql-driver/mysql
	"github.com/gomodule/redigo/redis" // applied differently from "database/sql"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/basicauth"
	"os"
	"regexp"
	"time"
)

const format = "2006-01-02 15:04:05"
const webuser = "www"
const webpsw = "www"

func newApp(db *sql.DB) *iris.Application {
	// redis using conn pool
	pool := newPool()
	// timely record the redis data to mysql
	done := make(chan bool)

	go jobInterval(5, done, db, pool)
	app := iris.New()
	// ######  bindata applied  https://github.com/kataras/iris/blob/master/_examples/file-server/embedding-files-into-app/main.go
	// https://github.com/shuLhan/go-bindata
	// Follow these steps first:
	// $ go get -u github.com/shuLhan/go-bindata/...
	// $ go-bindata ./assets/...
	// $ go build
	app.StaticEmbedded("/public", "./public", Asset, AssetNames)

	// ######  basicauth part   https://github.com/kataras/iris/blob/master/_examples/authentication/basicauth/main.go
	authConfig := basicauth.Config{
		Users:   map[string]string{webuser: webpsw},
		Realm:   "Authorization Required", // defaults to "Authorization Required"
		Expires: time.Duration(30) * time.Minute,
	}

	authentication := basicauth.New(authConfig)
	// embedded part from https://github.com/kataras/iris/tree/master/view
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
			// using pool.Get to give conn
			c := pool.Get()
			defer c.Close()

			val, err := redis.String(c.Do("get", "time"))
			errCheck(err)
			ctx.Writef(string(val))
		})
		needAuth.Get("/data", func(ctx iris.Context) {
			// using pool.Get to give conn
			c := pool.Get()
			defer c.Close()

			var eqSlice []string
			var eqMap map[string]map[string]string
			eqMap = make(map[string]map[string]string)

			keyLs, getErr := redis.Strings(c.Do("KEYS", "*"))
			//fmt.Println(key_ls)
			errCheck(getErr)
			reValidKeys, _ := regexp.Compile("^[a-zA-Z]+\\d+[a-zA-Z]+$")
			reValidEqs, _ := regexp.Compile("^[a-zA-Z]+\\d+")
			reItem, _ := regexp.Compile("[a-zA-Z]+$")
			for _, element := range keyLs {
				if reValidKeys.MatchString(element) {
					// 处理设备列表
					eqSlice = append(eqSlice, element)
					// 处理设备映射
					tmpkey := reValidEqs.FindString(element)
					tmpitem := reItem.FindString(element)
					//fmt.Println("aaa",tmpkey)
					if _, exist := eqMap[tmpkey]; exist {
						val, _ := redis.String(c.Do("get", element))
						eqMap[tmpkey][tmpitem] = val
					} else {
						val, _ := redis.String(c.Do("get", element))
						eqMap[tmpkey] = map[string]string{tmpitem: val}
					}
				}
				//fmt.Println(index, element)
			}
			mapB, _ := json.Marshal(eqMap)
			ctx.Writef(string(mapB))
		})

		needAuth.Get("/tsdata", func(ctx iris.Context) {
			mapB := giveData(db)
			ctx.Writef(string(mapB))
		})

		needAuth.Get("/set/{key}/{value}", func(ctx iris.Context) {
			key, value := ctx.Params().Get("key"), ctx.Params().Get("value")
			// using pool.Get to give conn
			c := pool.Get()
			defer c.Close()

			val, err := redis.String(c.Do("SET", key, value))
			errCheck(err)
			// test if setted here
			ctx.Writef("All ok with the key: '%s' and val is: '%s'", key, val)
		})

		needAuth.Get("/get/{key}", func(ctx iris.Context) {
			key := ctx.Params().Get("key")
			// using pool.Get to give conn
			c := pool.Get()
			defer c.Close()

			val, err := redis.String(c.Do("GET", key))
			errCheck(err)
			// test if setted here
			ctx.Writef("The '%s' on the /get was: %v", key, val)

		})

		needAuth.Get("/delete/{key}", func(ctx iris.Context) {
			// delete a specific key
			key := ctx.Params().Get("key")
			// using pool.Get to give conn
			c := pool.Get()
			defer c.Close()

			_, err := c.Do("DEL", key)
			errCheck(err)
			// test if setted here
			ctx.Writef("The '%s' on the /delete was deleted: ", key)
		})

		needAuth.Get("/clear", func(ctx iris.Context) {
			// removes all entries
			// using pool.Get to give conn
			c := pool.Get()
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
	// ######  use go-sql-driver/mysql @ https://github.com/go-sql-driver/mysql/wiki/Examples
	// the sql.DB object is designed to be long-lived.
	// pass the sql.DB into that short-lived function as an argument.
	db, err := sql.Open("mysql", "root:wonderwhynot@tcp(127.0.0.1:3306)/sensor_data")
	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}
	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	app := newApp(db)
	//app.StaticWeb("/public", "./public")
	app.Run(iris.Addr(":80"))
}

func errCheck(err error) {
	if err != nil {
		fmt.Println("sorry,has some error:", err)
		os.Exit(-1)
	}
}

func queryJob(db *sql.DB, ID string) ([]string, []string, []string) {
	rows, err := db.Query("select date_format(datetime,'%Y-%m-%d %H:%i') as dt,ROUND(AVG(temp),2) as avgt,"+
		"ROUND(AVG(humi),2) as avgh from sensor_table where ID = ? GROUP BY date_format(datetime,'%Y-%m-%d %H:%i') "+
		"ORDER BY dt ASC", ID)
	if err != nil {
		panic(err.Error())
	}

	//获取列名称
	//columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}
	//fmt.Println(columns)

	var recordTimeSlice []string
	var avtSlice []string
	var avhSlice []string

	//遍历输出
	for rows.Next() {
		var recordTime string
		var avt string
		var avh string

		err = rows.Scan(&recordTime, &avt, &avh)
		if err != nil {
			panic(err.Error())
		}
		recordTimeSlice = append(recordTimeSlice, recordTime)
		avtSlice = append(avtSlice, avt)
		avhSlice = append(avhSlice, avh)

		//fmt.Println(recordTime, avt, avh)
		//fmt.Println(recordTimeSlice, avtSlice, avhSlice)
	}
	return recordTimeSlice, avtSlice, avhSlice
}

func giveData(db *sql.DB) []byte {

	// get all distinct eqs by eqSlice
	var eqMap map[string]map[string][]string
	eqMap = make(map[string]map[string][]string)
	//查询
	var IDSlice []string

	rows, err := db.Query("select DISTINCT ID from sensor_table")
	if err != nil {
		panic(err.Error())
	}

	// iter eqSlice to get all avtSlice and avhSlice
	for rows.Next() {
		var ID string
		err = rows.Scan(&ID)
		if err != nil {
			panic(err.Error())
		}
		IDSlice = append(IDSlice, ID)
		if _, exist := eqMap[ID]; exist {

		} else {
			recordTimeSlice, avtSlice, avhSlice := queryJob(db, ID)
			eqMap[ID] = map[string][]string{"RTS": recordTimeSlice, "ATS": avtSlice, "AHS": avhSlice}
		}

	}
	//fmt.Println(eqMap)
	mapB, _ := json.Marshal(eqMap)
	return mapB

}

//func chanStop(done chan bool, interval int) {
//	time.Sleep(time.Duration(interval) * time.Minute)
//	done <- true
//}

func jobInterval(interval int, done chan bool, db *sql.DB, pool *redis.Pool) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	// using pool.Get to give conn
	c := pool.Get()
	defer c.Close()

	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case ut := <-ticker.C:
			//fmt.Println("Current time: ", t)

			var eqMap map[string]bool
			eqMap = make(map[string]bool)

			keyLs, getErr := redis.Strings(c.Do("KEYS", "*"))
			//fmt.Println(key_ls)
			errCheck(getErr)
			reValidKeys, _ := regexp.Compile("^[a-zA-Z]+\\d+[a-zA-Z]+$")
			reValidEqs, _ := regexp.Compile("^[a-zA-Z]+\\d+")

			for _, element := range keyLs {
				if reValidKeys.MatchString(element) {
					// 处理设备映射
					tmpkey := reValidEqs.FindString(element)

					//fmt.Println("aaa",tmpkey)
					if _, exist := eqMap[tmpkey]; exist {
						//eqMap[tmpkey] = true
					} else {
						eqMap[tmpkey] = true
					}
				}
			}
			for eq := range eqMap {

				eqH, err := redis.String(c.Do("GET", eq+"H"))
				eqT, err := redis.String(c.Do("GET", eq+"T"))
				eqTime, err := redis.String(c.Do("GET", eq+"Time"))
				errCheck(err)
				//fmt.Println(eqH, eqT, eqTime)

				_, err = db.Exec("insert into sensor_table(ID,datetime,updatetime,temp,humi) values(?,?,?,?,?) "+
					"ON duplicate KEY UPDATE updatetime = ?", eq, eqTime, ut.Format(format), eqT, eqH, ut.Format(format))
				if err != nil {
					fmt.Println("insert err: ", err.Error())
				}
			}

		}
	}

}

//连接池的连接到服务的函数
func newPoolFunc() (redis.Conn, error) {
	return redis.Dial("tcp", ":6379")
}

//https://godoc.org/github.com/gomodule/redigo/redis#Pool
//https://github.com/pete911/examples-redigo
//生成一个连接池对象
func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle: 10,
		Dial:    newPoolFunc,
		Wait:    true,
	}
}
