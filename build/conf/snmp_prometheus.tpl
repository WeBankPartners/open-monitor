  - job_name: '{{snmp_exporter_id}}'
    static_configs:
      - targets: [{{snmp_target}}]
    metrics_path: /snmp
    params:
      module: [{{modules}}]
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: {{snmp_exporter_address}}