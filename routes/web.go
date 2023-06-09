package routes

import (
	"github.com/gorilla/mux"
	"goWeb/app/http/controllers"
	"net/http"
)

// 注册网页相关路由

func RegisterWebRoutes(r *mux.Router)()  {
	// 静态页面
	pc := new(controllers.PagesController)
	r.HandleFunc("/",pc.Home).Methods("GET").Name("home")
	r.HandleFunc("/about", pc.About).Methods("GET").Name("about")
	r.NotFoundHandler = http.HandlerFunc(pc.NotFound)

	//文章相关页面
	ac := new(controllers.ArticlesController)
	r.HandleFunc("/articles/{id:[0-9]+}", ac.Show).Methods("GET").Name("articles.show")
	// 文章列表
	r.HandleFunc("/articles",ac.Index).Methods("GET").Name("articles.index")
	//创建文章
	r.HandleFunc("/articles/create",ac.Create).Methods("GET").Name("articles.create")
	r.HandleFunc("/articles", ac.Store).Methods("POST").Name("articles.store")
}