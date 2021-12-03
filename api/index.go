package table2svg

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const (
	ScrapboxAccountURL = "https://scrapbox.io/api/table/lololololol/%s/account.csv"
	SiteTypeTwitter    = "twitter"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	// getパラメータの解析
	q := r.URL.Query()
	actorName := q.Get("name")
	getAccount(actorName)
	productionColor := "red"
	svgPage := fmt.Sprintf(`
	<svg width="500" height="500" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
    <circle cx="250" cy="250" r="100" />
    <text x="250" y="250" style="text-anchor:middle;font-size:30px;fill:%v">%v</text>
</svg>
	`, productionColor, actorName)

	// Content-Type: image/svg+xml
	// Vary: Accept-Encoding
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")

	fmt.Fprint(w, svgPage)
}

func getAccount(t string) []string {
	ids := make([]string, 0)

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
		// panic(err)
		return nil
	}
	for ii := 0; ii < len(records); ii++ {

		// 種類
		itemType := strings.ToLower(records[ii][0])

		regReplace := func(str string) string {
			rep := regexp.MustCompile(`\?.*`)
			ret := rep.ReplaceAllString(str, "")

			return ret
		}

		// ID
		itemID := records[ii][1]
		itemID = regReplace(itemID)
		// [|]を除外して、Twitterになっていれば、IDを取得する
		newType := strings.Replace(itemType, "[", "", -1)
		idType := strings.Replace(newType, "]", "", -1)
		twtId := ""
		idLen := len(itemID)
		// twitterが指定された場合
		if idType == SiteTypeTwitter {
			//IDがURL全部なら、IDのみにする
			urlPos := strings.LastIndex(itemID, "/")
			if urlPos > 0 {

				twtId = itemID[urlPos+1 : idLen]
				ids = append(ids, twtId)
			} else {
				twtId = itemID
				ids = append(ids, twtId)
			}

			//配列につめる
		} else {
			// Twitter以外の場合でも、IDのところがTwitterを示していれば取得する
			fmt.Println(idType)
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

	return ids
}
