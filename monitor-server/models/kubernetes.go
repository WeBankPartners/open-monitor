package models

import "time"

type KubernetesClusterTable struct {
	Id          int       `json:"id"`
	ClusterName string    `json:"cluster_name"`
	ApiServer   string    `json:"api_server"`
	Token       string    `json:"token"`
	Guid        string    `json:"guid"`
	CreateAt    time.Time `json:"create_at"`
}

type KubernetesEndpointRelTable struct {
	Id           int    `json:"id"`
	KuberneteId  int    `json:"kubernete_id"`
	EndpointGuid string `json:"endpoint_guid"`
	PodGuid      string `json:"pod_guid"`
	Namespace    string `json:"namespace"`
}

type KubernetesClusterParam struct {
	Id          int    `json:"id"`
	ClusterName string `json:"cluster_name" binding:"required"`
	Ip          string `json:"ip" binding:"required"`
	Port        string `json:"port" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Guid        string `json:"guid"`
}
