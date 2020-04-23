package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
)

//注意：body 只能获取一次，若是调用了该函数，就不能在获得body的内容了
func PrintHeadBody(c *gin.Context) {
	c.Request.Header.Write(os.Stdout)
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		os.Exit(-1)
	}
	fmt.Printf("body: \n%s", body)
}

func main() {

	//InitLog()

	r := gin.New()

	//ModLogFormat(r)
	Group(r)
	InitRouter(r)
	CustomValid(r)

	DownloadFile(r)

	r.Run("localhost:8888")

}
