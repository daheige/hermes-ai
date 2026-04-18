package entity

type Option struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value"`
}

// OptionTable options表
const OptionTable = "options"

func (Option) TableName() string {
	return OptionTable
}
