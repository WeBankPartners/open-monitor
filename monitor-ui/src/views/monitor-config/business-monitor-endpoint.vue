<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <div v-for="table in totalData" :key="table.guid">
        <Tag size="medium" color="blue" style="margin: 8px 0">{{$t('field.resourceLevel')}}：{{table.tableData.display_name}}</Tag>
        <PageTable :pageConfig="table.tableConfig">
          <div slot='tableExtend'>
            <div style="margin:8px;border:1px solid #2db7f5">
              <extendTable :detailConfig="table.tableConfig.table.isExtend.detailConfig"></extendTable>
            </div>
            <!-- <div style="margin:8px;border:1px solid #19be6b">
              <Tag type="border" color="success">{{$t('m_metric_regular')}}</Tag>
              <extendTable :detailConfig="table.tableConfig.table.isCustomMetricExtend.detailConfig"></extendTable>
            </div> -->
          </div>
        </PageTable>
      </div>
    </section>
    <Modal
      v-model="ruleModelConfig.isShow"
      :title="$t('m_json_regular')"
      :mask-closable="false"
      width="840"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <Form :label-width="100">
          <FormItem :label="$t('tableKey.regular')">
            <Input disabled v-model="ruleModelConfig.addRow.json_regular" style="width:100%"/>
          </FormItem>
          <FormItem :label="$t('tableKey.tags')">
            <Input disabled v-model="ruleModelConfig.addRow.tags" style="width:100%" />
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in ruleModelConfig.addRow.metric_list">
            <p :key="index + '3'" style="text-align: center;">
              <Tooltip :content="$t('m_key')" :delay="1000">
                <Input disabled v-model="item.json_key" style="width: 190px" :placeholder="$t('m_key') + ' e.g:[.*][.*]'" />
              </Tooltip>
              <Tooltip :content="$t('field.metric')" :delay="1000">
                <Input disabled v-model="item.metric" style="width: 190px" :placeholder="$t('field.metric') + ' , e.g:code'" />
              </Tooltip>
              <Tooltip :content="$t('field.aggType')" :delay="1000">
                <Select disabled v-model="item.agg_type" filterable clearable style="width:190px">
                  <Option v-for="agg in ruleModelConfig.aggOption" :value="agg" :key="agg">{{
                    agg
                  }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('tableKey.description')" :delay="1000">
                <Input disabled v-model="item.display_name" style="width: 160px" :placeholder="$t('tableKey.description')" />
              </Tooltip>
            </p>
            <div v-if="item.string_map.length > 0" :key="index + 1" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
              <template v-for="(stringMapItem, stringMapIndex) in item.string_map">
                <p :key="stringMapIndex + 2" style="text-align: center;">
                  <Tooltip :content="$t('tableKey.regular')" :delay="1000">
                    <Select disabled v-model="stringMapItem.regulative" filterable clearable style="width:120px">
                      <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                        regulation.label
                      }}</Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('m_business_object')" :delay="1000">
                    <Input disabled v-model="stringMapItem.target_value" style="width: 230px" :placeholder="$t('m_business_object')" />
                  </Tooltip>
                  <Tooltip :content="$t('m_log_server')" :delay="1000">
                    <Input disabled v-model="stringMapItem.source_value" style="width: 230px" :placeholder="$t('m_log_server')" />
                  </Tooltip>
                </p>
              </template>
            </div>
            <Divider :key="index + 'Q'" />
          </template>
        </div>
      </div>
      <div slot="footer">
        <Button @click="ruleModelConfig.isShow=false">{{$t('button.cancel')}}</Button>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('field.aggType')}}:</label>
          <Select disabled v-model="customMetricsModelConfig.addRow.agg_type" filterable clearable style="width:375px">
            <Option v-for="agg in customMetricsModelConfig.slotConfig.aggOption" :value="agg" :key="agg">{{
              agg
            }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in customMetricsModelConfig.addRow.string_map">
              <p :key="index" style="text-align: center;">
                <Tooltip :content="$t('tableKey.regular')" :delay="1000">
                  <Select disabled v-model="item.regulative" filterable clearable style="width:150px">
                    <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                      regulation.label
                    }}</Option>
                  </Select>
                </Tooltip>
                <Tooltip :content="$t('m_business_object')" :delay="1000">
                  <Input disabled v-model="item.target_value" style="width: 150px" :placeholder="$t('m_business_object')" />
                </Tooltip>
                <Tooltip :content="$t('m_log_server')" :delay="1000">
                  <Input disabled v-model="item.source_value" style="width: 150px" :placeholder="$t('m_log_server')" />  
                </Tooltip>
              </p>
            </template>
          </div>
        </div>
        <Button style="float:right" @click="cancelModal">{{$t('button.cancel')}}</Button>
      </div>
    </ModalComponent>
    <!-- DB config -->
    <section v-if="showDbManagement" style="margin-top: 16px;">
      <Tag size="medium" color="blue">{{$t('m_db')}}</Tag>
      <PageTable :pageConfig="pageDbConfig"></PageTable>
    </section>
    <Modal
      v-model="dbModelConfig.isShow"
      :title="$t('m_db')"
      width="680"
      :mask-closable="false"
      footer-hide
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <Form :label-width="100">
          <FormItem :label="$t('field.displayName')">
            <Input disabled v-model="dbModelConfig.addRow.display_name" style="width:520px"/>
          </FormItem>
          <FormItem :label="$t('field.metric')">
            <Input disabled v-model="dbModelConfig.addRow.metric" style="width:520px" />
          </FormItem>
          <FormItem label="SQL">
            <Input disabled v-model="dbModelConfig.addRow.metric_sql" type="textarea" style="width:520px" />
          </FormItem>
          <FormItem :label="$t('field.type')">
            <Select disabled v-model="dbModelConfig.addRow.monitor_type" @on-change="getEndpoint(dbModelConfig.addRow.monitor_type, 'mysql')" style="width: 520px">
              <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
            </Select>
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in dbModelConfig.addRow.endpoint_rel">
            <p :key="index + '3'" style="text-align: center;">
              <Tooltip :content="$t('m_db')" :delay="1000">
                <Input disabled v-model="item.target_endpoint" style="width:290px" />
                <!-- <Select disabled v-model="item.target_endpoint" style="width: 290px" :placeholder="$t('m_business_object')">
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select> -->
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Input disabled v-model="item.source_endpoint" style="width:290px" />
                <!-- <Select disabled v-model="item.source_endpoint" style="width: 290px" :placeholder="$t('m_log_server')">
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select> -->
              </Tooltip>
            </p>
          </template>
        </div>
      </div>
    </Modal>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="$t('button.view')"
      :mask-closable="false"
      :width="720"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <div>
          <span>{{$t('field.type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type" disabled style="width: 640px">
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div style="margin: 8px 0">
          <span>{{$t('tableKey.path')}}:</span>
          <Input style="width: 640px" disabled v-model="addAndEditModal.dataConfig.log_path" />
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;text-align:center">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Tooltip :content="$t('m_business_object')" :delay="1000">
                <Input disabled v-model="item.source_endpoint" style="width:315px" />
                <!-- <Select v-model="item.target_endpoint" disabled style="width: 315px" :placeholder="$t('m_business_object')">
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select> -->
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Input disabled v-model="item.source_endpoint" style="width:315px" />
                <!-- <Select v-model="item.source_endpoint" disabled style="width: 315px" :placeholder="$t('m_log_server')">
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select> -->
              </Tooltip>
            </p>
          </template>
        </div>
      </div>
      <div slot="footer">
        <Button @click="addAndEditModal.isShow = false">{{$t('button.cancel')}}</Button>
      </div>
    </Modal>
    <CustomRegex ref="customRegexRef"></CustomRegex>
    <BusinessMonitorGroupConfig ref="businessMonitorGroupConfigRef"></BusinessMonitorGroupConfig>
  </div>
</template>

<script>
import CustomRegex from '@/views/monitor-config/log-template-config/custom-regex.vue'
import BusinessMonitorGroupConfig from '@/views/monitor-config/business-monitor-group-config.vue'
import extendTable from '@/components/table-page/extend-table'
let tableEle = [
  {title: 'tableKey.logPath', value: 'log_path', display: true},
  {title: 'field.type', value: 'monitor_type', display: true},
]
const btn = [
  {btn_name: 'button.view', btn_func: 'editF'},
]
let tableDbEle = [
  {title: 'field.displayName', value: 'display_name', display: true},
  {title: 'field.metric', value: 'metric', display: true},
  {title: 'field.type', value: 'monitor_type', display: true}
]
const btnDb = [
  {btn_name: 'button.view', btn_func: 'editDbItem'}
]
export default {
  name: '',
  data () {
    return {
      MODALHEIGHT: 300,
      showManagement: false,
      totalData: [],
      regulationOption: [
        {label: this.$t('m_regular_match'), value: 1},
        {label: this.$t('m_irregular_matching'), value: 0}
      ],
      ruleModelConfig: {
        isShow: false,
        isAdd: true,
        addRow: {
          log_metric_monitor: null,
          json_regular: null,
          tags: null,
          metric_list: []
        },
        aggOption: ['sum', 'avg', 'count', 'max', 'min']
      },
      customMetricsModelConfig: {
        modalId: 'custom_metrics',
        isAdd: true,
        modalStyle: 'min-width:550px',
        modalTitle: 'm_metric_regular',
        noBtn: true,
        config: [
          {label: 'field.metric', value: 'metric', placeholder: '', disabled: true, type: 'text'},
          {label: 'tableKey.description', value: 'display_name', placeholder: '', disabled: true, type: 'text'},
          {label: 'tableKey.regular', value: 'regular', placeholder: 'tips.required', v_validate: 'required:true', disabled: true, type: 'text'},
          {name:'ruleConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          log_metric_monitor: '',
          display_name: '',
          agg_type: 'min',
          metric: null,
          regular: '',
          string_map: []
        },
        slotConfig: {
          aggOption: ['sum', 'avg', 'count', 'max', 'min'],
          regulationOption: [
            {label: this.$t('m_regular_match'), value: 1},
            {label: this.$t('m_irregular_matching'), value: 0}
          ]
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      // DB config 
      showDbManagement: false,
      pageDbConfig: {
        table: {
          tableData: [],
          tableEle: tableDbEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btnDb,
          handleFloat:true
        }
      },
      dbModelConfig: {
        isShow: false,
        isAdd: true,
        addRow: {
          service_group: '',
          metric_sql: '',
          metric: '',
          display_name: '',
          monitor_type: '',
          endpoint_rel: []
        }
      },
      monitorTypeOptions: [
        {label: 'process', value: 'process'},
        {label: 'java', value: 'java'},
        {label: 'nginx', value: 'nginx'},
        {label: 'http', value: 'http'}
      ],
      sourceEndpoints: [],
      targetEndpoints: [],
      addAndEditModal: {
        isShow: false,
        isAdd: false,
        dataConfig: {
          service_group: '',
          log_path: [],
          monitor_type: '',
          endpoint_rel: []
        },
        pathOptions: []
      },
    }
  },
  methods: {
    // DB config 
    getDbDetail (targetId) {
      const api = this.$root.apiCenter.getTargetDbDetail + '/endpoint/' + targetId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.pageDbConfig.table.tableData = responseData
        this.showDbManagement = true
      }, {isNeedloading:false})
    },
    editDbItem (rowData) {
      this.getEndpoint(rowData.monitor_type, 'mysql', rowData.service_group)
      this.dbModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.dbModelConfig.isAdd = false
      this.dbModelConfig.isShow = true
    },
    async getEndpoint (val, type, targrtId) {
      // await this.getDefaultConfig(val, type)
      // get source Endpoint
      const sourceApi = this.$root.apiCenter.getEndpointsByType + '/' + targrtId + '/endpoint/' + type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', sourceApi, '', (responseData) => {
        this.sourceEndpoints = responseData
      }, {isNeedloading:false})
      const targetApi = this.$root.apiCenter.getEndpointsByType + '/' + targrtId + '/endpoint/' + val
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', targetApi, '', (responseData) => {
        this.targetEndpoints = responseData
      }, {isNeedloading:false})
    },
    //other config 
    editF (rowData) {
      this.getEndpoint(rowData.monitor_type, 'host', rowData.service_group)
      this.addAndEditModal.isAdd = false
      this.addAndEditModal.addRow = rowData
      this.modelTip.value = rowData.guid
      this.addAndEditModal.dataConfig.guid = rowData.guid
      this.addAndEditModal.dataConfig.service_group = rowData.service_group
      this.addAndEditModal.dataConfig.monitor_type = rowData.monitor_type
      this.addAndEditModal.dataConfig.log_path = rowData.log_path
      this.addAndEditModal.dataConfig.endpoint_rel = rowData.endpoint_rel
      this.addAndEditModal.isShow = true
    },
    getDefaultConfig (val, type) {
      const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.targrtId}&sourceType=${type}&targetType=${val}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        const tmp = responseData.map(r => {
            return {
              source_endpoint: r.source_endpoint,
              target_endpoint: r.target_endpoint
            }
          })
        if (type === 'host') {
          tmp.forEach(t => {
            const find = this.addAndEditModal.dataConfig.endpoint_rel.find(rel => rel.source_endpoint === t.source_endpoint && rel.target_endpoint === t.target_endpoint)
            if (find === undefined) {
              this.addAndEditModal.dataConfig.endpoint_rel.push(t)
            }
          })
        } else {
          tmp.forEach(t => {
            const find = this.dbModelConfig.addRow.endpoint_rel.find(rel => rel.source_endpoint === t.source_endpoint && rel.target_endpoint === t.target_endpoint)
            if (find === undefined) {
              this.dbModelConfig.addRow.endpoint_rel.push(t)
            }
          })
        }
      })
    },
    cancelModal () {
      this.$root.JQ('#custom_metrics').modal('hide')
    },
    editCustomMetricItem (rowData) {
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.display_name
      this.customMetricsModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.$root.JQ('#custom_metrics').modal('show')
    },
    editRuleItem (rowData) {
      // this.ruleModelConfig.isAdd = false
      // this.ruleModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      // this.ruleModelConfig.isShow = true
      if (rowData.log_type === 'custom') {
        this.$refs.customRegexRef.loadPage('view', '', rowData.log_metric_monitor, rowData.guid)
      } else {
        this.$refs.businessMonitorGroupConfigRef.loadPage('view', rowData.log_monitor_template, rowData.log_metric_monitor, rowData.guid)
      }
    },
    getExtendInfo (item) {
      const guid = item.guid
      // // eslint-disable-next-line no-redeclare
      let index = null
      this.totalData.forEach((td, tdIndex) => {
        const res = td.tableData.config.findIndex(x => x.guid === guid)
        if (res !== -1) {
          index = tdIndex
        }
      })
      item.metric_groups.forEach(xx => xx.pId = item.guid)
      this.totalData[index].tableConfig.table.isExtend.detailConfig[0].data = item.metric_groups.map(group => {
        const typeToName = {
          custom: this.$t('m_custom_regex'),
          regular: this.$t('m_standard_regex'),
          json: this.$t('m_standard_json'),
        }
        group.log_type_display = typeToName[group.log_type]
        return group
      })
      this.totalData[index].tableConfig.table.isExtend.parentData = item
    },
    getDetail (targrtId) {
      this.showManagement = false
      this.showDbManagement = false
      this.targrtId = targrtId
      if (this.targrtId.endsWith('_mysql')) {
        this.getDbDetail(targrtId)
        return
      }
      const api = this.$root.apiCenter.getTargetDetail + '/endpoint/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.totalData = responseData.map(d => {
          let tmp = {}
          tmp.tableData = d
          tmp.tableConfig = {
            table: {
              tableData: d.config,
              tableEle: tableEle,
              // filterMoreBtn: 'filterMoreBtn',
              primaryKey: 'id',
              btn: btn,
              handleFloat:true,
              isExtend: {
                parentData: null,
                func: 'getExtendInfo',
                data: {},
                slot: 'tableExtend',
                detailConfig: [{
                  isExtendF: true,
                  title: '',
                  config: [
                    {title: 'm_configuration_name', value: 'name', display: true},
                    {title: 'm_associated_template', value: 'log_monitor_template_name', display: true},
                    {title: 'm_metric_config_type', value: 'log_type_display', display: true},
                    {title: 'm_updatedBy', value: 'update_user', display: true},
                    {title: 'title.updateTime', value: 'update_time', display: true},
                    {title: 'table.action',btn:[
                      {btn_name: 'button.view', btn_func: 'editRuleItem'}
                    ]}
                  ],
                  data: [1],
                  scales: ['25%', '20%', '15%', '20%', '20%']
                }]
              }
              // isCustomMetricExtend: {
              //   parentData: null,
              //   func: 'getExtendInfo',
              //   data: {},
              //   slot: 'rulesTableExtend',
              //   detailConfig: [{
              //     isExtendF: true,
              //     title: '',
              //     config: [
              //       {title: 'tableKey.regular', value: 'regular', display: true},
              //       {title: 'field.metric', value: 'metric', display: true},
              //       {title: 'field.aggType', value: 'agg_type', display: true},
              //       {title: 'table.action',btn:[
              //         {btn_name: 'button.view', btn_func: 'editCustomMetricItem'}
              //       ]}
              //     ],
              //     data: [1],
              //     scales: ['25%', '20%', '15%', '20%', '20%']
              //   }]
              // }
            }
          }
          return tmp
        })
        this.showManagement = true
        this.$root.$store.commit('changeTableExtendActive', -1)
      }, {isNeedloading:true})
    }
  },
  mounted () {
    this.MODALHEIGHT = document.body.scrollHeight - 300
  },
  components: {
    extendTable,
    CustomRegex,
    BusinessMonitorGroupConfig
  }
}
</script>

<style scoped lang="scss">
</style>
