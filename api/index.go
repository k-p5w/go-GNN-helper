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
	StartYEAR3         = "活動開始"
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
	ETCcolor           = "#cc44cc"
	KDcolor            = "#444444"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("start Ripple Stump UD Handler")

	// 1. パラメータの解析
	q := r.URL.Query()
	svgname := q.Get("name")
	if len(svgname) == 0 || !strings.HasSuffix(svgname, ".svg") {
		return
	}

	actorName := strings.Replace(svgname, ".svg", "", -1)
	urlName := strings.ReplaceAll(actorName, " ", "_")

	// 2. データの取得
	gi := getTableAccount(urlName)

	// 3. レイアウト計算
	nameLen := utf8.RuneCountInString(gi.Name)
	currentFontSize := 180
	if nameLen > 4 {
		currentFontSize = (850 / nameLen)
	}

	// --- 4. 年輪（Ripple）の計算 ---
	exp := float64(gi.PerformanceExperience)
	if exp < 1 {
		exp = 1
	}
	nStages := int(math.Log2(exp)) + 1
	if nStages > 7 {
		nStages = 7
	}

	// キャンバスサイズ
	rectWidth := (currentFontSize * nameLen) + 150
	rectHeight := 300
	canvasWidth := rectWidth + 50
	canvasHeight := rectHeight + 50

	baseColor := gi.ProductionColor.BaseColor

	// --- 5. 【修正】右端を中心として左へ広がる波紋年輪 ---
	var rippleRings strings.Builder
	// 中心をキャンバスの「右端の少し外」に置くことで、より綺麗な曲線が名前にかかります
	centerX := float64(canvasWidth + 20)
	centerY := float64(canvasHeight / 2)

	for i := 0; i < nStages; i++ {
		// 半径を等比数列で広げる。名前まで届くように初期値を調整
		// 1段階目から左へ向かって大きく広がっていく
		radius := float64(40 * int(math.Pow(2, float64(i))))

		opacity := 0.30 - (float64(i) * 0.04)
		if opacity < 0.05 {
			opacity = 0.05
		}

		// 右端を中心とした円（の左半分側が見える状態）を描画
		rippleRings.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" 
			fill="none" stroke="%s" stroke-width="3" stroke-opacity="%.2f" />
			`,
			centerX, centerY, radius, baseColor, opacity,
		))
	}

	// 6. SVG組み立て
	svgPage := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" style="background-color: transparent;">
	<defs>
		<filter id="strongShadow" x="-30%%" y="-30%%" width="160%%" height="160%%">
			<feDropShadow dx="0" dy="4" stdDeviation="6" flood-opacity="0.8"/>
		</filter>
		<clipPath id="frameClip">
			<rect x="0" y="0" width="%d" height="%d" rx="20" ry="20" />
		</clipPath>
	</defs>

	<g clip-path="url(#frameClip)">
		%s
	</g>

	<g style="font-family:'Hiragino Kaku Gothic ProN','Meiryo',sans-serif; font-size:%dpx; font-weight:900; filter:url(#strongShadow);">
		<text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
			stroke="black" stroke-width="18" stroke-linejoin="round" paint-order="stroke" fill="black">%s</text>
		<text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
			stroke="white" stroke-width="11" stroke-linejoin="round" paint-order="stroke" fill="white">%s</text>
		<text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
			fill="%s">%s</text>
	</g>
</svg>`,
		canvasWidth, canvasHeight, canvasWidth, canvasHeight,
		canvasWidth, canvasHeight, // clipPath用
		rippleRings.String(),
		currentFontSize, gi.Name, gi.Name, baseColor, gi.Name,
	)

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")
	fmt.Fprint(w, svgPage)
}

// getTableAccount は年月とブラケットを考慮してデータを取得します
func getTableAccount(t string) gnnInfo {
	var gi gnnInfo
	gi.PerformanceExperience = 1
	u := fmt.Sprintf(ScrapboxAccountURL, t)
	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	reader := csv.NewReader(res.Body)
	records, _ := reader.ReadAll()

	for _, record := range records {
		tblKey := strings.Trim(strings.ToLower(record[0]), "[]\ufeff ")
		tblValue := strings.Trim(record[1], "[]\ufeff ")

		switch tblKey {
		case ProductionKey:
			gi.ProductionName = tblValue
		case StartYEAR, StartYEAR2, StartYEAR3:
			// [1994年11月] のような形式から数字を抽出
			re := regexp.MustCompile(`([0-9]+)年(?:([0-9]+)月)?`)
			matches := re.FindStringSubmatch(tblValue)
			if len(matches) > 1 {
				debutYear, _ := strconv.Atoi(matches[1])
				debutMonth := 1
				if len(matches) > 2 && matches[2] != "" {
					debutMonth, _ = strconv.Atoi(matches[2])
				}
				now := time.Now()
				// 月単位の精密な芸歴計算
				totalMonths := (now.Year()-debutYear)*12 + int(now.Month()) - debutMonth
				expFloat := float64(totalMonths) / 12.0
				if expFloat < 1 {
					expFloat = 1
				}
				// floatをintとして保持（Handler側でfloatとして計算に利用）
				gi.PerformanceExperience = int(expFloat)
				gi.StartYear = tblValue
			}
		}
	}
	orgName := strings.ReplaceAll(t, "_", " ")
	gi.Name = width.Widen.String(orgName)
	gi.ProductionColor = getProductionColor(gi.ProductionName)
	return gi
}

func getProductionColor(name string) ColorInfo {

	fmt.Printf("getProductionColor//%v \n", name)

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
	case "ケイダッシュステージ":
		itemColor = KDcolor
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
