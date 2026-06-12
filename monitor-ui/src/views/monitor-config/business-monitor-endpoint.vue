<template>
  <div class=" ">
    <section v-if='!isEmpty(allPageContentData)' style="margin-top: 16px;">
      <div v-for="(single, i) in allPageContentData" :key="i">
        <div class="content-header">
          <div class="use-underline-title mr-4">
            {{ !isEmpty(single) ? single.display_name : ''}}
            <span class="underline"></span>
          </div>
          <Tag color="blue">{{ $t('m_field_resourceLevel') }}</Tag>
        </div>
        <div class="content-header">
          <div class="use-underline-title mr-4">
            {{$t('m_log_file')}}
            <span class="underline"></span>
          </div>
        </div>
        <Collapse v-model="single.logFileCollapseValue" v-if='!isEmpty(single) && !isEmpty(single.config)'>
          <Panel v-for="(item, index) in single.config"
                 :key="index"
                 :name="index + ''"
          >
            <div class="log-file-collapse-content">
              <div>
                <div class="use-underline-title mr-4">
                  {{item.log_path}}
                  <span class="underline"></span>
                </div>
                <Tag color="blue">{{ item.monitor_type }}</Tag>
              </div>
              <div class="log-file-collapse-button">
                <Button class="mr-5" size="small" type="info" @click.stop="editF(item)">
                  <Icon type="md-eye" size="16" />
                </Button>
              </div>

            </div>
            <template slot='content'>
              <Table
                class="log-file-table"
                size="small"
                :columns="logFileTableColumns"
                :data="item.metric_groups"
              />
            </template>
          </Panel>
        </Collapse>
        <div v-else class='no-logfile-data'>{{$t('m_table_noDataTip')}}</div>

        <div class="database-title">
          <div class="use-underline-title">
            {{$t('m_db')}}
            <span class="underline"></span>
          </div>
        </div>
        <Table
          size="small"
          :columns="dataBaseTableColumns"
          :data="!isEmpty(single) && !isEmpty(single.db_config) ? single.db_config : []"
        />
      </div>
    </section>
    <!-- <div v-else class='no-data-class'>{{$t('m_table_noDataTip')}}</div> -->
    <Modal
      v-model="ruleModelConfig.isShow"
      :title="$t('m_json_regular')"
      :mask-closable="false"
      width="840"
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <Form :label-width="100">
          <FormItem :label="$t('m_tableKey_regular')">
            <Input disabled v-model.trim="ruleModelConfig.addRow.json_regular" style="width:100%"/>
          </FormItem>
          <FormItem :label="$t('m_tableKey_tags')">
            <Input disabled v-model.trim="ruleModelConfig.addRow.tags" style="width:100%" />
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in ruleModelConfig.addRow.metric_list">
            <p :key="index + '3'" style="text-align: center;">
              <Tooltip :content="$t('m_key')" :delay="1000">
                <Input disabled v-model.trim="item.json_key" style="width: 190px" :placeholder="$t('m_key') + ' e.g:[.*][.*]'" />
              </Tooltip>
              <Tooltip :content="$t('m_field_metric')" :delay="1000">
                <Input disabled v-model.trim="item.metric" style="width: 190px" :placeholder="$t('m_field_metric') + ' , e.g:code'" />
              </Tooltip>
              <Tooltip :content="$t('m_field_aggType')" :delay="1000">
                <Select disabled v-model="item.agg_type" filterable clearable style="width:190px">
                  <Option v-for="agg in ruleModelConfig.aggOption" :value="agg" :key="agg">{{
                    agg
                  }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_tableKey_description')" :delay="1000">
                <Input disabled v-model.trim="item.display_name" style="width: 160px" :placeholder="$t('m_tableKey_description')" />
              </Tooltip>
            </p>
            <div v-if="item.string_map.length > 0" :key="index + 1" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
              <template v-for="(stringMapItem, stringMapIndex) in item.string_map">
                <p :key="stringMapIndex + 2" style="text-align: center;">
                  <Tooltip :content="$t('m_tableKey_regular')" :delay="1000">
                    <Select disabled v-model="stringMapItem.regulative" filterable clearable style="width:120px">
                      <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                        regulation.label
                      }}</Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('m_business_object')" :delay="1000">
                    <Input disabled v-model.trim="stringMapItem.target_value" style="width: 230px" :placeholder="$t('m_business_object')" />
                  </Tooltip>
                  <Tooltip :content="$t('m_log_server')" :delay="1000">
                    <Input disabled v-model.trim="stringMapItem.source_value" style="width: 230px" :placeholder="$t('m_log_server')" />
                  </Tooltip>
                </p>
              </template>
            </div>
            <Divider :key="index + 'Q'" />
          </template>
        </div>
      </div>
      <div slot="footer">
        <Button @click="ruleModelConfig.isShow = false">{{$t('m_button_cancel')}}</Button>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_field_aggType')}}:</label>
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
                <Tooltip :content="$t('m_tableKey_regular')" :delay="1000">
                  <Select disabled v-model="item.regulative" filterable clearable style="width:150px">
                    <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                      regulation.label
                    }}</Option>
                  </Select>
                </Tooltip>
                <Tooltip :content="$t('m_business_object')" :delay="1000">
                  <Input disabled v-model.trim="item.target_value" style="width: 150px" :placeholder="$t('m_business_object')" />
                </Tooltip>
                <Tooltip :content="$t('m_log_server')" :delay="1000">
                  <Input disabled v-model.trim="item.source_value" style="width: 150px" :placeholder="$t('m_log_server')" />
                </Tooltip>
              </p>
            </template>
          </div>
        </div>
        <Button style="float:right" @click="cancelModal">{{$t('m_button_cancel')}}</Button>
      </div>
    </ModalComponent>
    <Modal
      v-model="dbModelConfig.isShow"
      :title="$t('m_db')"
      width="680"
      :mask-closable="false"
      footer-hide
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <Form :label-width="100">
          <FormItem :label="$t('m_field_displayName')">
            <Input disabled v-model.trim="dbModelConfig.addRow.display_name" style="width:520px"/>
          </FormItem>
          <FormItem :label="$t('m_field_metric')">
            <Input disabled v-model.trim="dbModelConfig.addRow.metric" style="width:520px" />
          </FormItem>
          <FormItem label="SQL">
            <Input disabled v-model="dbModelConfig.addRow.metric_sql" type="textarea" style="width:520px" />
          </FormItem>
          <FormItem :label="$t('m_field_type')">
            <Select disabled v-model.trim="dbModelConfig.addRow.monitor_type" @on-change="getEndpoint(dbModelConfig.addRow.monitor_type, 'mysql')" style="width: 520px">
              <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
            </Select>
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in dbModelConfig.addRow.endpoint_rel">
            <p :key="index + '3'" style="text-align: center;">
              <Tooltip :content="$t('m_db')" :delay="1000">
                <Input disabled v-model.trim="item.target_endpoint" style="width:290px" />
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Input disabled v-model.trim="item.source_endpoint" style="width:290px" />
              </Tooltip>
            </p>
          </template>
        </div>
      </div>
    </Modal>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="$t('m_button_view')"
      :mask-closable="false"
      :width="720"
    >
      <div v-if="addAndEditModal.isShow" :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <div>
          <span>{{$t('m_field_type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type" disabled style="width: 640px">
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div style="margin: 8px 0">
          <span>{{$t('m_tableKey_path')}}:</span>
          <Input style="width: 640px" disabled v-model.trim="addAndEditModal.dataConfig.log_path" />
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;text-align:center">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Tooltip :content="$t('m_business_object')" :delay="1000">
                <Input disabled v-model.trim="item.source_endpoint" style="width:315px" />
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Input disabled v-model.trim="item.source_endpoint" style="width:315px" />
              </Tooltip>
            </p>
          </template>
        </div>
      </div>
      <div slot="footer">
        <Button @click="addAndEditModal.isShow = false">{{$t('m_button_cancel')}}</Button>
      </div>
    </Modal>
    <CustomRegex ref="customRegexRef"></CustomRegex>
    <BusinessMonitorGroupConfig ref="businessMonitorGroupConfigRef"></BusinessMonitorGroupConfig>
  </div>
</template>

<script>
import {map, isEmpty} from 'lodash'
import CustomRegex from '@/views/monitor-config/log-template-config/custom-regex.vue'
import BusinessMonitorGroupConfig from '@/views/monitor-config/business-monitor-group-config.vue'

export const custom_api_enum = [
  {
    getEndpointsByTypeByType: 'get'
  },
  {
    serviceGroupEendpoint: 'delete'
  },
  {
    logMetricListEndpoint: 'get'
  },
  {
    dbMetricListEndpoint: 'get'
  }
]

export default {
  name: '',
  data() {
    return {
      MODALHEIGHT: 300,
      // showManagement: false,
      regulationOption: [
        {
          label: this.$t('m_regular_match'),
          value: 1
        },
        {
          label: this.$t('m_irregular_matching'),
          value: 0
        }
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
          {
            label: 'm_field_metric',
            value: 'metric',
            placeholder: '',
            disabled: true,
            type: 'text'
          },
          {
            label: 'm_tableKey_description',
            value: 'display_name',
            placeholder: '',
            disabled: true,
            type: 'text'
          },
          {
            label: 'm_tableKey_regular',
            value: 'regular',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: true,
            type: 'text'
          },
          {
            name: 'ruleConfig',
            type: 'slot'
          }
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
            {
              label: this.$t('m_regular_match'),
              value: 1
            },
            {
              label: this.$t('m_irregular_matching'),
              value: 0
            }
          ]
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      // DB config
      // showDbManagement: false,
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
        {
          label: 'process',
          value: 'process'
        },
        {
          label: 'java',
          value: 'java'
        },
        {
          label: 'nginx',
          value: 'nginx'
        },
        {
          label: 'http',
          value: 'http'
        },
        {
          label: 'pod',
          value: 'pod'
        }
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
      allPageContentData: [],
      logFileTableColumns: [
        {
          title: this.$t('m_configuration_name'),
          width: 150,
          key: 'name',
          render: (h, params) => (
            params.row.name
              ? <Tooltip placement="right" max-width="300" content={params.row.name}>
                <div class="table-ellipsis" style="width: 130px">{params.row.name + (params.row.metric_prefix_code ? (' [' + params.row.metric_prefix_code + ']') : '') } </div>
              </Tooltip> : <div>-</div>
          )
        },
        {
          title: this.$t('m_associated_template'),
          width: 200,
          ellipsis: true,
          tooltip: true,
          key: 'log_monitor_template_name',
          render: (h, params) => (
            params.row.log_monitor_template_name
              ? <Tooltip placement="top" max-width="300" content={params.row.log_monitor_template_name}>
                <div class="table-ellipsis" style="width: 180px">{params.row.log_monitor_template_name}</div>
              </Tooltip> : <div>-</div>
          )
        },
        {
          title: this.$t('m_metric_key'),
          minWidth: 300,
          ellipsis: true,
          tooltip: true,
          key: 'log_type_display',
          render: (h, params) => {
            let metricStr = ''
            let metricHtml = ''
            if (params.row.metric_list.length) {
              const metricArr = map(params.row.metric_list, 'metric')
              metricStr = metricArr.join(',')
              metricHtml = '<p>' + metricArr.join('<br>') + '</p>'
            }
            return (
              metricStr
                ? <Tooltip placement="top" max-width="400">
                  <div slot="content" style="white-space: normal;">
                    <div domPropsInnerHTML={metricHtml}></div>
                  </div>
                  <div class="table-ellipsis" style="width: 280px">{metricStr}</div>
                </Tooltip> : <div>-</div>
            )
          }
        },
        {
          title: this.$t('m_metric_config_type'),
          width: 150,
          key: 'log_type_display'
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user'
        },
        {
          title: this.$t('m_title_updateTime'),
          minWidth: 100,
          key: 'update_time'
        },
        {
          title: this.$t('m_table_action'),
          width: 83,
          key: 'index',
          render: (h, params) => (
            <div>
              <Button size="small" class="mr-1" type="info" on-click={() => {
                this.currentLogFileIndex = params.row.logFileIndex; this.editRuleItem(params.row)
              }}>
                <Icon type="md-eye" size="16" />
              </Button>
            </div>
          )
        }
      ],
      typeToName: { // 模版枚举
        custom: this.$t('m_custom_regex'),
        regular: this.$t('m_standard_regex'),
        json: this.$t('m_standard_json'),
      },

      dataBaseTableColumns: [
        {
          title: this.$t('m_metric_key'),
          width: 350,
          key: 'metric'
        },
        {
          title: this.$t('m_field_type'),
          key: 'monitor_type',
          minWidth: 100,
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          width: 100,
        },
        {
          title: this.$t('m_title_updateTime'),
          minWidth: 100,
          key: 'update_time'
        },
        {
          title: this.$t('m_table_action'),
          width: 100,
          key: 'index',
          render: (h, params) => (
            <div>
              <Button size="small" type="info" on-click={() => {
                this.editDbItem(params.row)
              }}>
                <Icon type="md-eye" size="16" />
              </Button>
            </div>
          )
        }
      ],
      dataBaseTableData: [],
      isEmpty,
      metricKey: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  methods: {
    editDbItem(rowData) {
      this.getEndpoint(rowData.monitor_type, 'mysql', rowData.service_group)
      this.dbModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.dbModelConfig.isAdd = false
      this.dbModelConfig.isShow = true
    },
    async getEndpoint(val, type, targetId) {
      const sourceApi = this.apiCenter.getEndpointsByType + '/' + targetId + '/endpoint/' + type
      this.request('GET', sourceApi, '', responseData => {
        this.sourceEndpoints = responseData
      }, {isNeedloading: false})
      const targetApi = this.apiCenter.getEndpointsByType + '/' + targetId + '/endpoint/' + val
      this.request('GET', targetApi, '', responseData => {
        this.targetEndpoints = responseData
      }, {isNeedloading: false})
    },
    // other config
    editF(rowData) {
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
    getDefaultConfig(val, type) {
      const api = `${this.apiCenter.serviceGroupEendpoint}?serviceGroup=${this.targetId}&sourceType=${type}&targetType=${val}`
      this.request('GET', api, '', responseData => {
        const tmp = responseData.map(r => ({
          source_endpoint: r.source_endpoint,
          target_endpoint: r.target_endpoint
        }))
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
    cancelModal() {
      this.$root.JQ('#custom_metrics').modal('hide')
    },
    editCustomMetricItem(rowData) {
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.display_name
      this.customMetricsModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.$root.JQ('#custom_metrics').modal('show')
    },
    editRuleItem(rowData) {
      if (rowData.log_type === 'custom') {
        this.$refs.customRegexRef.loadPage('view', '', rowData.log_metric_monitor, rowData.guid)
      } else {
        this.$refs.businessMonitorGroupConfigRef.loadPage('view', rowData.log_monitor_template, rowData.log_metric_monitor, rowData.guid)
      }
    },
    async getLogKeyWordDetail() {
      return new Promise(resolve => {
        let api = this.apiCenter.getTargetDetail + '/endpoint/' + this.targetId
        if (this.metricKey) {
          api += `?metricKey=${this.metricKey}`
        }
        this.request('GET', api, '', responseData => {
          this.allPageContentData = responseData.map((res, index) => {
            if (index === 0) {
              res.logFileCollapseValue = ['0']
            } else {
              res.logFileCollapseValue = []
            }
            res.config.forEach(item => {
              item.metric_groups.forEach(group => {
                group.log_type_display = this.typeToName[group.log_type]
              })
            })
            return res
          })
          this.$root.$store.commit('changeTableExtendActive', -1)
          resolve(this.allPageContentData)
        }, {isNeedloading: true})
      })
    },
    async getDetail(targetId, metricKey = this.metricKey) {
      if ((metricKey || metricKey === '') && this.metricKey !== metricKey) {
        this.metricKey = metricKey
      }
      this.targetId = targetId
      await this.getLogKeyWordDetail()
      this.$emit('feedbackInfo', this.allPageContentData)
    }
  },
  mounted() {
    this.MODALHEIGHT = document.body.scrollHeight - 300
  },
  components: {
    CustomRegex,
    BusinessMonitorGroupConfig
  }
}
</script>

<style scoped lang="less">
.log-file-collapse-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
}

.use-underline-title {
    display: inline-block;
    font-weight: 700;
    margin: 0 10px;
    .underline {
      display: block;
      margin-top: -20px;
      margin-left: -6px;
      width: 100%;
      padding: 0 6px;
      height: 12px;
      border-radius: 12px;
      background-color: #c6eafe;
      -webkit-box-sizing: content-box;
      box-sizing: content-box;
    }
  }

.content-header {
  font-size: 16px;
  margin-top: 10px;
  margin-bottom: 10px;
  .use-underline-title {
    .underline {
      margin-top: -12px;
    }
  }
}

.database-title {
  font-size: 16px;
  font-weight: bold;
  margin: 10px 0;
  .underline {
    margin-top: -10px
  }
}
.no-logfile-data {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: 20px;
  font-size: 14px;
}
.no-data-class {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: 80px;
  font-size: 16px;
}
</style>
