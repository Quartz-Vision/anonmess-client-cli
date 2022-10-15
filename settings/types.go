package settings

import "errors"

var ErrEnvLoading = errors.New("error while loading the config")
var ErrEnvParsing = errors.New("error while parsing the config")
