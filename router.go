package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/fvbock/endless"
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

	//自定义中间件
	r.Use(Midware())

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

		//在协程中使用Context.Copy()结果
		api.GET("long_async", CopyContextUseInGroutin)

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
	//还有 ShouldBindHeader 将指定头部数据映射到特定结构中, 注意使用 tag 'header'

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
	//标签 binding:"-" 表示不用对该字段进行绑定

	//另外 ShouldBind 函数可以根据content type 自行决定使用哪种ShouldBind...来绑定数据

	r.POST("/upload_json", func(c *gin.Context) {
		type JsonContent struct {
			Content string `json:"content" binding:"required"`
			Pass    string `json:"pass" binding:"-"`
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

	//以上所有绑定其实都没有针对post 中 form data 来进行的绑定
	//要绑定 可以使用 Context.Bind 函数，他会自动根据 Context-Type 来进行对应类型的绑定
	//其中包含 form data
	//为此 yangfeng 编写了 github.com/smartwalle/binding 专门进行 post form data 的解析
	r.POST("/bind_form", func(c *gin.Context) {
		type StructA struct {
			FieldA string `form:"field_a"`
		}
		var sa StructA
		err := c.Bind(&sa)
		if err != nil {
			c.String(http.StatusInternalServerError, "err: %s", err)
			return
		}
		fmt.Println("============", sa)
		c.String(http.StatusOK, "sa: %v", sa)
	})
}

//使用 bind/shouldbind/mustbind 函数还有一个很重要的作用，即在映射过程中还会进行数据验证，这里使用的是如下三方库，功能很强大
//参见 https://godoc.org/github.com/go-playground/validator
//另外可以自行定义对于特定字段的验证
//todo: 该功能暂时貌似不能使用 来自 https://github.com/gin-gonic/gin  Custom Validators
//这里感觉其实并没有执行自定义检测函数 bookableDate 中的代码， 日志没有打印，并且随便修改，即便直接返回false 也不会对结果有影响
//代码逻辑是与当前时间进行比较，若是设定时间在当前时间之前应该不能通过验证，但是执行
//curl "localhost:8085/bookable?check_in=2018-04-16&check_out=2018-04-17" 居然能够通过验证，明显不正确
//curl "localhost:8085/bookable?check_in=2018-03-10&check_out=2018-03-09" 不能通过验证，是因为 tag gtfield=CheckIn 导致的

type Booking struct {
	CheckIn  time.Time `form:"check_in" binding:"required" time_format:"2006-01-02"`
	CheckOut time.Time `form:"check_out" binding:"required,gtfield=CheckIn" time_format:"2006-01-02"`
	Name     string    `form:"name" binding:"required"`
}

var bookableDate validator.Func = func(fl validator.FieldLevel) bool {
	fmt.Println("bookableDate")

	date, ok := fl.Field().Interface().(time.Time)
	if ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}

func getBookable(c *gin.Context) {
	fmt.Println("getBookable")

	var b Booking
	if err := c.ShouldBindWith(&b, binding.Query); err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Booking dates are valid!"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func CustomValid(r *gin.Engine) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		fmt.Println("--------------")
		err := v.RegisterValidation("bookableDate", bookableDate)
		if err != nil {
			fmt.Println("============err: ", err)
		}
	}
	r.GET("/bookable", getBookable)
}

//静态文件到获取
//StaticFile 用于单个文件
//StaticFS 用于一个目录下的指定文件
//注意 StaticFS 的第二个参数所指定的路径 需要与 上传文件接口中存入文件的路径保持一致
//c.File 将文件中的内容返回， 即 Serving data from file
//另外还有 Serving data from reader 参见 https://github.com/gin-gonic/gin
//http://127.0.0.1:8888/images.jpeg
//http://127.0.0.1:8888/getfile/file1
//http://127.0.0.1:8888/local/file2
func DownloadFile(r *gin.Engine) {
	r.StaticFS("/getfile", http.Dir("tmpfile/"))
	r.StaticFile("/images.jpeg", "./tmpfile/images.jpeg")

	r.GET("/local/file2", func(c *gin.Context) {
		c.File("tmpfile/file2")
	})

}

//重定向，可以定向到外部地址，也可以定向到内部地址
//http://127.0.0.1:8888/test
//http://127.0.0.1:8888/test1
func Redirect(r *gin.Engine) {

	//重定向到外部地址
	r.GET("/test", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
	})

	//重定向到内部地址
	r.GET("/test1", func(c *gin.Context) {
		c.Request.URL.Path = "/test2"
		r.HandleContext(c)
	})
	r.GET("/test2", func(c *gin.Context) {
		c.String(http.StatusOK, "hello,world")
	})
}

////自定义中间件
//需要在 header 中添加 Authorization
func Midware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.String(http.StatusUnauthorized, "err: invalid token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//验证token
		c.Next()
	}
}

//在新开协程中使用 gin.Context, 只能使用拷贝 Copy() 的对象，不能使用原对象
func CopyContextUseInGroutin(c *gin.Context) {
	cc := c.Copy()
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("request: ", cc.Request.URL.Path)
	}()

	c.String(http.StatusOK, "wait for 3 seconds.")
}

//使用特定参数运行http服务器
func RunHttpWithParam(router *gin.Engine) {
	s := http.Server{
		Addr:           ":8888",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()

}

//优雅停服和重启，这里使用如下三方库
//github.com/fvbock/endless
func RunGraceful(router *gin.Engine) {

	endless.DefaultWriteTimeOut = 10 * time.Second
	endless.DefaultReadTimeOut = 10 * time.Second
	endless.DefaultMaxHeaderBytes = 1 << 20
	endless.ListenAndServe(":8888", router)
}

//todo: 经试验 Run multiple service using Gin 存在问题，启动都服务无法访问
