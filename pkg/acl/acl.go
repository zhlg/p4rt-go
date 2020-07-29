package acl

import (
	iris "github.com/kataras/iris/v12"

)

type (
	acl struct {
		name       string
		aclEntries []*aclEntry
	}

	aclEntry struct {
		id   uint64        `json:"id"`
		conf configuration `json:"configuration"`
	}

	configuration struct {
		aceType  matchip `json:"ace_type"`
		srcIP    string  `json:"src_ip"`
		action   string  `json:"action"`
		evifName string  `json:"evif_name"`
		dstIp    string  `json:"dst_ip"`
		enCount  bool    `json:"en_count"`
		handle   string  `json:"handle"`
		vlanId   uint16  `json:"vlanid"`
	}

	matchip struct {
		isExactMatch bool   `json:"is_exact_match"`
		ipVersion    string `json:"ip_version"`
	}
)

func AddAcls(ctx iris.Context) {

}
func DeleteAcl(ctx iris.Context) {
	//id := ctx.Params().Get("id")

}
func ModifyAcl(ctx iris.Context) {
	//id := ctx.Params().Get("id")

}
func GetAllAcls(ctx iris.Context) {}
func GetAcl(ctx iris.Context) {
	//id := ctx.Params().Get("id")
}

func AddAclEntries(ctx iris.Context)    {}
func DeleteAclEntries(ctx iris.Context) {}
func ModifyAclEntries(ctx iris.Context) {}
func GetAllAclEntries(ctx iris.Context) {}
func GetAclEntries(ctx iris.Context)    {}
func UpdateAclEntries(ctx iris.Context) {}
