package conf

import (
	"fmt"

	"github.com/dtm-labs/dtmcli"
)

var DBConf = dtmcli.DBConf{
	Driver:   "mysql",
	Host:     "localhost",
	User:     "root",
	Password: "",
	Port:     3306,
}

var DtmServer = "http://localhost:36789/api/dtmsvr"

const BusiAPI = "/api/busi"
const BusiPort = 8081

var BusiUrl = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)
