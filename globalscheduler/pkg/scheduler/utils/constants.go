package utils

// ContentKey is defined to solve the goLint warning: should not use basic type string as ContentKey in context.WithValue
type ContentKey string

const (

	// ContextRegions is the request's context param regions
	ContextRegions ContentKey = "regions"

	//ContextNeedAdmin is admin
	ContextNeedAdmin ContentKey = "needAdmin"

	// GlobalRegion global
	GlobalRegion = "global"

	// RegionID region_id key
	RegionID = "region_id"

	// TimeFormat is time str format
	TimeFormat = "2006-01-02 15:04:05"
)

// DefaultRegion default region
var DefaultRegion = ""
