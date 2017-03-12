package types

// auth action type values
const (
	AuthActionTypeAllow = "allow"
	AuthActionTypeDeny  = "deny"
)

// auth driver type values
const (
	AuthDriverTypeRedis   = DictRedis
	AuthDriverTypeMongodb = DictMongoDB
)
