package ds

import (
	"github.com/hylent/sf/clients"
	"github.com/hylent/sf/db"
)

var (
	Db *db.AdapterMysql
	Es *clients.EsClient
)
