package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
)

var (
	globalServiceGroupMap = make(map[string]*models.ServiceGroupLinkNode)
)

func InitServiceGroup()  {

}

func ListServiceGroup()  {

}

func CreateServiceGroup()  {

}

func UpdateServiceGroup()  {

}

func DeleteServiceGroup()  {

}

func ListServiceGroupEndpoint(monitorType string)  {

}

func getSimpleServiceGroup(serviceGroupGuid string) (result models.ServiceGroupTable,err error) {
	var serviceGroupTable []*models.ServiceGroupTable
	err = x.SQL("select * from service_group where guid=?", serviceGroupGuid).Find(&serviceGroupTable)
	if err != nil {
		return result,fmt.Errorf("Query service_group table fail,%s ", err.Error())
	}
	if len(serviceGroupTable) == 0 {
		return result,fmt.Errorf("Can not find service_group with guid:%s ", serviceGroupGuid)
	}
	result = *serviceGroupTable[0]
	return
}
