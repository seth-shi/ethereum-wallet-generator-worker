package master

import (
	"encoding/json"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/consts"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/models"
	"os"
)

func getStatusByCache() *models.MasterRunStatusCache {

	statusData, err := os.ReadFile(consts.MasterRunStatusFile)
	if err != nil {
		return nil
	}

	var cacheStatus models.MasterRunStatusCache
	if err := json.Unmarshal(statusData, &cacheStatus); err != nil {
		return nil
	}

	return &cacheStatus
}

func setStatusToCache(data models.MasterRunStatusCache) error {

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(consts.MasterRunStatusFile, jsonData, 0666)
}
