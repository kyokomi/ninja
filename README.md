NINJA（忍者）
====================================

<img src="https://user-images.githubusercontent.com/1456047/63598488-9db6a780-c5fa-11e9-8383-d82ffc84b251.png" width="320">


## Usage
![image](https://user-images.githubusercontent.com/1456047/63650375-15184280-c785-11e9-899b-2a9d68680ce5.png)

slack botを作成し、視聴したいtl_xxxx / times_xxxx にinviteし、以下を起動する。

```shell script
$ ./ninja -token <slack bot token>
```

### Help

```shell script
$ ./ninja -h
Usage of ./ninja:
  -debug
    	slack bot debug flag
  -icon-dir string
    	icon file directory path (default "./icons")
  -image
    	window not support (default true)
  -log-prefix string
    	log prefix (default "slack-bot: ")
  -token string
    	slack bot token
```

## MEMO

- [ ] メンション時のチャンネル名やユーザー名を表示するようにする（IDが出てしまっている）
- [ ] カスタム絵文字サポート（うーん）
- [ ] アップロード画像のサイズを制限する（めちゃでかいと画面がうまる）
- [ ] URLをクリック可能にする
- [ ] Twitterの展開
- [ ] 文章中の画像URLの画像展開



