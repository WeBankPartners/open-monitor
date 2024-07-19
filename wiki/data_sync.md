自定义看板老数据迁移
[说明]之前老的自定义看板数据都保存在 custom_dashboard表的cfg字段中,里面包含看板的图表，图表配置等属性,
新的方式通过添加图库表,图库配置表等,并且通过接口形式同步老数据到新表中

[接口] /dashboard/data/sync  POST

调用逻辑,当前看板存在,并且cfg不为空,看板与图表的关系表(custom_dashboard_chart_rel)为空,就会新生成图表以及配置




