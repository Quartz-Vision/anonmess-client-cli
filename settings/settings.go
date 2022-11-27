package settings

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var Config = struct {
	AppName                    string
	AppConfigFileName          string
	AppDataDefaultDirName      string
	AppDataDirPath             string `env:"PROGRAM_DATA_DIR" envDefault:""`
	AppDownloadsDirPath        string `env:"PROGRAM_DOWNLOADS_DIR" envDefault:""`
	KeysBufferSizeKB           int64  `env:"KEYS_BUFFER_SIZE_KB" envDefault:"1024"`
	KeysBufferSizeB            int64
	KeysStartSizeMB            int64 `env:"KEYS_START_SIZE_MB" envDefault:"1"`
	KeysStartSizeB             int64
	ServerHost                 string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	ServerPort                 int64  `env:"SERVER_PORT" envDefault:"8081"`
	ServerAddr                 string
	KeysStorageDefaultDirName  string
	CacheStorageDefaultDirName string
}{
	AppName:                    "anonmess",
	AppConfigFileName:          "anonmess.conf",
	AppDataDefaultDirName:      "data",
	KeysStorageDefaultDirName:  "keys",
	CacheStorageDefaultDirName: "cache",
}

func tryLoadEnv(paths ...string) bool {
	err := godotenv.Load(paths...)
	return err == nil
}

func selectPath(path string, fallback string) (r string, err error) {
	if path == "" {
		return fallback, nil
	} else {
		path, err := filepath.Abs(path)
		if err != nil {
			return r, ErrEnvParsing
		}
		return path, nil
	}
}

func Init() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPaths := []string{
		filepath.Join("/etc", Config.AppConfigFileName),
		filepath.Join(userConfigDir, Config.AppName, Config.AppConfigFileName),
		filepath.Join(userHomeDir, "."+Config.AppName, Config.AppConfigFileName),
	}

	tryLoadEnv()
	for _, configPath := range configPaths {
		tryLoadEnv(configPath)
	}

	if err := env.Parse(&Config); err != nil {
		return ErrEnvParsing
	}

	// .ServerAddr
	Config.ServerAddr = Config.ServerHost + ":" + strconv.FormatInt(int64(Config.ServerPort), 10)

	// .AppDataDirPath
	Config.AppDataDirPath, err = selectPath(
		Config.AppDataDirPath,
		filepath.Join(userHomeDir, "."+Config.AppName, Config.AppDataDefaultDirName),
	)
	if err != nil {
		return err
	}

	// .AppDownloadsDirPath
	Config.AppDownloadsDirPath, err = selectPath(
		Config.AppDownloadsDirPath,
		filepath.Join(userHomeDir, "Downloads", Config.AppName),
	)
	if err != nil {
		return err
	}

	// .KeysBufferSizeB
	if Config.KeysBufferSizeKB < 1 {
		return ErrEnvParsing
	}
	Config.KeysBufferSizeB = Config.KeysBufferSizeKB << 10
	// .KeysStartSizeB
	if Config.KeysStartSizeMB < 1 {
		return ErrEnvParsing
	}
	Config.KeysStartSizeB = Config.KeysStartSizeMB << 20

	return nil
}
