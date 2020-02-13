package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"regexp"
	"time"
)

// --- model ---

// Article Object
type Article struct {
	ID    bson.ObjectId `bson:"_id" json:"id"`
	Title string        `bson:"title" json:"title"`
	Body  string        `bson:"body" json:"body"`
	CTime time.Time     `bson:"ctime" json:"ctime"`
}

// DB Database
type DB struct {
	Server   string
	Database string
	Table    string
}

var db *mgo.Database
var mdb = DB{Server: "mongodb://localhost:27017", Database: "blogo", Table: "article"}

// Connect database
func (m *DB) Connect() {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(m.Database)
}

// FindAll list
func (m *DB) FindAll() ([]Article, error) {
	var articles []Article
	err := db.C(m.Table).Find(bson.M{}).Sort("-$natural").All(&articles)
	return articles, err
}

// FindByID item
func (m *DB) FindByID(id string) (Article, error) {
	var article Article
	err := db.C(m.Table).FindId(bson.ObjectIdHex(id)).One(&article)
	return article, err
}

// Insert a article into database
func (m *DB) Insert(article Article) error {
	err := db.C(m.Table).Insert(&article)
	return err
}

// Delete a article into database
func (m *DB) Delete(article Article) error {
	err := db.C(m.Table).Remove(&article)
	return err
}

// Update a article into database
func (m *DB) Update(article Article) error {
	err := db.C(m.Table).UpdateId(article.ID, &article)
	return err
}

//--- WEBAPP ---

// Handler place
type Handler func(*Context)

// Route regex match handler
type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

// App object
type App struct {
	Routes       []Route
	DefaultRoute Handler
}

// NewApp for instance
func NewApp() *App {
	app := &App{
		DefaultRoute: func(ctx *Context) {
			ctx.JSON(http.StatusNotFound, "Not found")
		},
	}
	return app
}

// Handle router
func (a *App) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}
	a.Routes = append(a.Routes, route)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{Request: r, ResponseWriter: w}

	for _, rt := range a.Routes {
		if matches := rt.Pattern.FindStringSubmatch(ctx.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				ctx.Params = matches[1:]
			}

			rt.Handler(ctx)
			return
		}
	}
	a.DefaultRoute(ctx)
}

// Context object
type Context struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

// JSON send json response
func (c *Context) JSON(code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	c.ResponseWriter.Header().Set("Content-Type", "application/json")
	c.WriteHeader(code)
	c.ResponseWriter.Write(response)
}

// Success send normal json
func (c *Context) Success(msg string) {
	c.JSON(http.StatusOK, map[string]string{"error": msg})
}

func (c *Context) Error(code int, msg string) {
	c.JSON(code, map[string]string{"error": msg})
}

func init() {
	mdb.Connect()
}

func main() {
	app := NewApp()
	app.Handle(`^/api/save$`, app.saveHandler)
	app.Handle(`^/api/article/([\w\._-]+)$`, app.articleHandler)
	app.Handle(`^/api/index$`, app.indexHandler)
	app.Handle(`^/`, func(ctx *Context) {
		http.ServeFile(ctx.ResponseWriter, ctx.Request, "./static"+ctx.Request.URL.Path)
	})
	log.Fatal(http.ListenAndServe(":8081", app))
}

// --- API ---
func (a *App) indexHandler(ctx *Context) {
	items, err := mdb.FindAll()
	if err != nil {
		ctx.Error(http.StatusBadRequest, "Nothing")
		return
	}
	ctx.JSON(http.StatusOK, items)
}

func (a *App) saveHandler(ctx *Context) {
	defer ctx.Request.Body.Close()
	var article Article

	if err := json.NewDecoder(ctx.Request.Body).Decode(&article); err != nil {
		ctx.Error(http.StatusBadRequest, "invalid request")
		return
	}

	article.ID = bson.NewObjectId()
	article.CTime = time.Now()

	if err := mdb.Insert(article); err != nil {
		ctx.Error(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusCreated, article)
}

func (a *App) articleHandler(ctx *Context) {
	item, err := mdb.FindByID(ctx.Params[0])
	if err != nil {
		ctx.Error(http.StatusNotFound, "Not found")
		return
	}
	ctx.JSON(http.StatusOK, item)
}
