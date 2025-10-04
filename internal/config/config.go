// internal/config/config.go
package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DataPath string `json:"data_path"`
}

var cfg Config
var configFilePath string

var ResolvedDataPath string

func FindOrCreateDatePath() (bool, error) {
	// 設定ファイルの探索
	// OS標準の設定ディレクトリ
	usrConfigDir, err := os.UserConfigDir()
	if err != nil {
		return true, fmt.Errorf("ユーザ設定ディレクトリの取得に失敗しました: %w", err)
	}
	configDir := filepath.Join(usrConfigDir, "kk-invest")
	configFilePath = filepath.Join(configDir, "config.json")

	// 設定ファイルの読み込み
	file, err := os.Open(configFilePath)
	if err == nil {
		// ファイルが存在し, 読み込み成功
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&cfg); err != nil {
			return true, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
		}

		if cfg.DataPath != "" {
			ResolvedDataPath = cfg.DataPath
			fmt.Printf("設定書類から書類パスを取得しました: %s\n", ResolvedDataPath)
			return false, nil
		}
	}

	// 設定ファイルが存在しない場合、またはDataPathが空の場合、新規作成
	fmt.Println("KK-INVESTの初期設定を行います")
	fmt.Println("書類群を保存する場所を設定します")
	defaultDataPath := filepath.Join(configDir, "data")
	fmt.Printf("Returnキーの押下で規定の場所に設定します\n (規定の場所: %s)\n", defaultDataPath)
	fmt.Print("場所を指定> ")

	reader := bufio.NewReader(os.Stdin)
	inputPath, _ := reader.ReadString('\n')
	inputPath = trimNewline(inputPath)

	if inputPath == "" {
		ResolvedDataPath = defaultDataPath
	} else {
		ResolvedDataPath = inputPath
	}

	// ディレクトリの作成と設定の保存
	fmt.Printf("書類パスを %s に設定しました\n", ResolvedDataPath)
	if err := os.MkdirAll(ResolvedDataPath, 0755); err != nil {
		return true, fmt.Errorf("書類ディレクトリの作成に失敗しました: %w", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return true, fmt.Errorf("設定ディレクトリの作成に失敗しました: %w", err)
	}

	cfg.DataPath = ResolvedDataPath
	newFile, err := os.Create(configFilePath)
	if err != nil {
		return true, fmt.Errorf("設定ファイルの作成に失敗しました: %w", err)
	}
	defer newFile.Close()
	encoder := json.NewEncoder(newFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(cfg); err != nil {
		return true, fmt.Errorf("設定ファイルの保存に失敗しました: %w", err)
	}

	fmt.Println("設定が完了しました")
	return true, nil
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] == '\r' {
		s = s[:len(s)-1]
	}
	return s
}
