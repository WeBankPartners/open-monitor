{
  "http": {
    "port": "8088",
    "swagger": true,
    "cross": true,
    "alive": 2591000,
    "ldap": {
      "enable": false,
      "server": "127.0.0.1",
      "port": 389,
      "bindDN": "%s@test.com",
      "baseDN": "ou=Users_Employee",
      "filter": "(sAMAccountName=%s)",
      "attributes": ["name", "mail", "telephoneNumber"]
    },
    "log": {
      "enable": true,
      "level": "debug",
      "file": "prometheus_monitor.log",
      "stdout": true
    }
  },
  "store": [
    {
      "name": "default",
      "type": "mysql",
      "server": "127.0.0.1",
      "port": 3306,
      "user": "root",
      "pwd": "wecube",
      "database": "monitor",
      "maxOpen": 20,
      "maxIdle": 10,
      "timeout": 60
    }
  ],
  "datasource" : {
    "env" : "dev",
    "servers": [
      {
        "id": 1,
        "type": "prometheus",
        "env": "dev",
        "host": "127.0.0.1:9090",
        "token": ""
      }
    ],
    "divide_time": 1,
    "wait_time": 1
  },
  "limitIp": ["*"],
  "dependence": [
    {
      "name": "consul",
      "server": "http://127.0.0.1:8500",
      "username": "",
      "password": "",
      "expire": 0
    }
  ],
  "prometheus" : {
    "config_path": "{{MONITOR_BASE_PATH}}/prometheus/rules",
    "config_reload": "http://127.0.0.1:9090/-/reload"
  },
  "tag_blacklist" : ["veth"]
}