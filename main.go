package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	table2svg "github.com/k-p5w/go-table2svg/api"
	"golang.org/x/oauth2"
)

type PodcastEpisodeItem struct {
	AudioPreviewURL string `json:"audio_preview_url"`
	Description     string `json:"description"`
	HTMLDescription string `json:"html_description"`
	DurationMs      int    `json:"duration_ms"`
	Explicit        bool   `json:"explicit"`
	ExternalUrls    struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href   string `json:"href"`
	ID     string `json:"id"`
	Images []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	IsExternallyHosted   bool     `json:"is_externally_hosted"`
	IsPlayable           bool     `json:"is_playable"`
	Language             string   `json:"language"`
	Languages            []string `json:"languages"`
	Name                 string   `json:"name"`
	ReleaseDate          string   `json:"release_date"`
	ReleaseDatePrecision string   `json:"release_date_precision"`
	ResumePoint          struct {
		FullyPlayed      bool `json:"fully_played"`
		ResumePositionMs int  `json:"resume_position_ms"`
	} `json:"resume_point"`
	Type         string `json:"type"`
	URI          string `json:"uri"`
	Restrictions struct {
		Reason string `json:"reason"`
	} `json:"restrictions"`
	Show struct {
		AvailableMarkets []string `json:"available_markets"`
		Copyrights       []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"copyrights"`
		Description     string `json:"description"`
		HTMLDescription string `json:"html_description"`
		Explicit        bool   `json:"explicit"`
		ExternalUrls    struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			URL    string `json:"url"`
			Height int    `json:"height"`
			Width  int    `json:"width"`
		} `json:"images"`
		IsExternallyHosted bool     `json:"is_externally_hosted"`
		Languages          []string `json:"languages"`
		MediaType          string   `json:"media_type"`
		Name               string   `json:"name"`
		Publisher          string   `json:"publisher"`
		Type               string   `json:"type"`
		URI                string   `json:"uri"`
		TotalEpisodes      int      `json:"total_episodes"`
	} `json:"show"`
}

type PodcastEpisodeResponse struct {
	/*
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
	*/
	Items []struct {
		AudioPreviewURL string `json:"audio_preview_url"`
		Description     string `json:"description"`
		HTMLDescription string `json:"html_description"`
		DurationMs      int    `json:"duration_ms"`
		Explicit        bool   `json:"explicit"`
		ExternalUrls    struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			URL    string `json:"url"`
			Height int    `json:"height"`
			Width  int    `json:"width"`
		} `json:"images"`
		IsExternallyHosted   bool     `json:"is_externally_hosted"`
		IsPlayable           bool     `json:"is_playable"`
		Language             string   `json:"language"`
		Languages            []string `json:"languages"`
		Name                 string   `json:"name"`
		ReleaseDate          string   `json:"release_date"`
		ReleaseDatePrecision string   `json:"release_date_precision"`
		ResumePoint          struct {
			FullyPlayed      bool `json:"fully_played"`
			ResumePositionMs int  `json:"resume_position_ms"`
		} `json:"resume_point"`
		Type         string `json:"type"`
		URI          string `json:"uri"`
		Restrictions struct {
			Reason string `json:"reason"`
		} `json:"restrictions"`
		Show struct {
			AvailableMarkets []string `json:"available_markets"`
			Copyrights       []struct {
				Text string `json:"text"`
				Type string `json:"type"`
			} `json:"copyrights"`
			Description     string `json:"description"`
			HTMLDescription string `json:"html_description"`
			Explicit        bool   `json:"explicit"`
			ExternalUrls    struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				URL    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			IsExternallyHosted bool     `json:"is_externally_hosted"`
			Languages          []string `json:"languages"`
			MediaType          string   `json:"media_type"`
			Name               string   `json:"name"`
			Publisher          string   `json:"publisher"`
			Type               string   `json:"type"`
			URI                string   `json:"uri"`
			TotalEpisodes      int      `json:"total_episodes"`
		} `json:"show"`
	} `json:"items"`
}

// https://accounts.spotify.com/authorize?client_id=&response_type=code&redirect_uri=https://scrapbox.io/cybernote/&scope=user-read-currently-playing
const authorizeURL = "https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-currently-playing"
const accesstokenEndpoint = "https://accounts.spotify.com/api/token"

type Settings struct {
	ClientID    string `json:"clientID"`
	Secret      string `json:"clientSecret"`
	RedirectUri string `json:"redirectUri"`
	ShowID      string `json:"showID"`
	ShowName    string `json:"showName"`
}

// レスポンスボディをパースしてアクセストークンを取得
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func loadSetting() Settings {
	var ret Settings
	// setting.json ファイルを読み込む
	file, err := os.Open("setting.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ret
	}
	defer file.Close()

	// ファイル内容をバイトスライスとして読み込む
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ret
	}

	// 設定値を格納するための構造体を初期化
	settings := Settings{}

	// JSON デコードして設定値を取得
	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return ret
	}

	// 設定値を表示
	fmt.Println("ClientID:", settings.ClientID)
	fmt.Println("RedirectURL:", settings.RedirectUri)
	fmt.Println("Show ID:", settings.ShowID)

	return settings
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	fmt.Println("hey")

	// http.ListenAndServe(":"+port, nil)
	// debugのときはこれでファイアウォールの設定がでなくなるらしい
	http.HandleFunc("/", table2svg.Handler)
	http.ListenAndServe("localhost:"+port, nil)

	connOauth2()

	fmt.Println("end!")
	http.HandleFunc("/", table2svg.Handler)
	http.ListenAndServe("localhost:"+port, nil)

	/*
		// http.ListenAndServe(":"+port, nil)
		// debugのときはこれでファイアウォールの設定がでなくなるらしい
		http.ListenAndServe("localhost:"+port, nil)

	*/

}

func connOauth2() {
	fmt.Println("start-connOauth2!")
	jsonItem := loadSetting()
	// SpotifyのクライアントIDとクライアントシークレットを設定
	clientID := jsonItem.ClientID
	clientSecret := jsonItem.Secret
	redirectURL := jsonItem.RedirectUri

	// URLエンコードされたリダイレクトURLを生成
	encodedRedirectURL := url.QueryEscape(redirectURL)

	fmt.Printf("%v => %v \n", redirectURL, encodedRedirectURL)
	// OAuth2の設定を作成
	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user-read-private", "user-read-email"}, // 必要なスコープを指定
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	// 認可コードを取得するためのURLを生成
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("authURL:%v \n", authURL)
	// ユーザーに認可してもらうためにブラウザを開く
	fmt.Printf("Open the following URL in your browser and authorize the app:\n\n%s\n\n", authURL)

	// ユーザーが認可した後、コールバックを待ち受けるWebサーバを起動
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		// 認可コードを使用してトークンを取得
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		// トークンを表示
		fmt.Fprintf(w, "Access Token: %s\n", token.AccessToken)
		tokenItem := token.AccessToken
		fmt.Fprintf(w, "Refresh Token: %s\n", token.RefreshToken)
		getPodcastEpisodes(jsonItem, tokenItem)
		// サーバを終了
		// os.Exit(0)
	})

	// Webサーバを8080ポートで起動
	log.Fatal(http.ListenAndServe("localhost"+":8080", nil))

}

func getAccessToken() string {
	jsonItem := loadSetting()
	ret := ""
	// SpotifyアプリケーションのクライアントIDとクライアントシークレット
	clientID := jsonItem.ClientID
	clientSecret := jsonItem.Secret

	// 認可コードを取得するための認可エンドポイント
	authEndpoint := "https://accounts.spotify.com/authorize"

	/*
		// 認可エンドポイントのパラメータを設定
		params := url.Values{}
		params.Set("client_id", clientID)
		params.Set("response_type", "code")
		params.Set("redirect_uri", jsonItem.RedirectUri)
		params.Set("scope", "user-read-private user-read-email") // 必要なスコープを設定

		// 8c6f97a1d6dc4fe3a0ea0262a4c0d653
		// http://localhost:9999
		// https://accounts.spotify.com/authorize?client_id=8c6f97a1d6dc4fe3a0ea0262a4c0d653&redirect_uri=http://localhost:9999&response_type=code&scope=user-read-private+user-read-email
		// https://accounts.spotify.com/authorize?client_id=8c6f97a1d6dc4fe3a0ea0262a4c0d653&redirect_uri=http%3A%2F%2Flocalhost%3A9999&response_type=code&scope=user-read-private+user-read-email
		fmt.Println(params)
		// 認可URLを構築
		authURL := authEndpoint + "?" + params.Encode()
	*/
	authURL := fmt.Sprintf("?client_id=%v&response_type=%v&redirect_uri=%v", clientID, "code", jsonItem.RedirectUri)

	fmt.Println("Please visit the following URL to authorize the app:")
	fmt.Println(authEndpoint + authURL)
	// http.Redirect(w, r, authURL, http.StatusFound)

	// 認可コードを取得
	fmt.Print("Enter the code from the redirect URI: ")
	var authCode string
	fmt.Scanln(&authCode)
	// http://localhost:9999/?code=AQCb9ufhNjaHnfrXSI-72MEg3_yznMFeb1bwJqiEB_6x1aRalwBH9Sy37hXKDXoQ1YiCcBG-BxbP-RF835fQiHSBHgXUl9EV1jqlzf5dajbuALuJ5iWV4uKdW2GtXPzQc-LWAbjqtNy5tgcw0HIJ9Rh_VoktDmYeXQ
	// アクセストークンを取得するためのトークンエンドポイント

	// トークンエンドポイントに送信するリクエストボディを準備
	tokenParams := url.Values{}
	tokenParams.Set("grant_type", "client_credentials")
	tokenParams.Set("code", authCode)
	tokenParams.Set("redirect_uri", jsonItem.RedirectUri)

	// Basic認証用のヘッダーを設定
	authHeader := clientID + ":" + clientSecret
	fmt.Printf("encode! :%v \n", authHeader)
	encodedAuthHeader := "Basic " + base64Encode([]byte(authHeader))

	fmt.Println("トークンエンドポイントへのリクエストを作成.")
	tokenEndpoint := accesstokenEndpoint
	// トークンエンドポイントへのリクエストを作成
	req, err := http.NewRequest("POST", tokenEndpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ret
	}
	req.Header.Set("Authorization", encodedAuthHeader)
	// application/json
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = tokenParams

	fmt.Println("リクエストを送信してレスポンスを受け取る.")
	// リクエストを送信してレスポンスを受け取る
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ret
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("ERR-CODE :%v \n ", resp)
		return "err"
	}
	fmt.Printf("デコードする. :%v \n ", resp.StatusCode)
	// デコードする
	var tknItem TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tknItem)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return ret
	}

	fmt.Println("Access Token => ", tknItem.AccessToken)

	return tknItem.AccessToken
}

func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func getPodcastEpisodes(item Settings, accessToken string) {
	fmt.Printf("start-getPodcastEpisodes.:%v \n", accessToken)

	showID := item.ShowID // 対象のポッドキャスト番組のIDを入力してください
	// 0dxhjH0slmOSg8dGVdH0NK:オトステ
	//38oo7wLS16DbsNnAGhvRJM:ぶちラジ

	var allEpisodes []PodcastEpisodeItem
	var episodeItem PodcastEpisodeItem
	var getItem PodcastEpisodeResponse
	offset := 0
	limit := 50 // ページごとのエピソード数

	for i := 0; i < 20; i++ {

		fmt.Printf("%v/%v \n", offset, limit)

		apiURL := fmt.Sprintf("https://api.spotify.com/v1/shows/%s/episodes?offset=%d&limit=%d", showID, offset, limit)
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			fmt.Printf("ERR-CODE :%v \n ", resp)
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&getItem)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		for _, episode := range getItem.Items {
			episodeItem = episode
			allEpisodes = append(allEpisodes, episodeItem)
			fmt.Println(episode.Name)
		}

		// fmt.Println(episodeItem)
		// fmt.Println("エピソードを取得したので詰め替えるよ～")

		if len(allEpisodes) < limit {
			break
		}

		offset += limit

	}

	seqNo := len(allEpisodes)
	for _, v := range allEpisodes {

		episodeNo := getEpisodeNo(v.Name, seqNo)
		fmt.Printf("%v,%v,%v,\"%v\",,%v,%v \n", formatYMD(v.ReleaseDate), item.ShowName, episodeNo, v.Name, v.ExternalUrls.Spotify, getSeconds(v.DurationMs))
		// v.AudioPreviewURL
		// でサンプル音声が撮れる
		seqNo--
	}
}

func getSeconds(milliseconds int) int {

	// minutes := milliseconds / 60000
	seconds := (milliseconds) / 1000
	// fmt.Printf("エピソードの長さ（分単位）: %d分 %d秒\n", minutes, seconds)
	return seconds
}
func formatYMD(dateStr string) string {

	layout := "2006-01-02" // 入力される日付のフォーマット

	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil {
		ret := fmt.Sprintf("Error parsing date:%v \n", err)
		return ret
	}

	// 目的のフォーマットに変換して出力
	formattedDate := parsedDate.Format("2006年01月02日")
	// fmt.Println("Formatted Date:", formattedDate)
	return formattedDate
}

func getEpisodeNo(titleText string, seq int) int {

	// 正規表現パターンを定義
	// pattern := `#(\d+|N)`
	pattern := `#(\d+)`
	// fmt.Println(titleText)
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(titleText)

	if len(match) > 1 {
		numberStr := match[1]
		number, err := strconv.Atoi(numberStr)
		if err == nil {
			fmt.Println(number)
		} else {
			fmt.Println("Error converting to int")
		}
		return number
	} else {
		ret := seq
		return ret
	}
}
