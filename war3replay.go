package main

// 在线一键看war3录像
// go build -ldflags "-s -w -H windowsgui"

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

// const war3Path = "D:/GAME/Warcraft III/"
const war3Path = "./"
const war3Exe = "Frozen Throne.exe"
const replaySavePath = "replay/"

const httpAddr = "127.0.0.1:28080"
const httpListPattern = "/list"
const httpReplayPattern = "/replay"

func main() {
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/replay", replayHandler)

	http.HandleFunc("/locallist", localListHandler)
	http.HandleFunc("/localreplay", localReplayHandler)

	go startBrowser()

	log.Printf("listen at %s ...\n", httpAddr+httpListPattern)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

// 展示页面
func listHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("== list")

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
	                <td><a href="%s" target="_blank">L</a></td>
	                <td><a href="javascript:action('replay', '%s');">R</a></td></td>
	                <td><a href="javascript:action('download', '%s');">D</a></td></td>
	            </tr>
            `, rep.Date, rep.Race, rep.Player, rep.Map, rep.Link, rep.Link, rep.Link)
	}
	// <td><a href="/list?action=replay&link=%s">replay</a></td></td>
	// 展示
	response := fmt.Sprintf(`
            <html>
                <head>
                    <script src="http://lib.sinaapp.com/js/jquery/1.9.1/jquery-1.9.1.min.js"></script>
                    <style type="text/css">
                        table{width: 100%%;}
                        table,th,td{font-family: Consolas; font-size: 12px; border-collapse: collapse; border: #BBBBBB solid  1px;}
                        a{font-family: Consolas; font-size: 12px; }
                    </style>
                    <script>
                        function action(action, link) {
                            $.ajax({
                                url: "/"+action+"?link="+link
                            });
                        }
                    </script>
                </head>
                <body>
                	<a href="/locallist" target="_blank">local replays</a>
                    <table>
                      <thead><tr>
                        <th>Date</th>
                        <th>Race</th>
                        <th>Player</th>
                        <th>Map</th>
                        <th></th>
                        <th></th>
                        <th></th>
                      </tr></thead>
                      <tbody>%s</tbody>
                    </table>
                </body>
            </html>
        `, repbody)

	w.Write([]byte(response))
}

// 下载replay
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	getRep(r.FormValue("link"), false)
}

// 播放replay
func replayHandler(w http.ResponseWriter, r *http.Request) {
	getRep(r.FormValue("link"), true)
}

// 列出本地replay
func localListHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("== locallist")

	replist := getLocalReplays()

	// 组装页面内容
	repbody := ""
	for _, rep := range replist {
		repbody += fmt.Sprintf(`
	            <tr>
	                <td>%s</td>
	                <td><a href="javascript:action('localreplay', '%s');">R</a></td></td>
	            </tr>
            `, rep, rep)
	}
	// <td><a href="/list?action=replay&link=%s">replay</a></td></td>
	// 展示
	response := fmt.Sprintf(`
            <html>
                <head>
                    <script src="http://lib.sinaapp.com/js/jquery/1.9.1/jquery-1.9.1.min.js"></script>
                    <style type="text/css">
                        table{width: 100%%;}
                        table,th,td{font-family: Consolas; font-size: 12px; border-collapse: collapse; border: #BBBBBB solid  1px;}
                    </style>
                    <script>
                        function action(action, rep) {
                            $.ajax({
                                url: "/"+action+"?rep="+rep
                            });
                        }
                    </script>
                </head>
                <body>
                    <table border="1">
                      <tr>
                        <th>LocalFile</th>
                        <th></th>
                      </tr>
                      %s
                    </table>
                </body>
            </html>
    `, repbody)

	w.Write([]byte(response))
}

func localReplayHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("== localReplayHandler: %v\n", r.FormValue("rep"))
	startReplay(r.FormValue("rep"))
}

func getLocalReplays() []string {
	fileInfos, err := ioutil.ReadDir(replaySavePath)
	if err != nil {
		log.Printf("ReadDir error: %v\n", err)
		return nil
	}

	res := make([]string, 0)
	for _, file := range fileInfos {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".w3g") {
			res = append(res, file.Name())
		}
	}
	return res
}

//-------------------------------------------------------------------------------

func getRep(link string, replay bool) error {
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("reading repinfo body...")
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("reading repinfo body ok")

	content := string(buf)

	// 下载replay
	log.Println("====  download replay")
	replayPath := reFindAndReplaceAll(content,
		`<span id="ctl00_Content_labDown" class="download"><a href="(.*)">Download REP</a></span>`,
		"$1")
	//replayPath, err = url.QueryEscape(replayPath)
	//if err != nil {
	//    return err
	//}
	log.Printf("replayPath=%s\n", replayPath)
	//replayName := reReplaceAll(replayPath, `/Download.aspx\?ReplayID=.*&File=/ReplayFile/.*/(.*)`, "$1")
	replayName := reReplaceAll(replayPath, `/Download.aspx\?ReplayID=.*&File=%2fReplayFile%2f.*%2f(.*)`, "$1")
	replayName, err = url.QueryUnescape(replayName)
	if err != nil {
		return err
	}
	log.Printf("replayName=%s\n", replayName)

	// 如果replayName不存在，再下载
	replaySaveAbsPath := war3Path + replaySavePath + replayName
	_, err = os.Stat(replaySaveAbsPath)
	if err != nil && !os.IsExist(err) {
		respRep, err := http.Get("http://w3g.replays.net" + replayPath)
		if err != nil {
			return err
		}
		defer respRep.Body.Close()

		log.Println("reading rep body...")
		buf, err = ioutil.ReadAll(respRep.Body)
		if err != nil {
			return err
		}
		log.Println("reading rep body...")

		log.Printf("write replay file: %v\n", replaySaveAbsPath)
		err = ioutil.WriteFile(replaySaveAbsPath, buf, os.ModePerm)
		if err != nil {
			log.Printf("write replay file error: %v\n", err)
			return err
		}
	} else {
		log.Println("replay file already exists")
	}

	// 下载地图
	log.Println("==== download map")
	mapPath := reFindAndReplaceAll(content, `<span id="ctl00_Content_labMapname">([^<]*)</span>`, "$1")
	mapPath = strings.Replace(mapPath, "\\", "/", -1)
	log.Printf("mappath=%s\n", mapPath)

	mapAbsPath := war3Path + mapPath
	log.Printf("mapAbsPath=%s\n", mapAbsPath)

	// 获取本地地图的大小
	var localMapSize int64 = 0
	mapInfo, err := os.Stat(mapAbsPath)
	if err == nil {
		localMapSize = mapInfo.Size()
	}

	ind := strings.LastIndex(mapPath, "/")
	mapName := mapPath[ind+1:]
	log.Printf("mapName=%s\n", mapName)

	downPath := reFindAndReplaceAll(content, `javascript:getreplaymap\(.*,'(.*)','.*'\)`, "$1")
	log.Printf("downPath=%s\n", downPath)

	mapPathAbs := "http://w3g.replays.net/ReplayMap/download/" + downPath + "/" + mapName
	log.Printf("mapPathAbs=%s\n", mapPathAbs)

	respMap, err := http.Get(mapPathAbs)
	if err != nil {
		return err
	}
	defer respMap.Body.Close()

	// 如果服务器地图和本地大小不一致，再保存
	if respMap.ContentLength != localMapSize {
		log.Printf("map file different: local=%v, remote=%v\n", localMapSize, respMap.ContentLength)
		buf, err = ioutil.ReadAll(respMap.Body)
		if err != nil {
			return err
		}
		log.Println("reading map body ok")

		// 确认目录已存在
		ind = strings.LastIndex(mapAbsPath, "/")
		mapDir := mapAbsPath[:ind]
		log.Printf("mapDir2=%s\n", mapDir)
		err = os.MkdirAll(mapDir, 0777)
		if err != nil {
			return err
		}

		log.Printf("write map file: %v\n", mapAbsPath)
		err = ioutil.WriteFile(mapAbsPath, buf, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		log.Println("map file already exists")
	}

	if replay {
		startReplay(replayName)
	}

	return nil
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

	const left = `<ul class="datarow2">`
	const right = `<span id="ctl00_Content_labPage" class="cutpage">`
	var content2 *string
	strArr := strings.Split(content, left)
	if len(strArr) <= 1 {
		return make([]*repentry, 0)
	}

	content2 = &strArr[1]
	strArr2 := strings.Split(*content2, right)
	content2 = &strArr2[0]

	res := *content2

	res = reReplaceAll(res, `<li class="c_r"><a href=".*">(.*)</a></li>\r\n`, "$1|")
	res = reReplaceAll(res, `<li class="c_p"><a href="(.*)" target="_blank">(.*)</li>\r\n`, "$2|$1|")
	res = reReplaceAll(res, `<li class="c_m">(.*)</li>\r\n`, "$1|")
	res = reReplaceAll(res, `<li class="c_t">(.*)</li>\r\n`, "$1\n")
	res = reReplaceAll(res, `<(.*)>\r\n`, "")

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

// 在str中取出正则reStr，然后替换成replace
func reFindAndReplaceAll(str string, reStr string, replace string) string {
	re := regexp.MustCompile(reStr)
	res := re.FindString(str)
	return re.ReplaceAllString(res, replace)
}

// 直接把str中匹配正则reStr的替换成replace。不先取出。
func reReplaceAll(str string, reStr string, replace string) string {
	re := regexp.MustCompile(reStr)
	return re.ReplaceAllString(str, replace)
}
