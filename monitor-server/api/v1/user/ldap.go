package user

import (
	"github.com/go-ldap/ldap"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"strings"
	"fmt"
	"log"
)

func ldapAuth(username string, password string) bool {
	host := m.Config().Http.Ldap.Server
	port := m.Config().Http.Ldap.Port
	conn,err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err!=nil {
		log.Printf("ldap dial error : %v \n", err)
		return false
	}
	defer conn.Close()
	bindDN := fmt.Sprintf(m.Config().Http.Ldap.BindDN, username)
	err = conn.Bind(bindDN, password)
	if err!=nil {
		log.Printf("ldap bind error : %v \n", err)
		return false
	}
	var searchResult *ldap.SearchResult
	baseDN := m.Config().Http.Ldap.BaseDN
	searchReq := ldap.SearchRequest{
		BaseDN:       baseDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		Attributes: []string{
			"name",
			"mail",
			"cn",
			"company",
			"telephoneNumber",
		},
		Filter: strings.Replace(m.Config().Http.Ldap.Filter, "%s", ldap.EscapeFilter(username), -1),
	}
	searchResult,err = conn.Search(&searchReq)
	if err!=nil {
		log.Printf("ldap search error : %v \n", err)
		return false
	}
	if len(searchResult.Entries) == 1 {
		var cnName string
		var userEmail string
		for _,attr := range searchResult.Entries[0].Attributes {
			if attr.Name == "cn" {
				cnName = attr.Values[0]
			}
			if attr.Name == "mail" {
				userEmail = attr.Values[0]
			}
		}
		fmt.Sprintf("username: %s, cnname: %s, email: %s", username, cnName, userEmail)
	}else{
		log.Printf("ldap search result err, result num : %d \n", len(searchResult.Entries))
		return false
	}
	return true
}
