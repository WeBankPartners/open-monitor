 - job_name: 'k8s-kubelet-{{cluster_name}}'
    honor_timestamps: true
    metrics_path: /metrics
    scheme: https
    kubernetes_sd_configs:
    - api_server: https://{{api_server_ip}}:{{api_server_port}}
      role: node
      bearer_token_file: /app/monitor/prometheus/k8s_token/{{cluster_name}}
      tls_config:
        insecure_skip_verify: true
    bearer_token_file: /app/monitor/prometheus/k8s_token/{{cluster_name}}
    tls_config:
      insecure_skip_verify: true
    relabel_configs:
    - action: labelmap
      regex: __meta_kubernetes_node_label_(.+)
    - separator: ;
      regex: (.*)
      target_label: __address__
      replacement: {{api_server_ip}}:{{api_server_port}}
      action: replace
    - source_labels: [__meta_kubernetes_node_name]
      separator: ;
      regex: (.+)
      target_label: __metrics_path__
      replacement: /api/v1/nodes/${1}/proxy/metrics
      action: replace


 - job_name: 'k8s-cadvisor-{{cluster_name}}'
    honor_timestamps: true
    metrics_path: /metrics
    scheme: https
    kubernetes_sd_configs:
    - api_server: https://{{api_server_ip}}:{{api_server_port}}
      role: node
      bearer_token_file: /app/monitor/prometheus/k8s_token/{{cluster_name}}
      tls_config:
        insecure_skip_verify: true
    bearer_token_file: /app/monitor/prometheus/k8s_token/{{cluster_name}}
    tls_config:
      insecure_skip_verify: true
    relabel_configs:
    - action: labelmap
      regex: __meta_kubernetes_node_label_(.+)
    - separator: ;
      regex: (.*)
      target_label: __address__
      replacement: {{api_server_ip}}:{{api_server_port}}
      action: replace
    - source_labels: [__meta_kubernetes_node_name]
      separator: ;
      regex: (.+)
      target_label: __metrics_path__
      replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor
      action: replace