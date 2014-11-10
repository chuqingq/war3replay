package main

/*
项目：在线一键看war3录像

场景：
    war3replay.exe可以配置一个json的配置文件，说明：
        启动的war3目录（默认地图和replay下载到这个目录下）
    启动可执行程序，自动打开浏览器，看到replay页面，页面中展示最近的replay。
    展示形式如下：
        race players map date link replay
        UD vs HM [妖妖杯]Yumiko vs Xiaokai #2 LastRefuge1.3 11-07
    点击replay之后自动下载录像和地图到事先指定的目录下，并启动war3播放
    点击link后查看war3.replays.net上的页面

TODO
* 启动exe后先启动协程请求replist，然后拉起httpserver，协程拿到replist后就拉起浏览器
* rep和map要判断是否已经存在
* 容错性，不要任何错误挂掉
* 读取replist要在未读完时就时时展示：协程拿到replist后就拉起浏览器
* 点击replay之后不要打开新页面

*/
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type repentry struct {
	Race   string
	Player string
	Map    string
	Date   string
	Link   string
}

const war3Path = "D:/GAME/Warcraft III/"
const war3Exe = "Frozen Throne.exe"
const replaySavePath = "replay/"

const httpAddr = "127.0.0.1:8080"
const httpListPattern = "/list"
const httpReplayPattern = "/replay"

func main() {
	http.HandleFunc(httpListPattern, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access: %s\n", httpListPattern)
		replist := getReplays()

		// 组装页面内容
		repbody := ""
		for _, rep := range replist {
			repbody += fmt.Sprintf(`
                <tr>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s</td>
                    <td><a href="%s" target="_blank">link</a></td>
                    <td><a href="/replay?link=%s" target="_blank">replay</a></td></td>
                </tr>
            `, rep.Date, rep.Race, rep.Player, rep.Map, rep.Link, rep.Link)
		}
		// 展示
		fmt.Fprintf(w, `
            <html>
                <head></head>
                <body>
                    <table border="2">
                      <tr>
                        <th>Date</th>
                        <th>Race</th>
                        <th>Player</th>
                        <th>Map</th>
                        <th>Link</th>
                        <th>Replay</th>
                      </tr>
                      %s
                    </table>
                </body>
            </html>
        `, repbody)
	})

	http.HandleFunc(httpReplayPattern, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access: %s\n", httpReplayPattern)
		reqUrl, err := url.Parse(r.RequestURI)
		if err != nil {
			log.Fatal(err)
		}

		link := reqUrl.Query().Get("link")
		log.Printf("link: %s\n", link)

		go getRep(link)
		// TODO
		fmt.Fprintln(w, "hello")
	})

	go startBrowser()

	log.Printf("listen at %s ...\n", httpAddr+httpListPattern)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

func getRep(link string) {
	resp, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Println("reading repinfo body...")
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reading repinfo body ok")

	content := string(buf)
	// log.Printf("replay:%s\n", content)

	var res string

	// 下载replay
	replayPathRe := regexp.MustCompile(`<span id="ctl00_Content_labDown" class="download"><a href="(.*)">Download REP</a></span>`)
	res = replayPathRe.FindString(content)
	replayPath := replayPathRe.ReplaceAllString(res, "$1")
	replayPath, err = url.QueryUnescape(replayPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("replayPath=%s\n", replayPath)
	replayNameRe := regexp.MustCompile(`/Download.aspx\?ReplayID=.*&File=/ReplayFile/.*/(.*)`)
	replayName := replayNameRe.ReplaceAllString(replayPath, "$1")
	log.Printf("replayName=%s\n", replayName)

	respRep, err := http.Get("http://w3g.replays.net" + replayPath)
	if err != nil {
		log.Fatal(err)
	}
	defer respRep.Body.Close()

	log.Println("reading rep body...")
	buf, err = ioutil.ReadAll(respRep.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reading rep body...")

	// 无论是否存在
	err = ioutil.WriteFile(war3Path+replaySavePath+replayName, buf, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// 下载地图
	mapPathRe := regexp.MustCompile(`<span id="ctl00_Content_labMapname">([^<]*)</span>`)
	res = mapPathRe.FindString(content)
	mapPath := mapPathRe.ReplaceAllString(res, "$1")
	mapPath = strings.Replace(mapPath, "\\", "/", -1)
	log.Printf("mappath=%s\n", mapPath)

	ind := strings.LastIndex(mapPath, "/")
	mapName := mapPath[ind+1:]
	log.Printf("mapName=%s\n", mapName)

	downPathRe := regexp.MustCompile(`javascript:getreplaymap\(.*,'(.*)','.*'\)`)
	res = downPathRe.FindString(content)
	log.Printf("downpath res=%s\n", res)
	downPath := downPathRe.ReplaceAllString(res, "$1")
	log.Printf("downPath=%s\n", downPath)

	mapPathAbs := "http://w3g.replays.net/ReplayMap/download/" + downPath + "/" + mapName
	log.Printf("mapPathAbs=%s\n", mapPathAbs)

	respMap, err := http.Get(mapPathAbs)
	if err != nil {
		log.Fatal(err)
	}
	defer respMap.Body.Close()

	log.Println("reading map body...")
	buf, err = ioutil.ReadAll(respMap.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reading map body ok")

	// 无论是否存在
	mapDir := war3Path + mapPath
	log.Printf("mapDir=%s\n", mapDir)
	ind = strings.LastIndex(mapDir, "/")
	mapDir = mapDir[:ind]
	log.Printf("mapDir2=%s\n", mapDir)
	err = os.MkdirAll(mapDir, 0777)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(war3Path+mapPath, buf, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	startReplay(replayName)
}

func getReplays() []*repentry {
	// 获取replay页面内容
	resp, err := http.Get("http://w3g.replays.net")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 获取replist
	log.Println("reading replist body...")
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reading replist body ok")

	content := string(buf)

	// 处理页面内容，保存到replist中
	replist := make([]*repentry, 0)

	// 取"War3-left fl"和"War3-right fr"之间的内容
	// const left = `<div class="War3-rep-list">`
	const left = `<ul class="datarow2">`
	const right = `<span id="ctl00_Content_labPage" class="cutpage">`
	var content2 *string
	strArr := strings.Split(content, left)

	content2 = &strArr[1]
	strArr2 := strings.Split(*content2, right)
	content2 = &strArr2[0]

	res := *content2
	// log.Printf("body: %s\n", res)

	raceRe := regexp.MustCompile(`<li class="c_r"><a href=".*">(.*)</a></li>\r\n`)
	res = raceRe.ReplaceAllString(res, "$1|")

	playerRe := regexp.MustCompile(`<li class="c_p"><a href="(.*)" target="_blank">(.*)</li>\r\n`)
	res = playerRe.ReplaceAllString(res, "$2|$1|")

	mapRe := regexp.MustCompile(`<li class="c_m">(.*)</li>\r\n`)
	res = mapRe.ReplaceAllString(res, "$1|")

	dateRe := regexp.MustCompile(`<li class="c_t">(.*)</li>\r\n`)
	res = dateRe.ReplaceAllString(res, "$1\n")

	otherRe := regexp.MustCompile(`<(.*)>\r\n`)
	res = otherRe.ReplaceAllString(res, "")

	// log.Println(res)

	strArr = strings.Split(res, "\n")

	for _, line := range strArr {
		resArr := strings.Split(line, "|")
		if len(resArr) != 5 {
			continue
		}
		rep1 := repentry{
			Race:   resArr[0],
			Player: resArr[1],
			Map:    resArr[3],
			Link:   resArr[2],
			Date:   resArr[4],
		}
		// log.Printf("Race=%s,Player=%s,Map=%s,Link=%s,Date=%s\n", resArr[0], resArr[1], resArr[3], resArr[2], resArr[4])
		replist = append(replist, &rep1)
	}

	return replist
}

func startBrowser() {
	time.Sleep(2 * time.Second)
	cmd := exec.Command("cmd", "/c", "start http://"+httpAddr+httpListPattern)
	err := cmd.Run()
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}

func startReplay(replayName string) {
	log.Printf("startReplay: %s\n", replayName)
	cmd := exec.Command(war3Path+war3Exe, "-loadfile", war3Path+replaySavePath+replayName)
	err := cmd.Run()
	if err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}
