<template>
  <div class="all-page">
    <div class="content">
      <div class="chart-config_view">
        <div class="chart-view">
          <div class="use-underline-title mb-2">
            {{this.$t('m_preview')}}
            <span class="underline"></span>
          </div>
          <div v-if="isChartDataError || isChartSeriesEmpty" class="echart error-chart">
            {{this.$t('m_noData')}}
          </div>
          <div v-else>
            <div :id="elId" class="echart" ></div>
          </div>
        </div>
        <div class="chart-config">
          <div class="use-underline-title ml-4 mb-2">
            {{this.$t('m_chart_configuration')}}
            <span class="underline"></span>
          </div>

          <Form ref="formData" :model="chartConfigForm" :rules="ruleValidate" :label-width="100">
            <FormItem :label="$t('m_graph_name')" prop="name">
              <Input v-model.trim="chartConfigForm.name" :maxlength="100" show-word-limit/>
            </FormItem>
            <div v-if="isPieChart">
              <FormItem :label="$t('m_show_type')" prop="pieType">
                <Select
                  v-model="chartConfigForm.pieType"
                  filterable
                >
                  <Option
                    v-for="item in pieTypeOptions"
                    :value="item.value"
                    :key="item.label"
                  >
                    {{ item.label }}
                  </Option>
                </Select>
              </FormItem>
            </div>
            <div v-else>
              <FormItem :label="$t('m_chart_template')" prop="chartTemplate">
                <Select
                  v-model="chartConfigForm.chartTemplate"
                  filterable
                  @on-change="onChartTemplateSelected"
                >
                  <Option
                    v-for="item in chartTemplateOptions"
                    :value="item.key"
                    :key="item.label"
                    :label="item.label"
                  >
                    {{ item.label }}
                  </Option>
                </Select>
              </FormItem>
              <FormItem :label="$t('m_graph_type')" prop="lineType">
                <Select
                  filterable
                  v-model="chartConfigForm.lineType"
                  @on-change="onLineTypeChange"
                >
                  <Option
                    v-for="item in lineTypeOptions"
                    :value="item.value"
                    :label="item.label"
                    :key="item.value"
                  >
                    {{ item.label }}
                  </Option>
                </Select>
              </FormItem>
              <template v-if="chartConfigForm.lineType !== 'twoYaxes'">
                <FormItem :label="$t('m_computed_type')" prop="aggregate">
                  <Select
                    filterable
                    v-model="chartConfigForm.aggregate"
                    @on-change="onAggregateChange"
                  >
                    <Option
                      v-for="item in aggregateOptions"
                      :value="item"
                      :label="item"
                      :key="item"
                    >
                      {{ item }}
                    </Option>
                  </Select>
                </FormItem>
                <FormItem v-if="chartConfigForm.aggregate !== 'none'" :label="$t('m_calculation_period')" prop="aggStep">
                  <Select
                    filterable
                    v-model="chartConfigForm.aggStep"
                    @on-change="onAggStepChange"
                  >
                    <Option
                      v-for="item in aggStepOptions"
                      :value="item.value"
                      :label="item.label"
                      :key="item.value"
                    >
                      {{ item.label }}
                    </Option>
                  </Select>
                </FormItem>
                <FormItem :label="$t('m_unit')" prop="unit">
                  <Input v-model="chartConfigForm.unit" :maxlength="10"/>
                </FormItem>
              </template>
            </div>
          </Form>
        </div>
      </div>
      <div class="data-config">
        <div class="w-header" slot="title">
          <div class="title">
            {{this.$t('m_menu_metricConfiguration') }}
            <span class="underline"></span>
          </div>
        </div>
        <Table
          class="config-table"
          size="small"
          style="width:100%;"
          :border="false"
          :columns="finalTableColumns"
          :data="tableData"
        />

        <div v-if="!isPieChart" class="add-data-configuration">
          <Select
            v-model="endpointValue"
            filterable
            clearable
            class="mr-3"
            style="width: 260px"
            ref="select"
            :placeholder="$t('m_layer_endpoint')"
            @on-query-change="(e) => {this.getEndpointSearch = e; this.debounceGetEndpointList()}"
            @on-change="searchTypeByEndpoint"
          >
            <Option v-for="(option, index) in endpointOptions" :value="option.option_value" :label="option.option_text" :key="index">
              <TagShow :list="endpointOptions" name="option_type_name" :tagName="option.option_type_name" :index="index"/>
              {{option.option_text}}
            </Option>
          </Select>

          <Select
            v-model="monitorType"
            filterable
            class="mr-3"
            style="width: 150px"
            :placeholder="$t('m_endpoint_type')"
            @on-change="searchMetricByType"
          >
            <Option v-for="item in monitorTypeOptions" :value="item" :label="item" :key="item">
              {{item}}
            </Option>
          </Select>
          <Select
            v-model="metricGuid"
            class="metric-guid-select"
            filterable
            style="width: 300px"
            clearable
            :placeholder="$t('m_metric')"
            @on-change="searchTagOptions"
          >
            <Option v-for="(option, index) in metricOptions" :value="option.guid" :label="option.metric" :key="index">
              <Tag type="border" :color="metricTypeMap[option.metric_type].color">{{metricTypeMap[option.metric_type].label}}</Tag>
              {{option.metric}}
            </Option>
          </Select>
          <div v-if="chartAddTags.length" class="add-tag-configuration">
            <div v-for="(tag, index) in chartAddTags" :key="index" class="mb-1">
              <span class="mr-1">{{tag.tagName}}</span>
              <Select
                v-model="chartAddTags[index].tagValue"
                style="max-width:130px"
                filterable
                multiple
                clearable
              >
                <Option v-for="(option, key) in chartAddTagOptions[tag.tagName]" :key="key" :value="option.value">
                  {{option.key}}
                </Option>
              </Select>
            </div>
          </div>
          <div style="width: 250px" v-else></div>

          <Button
            v-if="operator !== 'view'"
            class="add-configuration-button"
            :disabled="!endpointValue || !monitorType || !metricGuid"
            @click="addConfiguration"
            type="success"
          >{{$t('m_button_add')}}</Button>
        </div>
      </div>
    </div>
    <div v-if="operator !== 'view'" class="config-footer">
      <Button class="mr-4" @click="resetChartConfig">{{$t('m_reset')}}</Button>
      <Button class="save-chart-library mr-4" v-if="!isNotSaveChartLibraryButtonShow" :disabled="chartPublic" @click="saveChartLibrary" :type="chartPublic ? 'default' : 'primary'">{{$t('m_save_chart_library')}}</Button>
      <Button type="primary" @click="saveChartConfig">{{$t('m_save')}}</Button>
    </div>
    <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveChartAuth" />
  </div>
</template>

<script>
import cloneDeep from 'lodash/cloneDeep'
import find from 'lodash/find'
import isEmpty from 'lodash/isEmpty'
import remove from 'lodash/remove'
import debounce from 'lodash/debounce'
import Vue from 'vue'
import TagShow from '@/components/Tag-show.vue'
import AuthDialog from '@/components/auth.vue'
import { readyToDraw, drawPieChart} from '@/assets/config/chart-rely'
import { generateUuid, getRandomColor } from '@/assets/js/utils'
import { changeSeriesColor } from '@/assets/config/random-color'
const initTableData = [
  {
    endpoint: '',
    serviceGroup: '',
    endpointName: '',
    monitorType: '',
    colorGroup: '',
    pieDisplayTag: '',
    endpointType: '',
    metricType: '',
    metricGuid: '',
    metric: '',
    tags: [],
    series: []
  }
]

const equalOptionList = [
  {
    name: 'm_include',
    value: 'in'
  },
  {
    name: 'm_not_include',
    value: 'notin'
  }
]

export default {
  name: '',
  props: {
    chartId: String,
    operator: String
  },
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      // chartId: "664eff8cbd85ad9a",
      elId: '',
      chartConfigForm: {
        name: '',
        chartTemplate: '',
        chartType: '',
        lineType: '',
        pieType: '',
        aggregate: '',
        aggStep: null,
        unit: ''
      },
      ruleValidate: {
        name: [
          {
            type: 'string',
            required: true,
            message: '请输入名称',
            trigger: 'blur'
          }
        ],
        chartTemplate: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'change'
          }
        ],
        lineType: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        pieType: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        aggregate: [
          {
            type: 'string',
            required: true,
            message: '请输入',
            trigger: 'blur'
          }
        ],
        aggStep: [
          {
            type: 'number',
            required: false,
            message: '请输入',
            trigger: 'blur'
          }
        ],
      },
      tableData: [],
      // modal中的线性表格配置
      lineChartConfigurationColumns: [
        {
          title: this.$t('m_layer_endpoint'),
          width: 300,
          render: (h, params) => params.row.endpointType.length ? (
            <div class="table-config-endpoint">
              <TagShow class="table-endpoint-tag" list={this.tableData} name="endpointType" tagName={params.row.endpointType} index={params.index} />
              <span class="table-endpoint-text">{params.row.endpointName}</span>
            </div>
          ) : (
            <div>{params.row.endpointName}</div>
          )
        },
        {
          title: this.$t('m_endpoint_type'),
          width: 150,
          key: 'monitorType',
          render: (h, params) => params.row.monitorType ? (
            <Button size="small" style="font-size: 12px">{params.row.monitorType}</Button>
          ) : <span>-</span>
        },
        {
          title: this.$t('m_indicator_color_system'),
          key: 'metric',
          width: 400,
          render: (h, params) => (
            <div class="indicator_color_system">
              {params.row.metricType ? <Tag class="indicator_system_tag" type="border" color={this.metricTypeMap[params.row.metricType].color}>{this.metricTypeMap[params.row.metricType].label}</Tag> : <span/>}
              <div class="metric-text ml-1 mr-1">{params.row.metric}</div>
              <ColorPicker value={params.row.colorGroup}
                on-on-open-change={
                  isShow => this.changeColorGroup(isShow, this.tableData[params.index], 'colorGroup')
                }
              />
            </div>
          )
        },
        {
          title: this.$t('m_label_value'),
          key: 'labelValue',
          width: 340,
          render: (h, params) => {
            this.joinTagValuesToOptions(params.row.tags, params.row.tagOptions, params.index)
            return (
              <div>
                {!isEmpty(params.row.tagOptions) && !isEmpty(params.row.tags)
                  ? (params.row.tags.map((i, selectIndex) => (
                    <div class="tags-show" key={selectIndex + '' + JSON.stringify(i)}>
                      <span>{i.tagName}</span>
                      <Select
                        style="width: 80px"
                        value={i.equal}
                        on-on-change={v => {
                          Vue.set(this.tableData[params.index].tags[selectIndex], 'equal', v)
                        }}
                        filterable>
                        {equalOptionList.map((item, index) => (
                          <Option value={item.value} key={item.value + index}>
                            {this.$t(item.name)}
                          </Option>
                        ))}
                      </Select>

                      <Select
                        style="maxWidth: 180px"
                        value={i.tagValue}
                        on-on-change={v => {
                          Vue.set(this.tableData[params.index].tags[selectIndex], 'tagValue', v)
                          this.updateAllColorLine(params.index)
                        }}
                        filterable
                        multiple
                        clearable
                      >
                        {!isEmpty(this.tableData[params.index].tagOptions)
                          && !isEmpty(this.tableData[params.index].tagOptions[i.tagName])
                            && this.tableData[params.index].tagOptions[i.tagName].map((item, index) => (
                              <Option value={item.value} key={item.key + index}>
                                {item.key}
                              </Option>
                            ))}
                      </Select>
                    </div>
                  ))) : '-' }
              </div>
            )
          }
        },
        {
          title: this.$t('m_generate_lines'),
          key: 'series',
          minWidth: 350,
          render: (h, params) => (
            <div>
              {!isEmpty(params.row.series)
                ? (params.row.series.map((item, selectIndex) => (
                  <div class="generate-lines">
                    {item.new ? <Tag class="new-line-tag" color="error">{this.$t('m_new')}</Tag> : <span/>}
                    <div class="series-name mr-2">{item.seriesName}</div>
                    <ColorPicker v-model={item.color}
                      on-on-open-change={
                        isShow => this.changeColorGroup(isShow, this.tableData[params.index].series[selectIndex], 'color')
                      }
                    />
                  </div>
                ))) : '-' }
            </div>
          )
        },
        {
          title: this.$t('m_table_action'),
          key: 'index',
          width: 80,
          align: 'center',
          fixed: 'right',
          render: (h, params) => (
            <Button disabled={this.operator === 'view'} class="ml-3" size="small" icon="md-trash" type="error" on-click={() => this.removeTableItem(params.index)} />
          )
        }
      ],
      pieChartConfigurationColumns: [
        {
          title: this.$t('m_endpoint'),
          minWidth: 220,
          key: 'endpointName',
          render: (h, params) => (
            <Select
              value={params.row.endpoint}
              on-on-change={v => {
                Vue.set(this.tableData[params.index], 'endpoint', v)
                this.endpointValue = v
                this.searchTypeByEndpointInPie(v, params.index)
              }}
              on-on-query-change={e => {
                this.getEndpointSearch = e
                this.debounceGetEndpointList()
              }}
              filterable
              clearable
            >
              {this.endpointOptions.map((option, index) => (
                <Option class="select-options-change" value={option.option_value} label={option.option_text} key={index}>
                  <TagShow list={this.endpointOptions} name="option_type_name" tagName={option.option_type_name} index={index}></TagShow>
                  {option.option_text}
                </Option>
              ))}
            </Select>
          )
        },
        {
          title: this.$t('m_type'),
          minWidth: 150,
          key: 'monitorType',
          render: (h, params) => (
            <Select
              value={params.row.monitorType}
              on-on-change={v => {
                Vue.set(this.tableData[params.index], 'monitorType', v)
                Vue.set(this.tableData[params.index], 'metricGuid', '')
                this.monitorType = v
                this.searchMetricByType()
              }}
              filterable
              clearable
            >
              {this.monitorTypeOptions.map((item, index) => (
                <Option value={item} label={item} key={index}>
                  {item}
                </Option>
              ))}
            </Select>
          )
        },
        {
          title: this.$t('m_metric'),
          key: 'metricGuid',
          minWidth: 250,
          render: (h, params) => (
            <Select
              value={params.row.metricGuid}
              on-on-change={v => {
                this.addConfigurationInPie(v)
                Vue.set(this.tableData[params.index], 'pieDisplayTag', '')
              }}
              filterable
              clearable
            >
              {this.metricOptions.map((option, index) => (
                <Option class="select-options-change" value={option.guid} label={option.metric} key={index}>
                  <Tag type="border" color={this.metricTypeMap[option.metric_type].color}>{this.metricTypeMap[option.metric_type].label}</Tag>
                  {option.metric}
                </Option>

              ))}
            </Select>
          )
        },
        {
          title: this.$t('m_label_value'),
          key: 'labelValue',
          minWidth: 350,
          render: (h, params) => {
            this.joinTagValuesToOptions(params.row.tags, params.row.tagOptions, params.index)
            return (
              <div>
                {!isEmpty(params.row.tagOptions) && !isEmpty(params.row.tags)
                  ? (params.row.tags.map((i, selectIndex) => (
                    <div class="tags-show" key={selectIndex + '' + JSON.stringify(i)}>
                      <span>{i.tagName}</span>

                      <Select
                        style="width: 80px"
                        value={i.equal}
                        on-on-change={v => {
                          Vue.set(this.tableData[params.index].tags[selectIndex], 'equal', v)
                        }}>
                        {equalOptionList.map((item, index) => (
                          <Option value={item.value} key={item.value + index}>
                            {this.$t(item.name)}
                          </Option>
                        ))}
                      </Select>

                      <Select
                        value={i.tagValue}
                        on-on-change={v => {
                          Vue.set(this.tableData[params.index].tags[selectIndex], 'tagValue', v)
                        }}
                        filterable
                        multiple
                        clearable
                      >
                        {!isEmpty(this.tableData[params.index].tagOptions)
                          && !isEmpty(this.tableData[params.index].tagOptions[i.tagName])
                            && this.tableData[params.index].tagOptions[i.tagName].map((item, index) => (
                              <Option value={item.value} key={item.key + index}>
                                {item.key}
                              </Option>
                            ))}
                      </Select>
                    </div>
                  ))) : '-' }
              </div>
            )
          }
        },
        {
          title: this.$t('m_show_metric'),
          key: 'metric',
          width: 250,
          render: (h, params) => {
            const options = []
            for (const key in params.row.tagOptions) {
              options.push(key)
            }
            return (
              <div>
                {
                  isEmpty(params.row.tags) ? '-'
                    : <Select
                      value={params.row.pieDisplayTag}
                      on-on-change={v => {
                        Vue.set(this.tableData[params.index], 'pieDisplayTag', v)
                      }}
                      filterable
                      clearable
                    >
                      {options.map((item, index) => (
                        <Option value={item} label={item} key={index}>
                          {item}
                        </Option>
                      ))}
                    </Select>
                }
              </div>
            )
          }
        },
      ],
      serviceGroup: '',
      endpointValue: '',
      endpointName: '',
      endpointType: '',
      endpointOptions: [],
      monitorType: '',
      monitorTypeOptions: [],
      metricGuid: '',
      metricOptions: [],
      chartTemplateOptions: [
        {
          label: this.$t('m_customize'),
          key: 'one',
          value: {
            aggregate: 'none',
            aggStep: null,
            chartType: 'line',
            lineType: 'line'
          }
        },
        {
          label: `${this.$t('m_volume')}: sum-60s-${this.$t('m_bar_chart')}`,
          key: 'two',
          value: {
            aggregate: 'sum',
            aggStep: 60,
            chartType: 'bar',
            lineType: 'bar'
          }
        },
        {
          label: `${this.$t('m_avg_consumed')}: avg-60s-${this.$t('m_line_chart')}-${this.$t('m_line_chart_s')}`,
          key: 'three',
          value: {
            aggregate: 'avg',
            aggStep: 60,
            chartType: 'line',
            lineType: 'line'
          }
        },
        {
          label: `${this.$t('m_max_consumed')}: max-60s-${this.$t('m_line_chart')}-${this.$t('m_line_chart_s')}`,
          key: 'four',
          value: {
            aggregate: 'max',
            aggStep: 60,
            chartType: 'line',
            lineType: 'line'
          }
        }
      ],
      aggStepOptions: [
        {
          label: '60S',
          value: 60
        },
        {
          label: '300S',
          value: 300
        },
        {
          label: '600S',
          value: 600
        },
        {
          label: '1800S',
          value: 1800
        },
        {
          label: '3600S',
          value: 3600
        }
      ],
      aggregateOptions: ['none', 'min', 'max', 'avg', 'p95', 'sum', 'avg_nonzero'],
      lineTypeOptions: [
        {
          label: this.$t('m_line_chart'),
          value: 'line'
        },
        {
          label: this.$t('m_bar_chart'),
          value: 'bar'
        },
        {
          label: this.$t('m_area_chart'),
          value: 'area'
        },
        {
          label: this.$t('m_two_y_axes'),
          value: 'twoYaxes'
        }

      ],
      pieTypeOptions: [
        {
          label: this.$t('m_tableKey_tags'),
          value: 'tag'
        },
        {
          label: this.$t('m_value'),
          value: 'value'
        }
      ],
      userRolesOptions: [],
      mgmtRolesOptions: [],
      mgmtRoles: [],
      useRoles: [],
      lineTypeOption: {
        twoYaxes: 2,
        line: 1,
        area: 0
      },
      metricTypeMap: {
        common: {
          label: this.$t('m_basic_type'),
          color: '#2d8cf0'
        },
        business: {
          label: this.$t('m_business_configuration'),
          color: '#81b337'
        },
        custom: {
          label: this.$t('m_metric_list'),
          color: '#b886f8'
        }
      },
      chartPublic: false,
      chartAddTagOptions: {},
      chartAddTags: [],
      getEndpointSearch: '',
      selectedEndpointOptionItem: {},
      isChartSeriesEmpty: false
    }
  },
  computed: {
    isPieChart() {
      return this.chartConfigForm.chartType === 'pie'
    },
    finalTableColumns() {
      const pieChartConfigurationColumnNoValue = cloneDeep(this.pieChartConfigurationColumns)
      pieChartConfigurationColumnNoValue.pop()
      return this.isPieChart
        ? (this.chartConfigForm.pieType === 'value' ? pieChartConfigurationColumnNoValue : this.pieChartConfigurationColumns)
        : this.lineChartConfigurationColumns
    },
    isChartDataError() {
      return isEmpty(this.tableData)
    },
    isNotSaveChartLibraryButtonShow() {
      return this.chartPublic && this.$route.path === '/viewConfigIndex/allChartList'
    }
  },
  watch: {
    tableData: {
      handler() {
        this.debounceDrawChart()
      },
      deep: true
    },
    chartConfigForm: {
      handler() {
        this.debounceDrawChart()
      },
      deep: true
    }
  },
  created() {
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  mounted() {
    this.getAllRolesOptions()
    this.getSingleChartAuth()
    this.getTableData()
  },
  methods: {
    async getTableData() {
      this.getEndpointList()
      this.request('GET', '/monitor/api/v2/chart/custom', {
        chart_id: this.chartId
      }, async res => {
        // public是true的时候，是引用态， public为false的时候，为非引用态
        this.chartPublic = res.public
        for (const key in this.chartConfigForm) {
          this.chartConfigForm[key] = res[key]
        }
        if (!this.chartConfigForm.chartTemplate) {
          this.chartConfigForm.chartTemplate = 'one'
        }
        this.tableData = cloneDeep(res.chartSeries)
        if (res.chartType === 'pie' && isEmpty(this.tableData)) {
          this.tableData = cloneDeep(initTableData)
        }
        await this.processRawTableData(this.tableData)
        this.drawChartContent()
      })
    },
    async processRawTableData(initialData) {
      return new Promise(async resolve => {
        if (isEmpty(initialData)) {
          return []
        }
        for (let i=0; i < initialData.length; i++) {
          const item = initialData[i]
          item.tagOptions = await this.findTagsByMetric(item.metricGuid, item.endpoint, item.serviceGroup)
          Vue.set(item, 'tags', this.initTagsFromOptions(item.tagOptions, item.tags))
          // 同环比修改
          if (item.comparison) {
            const basicParams = this.processBasicParams(item.metric, item.endpoint, item.serviceGroup, item.monitorType, item.tags, item.chartSeriesGuid, item)
            item.series = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/series/config', basicParams)
          } else {
            if (isEmpty(item.series) && this.chartConfigForm.chartType !== 'pie') {
              this.updateAllColorLine(i)
            }
          }
        }
        if (this.isPieChart && initialData.length === 1 && initialData[0].endpoint) {
          const selectedEndpointItem = find(cloneDeep(this.endpointOptions), {
            option_value: initialData[0].endpoint
          })
          this.endpointValue = initialData[0].endpoint
          this.serviceGroup = selectedEndpointItem.app_object

          this.request('GET', '/monitor/api/v1/dashboard/recursive/endpoint_type/list', {
            guid: selectedEndpointItem.option_value
          }, res => {
            if (!isEmpty(res)) {
              this.monitorType = initialData[0].monitorType
              this.monitorTypeOptions = res
              this.searchMetricByType()
            }
          })
        }
        resolve(initialData)
      })
    },
    getAllRolesOptions() {
      const params = {
        page: 1,
        size: 1000
      }
      this.request('GET','/monitor/api/v1/user/role/list', params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
        this.mgmtRolesOptions = this.processRolesList(res)
      })
    },
    getSingleChartAuth() {
      this.request('GET','/monitor/api/v2/chart/custom/permission', {
        chart_id: this.chartId
      }, res => {
        this.mgmtRoles = res.mgmtRoles
        this.useRoles = res.useRoles
      })
    },
    processRolesList(list = []) {
      if (isEmpty(list)) {
        return []
      }
      const resArr = cloneDeep(list).map(item => ({
        ...item,
        key: item.name,
        label: item.displayName || item.display_name
      }))
      return resArr
    },
    // 将tagOptions中不存在的tagValue拼到tagOptions中
    joinTagValuesToOptions(tags, tagOptions, index) {
      if (!isEmpty(tagOptions) && !isEmpty(tags)) {
        tags.forEach(item => {
          const tagValue = item.tagValue
          const tagName = item.tagName
          if (!isEmpty(tagValue)) {
            tagValue.forEach(value => {
              const findOptionItem = find(tagOptions[tagName], {
                value
              })
              if (isEmpty(findOptionItem)) {
                this.tableData[index].tagOptions[tagName].push({
                  key: value,
                  value
                })
              }
            })
          }
        })
      }
    },
    initTagsFromOptions(tagOptions, rawTags = []) {
      if (!tagOptions || isEmpty(tagOptions)) {
        return []
      }
      const tags = []
      for (const key in tagOptions) {
        const rawTagItem = find(rawTags, {
          tagName: key
        })
        const tagValue = isEmpty(rawTagItem) || isEmpty(rawTagItem.tagValue) ? [] : [...rawTagItem.tagValue]
        tags.push(
          {
            tagName: key,
            tagValue,
            equal: isEmpty(rawTagItem) || isEmpty(rawTagItem.equal) ? 'in' : rawTagItem.equal
          }
        )
      }
      return tags
    },

    debounceGetEndpointList: debounce(async function () {
      await this.getEndpointList()
    }, 200),

    getEndpointList() {
      return new Promise(resolve => {
        let search = '.'
        if (this.getEndpointSearch) {
          search = this.getEndpointSearch
        }
        const params = {
          search,
          page: 1,
          size: 10000
        }
        this.request('GET', '/monitor/api/v1/dashboard/search', params, res => {
          this.endpointOptions = res
          let tempItem
          if (!isEmpty(this.selectedEndpointOptionItem) && this.selectedEndpointOptionItem.option_value) {
            tempItem = find(res, {
              option_value: this.selectedEndpointOptionItem.option_value
            })
            if (!tempItem) {
              this.endpointOptions = [this.selectedEndpointOptionItem, ...res]
            }
            this.selectedEndpointOptionItem = {}
          }
          resolve(res)
        })
      })
    },
    resetSearchItem(index) {
      this.monitorType = ''
      this.metricGuid = ''
      this.metricOptions = []
      this.monitorTypeOptions = []
      this.tableData[index].monitorType = ''
      this.tableData[index].metric = ''
      this.tableData[index].metricGuid = ''
    },
    searchTypeByEndpointInPie(value, index) {
      this.resetSearchItem(index)
      const selectedItem = find(cloneDeep(this.endpointOptions), {
        option_value: value
      })
      if (!selectedItem) {
        return
      }
      this.serviceGroup = selectedItem.app_object
      this.endpointName = selectedItem.option_text
      this.endpointType = selectedItem.option_type_name
      this.request('GET', '/monitor/api/v1/dashboard/recursive/endpoint_type/list', {
        guid: selectedItem.option_value
      }, res => {
        this.monitorType = isEmpty(res) ? '' : res[0]
        Vue.set(this.tableData[index], 'monitorType', this.monitorType)
        Vue.set(this.tableData[index], 'metric', '')
        this.metricOptions = []
        this.monitorTypeOptions = res
        this.searchMetricByType()
      })

    },
    searchTypeByEndpoint(value) {
      this.selectedEndpointOptionItem = find(cloneDeep(this.endpointOptions), {
        option_value: value
      })
      setTimeout(() => {
        this.getEndpointSearch = ''
        this.debounceGetEndpointList()
      }, 0)
      this.monitorType = ''
      this.metricGuid = ''
      this.metricOptions = []
      this.monitorTypeOptions = []
      const selectedItem = find(cloneDeep(this.endpointOptions), {
        option_value: value
      })
      if (selectedItem && !isEmpty(selectedItem)) {
        this.endpointName = selectedItem.option_text
        this.endpointType = selectedItem.option_type_name
        this.serviceGroup = selectedItem.app_object
        const params = {
          guid: selectedItem.option_value
        }
        this.request('GET', '/monitor/api/v1/dashboard/recursive/endpoint_type/list', params, res => {
          this.monitorType = isEmpty(res) ? '' : res[0]
          this.monitorTypeOptions = res
          this.searchMetricByType()
        })
      }
    },
    searchMetricByType() {
      this.metricGuid = ''
      this.metricOptions = []
      if (!this.endpointValue || !this.monitorType) {
        return
      }

      const api = '/monitor/api/v2/monitor/metric/list'
      const params = {
        monitorType: this.monitorType,
        serviceGroup: this.endpointValue,
        query: this.chartConfigForm.lineType === 'twoYaxes' ? 'comparison' : 'all'
      }
      this.request('GET', api, params, res => {
        this.metricOptions = this.filterHasUsedMetric(res)
      })
    },

    // 删掉已经使用的指标
    filterHasUsedMetric(options) {
      if (this.isPieChart || isEmpty(this.tableData)) {
        return options
      }
      this.tableData.forEach(item => {
        remove(options, single => item.metric === single.metric && this.endpointValue === item.endpoint)
      })
      return options
    },

    processBasicParams(metric, endpoint, serviceGroup, monitorType, tags, chartSeriesGuid = '', allItem = {}) {
      let tempTags = tags
      if (allItem.comparison && !isEmpty(tags) && tags[0].tagName === 'calc_type' && isEmpty(tags[0].tagValue)) {
        tempTags = []
      }
      return {
        metric,
        endpoint,
        serviceGroup,
        monitorType,
        tags: tempTags,
        chartSeriesGuid
      }
    },
    async searchTagOptions() {
      this.chartAddTagOptions = await this.findTagsByMetric(this.metricGuid, this.endpointValue, this.serviceGroup)
      this.chartAddTags = this.initTagsFromOptions(this.chartAddTagOptions)
    },
    findTagsByMetric(metricId, endpoint, serviceGroup) {
      const api = '/monitor/api/v2/metric/tag/value-list'
      const params = {
        metricId,
        endpoint,
        serviceGroup
      }
      return new Promise(resolve => {
        this.request('POST', api, params, responseData => {
          const result = {}
          if (!isEmpty(responseData)) {
            responseData.forEach(item => {
              result[item.tag] = item.values
            })
          }
          resolve(result)
        })
      })
    },
    async addConfigurationInPie(metricGuid) {
      if (this.chartConfigForm.chartType === 'pie' && this.tableData.length === 1) {
        const item = this.tableData[0]
        const metricItem = find(this.metricOptions, {
          guid: metricGuid
        })
        if (!metricItem) {
          return
        }
        item.metricGuid = metricItem.guid
        Vue.set(item, 'metric', metricItem.metric)
        item.metricType = metricItem.metric_type
        item.tagOptions = await this.findTagsByMetric(metricItem.guid, this.endpointValue, this.serviceGroup)
        item.tags = this.initTagsFromOptions(item.tagOptions, item.tags)
        item.serviceGroup = this.serviceGroup
        item.endpointName = this.endpointName
        item.endpointType = this.endpointType
      }
    },
    async addConfiguration() {
      if (this.endpointValue && this.metricGuid) {
        const metricItem = find(this.metricOptions, {
          guid: this.metricGuid
        })
        const basicParams = this.processBasicParams(metricItem.metric, this.endpointValue, this.serviceGroup, this.monitorType, this.chartAddTags, '', {})
        const series = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/series/config', basicParams)
        const colorGroup = getRandomColor()
        const item = {
          endpoint: this.endpointValue,
          serviceGroup: this.serviceGroup,
          endpointName: this.endpointName,
          endpointType: this.endpointType,
          metricGuid: metricItem.guid,
          metricType: metricItem.metric_type,
          monitorType: this.monitorType,
          colorGroup,
          pieDisplayTag: '',
          metric: metricItem.metric,
          tags: this.chartAddTags,
          tagOptions: this.chartAddTagOptions
        }
        item.series = changeSeriesColor(series, colorGroup)
        this.tableData.push(item)
        this.metricGuid = ''
        this.endpointValue = ''
        this.monitorType = ''
        this.chartAddTagOptions = {}
        this.chartAddTags = []
        this.getEndpointSearch = ''
        this.debounceGetEndpointList()
      }
    },
    onChartTemplateSelected(key) {
      const template = find(this.chartTemplateOptions, {
        key
      })
      Object.assign(this.chartConfigForm, template.value)
    },
    requestReturnPromise(method, api, params) {
      return new Promise(resolve => {
        this.request(method, api, params, res => {
          resolve(res)
        })
      })
    },
    removeTableItem(index) {
      this.$delete(this.tableData, index)
      this.searchMetricByType()
    },
    resetChartConfig() {
      this.getTableData()
    },
    chartDuplicateNameCheck(chartId, chartName, isPublic = 0) {
      return new Promise(resolve => {
        if (!chartName) {
          resolve(true)
          return
        }
        const params = {
          chart_id: chartId,
          name: chartName
        }
        if (isPublic) {
          params.public = 1 // 是否存入图表库，1表示是
        }
        this.request('GET', '/monitor/api/v2/chart/custom/name/exist', params, res => {
          if (res) {
            this.$Message.error(isPublic ? (this.$t('m_chart_library') + this.$t('m_name') + this.$t('m_cannot_be_repeated')) : (this.$t('m_graph_name') + this.$t('m_cannot_be_repeated')))
          }
          resolve(res)
        })
      })
    },
    async beforeSaveValid() {
      const validResult = await this.$refs.formData.validate()
      const isDuplicateName = await this.chartDuplicateNameCheck(this.chartId, this.chartConfigForm.name)
      // const hasSetData = this.tableData.length > 0;
      const hasSetData = this.tableData.length > 0 && this.tableData[0].metric && this.tableData[0].monitorType && this.tableData[0].endpoint
      if (!hasSetData) {
        this.$Message.error(this.$t('m_configuration') + this.$t('m_cannot_be_empty'))
      }
      return new Promise(resolve => {
        resolve(validResult && !isDuplicateName && hasSetData)
      })
    },
    async saveChartConfig() {
      if (await this.beforeSaveValid()) {
        await this.submitChartConfig()
        this.$Message.success(this.$t('m_success'))
        this.$parent.$parent.showChartConfig = false
        this.$parent.$parent.closeChartInfoDrawer()
      }
    },
    async saveChartLibrary() {
      if (await this.beforeSaveValid()) {
        await this.submitChartConfig()
        const isDuplicateRefName = await this.chartDuplicateNameCheck(this.chartId, this.chartConfigForm.name, 1)
        if (isDuplicateRefName) {
          return
        }
        this.$refs.authDialog.startAuth(this.mgmtRoles, this.useRoles, this.mgmtRolesOptions, this.userRolesOptions)
      }
    },
    async submitChartConfig() {
      this.debounceDrawChart()
      return new Promise(resolve => {
        let chartSeries = cloneDeep(this.tableData)
        chartSeries = chartSeries.map(item => {
          delete item.tagOptions
          return item
        })
        const params = Object.assign({}, this.chartConfigForm, {
          id: this.chartId,
          chartSeries
        })
        this.request('PUT', '/monitor/api/v2/chart/custom', params, () => {
          resolve()
        })
      })
    },

    onLineTypeChange(lineType) {
      if (lineType === 'bar') {
        this.chartConfigForm.chartType = 'bar'
      } else if (lineType === 'twoYaxes') {
        this.chartConfigForm.aggStep = 60
        this.chartConfigForm.aggregate = 'none'
        this.chartConfigForm.unit = ''
      } else {
        this.chartConfigForm.chartType = 'line'
      }
      this.resetChartTemplate()
      this.searchMetricByType()
    },
    onAggregateChange() {
      this.resetChartTemplate()
    },
    onAggStepChange() {
      this.resetChartTemplate()
    },
    resetChartTemplate() {
      this.chartConfigForm.chartTemplate = 'one'
    },
    saveChartAuth(mgmtRoles, useRoles) {
      this.mgmtRoles = mgmtRoles
      this.useRoles = useRoles
      const path = '/monitor/api/v2/chart/custom/permission'
      this.request('POST', path, {
        chartId: this.chartId,
        mgmtRoles,
        useRoles
      }, () => {
        this.$Message.success(this.$t('m_success'))
        this.chartPublic = true
      })
    },
    // 将数据拼好，请求数据并画图
    drawChartContent() {
      if (this.isPieChart) {
        const params = this.generateLineParamsData()
        params[0].pieType = this.chartConfigForm.pieType
        if (!params[0].metric) {
          return
        }
        this.request(
          'POST',
          '/monitor/api/v1/dashboard/pie/chart',
          params,
          res => {
            drawPieChart(this, res)
          }
        )
      } else {
        const params = {
          aggregate: this.chartConfigForm.aggregate || 'none',
          agg_step: this.chartConfigForm.aggStep || 60,
          chartType: this.chartConfigForm.chartType || 'line',
          lineType: this.lineTypeOption[this.chartConfigForm.lineType],
          time_second: -1800,
          start: 0,
          end: 0,
          title: '',
          unit: '',
          data: this.generateLineParamsData()
        }
        if (isEmpty(params.data)) {
          return
        }
        if (!isEmpty(params.data) && params.data.every(item => isEmpty(item.series))) {
          this.isChartSeriesEmpty = true
          return
        }
        this.request(
          'POST',
          '/monitor/api/v1/dashboard/chart',
          params,
          responseData => {
            if (isEmpty(responseData.series)) {
              this.isChartSeriesEmpty = true
              return
            }
            responseData.yaxis.unit = this.chartConfigForm.unit
            this.isChartSeriesEmpty = false
            setTimeout(() => {
              readyToDraw(this, responseData, 1, {
                eye: false,
                lineBarSwitch: true,
                chartType: this.chartConfigForm.chartType,
                params
              })
            }, 100)
          }
        )
      }
    },

    debounceDrawChart: debounce(function () {
      this.drawChartContent()
    }, 50),

    generateLineParamsData() {
      if (isEmpty(this.tableData)) {
        return []
      }
      const data = cloneDeep(this.tableData).map(item => {
        item.app_object = item.serviceGroup
        item.defaultColor = item.colorGroup

        if (item.series && !isEmpty(item.series)) {
          item.metricToColor = cloneDeep(item.series).map(one => {
            one.metric = one.seriesName
            delete one.seriesName
            return one
          })
        } else {
          item.metricToColor = []
        }
        delete item.colorGroup
        delete item.tagOptions
        return item
      })
      return data
    },

    async updateAllColorLine(index) {
      const item = this.tableData[index]
      const basicParams = this.processBasicParams(item.metric, item.endpoint, item.serviceGroup, item.monitorType, item.tags, item.chartSeriesGuid, item)
      const series = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/series/config', basicParams)
      this.tableData[index].series = changeSeriesColor(series, this.tableData[index].colorGroup)
    },
    changeColorGroup(isShow = true, data, key) {
      if (isShow) {
        this.$nextTick(() => {
          const confirmButtonList = document.querySelectorAll('.ivu-color-picker-confirm .ivu-btn-primary')
          const resetButtonList = document.querySelectorAll('.ivu-color-picker-confirm .ivu-btn-default')
          if (isEmpty(confirmButtonList)) {
            return
          }
          confirmButtonList[0].addEventListener('click', () => {
            const inputList = document.querySelectorAll('.ivu-color-picker-confirm .ivu-input')
            if (isEmpty(inputList)) {
              return
            }
            const color = inputList[0].value
            data[key] = color
            if (key === 'colorGroup') {
              if (Array.isArray(data.series) && !isEmpty(data.series)) {
                changeSeriesColor(data.series, color)
              }
            }
          })
          if (isEmpty(resetButtonList)) {
            return
          }
          resetButtonList[0].addEventListener('click', () => {
            data[key] = ''
          })
        })
      }
    }
  },
  components: {
    TagShow,
    AuthDialog
  }
}

</script>

<style lang="less">
.chart-config {
  .ivu-form-item {
    margin-bottom: 10px !important;
  }
  .ivu-form-item-error-tip {
    padding-top: 2px;
  }
}

// .add-data-configuration > div {
//   width: 20%;
// }

.save-chart-library.ivu-btn-primary {
  background-color: #b088f1;
  border-color: #b088f1;
}

.indicator_color_system {
  display: flex;
  flex-direction: row;
  align-items: center;
  .indicator_system_tag {
    width: fit-content;
    padding: 0 10px;
  }
  .metric-text {
    // width: 40px;
    max-width: 220px;
  }
}

.config-table {
  // .ivu-table-cell {
  //   padding-left: 0px;
  //   padding-right: 0px
  // }
}

.table-config-endpoint {
  display: flex;
  align-items: center;
  .diy-tag {
    width: 105px;
  }
  .table-endpoint-tag {
    width: 105px;
    padding: 0 10px;
    display: flex;
    justify-content: center;
    align-items: center;
  }
}

.generate-lines {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  .series-name {
    max-width: 85%;
    overflow: hidden;

  }
  .new-line-tag {
    min-width: 28px;
  }
}

.select-options-change {
  display: flex;
  align-items: center;
}

.tags-show {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: 5px;
}
.tags-show > span {
  width: 70px;
}

.tags-show .ivu-select-item.ivu-select-item-selected::after {
    top: 6px;
    right: -6px
}

.ivu-table-wrapper {
  overflow: inherit;
}

.ivu-color-picker {
  .ivu-icon.ivu-icon-ios-close::before {
    content: "\f193"
  }
}

.metric-guid-select {
  .ivu-select-item {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
  }
}

</style>

<style scoped lang='less'>

.all-page {
  .content {
    min-height: 80vh;
    margin-bottom: 50px;
    .chart-config_view {
      display: flex;
      flex-direction: row;
      border-bottom: 1px solid #dcdee2;
      .chart-view {
        flex-direction: column;
        width: 70%;
        padding-right: 20px;
        border-right: 1px solid #dcdee2;
        .echart {
          width: 100%;
          height: 300px;
        }
        .error-chart {
          display: flex;
          justify-content: center;
          align-items: center;
        }
      }
      .chart-config {
        width: 30%;
      }
    }

    .data-config {
      display: flex;
      flex-direction: column;
      margin-top: 10px;
      .use-underline-title {
        width: 66px
      }
      .add-data-configuration {
        // position: relative;
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
        justify-content: flex-start;
        align-items: center;
        margin-top: 15px;
        padding: 20px;
        width: 100%;
        background-color: #efefef;
        .add-tag-configuration {
          width: 250px;
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          margin-left: 20px
        }
        .add-configuration-button {
          margin-left: auto;
        }
      }
    }

  }
  .config-footer {
    position: fixed;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 50px;
    margin-top: 30px;
    bottom: 0px;
    width: 100%;
    background-color: #fff;
    opacity: 1;
    z-index: 100000;
  }

}

.use-underline-title {
  display: inline-block;
  font-size: 16px;
  font-weight: 700;
  margin: 0 10px;
  .underline {
    display: block;
    margin-top: -10px;
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

.w-header {
  display: flex;
  align-items: center;
  .title {
    font-size: 16px;
    font-weight: bold;
    margin: 0 10px;
    .underline {
      display: block;
      margin-top: -10px;
      margin-left: -6px;
      width: 100%;
      padding: 0 6px;
      height: 12px;
      border-radius: 12px;
      background-color: #c6eafe;
      box-sizing: content-box;
    }
  }
}
</style>
