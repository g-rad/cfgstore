package cfgstore

type ConfigKeyValue struct {
	ApplicationId int       `db:"ApplicationId"`
	EnvironmentId int       `db:"EnvironmentId"`
	KeyId         int       `db:"KeyId"`
	Key           string    `db:"Name"`
	Value         string    `db:"Value"`
}
