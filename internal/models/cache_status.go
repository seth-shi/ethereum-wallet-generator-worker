package models

import "time"

type MasterRunStatusCache struct {
	Workers []*WorkStatusRequest `json:"workers"`
	StartAt time.Time            `json:"start_at"`
}
