package utils

import (
	"fmt"
)

func LogQuota(quota int64, quotaPerUnit float64, displayInCurrencyEnabled bool) string {
	if displayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f 额度", float64(quota)/quotaPerUnit)
	} else {
		return fmt.Sprintf("%d 点额度", quota)
	}
}
