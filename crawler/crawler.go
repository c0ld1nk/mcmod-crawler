package crawler

import (
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ModPack struct {
	Title    string   `json:"title"`
	Views    string   `json:"views"`
	Points   float32  `json:"points"`
	Scores   float32  `json:"scores"`
	Category []string `json:"category"`
	URL      string   `json:"url"`
}

type parsedNodes struct {
	pointsDivs   []soup.Root
	scoresDivs   []soup.Root
	titleDivs    []soup.Root
	categoryDivs []soup.Root
	viewDivs     []soup.Root
}

type resultModPack struct {
	pack ModPack
	err  error
	id   int
}

var userAgents = []string{
	// Chrome
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.97 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.97 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.97 Safari/537.36",
	// Firefox
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:116.0) Gecko/20100101 Firefox/116.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.5; rv:116.0) Gecko/20100101 Firefox/116.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:116.0) Gecko/20100101 Firefox/116.0",
	// Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1",
	// Edge
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.97 Safari/537.36 Edg/116.0.1938.81",
	// Android Chrome
	"Mozilla/5.0 (Linux; Android 13; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.98 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 12; SM-G991U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.98 Mobile Safari/537.36",
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

var client = &http.Client{
	Timeout: 10 * time.Second, // 避免请求长时间挂起
}

func randomUA() string {
	return userAgents[rnd.Intn(len(userAgents))]
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		return
	}
}

func parseNodes(root soup.Root) parsedNodes {
	return parsedNodes{
		pointsDivs:   root.FindAll("div", "class", "block-right"),
		scoresDivs:   root.FindAll("div", "class", "star"),
		titleDivs:    root.FindAll("div", "class", "class-title"),
		categoryDivs: root.FindAll("div", "class", "class-category"),
		viewDivs:     root.FindAll("div", "class", "infos"),
	}
}

func findPoint(root []soup.Root) float32 {
	point := root[0].FindAll("div")[1].Children()[0].NodeValue
	fields := strings.Fields(point)
	numStr := fields[len(fields)-1]
	end, err := strconv.ParseFloat(numStr, 32)
	checkErr(err)
	return float32(end)
}

func findView(root []soup.Root) string {
	return root[0].Find("p").Text()
}

func findCategory(root []soup.Root) []string {
	category := root[0].FindAll("a")
	temp := make([]string, len(category))
	for i, c := range category {
		temp[i] = c.Text()
	}
	return temp
}

func findScores(root []soup.Root) float32 {
	star := root[0].Find("p").Text()
	end, err := strconv.ParseFloat(star, 32)
	checkErr(err)
	return float32(end)
}

func findTitle(root []soup.Root) string {
	return root[0].Find("h3").Text()
}

func contentParser(content string) *ModPack {
	tempPack := &ModPack{}
	html := soup.HTMLParse(content)
	nodes := parseNodes(html)
	tempPack.Title = findTitle(nodes.titleDivs)
	tempPack.Category = findCategory(nodes.categoryDivs)
	tempPack.Scores = findScores(nodes.scoresDivs)
	tempPack.Views = findView(nodes.viewDivs)
	tempPack.Points = findPoint(nodes.pointsDivs)
	return tempPack
}

func sleepRandom() {
	t := rnd.Float32() + 0.05 // 随机 0.05~1.05 秒
	time.Sleep(time.Duration(t) * time.Second)
}

func (pack *ModPack) String() {
	fmt.Println("整合包名称：", pack.Title)
	fmt.Println("标签：", pack.Category)
	fmt.Println("分数：", pack.Scores)
	fmt.Println("总浏览：", pack.Views)
	fmt.Println("昨日指数：", pack.Points)
}

func request(id int, url string) resultModPack {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("User-Agent", randomUA())
	rsp, err := client.Do(req)
	if err != nil {
		return resultModPack{err: err, id: id}
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return resultModPack{err: errors.New(rsp.Status), id: id}
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return resultModPack{err: err, id: id}
	}

	content := string(body)
	pack := contentParser(content)
	pack.URL = url
	sleepRandom()

	return resultModPack{pack: *pack, err: nil, id: id}
}

func getNumbers() int {
	url := "https://www.mcmod.cn/modpack.html?sort=createtime"
	rsp, err := http.Get(url)
	checkErr(err)
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	checkErr(err)

	html := soup.HTMLParse(string(body))
	subDoc := html.Find("div", "class", "modlist-block")
	number := subDoc.Find("a").Attrs()["href"]

	parts := strings.Split(number, "/")
	filename := parts[len(parts)-1]
	numStr := strings.TrimSuffix(filename, ".html")
	num, _ := strconv.Atoi(numStr)
	return num
}

func StartCrawler(workerCount int) []ModPack {
	// http请求网页内容
	fmt.Println("启动爬虫程序ing...")
	start := time.Now()
	maxNumbers := getNumbers()
	fmt.Printf("整合包总数量：%d, 基于最新的整合包url获取\n", maxNumbers)

	fmt.Println("当前线程数：", workerCount)
	jobs := make(chan int, workerCount)
	results := make(chan resultModPack, maxNumbers)

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				url := "https://www.mcmod.cn/modpack/" + strconv.Itoa(j) + ".html"
				results <- request(j, url)
			}
		}()
	}

	go func() {
		for i := 1; i < maxNumbers; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var modPacks []ModPack
	for res := range results {
		if res.err != nil {
			fmt.Printf("ID %d 出错: %v\n", res.id, res.err)
			continue
		}
		modPacks = append(modPacks, res.pack)
	}
	n := len(modPacks)
	if n > 0 {
		fmt.Println("测试输出")
		modPacks[0].String()
	}

	end := time.Now()
	fmt.Println("耗时：", end.Sub(start))

	fmt.Printf("%d 整合包获取成功,已保存到", n)

	return modPacks
}
