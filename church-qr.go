package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"image"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type Church struct {
	Name string
	Logo string
}

var Churches = map[string]Church{
	"doxa":          {"Doxa Deo", "doxa.png"},
	"waterkloofags": {"Waterkloof AGS", "waterkloofags.png"},
}

func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("Running with %d CPUs\n", nuCPU)
}

func StartGin() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.LoadHTMLGlob("templates/*.gohtml")
	router.GET("/", index)
	router.GET("/:name", church)
	router.POST("/:name", generateQRCode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
}

func index(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/doxa")
}

func church(c *gin.Context) {
	selectedChurch, okay := Churches[c.Param("name")]
	if !okay {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "index.gohtml", gin.H{
		"Church": selectedChurch.Name,
	})
}

func generateQRCode(c *gin.Context) {
	selectedChurch, okay := Churches[c.Param("name")]
	if !okay {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	downloadName := fmt.Sprintf("qrcode-%d.jpeg", time.Now().Unix())

	//  Uncomment the following three lines to force download.
	//header := c.Writer.Header()
	//header["Content-type"] = []string{"application/octet-stream"}
	//header["Content-Disposition"] = []string{"attachment; filename= " + downloadName}

	imageBytes, err := os.ReadFile(selectedChurch.Logo)
	if err != nil {
		log.Fatalln(err)
	}
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		log.Fatalln(err)
	}

	//qrc, err := qrcode.New("https://doxadeojhb.churchsuite.com/events/egfn6xkr")
	qrc, err := qrcode.New(c.PostForm("url"))
	if err != nil {
		fmt.Printf("could not generate QRCode: %v", err)
		return
	}

	w, err := standard.New(downloadName, standard.WithLogoImage(img))
	if err != nil {
		fmt.Printf("standard.New failed: %v", err)
		return
	}

	width := w.Attribute(qrc.Dimension()).W
	calculateWidth := width / 5
	if calculateWidth < 160 {
		img = resize.Resize(uint(calculateWidth), uint(calculateWidth), img, resize.Lanczos3)
	}

	w, err = standard.New(downloadName, standard.WithLogoImage(img))

	// save file
	if err = qrc.Save(w); err != nil {
		fmt.Printf("could not save image: %v", err)
	}

	c.File(downloadName)

	os.Remove(downloadName)
}

func main() {
	ConfigRuntime()
	StartGin()
}
