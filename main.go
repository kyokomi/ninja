package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/kyokomi/emoji"
	sixel "github.com/mattn/go-sixel"
	"github.com/nfnt/resize"
	"github.com/nlopes/slack"
)

func main() {
	botToken := flag.String("token", os.Getenv("SLACK_BOT_TOKEN"), "slack bot token")
	isDebug := flag.Bool("debug", false, "slack bot debug flag")
	logPrefix := flag.String("log-prefix", "slack-bot: ", "log prefix")
	isImage := flag.Bool("image", true, "window not support")
	maxImageWidth := flag.Uint("max-image-width", 480, "max image size(width)")
	iconDir := flag.String("icon-dir", "./icons", "icon file directory path")
	flag.Parse()

	api := slack.New(
		*botToken,
		slack.OptionDebug(*isDebug),
		slack.OptionLog(log.New(os.Stdout, *logPrefix, log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	s := Service{api: api, config: Config{
		token:         *botToken,
		isDebug:       *isDebug,
		isImage:       *isImage,
		iconDir:       *iconDir,
		maxImageWidth: *maxImageWidth,
	}, syncFile: &sync.Map{}}

	for msg := range rtm.IncomingEvents {
		//fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello
		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			s.messageHandler(ev)

		case *slack.PresenceChangeEvent:

		case *slack.LatencyReport:

		case *slack.RTMError:

		case *slack.InvalidAuthEvent:
			return

		case *slack.FileSharedEvent:
			if s.config.isImage {
				if s.config.isDebug {
					fmt.Printf("FileShare: %+v\n", msg.Data)
				}

				s.fireShareHandler(ev)
			}
		default:
			// Ignore other events..
			if s.config.isDebug {
				fmt.Printf("Unexpected: %+v\n", msg.Data)
			}

			// TODO: 画像を直接uploadしたやつも画像表示したい
		}
	}
}

// Config 設定
type Config struct {
	token         string
	isDebug       bool
	isImage       bool
	iconDir       string
	maxImageWidth uint
}

// Service いろいろやっていき
type Service struct {
	api    *slack.Client
	config Config

	// 同じファイルを何回もdownloadして表示しないようにするlock
	syncFile *sync.Map
}

func (s *Service) messageHandler(ev *slack.MessageEvent) {
	c, _ := s.api.GetChannelInfo(ev.Channel)
	u, _ := s.api.GetUserInfo(ev.User)

	var channelName string
	if c != nil {
		channelName = c.Name
	}
	var username string
	if u != nil {
		username = u.Name

		if s.config.isImage {
			if err := os.MkdirAll(s.config.iconDir, 0755); err != nil {
				// スルー
			}
			userFilepath := filepath.Join(s.config.iconDir, username+filepath.Ext(u.Profile.Image48))
			f, err := os.Open(userFilepath)
			if err != nil {
				// ファイルがなかったりエラーなら取得しなおす
				s.downloadImage(userFilepath, u.Profile.Image48)
				f, err = os.Open(userFilepath)
			}

			// 最終的にファイルがエラーじゃなければdecodeして表示する
			if err == nil {
				renderImageSize(f, 40, 40)
			}
		}
	}

	fmt.Println(channelName, username, emoji.Sprint(ev.Text))

	// TODO: TextにURLが含まれてるケースも画像を展開したい
	if s.config.isImage {

	}
}

func (s *Service) fireShareHandler(ev *slack.FileSharedEvent) {
	if _, ok := s.syncFile.Load(ev.FileID); ok {
		return // すでに表示ずみはskip
	}

	fileInfo, _, _, err := s.api.GetFileInfo(ev.FileID, 1, 1)
	if err != nil {
		return
	}

	switch filepath.Ext(fileInfo.URLPrivateDownload) {
	case ".png", ".jpg", ".jpeg", ".gif":
		break
	default:
		return
	}

	if err := os.MkdirAll("files", 0755); err != nil {
		// スルー
	}
	fPath := filepath.Join("files", fileInfo.Name)
	s.downloadImageAuth(fPath, fileInfo.URLPrivateDownload)
	f, err := os.Open(fPath)
	if err != nil {
		panic(err) // TODO: あとで
	}
	s.renderImage(f)
	fmt.Println()
	s.syncFile.Store(ev.FileID, nil)
}

func (s *Service) renderImage(f *os.File) {
	renderImageSize(f, s.config.maxImageWidth, 0) // でかすぎる画像がしんどいのでw=480を最大にする
}

func renderImageSize(f *os.File, w, h uint) {
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err) // TODO: あとで
	}
	defer f.Close()

	buf := bufio.NewWriter(os.Stdout)
	defer buf.Flush()

	enc := sixel.NewEncoder(buf)
	enc.Dither = true
	if err := enc.Encode(resize.Resize(w, h, img, resize.Bicubic)); err != nil {
		panic(err) // TODO: あとで
	}
}

// ファイルをダウンロードしてローカルに保存します
func (s *Service) downloadImageAuth(outputFilePath string, downloadURL string) {
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		panic(err) // TODO: あとで
	}
	req.Header.Set("Authorization", "Bearer "+s.config.token)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err) // TODO: あとで
	}
	defer response.Body.Close()

	file, err := os.Create(outputFilePath)
	if err != nil {
		panic(err) // TODO: あとで
	}
	defer file.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		panic(err) // TODO: あとで
	}
}

// ファイルをダウンロードしてローカルに保存します
func (s *Service) downloadImage(outputFilePath string, downloadURL string) {
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		panic(err) // TODO: あとで
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err) // TODO: あとで
	}
	defer response.Body.Close()

	file, err := os.Create(outputFilePath)
	if err != nil {
		panic(err) // TODO: あとで
	}
	defer file.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		panic(err) // TODO: あとで
	}
}
