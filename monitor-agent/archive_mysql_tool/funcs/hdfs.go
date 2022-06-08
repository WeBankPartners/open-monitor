package funcs

import (
	"fmt"
	"log"
	"os"
	"time"

	hdfs "github.com/colinmarc/hdfs/v2"

	"github.com/colinmarc/hdfs/v2/hadoopconf"
	krb "github.com/jcmturner/gokrb5/v8/client"
	krconfig "github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/keytab"
)

func getKerberosClient() (*krb.Client, error) {
	cfg, err := krconfig.Load(Config().Hdfs.Krb5)
	//fmt.Println(cfg.JSON())
	if err != nil {
		log.Printf("Couldn't load krb config:", err)
		return nil, err
	}

	//TODO:
	keyTab, err := keytab.Load(Config().Hdfs.Keytab)
	if err != nil {
		log.Printf("Couldn't load keytab:", err)
		return nil, err
	}
	l := log.New(os.Stderr, "krb5-client", log.Lshortfile)
	c := krb.NewWithKeytab(Config().Hdfs.Username, Config().Hdfs.Realm, keyTab, cfg, krb.DisablePAFXFAST(true), krb.Logger(l))
	err = c.Login()
	if err != nil {
		fmt.Errorf("Login failed: %w", err)
		return nil, err
	}
	return c, nil
}

func initHdfsClient() (*hdfs.Client, error) {
	os.Setenv("HADOOP_CONF_DIR", Config().Hdfs.ConfDir)
	conf, err := hadoopconf.LoadFromEnvironment()
	if err != nil || conf == nil {
		log.Printf("Couldn't load hadoop config", err)
		return nil, err
	}

	options := hdfs.ClientOptionsFromConf(conf)

	options.KerberosClient, err = getKerberosClient()

	if err != nil {
		log.Printf("Can't get client from kerberos", err)
		return nil, err
	}

	hdfsClient, err := hdfs.NewClient(options)
	if err != nil {
		log.Printf("Can't get hdfs client,", err)
		return nil, err
	}
	return hdfsClient, nil
}

func CopyLocalToHdfs(filename string) {
	hdfsClient, err := initHdfsClient()
	if err != nil {
		log.Printf("init hdfs client error,", err)
		return
	}
	defer hdfsClient.Close()

	hdfsDestDir := Config().Hdfs.DestDir
	now := time.Now()
	d, _ := time.ParseDuration("-24h")
	datetime := now.Add(d).Format("20060102")
	fullDestPath := hdfsDestDir + datetime + "/" + filename
	err = hdfsClient.Remove(fullDestPath)
	if err != nil {
		log.Printf("Can't remove path,", err)
	}

	err = hdfsClient.MkdirAll(hdfsDestDir+datetime, 0750)
	if err != nil {
		log.Printf("make hdfs directory error,", err)
	}
	tmp_dir := Config().Hdfs.LocalTempDir
	err = hdfsClient.CopyToRemote(tmp_dir+"/"+filename, fullDestPath)
	if err != nil {
		log.Printf("Copy local file to hdfs error,", err)
	}
	err = os.Remove(tmp_dir + "/" + filename)
	if err != nil {
		log.Printf("Dlete local file error", err)
	} else {
		log.Printf(filename + " remove success")
	}
}
