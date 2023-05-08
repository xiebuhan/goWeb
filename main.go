package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"html/template"
	"log"
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
	checkError(err)

	// 设置最大连接数
	db.SetMaxOpenConns(25)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)
	// 设置每个链接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	//尝试连接，失败报错
	err = db.Ping()
	checkError(err)

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//func createTables() {
//	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
//    id bigint(10) PRIMARY KEY AUTO_INCREMENT NOT NULL,
//    title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
//    body longtext COLLATE utf8mb4_unicode_ci
//); `
//
//	_, err := db.Exec(createArticlesSQL)
//	checkError(err)
//}

type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello, 欢迎来到 goblog！</h1>")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 "+
		"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}

type Article struct {
	Title, Body string
	ID          int64
}


func articlesShowHandler(w http.ResponseWriter, r *http.Request) {

   id := getRouteVariable("id",r)
	// 2 读取对应的文章数据
	article,err := getArticlesByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			//3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w,"404数据未找到")
		} else{
			// 3.2 数据库错误
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,"500服务器内部错误")
		}
	} else {
		// 读取成功
		tmpl,err := template.ParseFiles("resources/views/articles/show.gohtml")
		checkError(err)
		err = tmpl.Execute(w,article)
		checkError(err)
	}


}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "访问文章列表")
}
//接受文章表单的数据
func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

    errors:= ValidateArticleFormData(title,body)

	// 检查是否有错误
	if len(errors) == 0 {
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprint(w, "插入成功，ID 为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,  "500 服务器内部错误")
		}
	} else {
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>创建文章 —— 我的技术博客</title>
    <style type="text/css">.error {color: red;}</style>
</head>
<body>
    <form action="{{ .URL }}" method="post">
        <p><input type="text" name="title" value="{{ .Title }}"></p>
        {{ with .Errors.title }}
        <p class="error">{{ . }}</p>
        {{ end }}
        <p><textarea name="body" cols="30" rows="10">{{ .Body }}</textarea></p>
        {{ with .Errors.body }}
        <p class="error">{{ . }}</p>
        {{ end }}
        <p><button type="submit">提交</button></p>
    </form>
</body>
</html>
`
		storeURL, _ := router.Get("articles.store").URL()

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}

		tmpl, err := template.New("create-form").Parse(html)
		if err != nil {
			panic(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err)
		}
	}

}

//保存文章
func saveArticleToDB(title string,body string) (int64,error)  {
	//变量初始化
	var(
		id int64
		err error
		rs sql.Result
		stmt *sql.Stmt
	)

	stmt, err = db.Prepare("INSERT INTO articles (title, body) VALUES(?,?)")
	// 例行的错误检测
	if err != nil {
		return 0, err
	}

	// 2. 在此函数运行结束后关闭此语句，防止占用 SQL 连接
	defer stmt.Close()
	// 3. 执行请求，传参进入绑定的内容
	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	// 4. 插入成功的话，会返回自增 ID
	if id, err = rs.LastInsertId(); id > 0 {
		return id, nil
	}

	return 0, err
}

func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续处理请求
		next.ServeHTTP(w, r)
	})
}
// 创建博文表单
func articlesCreateHandler(w http.ResponseWriter,r *http.Request)  {
	StoreURL,_:=router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    StoreURL,
		Errors: nil,
	}

	tmpl,err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w,data)
	if err != nil {
		panic(err)
	}


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
			checkError(err)
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
		checkError(err)

		err = tmpl.Execute(w,data)
		checkError(err)

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
				checkError(err)
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
			checkError(err)
			err = tmpl.Execute(w,data)
			checkError(err)

		}

	}
}

//获取url
func getRouteVariable(parameterName string,r *http.Request) string  {
	vars := mux.Vars(r)
	return  vars[parameterName]
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



func main() {
	initDB()
	//createTables()
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")

	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create",articlesCreateHandler).Methods("GET").Name("articles.create")
	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")

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