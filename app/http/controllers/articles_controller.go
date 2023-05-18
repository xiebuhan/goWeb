package controllers

import (
	"fmt"
	article2 "goWeb/app/models/article"
	"goWeb/pkg/logger"
	"goWeb/pkg/route"
	"goWeb/pkg/types"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"strconv"
	"unicode/utf8"
)

type ArticlesController struct {
	
}
//创建博文表单数据
type ArticlesFormData  struct {
	Title,Body string
	URL		   string
	Errors		map[string]string
}

// show 文章详情页面
func (*ArticlesController) Show(w http.ResponseWriter,r *http.Request)  {

	id := route.GetRouteVariable("id",r)
	// 2 读取对应的文章数据
	article,err := article2.Get(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w,"404数据未找到")
		} else{
			// 3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,"500服务器内部错误")
		}
	} else {
		// 读取成功
		tmpl,err := template.New("show.gohtml").Funcs(template.FuncMap{
			"RouteName2URL" :route.Name2URL,
			"Uint64ToString" :types.Uint64ToString,
		}).ParseFiles("resources/views/articles/show.gohtml")
		logger.LogError(err)
		err = tmpl.Execute(w,article)
		logger.LogError(err)
	}
}

//文章列表页
func (*ArticlesController) Index(w http.ResponseWriter, r *http.Request) {
	// 1 执行查询语句，返回一个结果集
	articles,err := article2.GetAll()

	logger.LogError(err)

	if err != nil{
		//数据库错误
		logger.LogError(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w,"500服务器内部错误")
	} else {
		// 加载模板
		tmpl,err := template.ParseFiles("resources/views/articles/index.gohtml")
		logger.LogError(err)

		//渲染模板
		err = tmpl.Execute(w,articles)
		logger.LogError(err)
	}



}

// 创建博文表单
func  (*ArticlesController) Create(w http.ResponseWriter,r *http.Request)  {
	storeURL := route.Name2URL("articles.store")

	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
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


//接受文章表单的数据
func (*ArticlesController) Store(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	errors:= ValidateArticleFormData(title,body)

	// 检查是否有错误
	if len(errors) == 0 {
		_article := article2.Article{
			Title: title,
			Body:  body,
		}
		_article.Create()

		if _article.ID > 0 {
			fmt.Fprint(w, "插入成功，ID 为"+strconv.FormatUint(_article.ID, 10))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w,  "500 服务器内部错误")
		}
	} else {

		storeURL := route.Name2URL("articles.store")

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")

		logger.LogError(err)

		err = tmpl.Execute(w, data)
		logger.LogError(err)
	}

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