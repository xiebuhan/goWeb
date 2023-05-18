package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"goWeb/bootstrap"
	"goWeb/pkg/logger"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
)


var router = mux.NewRouter()
var db *sql.DB

func initDB()  {
	var err error
	config := mysql.Config{
		User:                 "root",
		Passwd:               "root",
		Addr:                 "127.0.0.1:3306",
		Net:                  "tcp",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}
	// 准备数据库连接池
	db,err =  sql.Open("mysql",config.FormatDSN())
	logger.LogError(err)

	// 设置最大连接数
	db.SetMaxOpenConns(25)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)
	// 设置每个链接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	//尝试连接，失败报错
	err = db.Ping()
	logger.LogError(err)

}


type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}

type Article struct {
	Title, Body string
	ID          int64
}


func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续处理请求
		next.ServeHTTP(w, r)
	})
}



func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}

//编辑文章
func articlesEditHandler(w http.ResponseWriter,r *http.Request)  {
	//获取url参数
	id := getRouteVariable("id",r)
	//2 读取对应的文章数据
	article,err := getArticlesByID(id)
	if err != nil {
		if err == sql.ErrNoRows{
			//3.1 数据未找到
			w.WriteHeader(http.StatusFound)
			fmt.Fprint(w,"404文章未找到")
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,"500 服务器内部错误")
		}
	} else {
		// 4 读取成功 显示表单
		updateURL,_:= router.Get("articles.update").URL("id",id)
		data := ArticlesFormData{
			Title:  article.Title,
			Body:   article.Body,
			URL:    updateURL,
			Errors: nil,
		}

		tmpl,err := template.ParseFiles("resources/views/articles/edit.gohtml")
		logger.LogError(err)

		err = tmpl.Execute(w,data)
		logger.LogError(err)

	}




}
//更新数据
func articlesUpdateHandler(w http.ResponseWriter,r *http.Request){
	// 1 获取RUL参数
	id := getRouteVariable("id",r)
	// 2 读取对应文章的数据
	_,err:=getArticlesByID(id)

	//3 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusFound)
			fmt.Fprint(w,"文章没有找到")
		}else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,"505 系统错误")
		}

	} else {
		// 4 未出现错误

		// 4.1 表单验证
		title := r.PostFormValue("title")
		body := r.PostFormValue("body")

		errors := ValidateArticleFormData(title, body)

		if len(errors) == 0{
			// 表单验证通过，更新数据
			query := "update articles set title = ?,body = ? where id = ?"
			rs,err := db.Exec(query,title,body,id)
			if err != nil {
				logger.LogError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w,"500服务器内部错误")
			}

			 // 更新成功，跳转到文章详情页
			if n,_ := rs.RowsAffected(); n>0 {
			 showURL,_ := router.Get("articles.show").URL("id",id)
			 http.Redirect(w,r,showURL.String(),http.StatusFound)
			} else {
				fmt.Fprint(w,"您没有做任何更改")
			}

		}else {
			// 表单没有通过 显示原因
			updateUrl,_ :=  router.Get("articles.update").URL("id",id)
			data := ArticlesFormData{
				Title:  title,
				Body:   body,
				URL:    updateUrl,
				Errors: errors,
			}

			tmpl,err := template.ParseFiles("resources/views/articles/edit.gohtml")
			logger.LogError(err)
			err = tmpl.Execute(w,data)
			logger.LogError(err)

		}

	}
}

// 删除文章
func articlesDeleteHandler(w http.ResponseWriter,r *http.Request)  {
	// 1 获取URL参数
	id := getRouteVariable("id",r)

	// 2 读取对应的文章数据
	article,err := getArticlesByID(id)

	//3 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w,"404文章没有找到")
		} else {
			// 3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,"500服务器内部错误")
		}
	} else{
		// 4 未出现错误 执行操作
		rowsAffected,err := article.Delete()

		//4.1 发生错误
		if err != nil {
			// 应该是SQL报错了
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,"500服务器内部错误'")
		}else{
			// 4.2 未发生错误
			if rowsAffected >0 {
				// 重定向到文章列表页
				indexURL,_ := router.Get("articles.index").URL()
				http.Redirect(w,r,indexURL.String(),http.StatusFound)
			} else {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w,"404文章未找到")
			}
		}
	}

}


// 获取文章ID
func getArticlesByID(id string)(Article,error)  {
	article := Article{}
	query := "select * from articles where id = ?"
	err := db.QueryRow(query,id).Scan(&article.ID,&article.Title,&article.Body)
	return article,err

}

func ValidateArticleFormData(title string,body string) map[string]string{
	errors := make(map[string]string)
	// 验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于 3-40"
	}

	// 验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于 10 个字节"
	}

	return errors
}



func (a Article) Delete() (rowsAffected int64,err error)  {
 	rs,err := db.Exec("DELETE FROM  articles where id =" + strconv.FormatInt(a.ID,10))
	if err != nil {
		return  0,err
	}

	//删除成功
	if n,_:=rs.RowsAffected();n>0{
		return  n,nil
	}

	return  0,nil
}

// Int64Tostring 将int64 转换为string
func Int64ToString(num int64)string  {
	return strconv.FormatInt(num,10)
}

func getRouteVariable(parameterName string,r *http.Request) string  {
	vars := mux.Vars(r)
	return  vars[parameterName]
}


func main() {
	//database.Initialize()
	//db = database.DB

	bootstrap.SetDB()
	router = bootstrap.SetupRoute()
	//createTables()

	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")
	// 自定义 404 页面
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	// 通过命名路由获取 URL 示例
	homeURL, _ := router.Get("home").URL()
	fmt.Println("homeURL: ", homeURL)
	articleURL, _ := router.Get("articles.show").URL("id", "1")
	fmt.Println("articleURL: ", articleURL)


	http.ListenAndServe(":3000", removeTrailingSlash(router))
}