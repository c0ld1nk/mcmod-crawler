package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"mcmod-crawler/crawler"
	"mcmod-crawler/jsonProcess"
	"net/http"
	"os"
	"time"
)

var workerCount = flag.Int("thread", 25, "线程数，默认25")

func startWebServer() {

	r := gin.Default()
	packs, err := jsonProcess.LoadJson("modpacks.json")
	if err != nil {
		panic(err)
	}
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>ModPack 列表</title>
</head>
<body>
    <h1>ModPack 列表</h1>
	<a>因为只有go单文件没有模板文件，所以加载数据需要一段时间</a>
    <form>
        排序字段：
        <select id="field">
            <option value="title">标题</option>
            <option value="views">浏览数</option>
            <option value="points">昨日指数</option>
            <option value="scores">热度</option>
        </select>
        升序/降序：
        <select id="order">
            <option value="asc">升序</option>
            <option value="desc">降序</option>
        </select>
        <button type="button" onclick="applySort()">排序</button>
    </form>
    <table border="1" id="table">
        <thead>
            <tr>
                <th>标题</th><th>浏览数</th><th>昨日指数</th><th>热度</th><th>分类</th><th>URL</th>
            </tr>
        </thead>
        <tbody></tbody>
    </table>
    <script>
        let data = [];
        fetch('/data')
            .then(function(r){ return r.json(); })
            .then(function(d){ data = d; renderTable(data); });

        function renderTable(d) {
            const tbody = document.querySelector("#table tbody");
            tbody.innerHTML = '';
            d.forEach(function(p) {
                const row = "<tr>" +
                    "<td>" + p.title + "</td>" +
                    "<td>" + p.views + "</td>" +
                    "<td>" + p.points + "</td>" +
                    "<td>" + p.scores + "</td>" +
                    "<td>" + p.category.join(',') + "</td>" +
                    "<td><a href='" + p.url + "' target='_blank'>链接</a></td>" +
                    "</tr>";
                tbody.innerHTML += row;
            });
        }

        function applySort() {
            const field = document.getElementById('field').value;
            const order = document.getElementById('order').value;
            fetch('/data?field=' + field + '&order=' + order)
                .then(function(r){ return r.json(); })
                .then(function(d){ renderTable(d); });
        }
    </script>
</body>
</html>`)
	})
	r.GET("/data", func(c *gin.Context) {
		field := c.DefaultQuery("field", "title")
		order := c.DefaultQuery("order", "asc")
		desc := order == "desc"
		tmp := make([]crawler.ModPack, len(packs))

		copy(tmp, packs)
		jsonProcess.SortJson(tmp, field, desc)

		c.JSON(http.StatusOK, tmp)
	})

	fmt.Println("服务器已启动: http://localhost:8080")
	r.Run(":8080")
}

func loadOrFetch() {
	if _, err := os.Stat("modpacks.json"); err == nil {
		fmt.Println("存在Json文件，直接启动，如果希望重新爬取，请删除json文件")
		return
	}
	fmt.Println("json不存在，开始爬取")
	packs := crawler.StartCrawler(*workerCount)
	filename := "modpacks.json"
	err := jsonProcess.SaveJSON(filename, packs)
	if err != nil {
		panic(err)
	}
}

func scheduleWeeklyUpdate() {
	ticker := time.NewTicker(7 * 24 * time.Hour)
	go func() {
		for range ticker.C {
			fmt.Println("每周自动更新数据")
			packs := crawler.StartCrawler(*workerCount)
			filename := "modpacks.json"
			err := jsonProcess.SaveJSON(filename, packs)
			if err != nil {
				fmt.Printf("json保存失败: %s", err.Error())
			}
		}
	}()
}

func main() {
	scheduleWeeklyUpdate()

	flag.PrintDefaults()
	flag.Parse()

	loadOrFetch()
	startWebServer()

}
