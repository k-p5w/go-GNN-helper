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
}

type ColorInfo struct {
	StrokeColor string // 事務所カラー
	TextColor   string // 文字色
}

const (
	ProductionKey      = "所属"
	StartYEAR          = "結成"
	StartYEAR2         = "結成年"
	StartYEAR3         = "活動開始"
	ScrapboxAccountURL = "https://scrapbox.io/api/table/lololololol/%s/account.csv"

	// 事務所カラー定数
	MSKcolor  = "#F39800"
	WTNBcolor = "#19a0c4"
	SMGcolor  = "#7fb2e5"
	TTNcolor  = "#f7b916"
	YSMTcolor = "#e94609"
	HRPCcolor = "#002e73"
	OTPcolor  = "#ff4c00"
	SMAcolor  = "#d9006c"
	JRKcolor  = "#eeeeee"
	SCGcolor  = "#d80c18"
	GRPcolor  = "#7e3f98"
	NLEcolor  = "#231815"
	KDcolor   = "#444444"
	ETCcolor  = "#cc44cc"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	svgname := q.Get("name")
	if len(svgname) == 0 || !strings.HasSuffix(svgname, ".svg") {
		return
	}

	actorName := strings.Replace(svgname, ".svg", "", -1)
	urlName := strings.ReplaceAll(actorName, " ", "_")
	gi := getTableAccount(urlName)

	nameLen := utf8.RuneCountInString(gi.Name)
	currentFontSize := 180
	if nameLen > 4 {
		currentFontSize = (850 / nameLen)
	}

	exp := float64(gi.PerformanceExperience)
	if exp < 1 {
		exp = 1
	}
	nStages := int(math.Log2(exp)) + 1
	if nStages > 7 {
		nStages = 7
	}

	rectWidth := (currentFontSize * nameLen) + 150
	rectHeight := 300
	canvasWidth, canvasHeight := rectWidth+50, rectHeight+50

	var rippleRings strings.Builder
	centerX, centerY := float64(canvasWidth+20), float64(canvasHeight/2)
	for i := 0; i < nStages; i++ {
		radius := float64(40 * int(math.Pow(2, float64(i))))
		opacity := 0.30 - (float64(i) * 0.04)
		if opacity < 0.05 {
			opacity = 0.05
		}
		rippleRings.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="3" stroke-opacity="%.2f" />`,
			centerX, centerY, radius, gi.ProductionColor.StrokeColor, opacity,
		))
	}

	// レイヤ構成の組み立て
	var textLayers string
	if gi.ProductionName == "太田プロダクション" {
		// 【太田プロ特別仕様】文字:黒、ふち:白→事務所色→黒
		textLayers = fmt.Sprintf(`
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            stroke="black" stroke-width="26" stroke-linejoin="round" paint-order="stroke" fill="black">%s</text>
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            stroke="%s" stroke-width="18" stroke-linejoin="round" paint-order="stroke" fill="%s">%s</text>
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            stroke="white" stroke-width="10" stroke-linejoin="round" paint-order="stroke" fill="white">%s</text>
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            fill="#000000">%s</text>`,
			gi.Name, gi.ProductionColor.StrokeColor, gi.ProductionColor.StrokeColor, gi.Name, gi.Name, gi.Name)
	} else {
		// 【通常仕様】文字:事務所色、ふち:白→黒
		textLayers = fmt.Sprintf(`
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            stroke="black" stroke-width="22" stroke-linejoin="round" paint-order="stroke" fill="black">%s</text>
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            stroke="white" stroke-width="12" stroke-linejoin="round" paint-order="stroke" fill="white">%s</text>
        <text x="45%%" y="50%%" text-anchor="middle" dominant-baseline="central" 
            fill="%s">%s</text>`,
			gi.Name, gi.Name, gi.ProductionColor.TextColor, gi.Name)
	}

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
    <g clip-path="url(#frameClip)">%s</g>
    <g style="font-family:'Hiragino Kaku Gothic ProN','Meiryo',sans-serif; font-size:%dpx; font-weight:900; filter:url(#strongShadow);">
        %s
    </g>
</svg>`,
		canvasWidth, canvasHeight, canvasWidth, canvasHeight,
		canvasWidth, canvasHeight,
		rippleRings.String(),
		currentFontSize,
		textLayers,
	)

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	fmt.Fprint(w, svgPage)
}

func getTableAccount(t string) gnnInfo {
	var gi gnnInfo
	gi.PerformanceExperience = 1
	u := fmt.Sprintf(ScrapboxAccountURL, t)
	res, err := http.Get(u)
	if err != nil {
		return gi
	}
	defer res.Body.Close()
	reader := csv.NewReader(res.Body)
	records, _ := reader.ReadAll()
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		tblKey := strings.Trim(strings.ToLower(record[0]), "[]\ufeff ")
		tblValue := strings.Trim(record[1], "[]\ufeff ")
		switch tblKey {
		case ProductionKey:
			gi.ProductionName = tblValue
		case StartYEAR, StartYEAR2, StartYEAR3:
			re := regexp.MustCompile(`([0-9]+)年`)
			matches := re.FindStringSubmatch(tblValue)
			if len(matches) > 1 {
				debutYear, _ := strconv.Atoi(matches[1])
				gi.PerformanceExperience = time.Now().Year() - debutYear
			}
		}
	}
	gi.Name = width.Widen.String(strings.ReplaceAll(t, "_", " "))
	gi.ProductionColor = getProductionColor(gi.ProductionName)
	return gi
}

func getProductionColor(name string) ColorInfo {
	itemColor := ETCcolor
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
	}
	return ColorInfo{StrokeColor: itemColor, TextColor: itemColor}
}
