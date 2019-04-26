# goenv
goenvはGoのバージョンを管理するための単純なコマンドです
goenvは以下の順番でGoコマンドのバージョンを決定します

## 探索場所
環境変数GOENV_LOCATIONのディレクトリ(デフォルトでは/usr/local/go/bin)

## 探索順序
1. 環境変数GOCMDを参照し、値がセットされていればそれを使用します  
2. カレントディレクトリがgitリポジトリである場合、リポジトリのトップの.go-versionファイルを参照する。ファイルが存在し中身が空ではないならばそのバージョンを使用します  
3. GOENV_LOCATIONにsemverに従ってgoX.Y.Zの命名規則でGoが配置されていると仮定し、一番新しいものを使用します  


