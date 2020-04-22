package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func PrintHeadBody(c *gin.Context) {
	c.Request.Header.Write(os.Stdout)
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		os.Exit(-1)
	}
	fmt.Printf("body: \n%s", body)
}

func main() {
	r := gin.Default()
	//http://127.0.0.1:8888/user/lilei/send
	//可以使用 c.Param() 获取单个指定参数
	//可以使用 c.Params 获取所有参数
	//可以使用 c.ShouldBindUri 将指定参数映射到特定结构中, 注意使用 tag 'uri'

	r.POST("/user_param/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		c.String(http.StatusOK, "hello, %s, %s", name, action)
	})

	r.POST("/user_params/:name/*action", func(c *gin.Context) {
		params := c.Params
		c.String(http.StatusOK, "hello, %s, %s", params[0].Value, params[1].Value)
	})

	r.POST("/user_bind/:name/*action", func(c *gin.Context) {
		type Tmp struct {
			Name   string `uri:"name"`
			Action string `uri:"-"`
		}
		var tmp Tmp
		err := c.ShouldBindUri(&tmp)
		if err != nil {
			fmt.Println(err)
		}
		c.String(http.StatusOK, "hello, %s, %s", tmp.Name, tmp.Action)
	})

	//POST /postquery?id=1234&page=1 HTTP/1.1
	//可以使用 c.Query() 获取单个指定参数
	//可以使用 c.ShouldBindQuery 将指定参数映射到特定结构中, 注意使用 tag 'form'
	//另外还可以使用 c.QueryArray c.QueryMap 来获取query 中到数组和map

	r.POST("/postquery", func(c *gin.Context) {
		id := c.Query("id")
		page := c.Query("page")

		c.String(http.StatusOK, "%s, %s", id, page)
	})

	r.POST("postquery_bind", func(c *gin.Context) {
		type Tmp struct {
			Id   string `form:"id"`
			Page string `form:"page"`
		}
		var tmp Tmp
		c.ShouldBindQuery(&tmp)
		c.String(http.StatusOK, "%s, %s", tmp.Id, tmp.Page)

	})

	//POST /post?ids[a]=1234&ids[b]=hello HTTP/1.1
	//Content-Type: application/x-www-form-urlencoded
	//
	//names[first]=thinkerou&names[second]=tianou
	r.POST("/postmap", func(c *gin.Context) {
		ids := c.QueryMap("ids")
		names := c.PostFormMap("names")
		c.String(http.StatusOK, "ids: %v, names: %v", ids, names)
	})

	//POST /post?ids=1234&ids=hello HTTP/1.1
	//Content-Type: application/x-www-form-urlencoded
	//
	//names=thinkerou&names=tianou
	r.POST("/postarray", func(c *gin.Context) {
		ids := c.QueryArray("ids")
		names := c.PostFormArray("names")
		c.String(http.StatusOK, "ids: %v, names: %v", ids, names)
	})

	//POST http://127.0.0.1:8888/upload_one_file
	//Content-Type: multipart/form-data
	//使用 form-data 上传key=file 的文件， 单个文件
	r.POST("/upload_one_file", func(c *gin.Context) {
		head, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusInternalServerError, "err: %s', err")
			return
		}
		c.SaveUploadedFile(head, "./file")
		c.String(http.StatusOK, "upload file: %s", head.Filename)
	})

	//POST http://127.0.0.1:8888/upload_multi_files
	//Content-Type: multipart/form-data
	//使用 form-data 上传key=files 的文件， 多个文件
	r.POST("/upload_multi_files", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		files := form.File["files"]
		filenames := make([]string, 0)
		for _, file := range files {
			filenames = append(filenames, file.Filename)
			c.SaveUploadedFile(file, file.Filename)
		}
		c.String(http.StatusOK, "upload files: %v", filenames)
	})

	//POST http://127.0.0.1:8888/upload_bin?filename=fn
	//使用 binary 上传二进制文件，可能需要依靠参数等位置来指定名字
	r.POST("/upload_bin", func(c *gin.Context) {
		filename := c.Query("filename")
		file, err := c.GetRawData()
		if err != nil {
			c.String(http.StatusInternalServerError, "err: %s', err")
			return
		}
		ioutil.WriteFile(filename, file, os.ModePerm)
		c.String(http.StatusOK, "upload file: %s", filename)
	})

	//POST http://127.0.0.1:8888/upload_text
	//使用 raw 上传文本text
	r.POST("/upload_text", func(c *gin.Context) {
		content, err := c.GetRawData()
		if err != nil {
			c.String(http.StatusInternalServerError, "err: %s', err")
			return
		}
		c.String(http.StatusOK, "upload text: %s", content)
	})

	//POST http://127.0.0.1:8888/upload_json
	//使用 raw 上传文本json
	//{
	//	"content": "this is a json."
	//}
	//c.ShouldBind... 用于将特定内容（body/head/query字段）绑定到定义到结构体中
	//c.Bind... 与ShouldBind... 相似，差别在于 Bind 中出现错误会直接c.Abortxxx 即不在执行后续流程
	//因此在 BindJSON 返回结果中无需再使用 c.Json进行回复
	//Bind... 函数应该在中间件中使用
	r.POST("/upload_json", func(c *gin.Context) {
		type JsonContent struct {
			Content string `json:"content"`
		}
		var jc JsonContent
		err := c.ShouldBindJSON(&jc)
		if err != nil {
			c.String(http.StatusInternalServerError, "err: %s", err)
			return
		}

		//err := c.BindJSON(jc)
		//if err != nil {
		//	fmt.Println("err: ", err)
		//	return
		//}

		fmt.Println("============")
		c.String(http.StatusOK, "upload json: %v", jc)
	})

	r.Run("localhost:8888")

}
