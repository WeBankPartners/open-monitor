<template>
  <div class="all-page">
    <div class="content">
      <div class="chart-config_view">
        <div class="chart-view">
          <div class="use-underline-title mb-2">
            {{this.$t('m_preview')}}
            <span class="underline"></span>
          </div>
          <div v-if="isChartDataError" class="echart error-chart">
            {{this.$t('m_noData')}}
          </div>
          <div v-else>
            <div :id="elId" class="echart" />
          </div>
        </div>
        <div class="chart-config">
          <div class="use-underline-title ml-4 mb-2">
            {{this.$t('m_chart_configuration')}}
            <span class="underline"></span>
          </div>

           <Form ref="formData" :model="chartConfigForm" :rules="ruleValidate" :label-width="100">
            <FormItem :label="$t('m_graph_name')" prop="name">
              <Input v-model.trim="chartConfigForm.name" :maxlength="30" show-word-limit></Input>
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
                      :key="item.label">
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
                      :label="item.label">
                      {{ item.label }}
                    </Option>
                  </Select>
              </FormItem>
              <FormItem :label="$t('m_type')" prop="lineType">
                  <Select 
                    filterable
                    v-model="chartConfigForm.lineType"
                    @on-change="onLineTypeChange"
                    >
                    <Option 
                      v-for="item in lineTypeOptions" 
                      :value="item.value" 
                      :label="item.label"
                      :key="item.value">
                      {{ item.label }}
                    </Option>
                  </Select>
              </FormItem>
              <FormItem :label="$t('m_computed_type')" prop="aggregate">
                  <Select 
                    filterable
                    v-model="chartConfigForm.aggregate"
                    @on-change="onAggregateChange" >
                    <Option 
                      v-for="item in aggregateOptions" 
                      :value="item"
                      :label="item" 
                      :key="item">
                      {{ item }}
                    </Option>
                  </Select>
              </FormItem>
              <FormItem v-if="chartConfigForm.aggregate !== 'none'" :label="$t('m_calculation_period')" prop="aggStep">
                  <Select 
                    filterable
                    v-model="chartConfigForm.aggStep"
                    @on-change="onAggStepChange" >
                    <Option 
                      v-for="item in aggStepOptions" 
                      :value="item.value" 
                      :label="item.label" 
                      :key="item.value">
                      {{ item.label }}
                    </Option>
                  </Select>
              </FormItem>
              <FormItem :label="$t('m_unit')" prop="unit">
                  <Input v-model="chartConfigForm.unit" :maxlength="10"></Input>
              </FormItem>
            </div>
          </Form>
        </div>
      </div>

      <div class="data-config">
        <div class="use-underline-title mb-2">
          {{this.$t('m_data_configuration')}}
          <span class="underline"></span>
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
            ref="select"
            :placeholder="$t('m_endpoint')"
            @on-query-change="(e) => {this.getEndpointSearch = e; this.debounceGetEndpointList()}"
            @on-change="searchTypeByEndpoint"
            >
            <Option v-for="(option, index) in endpointOptions" :value="option.option_value" :label="option.option_text" :key="index">
              <TagShow :list="endpointOptions" name="option_type_name" :tagName="option.option_type_name" :index="index"></TagShow> 
              {{option.option_text}}
            </Option>
          </Select>

          <Select
            v-model="monitorType"
            filterable
            ref="select"
            :placeholder="$t('m_type')"
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
          <div v-else></div>

          <Button :disabled="!endpointValue || !monitorType || !metricGuid" @click="addConfiguration" type="primary">{{$t('m_add_configuration')}}</Button>
        </div>
      </div>
    </div>
    <div class="config-footer">
      <Button class="mr-4" @click="resetChartConfig">{{$t('m_reset')}}</Button>
      <Button class="save-chart-library mr-4" v-if="!isNotSaveChartLibraryButtonShow" :disabled="chartPublic" @click="saveChartLibrary" :type="chartPublic ? 'default' : 'primary'">{{$t('m_save_chart_library')}}</Button>
      <Button class="mr-4" type="primary" @click="saveChartConfig">{{$t('m_save')}}</Button>
    </div>
    <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveChartAuth" />
  </div>
  
</template>

<script>
import cloneDeep from 'lodash/cloneDeep';
import find from 'lodash/find';
import isEmpty from 'lodash/isEmpty';
import remove from 'lodash/remove';
import debounce from 'lodash/debounce';
import Vue from 'vue';
import TagShow from '@/components/Tag-show.vue'
import AuthDialog from '@/components/auth.vue';
import { readyToDraw, drawPieChart} from "@/assets/config/chart-rely";
import { generateUuid } from "@/assets/js/utils"

const initTableData = [
    {
      "endpoint": "",
      "serviceGroup": "",
      "endpointName": "",
      "monitorType": "",
      "colorGroup": "",
      "pieDisplayTag": "",
      "endpointType": "",
      "metricType": "",
      "metricGuid": "",
      "metric": "",
      "tags": [],
      "series": []
    }
  ]

export default {
  name: "",
  props: {
    chartId: String
  },
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      // chartId: "664eff8cbd85ad9a",
      elId: "",
      chartConfigForm: {
        name: "",
        chartTemplate: "",
        chartType: "",
        lineType: "",
        pieType: "",
        aggregate: "",
        aggStep: null,
        unit: ""
      },
      ruleValidate: {
        name: [
          {type: 'string', required: true, message: '请输入名称', trigger: 'blur' }
        ],
        chartTemplate: [
          {type: 'string', required: true, message: '请输入', trigger: 'change' }
        ],
        lineType: [
          {type: 'string', required: true, message: '请输入', trigger: 'blur' }
        ],
        pieType: [
          {type: 'string', required: true, message: '请输入', trigger: 'blur' }
        ],
        aggregate: [
          {type: 'string', required: true, message: '请输入', trigger: 'blur' }
        ],
        aggStep: [
          {type: 'number', required: false, message: '请输入', trigger: 'blur' }
        ],
      },
      tableData: [],
      lineChartConfigurationColumns: [
        {
          title: this.$t('table.action'),
          key: 'index',
          width: 80,
          render: (h, params) => {
            return (
                <Button size="small" icon="md-trash" type="error" on-click={() => this.removeTableItem(params.index)} />
            )
          }
        },
        {
            title: this.$t('m_endpoint'),
            minWidth: 220,
            render: (h, params) => {
              return params.row.endpointType.length ?  (
                <div class="table-config-endpoint">
                  <TagShow class="table-endpoint-tag" list={this.tableData} name="endpointType" tagName={params.row.endpointType} index={params.index} /> 
                  <span class="table-endpoint-text">{params.row.endpointName}</span>
                </div>
              ) : (
                <div>{params.row.endpointName}</div>
              )
            }
        },
        {
            title: this.$t('m_type'),
            minWidth: 150,
            key: 'monitorType',
            render: (h, params) => {
              return params.row.monitorType ? (
                <Button size="small">{params.row.monitorType}</Button>
              ) : <span>-</span>
            }
        },
        {
          title: this.$t('m_indicator_color_system'),
          key: 'metric',
          minWidth: 350,
          render: (h, params) => {
            return (
              <div class="indicator_color_system">
                {params.row.metricType ? <Tag class="indicator_system_tag" type="border" color={this.metricTypeMap[params.row.metricType].color}>{this.metricTypeMap[params.row.metricType].label}</Tag> : <span/>}
                <div class="metric-text ml-1 mr-1">{params.row.metric}</div>
                <ColorPicker v-model={params.row.colorGroup} on-on-change={e => {
                  this.tableData[params.index].colorGroup = e
                }}  />
              </div>
            )
          }
        },
        {
          title: this.$t('m_label_value'),
          key: 'labelValue',
          minWidth: 300,
          render: (h, params) => {
            this.joinTagValuesToOptions(params.row.tags, params.row.tagOptions, params.index);
            return (
              <div>
                {!isEmpty(params.row.tagOptions) && !isEmpty(params.row.tags) ?
                  (params.row.tags.map((i, selectIndex) => (
                    <div class="tags-show" key={selectIndex + '' + JSON.stringify(i)}>
                      <span>{i.tagName}</span>
                      <Select
                        style="maxWidth: 200px"
                        value={i.tagValue}
                        on-on-change={v => {
                          Vue.set(this.tableData[params.index].tags[selectIndex], 'tagValue', v)
                          this.updateAllColorLine(params.index);
                        }}
                        filterable
                        multiple
                        clearable
                      >
                        {!isEmpty(this.tableData[params.index].tagOptions) && 
                          !isEmpty(this.tableData[params.index].tagOptions[i.tagName]) &&
                            this.tableData[params.index].tagOptions[i.tagName].map((item, index) => (
                            <Option value={item.value} key={item.key + index}>
                              {item.key}
                            </Option>
                          ))}
                      </Select>
                    </div>
                  )) ) : "-" } 
              </div>
            )
          }
        },
        {
          title: this.$t('m_generate_lines'),
          key: 'series',
          minWidth: 350,
          render: (h, params) => {
            return (
              <div>
                {!isEmpty(params.row.series) ?
                  (params.row.series.map((item, selectIndex) => (
                    <div class="generate-lines">
                      {item.new ? <Tag class="new-line-tag" color="error">{this.$t('m_new')}</Tag> : <span/>}
                      <div class="series-name mr-2">{item.seriesName}</div>
                      <ColorPicker v-model={item.color} on-on-change={e => {
                        this.tableData[params.index].series[selectIndex].color = e
                      }}  />
                    </div>
                  )) ) : "-" } 
              </div>
            )
          }
        }
      ],
      pieChartConfigurationColumns: [
        {
            title: this.$t('m_endpoint'),
            minWidth: 220,
            key: 'endpointName',
            render: (h, params) => {
              return (
                <Select
                  value={params.row.endpoint}
                  on-on-change={v => {
                    Vue.set(this.tableData[params.index], 'endpoint', v);
                    this.endpointValue = v;
                    this.searchTypeByEndpointInPie(v, params.index)
                  }}
                  on-on-query-change={e => {
                    this.getEndpointSearch = e; 
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
            }
        },
        {
            title: this.$t('m_type'),
            minWidth: 150,
            key: 'monitorType',
            render: (h, params) => {
              return (
                <Select
                  value={params.row.monitorType}
                  on-on-change={v => {
                    Vue.set(this.tableData[params.index], 'monitorType', v);
                    this.monitorType = v;
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
            }
        },
        {
          title: this.$t('m_metric'),
          key: 'metricGuid',
          minWidth: 250,
          render: (h, params) => {
            return (
              <Select
                value={params.row.metricGuid}
                on-on-change={v => {
                    this.addConfigurationInPie(v)
                    Vue.set(this.tableData[params.index], 'pieDisplayTag', '');
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
          }
        },
        {
          title: this.$t('m_label_value'),
          key: 'labelValue',
          minWidth: 350,
          render: (h, params) => {
            this.joinTagValuesToOptions(params.row.tags, params.row.tagOptions, params.index);
            return (
              <div>
                {!isEmpty(params.row.tagOptions) && !isEmpty(params.row.tags) ?
                  (params.row.tags.map((i, selectIndex) => (
                    <div class="tags-show" key={selectIndex + '' + JSON.stringify(i)}>
                      <span>{i.tagName}</span>
                      <Select
                        value={i.tagValue}
                        on-on-change={v => {
                          Vue.set(this.tableData[params.index].tags[selectIndex], 'tagValue', v)
                        }}
                        filterable
                        multiple
                        clearable
                      >
                        {!isEmpty(this.tableData[params.index].tagOptions) && 
                          !isEmpty(this.tableData[params.index].tagOptions[i.tagName]) &&
                            this.tableData[params.index].tagOptions[i.tagName].map((item, index) => (
                            <Option value={item.value} key={item.key + index}>
                              {item.key}
                            </Option>
                          ))}
                      </Select>
                    </div>
                  )) ) : "-" } 
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
            for(let key in params.row.tagOptions) {
              options.push(key)
            }
            return (
              <div>
                {
                  isEmpty(params.row.tags) ? '-' : 
                  <Select
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
      serviceGroup: "",
      endpointValue: "",
      endpointName: "",
      endpointType: "",
      endpointOptions: [],
      monitorType: "",
      monitorTypeOptions: [],
      metricGuid: '',
      metricOptions: [],
      chartTemplateOptions: [
        {
          label: this.$t('m_customize'),
          key: 'one',
          value: {
            'aggregate': 'none',
            'aggStep': null,
            'chartType': 'line',
            'lineType': 'line'
          }
        },
        {
          label: `${this.$t('volume')}: sum-60s-${this.$t('m_bar_chart')}`,
          key: 'two',
          value: {
            'aggregate': 'sum',
            'aggStep': 60,
            'chartType': 'line',
            'lineType': 'bar'
          }
        },
        {
          label: `${this.$t('t_avg_consumed')}: avg-60s-${this.$t('m_line_chart')}-${this.$t('m_line_chart_s')}`,
          key: 'three',
          value: {
            'aggregate': 'avg',
            'aggStep': 60,
            'chartType': 'line',
            'lineType': 'line'
          }
        },
        {
          label: `${this.$t('t_max_consumed')}: max-60s-${this.$t('m_line_chart')}-${this.$t('m_line_chart_s')}`,
          key: 'four',
          value: {
            'aggregate': 'max',
            'aggStep': 60,
            'chartType': 'line',
            'lineType': 'line'
          }
        }
      ],
      aggStepOptions: [
        {label: '60S', value: 60},
        {label: '300S', value: 300},
        {label: '600S', value: 600},
        {label: '1800S', value: 1800},
        {label: '3600S', value: 3600}
      ],
      aggregateOptions: ['none', 'min', 'max', 'avg', 'p95', 'sum', 'avg_nonzero'],
      lineTypeOptions: [
        {label: this.$t('m_line_chart'), value: 'line'},
        {label: this.$t('m_bar_chart'), value: 'bar'},
        {label: this.$t('m_area_chart'), value: 'area'},
      ],
      pieTypeOptions: [
        {
          label: "标签",
          value: "tag"
        },
        {
          label: "值",
          value: "value"
        }
      ],
      userRolesOptions: [],
      mgmtRolesOptions: [],
      mgmtRoles: [],
      useRoles: [],
      lineTypeOption: {
        line: 1,
        area: 0
      },
      metricTypeMap: {
        common: {
          label: this.$t('m_basic_group'),
          color: '#2d8cf0'
        },
        business: {
          label: this.$t('m_business_configuration'),
          color: '#81b337'
        },
        custom: {
          label: this.$t('m_customize'),
          color: '#b886f8'
        }
      },
      chartPublic: false,
      chartAddTagOptions: {},
      chartAddTags: [],
      getEndpointSearch: ''
    }
  },
  computed: {
    isPieChart() {
      return this.chartConfigForm.chartType === 'pie'
    },
    finalTableColumns() {
      const pieChartConfigurationColumnNoValue = cloneDeep(this.pieChartConfigurationColumns);
      pieChartConfigurationColumnNoValue.pop()
      return this.isPieChart ?
        (this.chartConfigForm.pieType === 'value' ? pieChartConfigurationColumnNoValue : this.pieChartConfigurationColumns) 
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
      handler () {
        this.debounceDrawChart()
      },
      deep: true
    },
    chartConfigForm: {
      handler () {
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
    this.debounceGetEndpointList();
    this.getAllRolesOptions();
    this.getSingleChartAuth();
    this.getTableData();
  },
  methods: {
    getTableData() {
      this.request('GET', '/monitor/api/v2/chart/custom', {
        chart_id: this.chartId
      }, res => {
        // public是true的时候，是引用态， public为false的时候，为非引用态
        this.chartPublic = res.public;
        for(let key in this.chartConfigForm) {
          this.chartConfigForm[key] = res[key]
        }
        if (!this.chartConfigForm.chartTemplate) {
          this.chartConfigForm.chartTemplate = 'one';
        }
        this.tableData = cloneDeep(res.chartSeries);

        if (res.chartType === "pie" && isEmpty(this.tableData)) {
          this.tableData = cloneDeep(initTableData)
        }
        this.processRawTableData(this.tableData);
        this.drawChartContent();
      })
    },
    async processRawTableData(initialData) {
      if (isEmpty(initialData)) return [];
      for(let i=0; i < initialData.length; i++) {
        const item = initialData[i];
        item.tagOptions = await this.findTagsByMetric(item.metricGuid, item.endpoint, item.serviceGroup);
        Vue.set(item, 'tags', this.initTagsFromOptions(item.tagOptions, item.tags));
      }
      if (this.isPieChart && initialData.length === 1 && initialData[0].endpoint) {
        const selectedEndpointItem = find(cloneDeep(this.endpointOptions), {
          option_value: initialData[0].endpoint
        })
        this.endpointValue = initialData[0].endpoint;
        this.request('GET', '/monitor/api/v1/dashboard/recursive/endpoint_type/list', {
          guid: selectedEndpointItem.option_value
        }, res => {
          if (!isEmpty(res)) {
            this.monitorType = initialData[0].monitorType;
            this.monitorTypeOptions = res;
            this.searchMetricByType();
          }
        })
      } 
    },
    getAllRolesOptions () {
      const params = {
        page: 1,
        size: 1000
      }
      this.request('GET','/monitor/api/v1/user/role/list', params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
        this.mgmtRolesOptions = this.processRolesList(res);
      })
    },
    getSingleChartAuth() {
      this.request('GET','/monitor/api/v2/chart/custom/permission', {
        chart_id: this.chartId
      }, res => {
        this.mgmtRoles = res.mgmtRoles;
        this.useRoles = res.useRoles;
      })
    },
    processRolesList(list = []) {
      if (isEmpty(list)) return [];
      const resArr = cloneDeep(list).map(item => {
        return {
          ...item,
          key: item.name,
          label: item.displayName || item.display_name
        }
      })
      return resArr
    },
    // 将tagOptions中不存在的tagValue拼到tagOptions中
    joinTagValuesToOptions(tags, tagOptions, index) {
      if (!isEmpty(tagOptions) && !isEmpty(tags)) {
        tags.forEach(item => {
          const tagValue = item.tagValue
          const tagName = item.tagName;
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
        });  
      }
    },
    initTagsFromOptions(tagOptions, rawTags = []) {
      if (!tagOptions || isEmpty(tagOptions)) return [];
      const tags = [];
      for(let key in tagOptions) {
        const rawTagItem = find(rawTags, {
          tagName: key
        })
        const tagValue = isEmpty(rawTagItem) || isEmpty(rawTagItem.tagValue) ? [] : [...rawTagItem.tagValue]
        tags.push(
          {
            tagName: key,
            tagValue
          }
        )
      }
      return tags
    },

    debounceGetEndpointList: debounce(function() {
      this.getEndpointList()
    }, 300),

    getEndpointList() {
      let search = '.'
      if (this.getEndpointSearch) {
        search = this.getEndpointSearch
      }
      let params = {
        search,
        page: 1,
        size: 10000
      }
      this.request('GET', '/monitor/api/v1/dashboard/search', params, res => {
          this.endpointOptions = res;
        }
      )
    },
    resetSearchItem(index) {
      this.monitorType = '';
      this.metricGuid = '';
      this.metricOptions = [];
      this.monitorTypeOptions = [];
      this.tableData[index].monitorType = '';
      this.tableData[index].metric = '';
      this.tableData[index].metricGuid = '';
    },
    searchTypeByEndpointInPie(value, index) {
      this.resetSearchItem(index);
      const selectedItem = find(cloneDeep(this.endpointOptions), {
        option_value: value
      })
      this.serviceGroup = selectedItem.app_object;
      this.endpointName = selectedItem.option_text;
      this.endpointType = selectedItem.option_type_name;
      this.request('GET', '/monitor/api/v1/dashboard/recursive/endpoint_type/list', {
        guid: selectedItem.option_value
      }, res => {
        this.monitorType = isEmpty(res) ? '' : res[0];
        Vue.set(this.tableData[index], 'monitorType', this.monitorType);
        Vue.set(this.tableData[index], 'metric', '');
        this.metricOptions = [];
        this.monitorTypeOptions = res;
        this.searchMetricByType();
      })

    },
    searchTypeByEndpoint(value) {
      this.monitorType = '';
      this.metricGuid = '';
      this.metricOptions = [];
      this.monitorTypeOptions = [];
      const selectedItem = find(cloneDeep(this.endpointOptions), {
        option_value: value
      })
      if (selectedItem && !isEmpty(selectedItem)) {
        this.endpointName = selectedItem.option_text;
        this.endpointType = selectedItem.option_type_name;
        this.serviceGroup = selectedItem.app_object;
        let params = {
          guid: selectedItem.option_value
        }
        this.request('GET', '/monitor/api/v1/dashboard/recursive/endpoint_type/list', params, res => {
          this.monitorType = isEmpty(res) ? '' : res[0];
          this.monitorTypeOptions = res;
          this.searchMetricByType();
        })
      }
    },
    searchMetricByType() {
      this.metricGuid = '';
      this.metricOptions = [];
      if (!this.endpointValue || !this.monitorType) return
      this.request('GET', '/monitor/api/v2/monitor/metric/list', {
        monitorType: this.monitorType,
        serviceGroup: this.endpointValue
      }, res => {
        this.metricOptions = this.filterHasUsedMetric(res);
      })
    },

    // 删掉已经使用的指标
    filterHasUsedMetric(options) {
      if (this.isPieChart || isEmpty(this.tableData)) return options
      this.tableData.forEach(item => {
        remove(options, single => {
          return item.metric === single.metric && this.endpointValue === item.endpoint
        })
      })
      return options
    },

    processBasicParams(metric, endpoint, serviceGroup, monitorType, tags, chartSeriesGuid = '') {
      return {
        metric,
        endpoint,
        serviceGroup,
        monitorType,
        tags,
        chartSeriesGuid
      }
    },
    async searchTagOptions() {
      this.chartAddTagOptions = await this.findTagsByMetric(this.metricGuid, this.endpointValue, this.serviceGroup);
      this.chartAddTags = this.initTagsFromOptions(this.chartAddTagOptions);
    }, 
    findTagsByMetric(metricId, endpoint, serviceGroup) {
      const api = '/monitor/api/v2/metric/tag/value-list';
      const params = {
        metricId,
        endpoint,
        serviceGroup
      }
      return new Promise(resolve => {
        this.request('POST', api, params, responseData => {
          let result = {}
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
        const item = this.tableData[0];
        const metricItem = find(this.metricOptions, {
            guid: metricGuid
        })
        item.metricGuid = metricItem.guid;
        Vue.set(item, 'metric', metricItem.metric);
        item.metricType = metricItem.metric_type;
        item.tagOptions = await this.findTagsByMetric(metricItem.guid, this.endpointValue, this.serviceGroup)
        item.tags = this.initTagsFromOptions(item.tagOptions, item.tags);
        item.serviceGroup = this.serviceGroup;
        item.endpointName = this.endpointName;
        item.endpointType = this.endpointType;
      }
    },
    getRandomColor() {
      var letters = '0123456789ABCDEF';
      var color = '#';
      for (var i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)];
      }
      return color;
    },
    async addConfiguration() {
      if (this.endpointValue && this.metricGuid) {
        const metricItem = find(this.metricOptions, {
          guid: this.metricGuid
        })
        const basicParams = this.processBasicParams(metricItem.metric, this.endpointValue, this.serviceGroup, this.monitorType, this.chartAddTags, '');
        const series = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/series/config', basicParams);
        this.tableData.push({
          endpoint: this.endpointValue,
          serviceGroup: this.serviceGroup,
          endpointName: this.endpointName,
          endpointType: this.endpointType,
          metricGuid: metricItem.guid,
          metricType: metricItem.metric_type,
          monitorType: this.monitorType,
          colorGroup: this.getRandomColor(),
          pieDisplayTag: "",
          metric: metricItem.metric,
          tags: this.chartAddTags,
          series: series.map(item => {
            item.color = this.getRandomColor();
            return item
          }),
          tagOptions: this.chartAddTagOptions
        }) 
        this.metricGuid = '';
        this.endpointValue = '';
        this.monitorType = '';
        this.chartAddTagOptions = {};
        this.chartAddTags = [];
        this.getEndpointSearch = '';
        this.debounceGetEndpointList();
      }
    },
    onChartTemplateSelected(key) {
      const template = find(this.chartTemplateOptions, {
        key
      })
      Object.assign(this.chartConfigForm, template.value);
    },
    requestReturnPromise(method, api, params) {
      return new Promise(resolve => {
        this.request(method, api, params, res => {
            resolve(res)
          })
      })
    },
    removeTableItem(index) {
      this.$delete(this.tableData, index);
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
            this.$Message.error(isPublic ? (this.$t('m_chart_library') + this.$t('m_name') + this.$t('m_cannot_be_repeated')) : (this.$t('m_graph_name') + this.$t('m_cannot_be_repeated')));
          }
          resolve(res)
        })
      })
    },
    async beforeSaveValid() {
      const validResult = await this.$refs.formData.validate();
      const isDuplicateName = await this.chartDuplicateNameCheck(this.chartId, this.chartConfigForm.name);
      // const hasSetData = this.tableData.length > 0;
      const hasSetData = this.tableData.length > 0 && this.tableData[0].metric && this.tableData[0].monitorType && this.tableData[0].endpoint
      if (!hasSetData) {
        this.$Message.error(this.$t('m_configuration') + this.$t('m_cannot_be_empty'));
      }
      return new Promise(resolve => {
        resolve(validResult && !isDuplicateName && hasSetData)
      })
    },
    async saveChartConfig() {
      if (await this.beforeSaveValid()) {
        await this.submitChartConfig();
        this.$Message.success(this.$t('m_success'));
        this.$parent.$parent.showChartConfig = false;
        this.$parent.$parent.closeChartInfoDrawer();
      }
    },
    async saveChartLibrary() {
      if (await this.beforeSaveValid()) {
        await this.submitChartConfig();
        const isDuplicateRefName = await this.chartDuplicateNameCheck(this.chartId, this.chartConfigForm.name, 1);
        if (isDuplicateRefName) return;
        this.$refs.authDialog.startAuth(this.mgmtRoles, this.useRoles, this.mgmtRolesOptions, this.userRolesOptions);
      }
    },
    async submitChartConfig() {
      this.debounceDrawChart();
      return new Promise(resolve => {
        let chartSeries = cloneDeep(this.tableData);
        chartSeries = chartSeries.map(item => {
          delete item.tagOptions;
          return item
        })
        const params = Object.assign({}, this.chartConfigForm, {
          id: this.chartId,
          chartSeries
        })
        this.request('PUT', '/monitor/api/v2/chart/custom', params, res => {
          resolve()
        })
      })
    },

    onLineTypeChange(lineType) {
      if (lineType === 'bar') {
        this.chartConfigForm.chartType = 'bar';
      } else {
        this.chartConfigForm.chartType = 'line';
      }
      this.resetChartTemplate();
    },
    onAggregateChange() {
      this.resetChartTemplate();
    },
    onAggStepChange() {
      this.resetChartTemplate();
    },
    resetChartTemplate() {
      this.chartConfigForm.chartTemplate = 'one'
    },
    saveChartAuth(mgmtRoles, useRoles) {
      this.mgmtRoles = mgmtRoles;
      this.useRoles = useRoles;
      const path = '/monitor/api/v2/chart/custom/permission';
      this.request('POST', path, {
        chartId: this.chartId,
        mgmtRoles,
        useRoles
      }, () => {
        this.$Message.success(this.$t('m_success'));
        this.chartPublic = true;
      })
    },
    // 将数据拼好，请求数据并画图
    drawChartContent() {
      if (this.isPieChart) {
        const params = this.generateLineParamsData();
        if (!params[0].metric) return;
        this.request('POST', '/monitor/api/v1/dashboard/pie/chart', params,
          res => {
            drawPieChart(this, res)
        })
      } else {
        const params = {
          aggregate: this.chartConfigForm.aggregate || 'none',
          agg_step: this.chartConfigForm.agg_step || 60,
          time_second: -1800,
          lineType: this.lineTypeOption[this.chartConfigForm.lineType],
          start: 0,
          end: 0,
          title: '',
          unit: '',
          data: this.generateLineParamsData()
        }
        if (isEmpty(params.data)) return
        this.request('POST', '/monitor/api/v1/dashboard/chart', params,
          responseData => {
            responseData.yaxis.unit = this.chartConfigForm.unit;
            readyToDraw(this, responseData, 1, {eye: false, lineBarSwitch: true, chartType: this.chartConfigForm.chartType, params: params})
          }
        )
      }
    },

    debounceDrawChart: debounce(function() {
      this.drawChartContent()
    }, 500),

    generateLineParamsData() {
      if (isEmpty(this.tableData)) return [];
      const data = cloneDeep(this.tableData).map(item => {
        item.app_object = item.serviceGroup;
        item.defaultColor = item.colorGroup;

        if (item.series && !isEmpty(item.series)) {
          item.metricToColor = cloneDeep(item.series).map(one => {
            one.metric = one.seriesName;
            delete one.seriesName
            return one
          })
        } else {
          item.metricToColor = []
        }

        delete item.colorGroup;
        delete item.tagOptions
        return item
      })
      return data
    },

    async updateAllColorLine(index) {
      const item = this.tableData[index];
      const basicParams = this.processBasicParams(item.metric, item.endpoint, item.serviceGroup, item.monitorType, item.tags, item.chartSeriesGuid);
      const series = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/series/config', basicParams);
      
      this.tableData[index].series = series.map(item => {
        item.color = this.getRandomColor();
        return item
      })
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

.add-data-configuration > div {
  width: 20%;
}

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
    max-width: 200px;
  }
}

.config-table {
  .ivu-table-cell {
    padding-left: 0px;
    padding-right: 0px
  }
}

.table-config-endpoint {
  display: flex;
  align-items: center;
  .table-endpoint-tag {
    width: fit-content;
    padding: 0 10px;
    display: flex;
    justify-content: center;
    align-items: center;
  }
}

.generate-lines {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  .series-name {
    max-width: 85%;
    overflow: hidden;

  }
  .new-line-tag {
    width: 30px
  }
}

.select-options-change {
  display: flex;
  align-items: center;
}

.tags-show {
  display: flex;
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
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
        justify-content: space-between;
        align-items: center;
        margin-top: 15px;
        padding: 20px;
        width: 100%;
        background-color: #efefef;
        .add-tag-configuration {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
        }
      }
    } 

  }
  .config-footer {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-top: 30px
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


</style>

