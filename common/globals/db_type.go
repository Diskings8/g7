package globals

const (
	DBMysql = "mysql"
	DBMongo = "mongo"
)

const (
	SaveDataKindCornCache = int(iota)
	SaveDataKindCornDb
	SaveDataKindLoginOut
)
