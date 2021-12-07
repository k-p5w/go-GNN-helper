package table2svg

import (
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/width"
)

// gnnInfo is 芸人info
type gnnInfo struct {
	Name                  string
	ProductionColor       ColorInfo
	ProductionName        string
	StartYear             string
	PerformanceExperience int
	SNSAccount            []string
}

type ColorInfo struct {
	BaseColor          string
	ComplementaryColor string
	InvertColor        string
}
type RGB struct {
	R, G, B float64
}

// YMDstring is 年月日取り出し用の正規表現
const YMDstring = `([0-9]+)年`
const layoutYYYY = "2006"

// CANVAS向け定数
const (
	FrameXY        = 40
	FrameRoundness = 20
	FrameBase      = 200
	FontSize       = 150
	FrameTextXY    = 1300
	FrameHeight    = FontSize + 100
	TextBaseX      = 30
	TextBaseY      = 220
)

const (
	ProductionKey      = "所属"
	StartYEAR          = "結成"
	StartYEAR2         = "結成年"
	ScrapboxAccountURL = "https://scrapbox.io/api/table/lololololol/%s/account.csv"
	SiteTypeTwitter    = "twitter"
	MSKcolor           = "#F39800"
	WTNBcolor          = "#19a0c4"
	SMGcolor           = "#7fb2e5"
	TTNcolor           = "#f7b916"
	YSMTcolor          = "#e94609"
	HRPCcolor          = "#002e73"
	OTPcolor           = "#ff4c00"
	SMAcolor           = "#d9006c"
	JRKcolor           = "#eeeeee"
	SCGcolor           = "#d80c18"
	GRPcolor           = "#7e3f98"
	NLEcolor           = "#231815"
	ETCcolor           = "#505050"
)

// Handler is /APIから呼ばれる
func Handler(w http.ResponseWriter, r *http.Request) {

	// getパラメータの解析
	q := r.URL.Query()
	svgname := q.Get("name")
	// スペースだった場合は、"_"が設定されてくる

	fmt.Println(q)
	if len(svgname) == 0 {
		return
	}
	actorName := ""
	// SVGで終わっていること
	if strings.HasSuffix(svgname, ".svg") {
		actorName = strings.Replace(svgname, ".svg", "", -1)
		// actorName = filepath.Base(svgname)
		fmt.Printf("%v => %v", svgname, actorName)
	} else {
		return
	}

	fmt.Println(actorName)
	// テーブルからデータの取り出し
	// " "であれば"_"に置換する
	urlName := strings.ReplaceAll(actorName, " ", "_")

	gi := getTableAccount(urlName)

	// フォントサイズの導出
	nameLen := utf8.RuneCountInString(gi.Name)
	frameWidth := FontSize * nameLen
	// canvasText := canvasBase / 2
	canvasFont := (FontSize / nameLen) / 3
	fmt.Printf("[%v]len=%v,size=%v \n", gi.Name, nameLen, canvasFont)
	// 芸歴

	// circle
	TextShadowX := TextBaseX + 10
	TextShadowY := TextBaseY + 5
	canvasWidth := frameWidth + 100
	canvasHeight := FrameHeight + 80
	svgPage := fmt.Sprintf(`
	<svg width="%v" height="%v" xmlns="http://www.w3.org/2000/svg" 		xmlns:xlink="http://www.w3.org/1999/xlink"		>
		<rect x="%v" y="%v" rx="%v" ry="%v" width="%v" 	height="%v" 			stroke="%v" 			fill="transparent" stroke-width="%v" 			/>
		<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;"			>			
		%v
		</text>
		<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:RGB(2,2,2);font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;">
        %v
    	</text>
	</svg>
	`, canvasWidth, canvasHeight,
		FrameXY, FrameXY, gi.PerformanceExperience, gi.PerformanceExperience, frameWidth, FrameHeight,
		gi.ProductionColor.BaseColor,
		gi.PerformanceExperience*3,
		TextShadowX, TextShadowY, FontSize,
		gi.ProductionColor.InvertColor,
		gi.Name,
		TextBaseX, TextBaseY, FontSize, gi.Name)
	fmt.Println(gi)
	// Content-Type: image/svg+xml
	// Vary: Accept-Encoding
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")

	fmt.Fprint(w, svgPage)
}

// getTableAccount is テーブルから情報を取得する
func getTableAccount(t string) gnnInfo {
	ids := make([]string, 0)
	var gi gnnInfo
	gi.PerformanceExperience = 1
	u := fmt.Sprintf(ScrapboxAccountURL, t)
	// URLにアクセスしデータを取得する
	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(res.Body)
	reader.Comma = ','
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("表形式データの取得に失敗しました。詳細（以下CSVファイルの中身が想定外です.)：\n %v \n\n", u)
		panic(err)
		// return nil
	}
	for ii := 0; ii < len(records); ii++ {

		// 種類
		itemType := strings.ToLower(records[ii][0])

		regReplace := func(str string) string {
			rep := regexp.MustCompile(`\?.*`)
			ret := rep.ReplaceAllString(str, "")

			return ret
		}

		linkReplace := func(str string) string {
			newstr := strings.Replace(str, "[", "", -1)
			ret := strings.Replace(newstr, "]", "", -1)
			return ret
		}

		// ID
		tblValue := records[ii][1]
		tblValue = regReplace(tblValue)
		// [|]を除外して、Twitterになっていれば、IDを取得する
		newType := strings.Replace(itemType, "[", "", -1)
		tblKey := strings.Replace(newType, "]", "", -1)
		twtId := ""
		idLen := len(tblValue)
		switch tblKey {
		case SiteTypeTwitter:
			//IDがURL全部なら、IDのみにする
			urlPos := strings.LastIndex(tblValue, "/")
			if urlPos > 0 {

				twtId = tblValue[urlPos+1 : idLen]
				ids = append(ids, twtId)
			} else {
				twtId = tblValue
				ids = append(ids, twtId)
			}

		case ProductionKey:
			// 事務所を取得する

			if idLen > 0 {
				gi.ProductionName = linkReplace(tblValue)
			}
		case StartYEAR, StartYEAR2:
			// 結成年を取得する
			performanceExperience := linkReplace(tblValue)
			fmt.Printf("%v=>%v \n", StartYEAR, gi.StartYear)
			ret := ""
			re := regexp.MustCompile(YMDstring)
			val := re.FindStringSubmatch(performanceExperience)
			// fmt.Printf("----- %v ----- \n", regval)
			if len(val) > 1 {
				// fmt.Println(val[1])
				ret = val[1]

			}

			str2num := func(s string) int {
				v, err := strconv.Atoi(s)
				if err != nil {
					v = 2020
				}
				return v
			}
			day := time.Now()
			debutYear := str2num(ret)

			nowYear := day.Format(layoutYYYY)
			fmt.Printf("%v-%v", nowYear, debutYear)
			gi.PerformanceExperience = str2num(nowYear) - debutYear
			gi.StartYear = performanceExperience
			fmt.Println(gi.PerformanceExperience)
		default:
			fmt.Printf("%v,%v\n", tblKey, tblValue)
		}

	}

	// 分解する
	// もしIDがなかったらエラーにする
	if len(ids) == 0 {
		fmt.Println("検索対象のアカウントIDの取得に失敗しました")
		// panic(err)
	} else {
		fmt.Printf("getAccount//%s:%v \n", t, len(ids))
	}
	orgName := strings.ReplaceAll(t, "_", " ")
	gi.Name = width.Widen.String(orgName)
	gi.ProductionColor = getProductionColor(gi.ProductionName)
	gi.SNSAccount = ids

	fmt.Println(gi)
	return gi
}

func getProductionColor(name string) ColorInfo {

	itemColor := ""
	switch name {
	case "ワタナベエンターテインメント":
		itemColor = WTNBcolor
	case "サンミュージック":
		itemColor = SMGcolor
	case "タイタン":
		itemColor = TTNcolor
	case "吉本興業":
		itemColor = YSMTcolor
	case "マセキ芸能社":
		itemColor = MSKcolor
	case "ホリプロコム":
		itemColor = HRPCcolor
	case "太田プロダクション":
		itemColor = OTPcolor
	case "SMA":
		itemColor = SMAcolor
	case "プロダクション人力舎":
		itemColor = JRKcolor
	case "松竹芸能":
		itemColor = SCGcolor
	case "グレープカンパニー":
		itemColor = GRPcolor
	case "ナチュラルエイト":
		itemColor = NLEcolor

	default:
		// 対応できていないとき
		itemColor = ETCcolor
	}

	ret := getColorPallet(itemColor)

	return ret
}

// getColorPallet is 補色や反対色を取得するメソッド
func getColorPallet(c string) ColorInfo {

	var cp ColorInfo

	// 16進数→10進数
	rPt, _ := strconv.ParseUint(c[1:3], 16, 0)
	gPt, _ := strconv.ParseUint(c[3:5], 16, 0)
	bPt, _ := strconv.ParseUint(c[5:7], 16, 0)

	// int->float
	r := float64(rPt)
	g := float64(gPt)
	b := float64(bPt)

	min := math.Min(r, math.Min(g, b)) //Min. value of RGB
	max := math.Max(r, math.Max(g, b)) //Max. value of RGB
	pt := max + min                    //Delta RGB value

	newColorRGB := &RGB{pt - r, pt - g, pt - b}
	newColorRGB2 := &RGB{255 - r, 255 - g, 255 - b}

	// // float->int
	newR := int(newColorRGB.R)
	newG := int(newColorRGB.G)
	newB := int(newColorRGB.B)

	newR2 := int(newColorRGB2.R)
	newG2 := int(newColorRGB2.G)
	newB2 := int(newColorRGB2.B)

	cp.BaseColor = c
	cp.ComplementaryColor = fmt.Sprintf("RGB(%v,%v,%v)", newR, newG, newB)
	cp.InvertColor = fmt.Sprintf("RGB(%v,%v,%v)", newR2, newG2, newB2)
	return cp
}
