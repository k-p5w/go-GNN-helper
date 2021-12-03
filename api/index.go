package table2svg

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// gnnInfo is 芸人info
type gnnInfo struct {
	Name            string
	ProductionColor string
	ProductionName  string
	StartYear       string
	SNSAccount      []string
}

const (
	ProductionKey      = "所属"
	StartYEAR          = "結成"
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
	JRKcolor           = "#222021"
)

// Handler is /APIから呼ばれる
func Handler(w http.ResponseWriter, r *http.Request) {

	// getパラメータの解析
	q := r.URL.Query()
	actorName := q.Get("name")

	if len(actorName) == 0 {
		return
	}
	gi := getTableAccount(actorName)

	svgPage := fmt.Sprintf(`
	<svg width="500" height="500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <circle  style="fill:%v" cx="250" cy="250" r="100" />
    <text x="250" y="250" style="text-anchor:middle;font-size:30px;fill:black">%v</text>
</svg>
	`, gi.ProductionColor, gi.Name)

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
		case StartYEAR:
			// 結成年を取得する
			gi.StartYear = linkReplace(tblValue)
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

	gi.Name = t
	gi.ProductionColor = getProductionColor(gi.ProductionName)
	gi.SNSAccount = ids

	fmt.Println(gi)
	return gi
}

func getProductionColor(name string) string {

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
	default:
		// 対応できていないとき
		itemColor = "#808080"
	}

	return itemColor
}
