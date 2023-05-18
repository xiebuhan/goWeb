package tests

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestAllPages(t *testing.T)  {
	baseURL := "http://localhost:3000"

	//1 声明加初始化数据
	var tests = []struct{
		method string
		url 	string
		expected int
	}{
		{"GET", "/", 200},
		{"GET", "/about", 200},
		{"GET", "/notfound", 404},
		{"GET", "/articles", 200},
		{"GET", "/articles/create", 200},
		{"GET", "/articles/3", 200},
		{"GET", "/articles/3/edit", 200},
		{"POST", "/articles/3", 200},
		{"POST", "/articles", 200},
		{"POST", "/articles/1/delete", 404},
	}
	//2 遍历所有测试
	for _,test:=range tests{
		t.Logf("当前请求URL: %v \n",test.url)

		var (
			resp *http.Response
			err  error
		)
		// 请求以获取相应
		switch {
			case test.method == "POST":
				data := make(map[string][]string)
				resp,err = http.PostForm(baseURL+test.url,data)
			default:
				resp, err = http.Get(baseURL + test.url)
		}
		// 2 断言
		assert.NoError(t,err,"请求" + test.url)
		assert.Equal(t,test.expected,resp.Status,test.url+" 应返回状态码 "+strconv.Itoa(test.expected))

	}



}