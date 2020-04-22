package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

//MultiWriter 即将多个 writer 融合为一个，
//参见 https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter01/01.1.html
func InitLog() {
	//一般要写到文件中需要设置无色，否则会将颜色编码也写入文件，影响日志内容
	gin.DisableConsoleColor()
	//gin.ForceConsoleColor()
	ginlog, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(os.Stdout, ginlog)
	ginerrlog, _ := os.Create("gin.err.log")
	gin.DefaultErrorWriter = io.MultiWriter(ginerrlog)

	gin.SetMode(gin.DebugMode)

	return
}

func ModLogFormat(r *gin.Engine) {
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
}

func apiMiddle(c *gin.Context) {

}

func xxxMiddle(c *gin.Context) {

}

func specifyMiddle(c *gin.Context) {

}

func setSpecifyMiddle(param int) func(c *gin.Context) {
	return func(c *gin.Context) {
		fmt.Println(param)
	}
}

func Group(r *gin.Engine) {
	//指定全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	//指定分组中间件
	api := r.Group("/api")
	api.Use(apiMiddle)
	{
		//指定特定接口的中间件,无需传入参数
		api.GET("/getapi", specifyMiddle, func(c *gin.Context) {

		})
		//指定特定接口的中间件,需传入参数
		api.POST("/postapi", setSpecifyMiddle(666), func(c *gin.Context) {

		})

		//建立分组之下的分组
		xxx := api.Group("/xxx")
		xxx.Use(xxxMiddle)
		xxx.GET("/getxxx", func(c *gin.Context) {

		})
	}
}

func InitRouter(r *gin.Engine) {
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
	//c.MustBind... 与ShouldBind... 相似，差别在于 Bind 中出现错误会直接c.Abortxxx 即不在执行后续流程
	//因此在 BindJSON 返回结果中无需再使用 c.Json进行回复，若使用，会出现 warning
	//使用哪个函数 就看是否需要进行更精确的返回值控制
	//参见 https://github.com/gin-gonic/gin   Model binding and validation

	//另外函数 c.Bind 会自行根据 header 内容来判断应该使用哪种方式进行bind, 当然需要结构体中具有相应 tag
	//可以在结构体中编写多个 tag  json xml yaml form, 这样能够适应多种客户端形式

	//标签 binding:"required" 表示该字段为必须， 若是没有该字段，则会绑定返回错误
	?标签 binding:"-" 表示不用对该字段进行绑定

	r.POST("/upload_json", func(c *gin.Context) {
		type JsonContent struct {
			Content string `json:"content" binding:"-"`
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

		fmt.Println("============", jc)
		c.String(http.StatusOK, "upload json: %v", jc)
	})
}