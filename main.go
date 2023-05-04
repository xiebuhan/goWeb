package main

import (
	"fmt"
	"net/http"
)

func handlefunc(w http.ResponseWriter, r *http.Request)  {

	fmt.Fprint(w,"请求路径为:" + r.URL.Path)
	if r.URL.Path == "/"{
		fmt.Fprint(w, "<h1>Hello, 这里是 goblog</h1>")
	} else if r.URL.Path == "/about" {
		fmt.Fprint(w,"此博客是用以记录编程笔记，如您有反馈或建议，请联系 "+
			"<a href=\"www.baidu.com\">baiduddd</a>")
	} else {
		fmt.Fprint(w, "<h1>请求页面未找到 :(</h1>"+
			"<p>如有疑惑，请联系我们。</p>")
	}

}

func main() {
	http.HandleFunc("/", handlefunc)
	http.ListenAndServe(":3000", nil)
}