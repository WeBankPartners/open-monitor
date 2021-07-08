package db

import (
	"fmt"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"io/ioutil"
	"strings"
	"time"
)

func SnmpExporterList() (result []*models.SnmpExporterTable,err error) {
	result = []*models.SnmpExporterTable{}
	err = x.SQL("select * from snmp_exporter").Find(&result)
	return
}

func SnmpExporterCreate(input models.SnmpExporterTable) error {
	if input.Modules == "" {
		input.Modules = "if_mib"
	}
	_,err := x.Exec("insert into snmp_exporter(id,address,modules,create_at) value (?,?,?,?)", input.Id, input.Address, input.Modules, time.Now().Format(models.DatetimeFormat))
	return err
}

func SnmpExporterUpdate(input models.SnmpExporterTable) error {
	if input.Modules == "" {
		input.Modules = "if_mib"
	}
	_,err := x.Exec("update snmp_exporter set address=?,modules=? where id=?", input.Address, input.Modules, input.Id)
	if err != nil {
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	err = SyncSnmpPrometheusConfig()
	if err != nil {
		return fmt.Errorf("Sync prometheus config fail,%s ", err.Error())
	}
	return nil
}

func SnmpExporterDelete(id string) error {
	_,err := x.Exec("delete from snmp_exporter where id=?", id)
	if err != nil {
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	err = SyncSnmpPrometheusConfig()
	if err != nil {
		return fmt.Errorf("Sync prometheus config fail,%s ", err.Error())
	}
	return nil
}

func SnmpEndpointAdd(snmpExporter,endpointGuid,target string) error {
	if checkSnmpEndpointExists(snmpExporter,endpointGuid) {
		return nil
	}
	_,err := x.Exec("insert into snmp_endpoint_rel(snmp_exporter,endpoint_guid,target) value (?,?,?)", snmpExporter, endpointGuid, target)
	if err != nil {
		return fmt.Errorf("Insert database fail,%s ", err.Error())
	}
	err = SyncSnmpPrometheusConfig()
	return err
}

func SnmpEndpointDelete(snmpExporter,endpointGuid string) error {
	if !checkSnmpEndpointExists(snmpExporter,endpointGuid) {
		return nil
	}
	_,err := x.Exec("delete from snmp_endpoint_rel where snmp_exporter=? and target=?", snmpExporter, endpointGuid)
	if err != nil {
		return fmt.Errorf("Delete database fail,%s ", err.Error())
	}
	err = SyncSnmpPrometheusConfig()
	return err
}

func checkSnmpEndpointExists(snmpExporter,endpointGuid string) bool {
	var snmpEndpointTable []*models.SnmpEndpointRelTable
	x.SQL("select id from snmp_endpoint_rel where snmp_exporter=? and target=?", snmpExporter, endpointGuid).Find(&snmpEndpointTable)
	if len(snmpEndpointTable) > 0 {
		return true
	}
	return false
}

func SyncSnmpPrometheusConfig() error {
	snmpList,err := SnmpExporterList()
	if err != nil {
		return err
	}
	if len(snmpList) == 0 {
		return nil
	}
	var snmpEndpointList []*models.SnmpEndpointRelTable
	err = x.SQL("select * from snmp_endpoint_rel order by snmp_exporter").Find(&snmpEndpointList)
	if err != nil {
		return fmt.Errorf("Try to get snmp endpoint table fail,%s ", err.Error())
	}
	var exporterTargetMap = make(map[string][]string)
	for _,v := range snmpEndpointList {
		if e,b := exporterTargetMap[v.SnmpExporter]; b {
			exporterTargetMap[v.SnmpExporter] = append(e, v.Target)
		}else{
			exporterTargetMap[v.SnmpExporter] = []string{v.Target}
		}
	}
	backupConfigName,backupErr := backupPrometheusConfig()
	if backupErr != nil {
		return backupErr
	}
	promBytes,err := ioutil.ReadFile("/app/monitor/prometheus/prometheus.yml")
	if err != nil {
		return fmt.Errorf("Read prometheus tpl file fail,%s ", err.Error())
	}
	promString := string(promBytes)
	startIndex := strings.Index(promString, "#Snmp_start")
	endIndex := strings.Index(promString, "#Snmp_end")
	tplBytes,err := ioutil.ReadFile("/app/monitor/prometheus/snmp_prometheus.tpl")
	if err != nil {
		return fmt.Errorf("Read snmp prometheus tpl file fail,%s ", err.Error())
	}
	snmpConfigString := ""
	tplString := string(tplBytes)
	for _,exporter := range snmpList {
		if len(exporterTargetMap[exporter.Id]) <= 0 {
			continue
		}
		tmpConfigString := tplString + "\n"
		tmpConfigString = strings.ReplaceAll(tmpConfigString, "{{snmp_exporter_id}}", exporter.Id)
		tmpConfigString = strings.ReplaceAll(tmpConfigString, "{{modules}}", exporter.Modules)
		tmpConfigString = strings.ReplaceAll(tmpConfigString, "{{snmp_exporter_address}}", exporter.Address)
		tmpConfigString = strings.ReplaceAll(tmpConfigString, "{{snmp_target}}", "'"+strings.Join(exporterTargetMap[exporter.Id], "','")+"'")
		snmpConfigString += tmpConfigString + "\n"
	}
	promString = promString[:startIndex+11] + "\n" + snmpConfigString + promString[endIndex:]
	err = ioutil.WriteFile("/app/monitor/prometheus/prometheus.yml", []byte(promString), 0644)
	if err != nil {
		err = fmt.Errorf("Write snmp config to prometheus fail,%s ", err.Error())
		recoverPrometheusConfig(backupConfigName)
		return err
	}
	err = prom.ReloadConfig()
	if err != nil {
		err = fmt.Errorf("Reload prometheus config fail,%s ", err.Error())
		recoverPrometheusConfig(backupConfigName)
		prom.ReloadConfig()
		return err
	}
	return nil
}