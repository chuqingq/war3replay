package main

import (
	"fmt"
	"regexp"
)

func main() {
	src := `<li>地图位置：<span id="ctl00_Content_labMapname">Maps\e-ewcl\1V1\(2)EchoIsles.w3x</span></li><li>使用地图：<span id="ctl00_Content_labUsemap">EchoIsles<a id='downmap' href="javascript:getreplaymap(26,'815248024','EE3A09BD2ABE4DD293E40D9D808AC564')" class='rnmap'>下载录像地图</a></span></li><li>游戏名称：<span id="ctl00_Content_labName">当地局域网内的游戏 (si</span></li><li>游戏类型：<span id="ctl00_Content_labType">custom</span></li><li>游戏时间：<span id="ctl00_Content_labTime">18:02</span></li><li>游戏人数：<span id="ctl00_Content_labUserno">2/12</span></li><li>游戏版本：<span id="ctl00_Content_labVer">1.26</span></li><li>更新时间：<span id="ctl00_Content_labUpdate">2014-11-8 21:03:36</span></li><li><span class="partbtn"><a href="javascript:showwinner();">点击查看胜负</a><a href="javascript:nicegame()">很好,很强大</a></span></li></ul>`
	re := regexp.MustCompile(`<span id="ctl00_Content_labMapname">([^<]*)</span>`)
	res := re.FindAllString(src, 1)
	fmt.Printf("%+v\n", res)

	// src := `<span id="ctl00_Content_labMapname">Maps\e-ewcl\1V1\(2)EchoIsles.w3x</span>`
	// re := regexp.MustCompile(`<span id="ctl00_Content_labMapname">(.*)</span>`)
	// name := re.ReplaceAllString(src, "$1")
	// println(name)

	// src := "/Download.aspx?ReplayID=62271&File=/ReplayFile/2014-11-8/141108_[NE]_VS_[UD]Lusiria_EchoIsles_RN.w3g"
	// re := regexp.MustCompile(`/Download.aspx\?ReplayID=.*&File=/ReplayFile/.*/(.*)`)
	// name := re.ReplaceAllString(src, "$1")
	// println(name)

	// src := "hello world!\nhello1 chuqq!\n1234567890"
	// re := regexp.MustCompile("(?m:hello1 (.*)qq!)")
	// // res := re.FindString(src)
	// res := re.FindString(src)
	// println("=" + res)
	// res = re.ReplaceAllString(res, "$1")
	// println("=" + res)
}

// func test() {
// 	rep := `<ul class="replayitem" id="rep_62257" onclick="showrepdata(this)" style="background-image: none;">
// }
// <li class="c_r"><a href="http://w3g.replays.net/replaylist.aspx?Gamerace=9">UD vs HM</a></li>
// <li class="c_p"><a href="http://w3g.replays.net/doc/cn/2014-11-7/14153367620713574252.html" target="_blank">[妖妖杯]Yumiko vs Xiaokai #2</a> (0)  </li>
// <li class="c_m">LastRefuge1.3</li>
// <li class="c_t">11-07</li>
// <li class="c_d" title="524" onclick="window.location.href='http://w3g.replays.net/doc/cn/2014-11-7/14153367620713574252.html';"><span style="width:2%"></span></li></ul>`

// 	replayitemRe := regexp.MustCompile(`<ul class="replayitem" id="rep_62257" onclick="showrepdata(this)" style="background-image: none;">`)
// 	res := replayitemRe.ReplaceAllString(rep, "$1")

// 	raceRe := regexp.MustCompile(`<li class="c_r"><a href=".*">(.*)</a></li>`)
// 	res = raceRe.ReplaceAllString(res, "race=$1,")

// 	playerRe := regexp.MustCompile(`<li class="c_p"><a href="(.*)" target="_blank">(.*)</li>`)
// 	res = playerRe.ReplaceAllString(res, "player=$2,link=$1,")

// 	mapRe := regexp.MustCompile(`<li class="c_m">(.*)</li>`)
// 	res = mapRe.ReplaceAllString(res, "map=$1,")

// 	dateRe := regexp.MustCompile(`<li class="c_t">(.*)</li>`)
// 	res = dateRe.ReplaceAllString(res, "date=$1\n")

// 	linkRe := regexp.MustCompile(`<li class="c_d" title="524" onclick="window.location.href='(.*)';"><span style="width:2%"></span></li></ul>`)
// 	res = linkRe.ReplaceAllString(res, "")

// 	fmt.Println(res)
// }
