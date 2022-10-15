package settings

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var Config = struct {
	AppName               string
	AppConfigFileName     string
	AppDataDefaultDirName string
	ProgramDataDir        string `env:"PROGRAM_DATA_DIR" envDefault:""`
	KeysBufferSizeKB      int64  `env:"KEYS_BUFFER_SIZE_KB" envDefault:"1024"`
	KeysBufferSizeB       int64
}{
	AppName:               "anonmess",
	AppConfigFileName:     "app.conf",
	AppDataDefaultDirName: "data",
}

func tryLoadEnv(paths ...string) bool {
	err := godotenv.Load(paths...)
	return err == nil
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
		filepath.Join("/etc", Config.AppName+".conf.d", Config.AppConfigFileName),
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

	// Config.ServerAddr = Config.ServerHost + ":" + strconv.FormatInt(int64(Config.ServerPort), 10)

	// .ProgramDataDir
	if Config.ProgramDataDir == "" {
		Config.ProgramDataDir = filepath.Join(userHomeDir, "."+Config.AppName, Config.AppDataDefaultDirName)
	} else {
		path, err := filepath.Abs(Config.ProgramDataDir)
		if err != nil {
			return ErrEnvParsing
		}
		Config.ProgramDataDir = path
	}

	// .KeysBufferSizeB
	if Config.KeysBufferSizeKB < 1 {
		return ErrEnvParsing
	}
	Config.KeysBufferSizeB = Config.KeysBufferSizeKB << 10

	return nil
}
