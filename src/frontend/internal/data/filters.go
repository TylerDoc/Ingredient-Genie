package data

type Filters struct {
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
	Sort     SortType `json:"sort"`
}

// TODO: go back to the backend sometime and make this an enum
type SortType = string
