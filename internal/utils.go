package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/taka1156/cms-cli/internal/entity"
)

// cmsc.json を読み込むだけの共通処理
func loadConfig() (entity.CMSConfig, error) {
	var config entity.CMSConfig

	configFile, err := os.ReadFile("cmsc.json")
	if err != nil {
		return config, fmt.Errorf("cmsc.json not found. Run './cmsc init' to create a default configuration")
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		return config, fmt.Errorf("failed to parse cmsc.json: %w", err)
	}

	return config, nil
}
