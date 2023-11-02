<template>
  <div class>
    <div class="zone zone-chart">
      <div class="zone-chart-title">{{panalTitle}}</div>
      <div v-if="!noDataTip">
        <div :id="elId" class="echart"></div>
      </div>
      <div v-else class="echart echart-no-data-tip">
        <span>~~~No Data!~~~</span>
      </div>
    </div>
    <div class="zone zone-config c-dark">
      <div class="flex">
        <div class="left">
          <Select class="select-option" filterable v-model="quickQueryValue" style="width:250px" @on-change="syncCondition">
            <Option
              v-for="agg in quickOptions"
              :value="agg.label"
              :key="agg.label"
            >{{agg.label}}</Option>
          </Select>
        </div>
        <div class="tool-save">
          <div class="condition">
            <Tooltip :content="$t('field.aggType')" :delay="1000">
              <Select filterable  class="select-option" v-model="templateQuery.aggregate" style="width:100px" @on-change="switchChartType">
                <Option
                  v-for="(type) in ['min', 'max', 'avg', 'p95', 'sum', 'none', 'avg_nonzero']"
                  :value="type"
                  :key="type"
                >{{type}}</Option>
              </Select>
            </Tooltip>
          </div>
          <div class="condition" v-if="templateQuery.aggregate !== 'none'">
            <Tooltip :content="$t('field.aggStep')" :delay="1000">
              <Select filterable  class="select-option" v-model="templateQuery.agg_step" style="width:120px" @on-change="switchChartType">
                <Option
                  v-for="agg in aggStepOptions"
                  :value="agg.value"
                  :key="agg.value"
                >{{agg.label}}</Option>
              </Select>
            </Tooltip>
          </div>
          <div class="condition">
            <Tooltip :content="$t('m_chart_type')" :delay="1000">
              <Select filterable class="select-option" v-model="templateQuery.chartType" style="width:120px" @on-change="switchChartType">
                <Option
                  v-for="(option, index) in chartTypeOption"
                  :value="option.value"
                  :key="index"
                >{{option.label}}</Option>
              </Select>
            </Tooltip>
          </div>
          <div class="condition" v-if="templateQuery.chartType === 'line'">
            <Tooltip :content="$t('m_line_type')" :delay="1000">
              <Select filterable class="select-option" v-model="templateQuery.lineType" style="width:120px" @on-change="switchChartType">
                <Option
                  v-for="(option, index) in lineOption"
                  :value="option.value"
                  :key="index"
                >{{option.label}}</Option>
              </Select>
            </Tooltip>
          </div>
          <button class="btn btn-sm btn-confirm-f" @click="saveConfig">{{$t('button.saveConfig')}}</button>
          <button class="btn btn-sm btn-cancel-f" @click="goback()">{{$t('button.cancel')}}</button>
        </div>
      </div>
      <div>
        <section class="zone-config-operation">
            <div class="tag-display"  v-for="(query, queryIndex) in chartQueryList" :key="queryIndex">
              <Tag
                color="primary"
                type="border"
                :name="queryIndex"
                closable
                @click.native="editQueryParams(query, queryIndex)"
                @on-close="removeQuery(queryIndex)"
              >{{$t('field.endpoint')}}：{{query.endpointName || query.endpoint}}; {{$t('field.metric')}}：{{query.metric}}</Tag>
            </div>
            <div v-if="fixSelect" class="condition-zone">
              <ul>
                <!--对象-->
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('field.endpoint')}}</div>
                  <div class="condition">
                    <Select
                      style="width:300px"
                      v-model="templateQuery.endpoint"
                      filterable
                      clearable
                      :placeholder="$t('requestMoreData')"
                      @on-open-change="getEndpointList('.', $event)"
                      @on-change="selectEndpoint"
                      @on-query-change="handleRemoteEndpoint"
                      @on-clear="metricDefaultColor=[]"
                    >
                      <Option
                        v-for="(option, index) in options"
                        :value="option.option_value"
                        :label="option.option_text"
                        :key="index"
                      >
                        <TagShow :tagName="option.option_type_name" :index="index"></TagShow>{{option.option_text}}</Option>
                      <Option value="moreTips" disabled>{{$t('tips.requestMoreData')}}</Option>
                    </Select>
                  </div>
                </li>
                <!--类型-->
                <li v-if="showRecursiveType">
                  <div class="condition condition-title c-black-gray">{{$t('field.type')}}</div>
                  <div class="condition">
                    <Select
                      v-model="templateQuery.endpoint_type"
                      style="width:300px"
                      filterable
                      clearable
                      @on-clear="metricDefaultColor=[]"
                    >
                      <Option
                        v-for="(item,index) in recursiveTypeOptions"
                        :value="item"
                        :key="item+index"
                      >{{ item }}</Option>
                    </Select>
                  </div>
                </li>
                <!--指标-->
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('field.metric')}}</div>
                  <div class="condition">
                    <Select
                      v-model="templateQuery.metric"
                      style="width:300px"
                      filterable
                      multiple
                      clearable
                      :label-in-value="true"
                      @on-change="changeMetric"
                      @on-open-change="metricSelectOpen(templateQuery.endpoint)"
                    >
                      <Option
                        v-for="(item,index) in metricList"
                        :value="item.metric"
                        :key="item.metric+index"
                      >{{ item.metric }}</Option>
                    </Select>
                  </div>
                </li>
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('m_default_color')}}</div>
                  <div class="condition">
                    <template v-for="mdc in metricDefaultColor">
                      <div :key="mdc.metric">
                        <Tooltip :content="mdc.metric" max-width="300">
                          <Tag>{{mdc.metric.length > 50 ? mdc.metric.substring(0,50) + '...' : mdc.metric}}</Tag>
                          <div slot="content" style="white-space: normal;">
                            <p>{{mdc.metric}}</p>
                          </div>
                        </Tooltip>
                        <ColorPicker v-model="mdc.defaultColor" />
                        {{mdc.defaultColor}}
                      </div>
                    </template>
                  </div>
                </li>
                <li v-if="templateQuery.metricToColor.length >0">
                  <div class="condition condition-title" style="vertical-align: top;">{{$t('m_custom_config')}}</div>
                  <div class="condition">
                    <template v-for="(mc, index) in templateQuery.metricToColor">
                      <div :key="index">
                        <Tooltip :content="mc.metric" max-width="300">
                          <Tag>{{mc.metric.length > 50 ? mc.metric.substring(0,50) + '...' : mc.metric}}</Tag>
                          <div slot="content" style="white-space: normal;">
                            <p>{{mc.metric}}</p>
                          </div>
                        </Tooltip>
                        <ColorPicker v-model="mc.color" />
                        {{mc.color}}
                      </div>
                    </template>
                  </div>
                </li>
                <!--单位-->
                <li>
                  <div class="condition condition-title">{{$t('field.unit')}}</div>
                  <div class="condition">
                    <Input v-model="panalUnit" placeholder="" style="width: 300px" />
                  </div>
                  <button class="btn btn-cancel-f" @click="addQuery()">{{$t('button.addConfig')}}</button>
                  <button class="btn btn-cancel-f" @click="clearParams">{{$t('m_clear')}}</button>
                </li>
              </ul>
            </div>
            <!-- <div class="loading-zone" v-else>
              <Spin fix></Spin>
            </div> -->
        </section>
      </div>
    </div>
  </div>
</template>

<script>
import { generateUuid } from "@/assets/js/utils"
import { readyToDraw } from "@/assets/config/chart-rely"
import TagShow from '@/components/Tag-show.vue'
import lodash from 'lodash'
export default {
  name: "",
  props: {
    activeGridConfig: {
      type: Object,
      default: () => {}
    },
    parentRouteData: {
      type: Object,
      default: () => {}
    }
  },
  data() {
    return {
      fixSelect: true, // 修复点击tag标签，数据回显不出来问题,等下拉列表数据全部请求完成，再重新加载下拉组件
      viewData: null,
      panalIndex: null,
      panalData: null,
      elId: null,
      noDataTip: false,
      endpointType: null,
      templateQuery: {
        endpoint: '',
        metric: [],
        chartType: '',
        lineType: 1,
        aggregate: '',
        agg_step: 60,
        endpoint_type: '',
        app_object: '',
        metricToColor: []
      },
      metricDefaultColor: [],
      quickQueryValue: null,
      quickOptions: Object.freeze([
        {
          label: `${this.$t('volume')}: sum-60s-${this.$t('m_bar_chart')}`,
          value: {
            'aggregate': 'sum',
            'agg_step': 60,
            'chartType': 'bar',
          }
        },
        {
          label: `${this.$t('t_avg_consumed')}: avg-60s-${this.$t('m_line_chart')}-${this.$t('m_line_chart_s')}`,
          value: {
            'aggregate': 'avg',
            'agg_step': 60,
            'chartType': 'line',
            'lineType': 1
          }
        },
        {
          label: `${this.$t('t_max_consumed')}: max-60s-${this.$t('m_line_chart')}-${this.$t('m_line_chart_s')}`,
          value: {
            'aggregate': 'max',
            'agg_step': 60,
            'chartType': 'line',
            'lineType': 1
          }
        },
        {
          label: this.$t('other'),
          value: {
            'aggregate': 'min',
            'agg_step': 60,
            'chartType': 'line',
            'lineType': 1
          }
        },
      ]),
      aggStepOptions: [
        {label: '60S', value: 60},
        {label: '300S', value: 300},
        {label: '600S', value: 600},
        {label: '1800S', value: 1800},
        {label: '3600S', value: 3600}
      ],
      chartTypeOption: [
        {label: this.$t('m_line_chart'), value: 'line'},
        {label: this.$t('m_bar_chart'), value: 'bar'}
      ],
      lineOption: [
        {label: this.$t('m_line_chart_s'), value: 1},
        {label: this.$t('m_area_chart'), value: 0}
      ],
      chartQueryList: [
        // {
        //   endpoint: '',
        //   metric: ''
        // }
      ],

      options: [],
      metricList: [],
      showRecursiveType: false,
      recursiveTypeOptions: [],
      panalTitle: "Default title",
      panalUnit: "",

      oriParams: null,
      params: '', // 保存增加及返回时参数，返回时直接取该值,
      editIndex: -1,
      oldMetricToColor: []
    }
  },
  watch: {
    // chartQueryList: {
    //   handler(data) {
    //     console.log(111)
    //     this.noDataTip = false
    //     let params = {
    //       aggregate: this.templateQuery.aggregate || 'none',
    //       agg_step: this.templateQuery.agg_step || 60,
    //       lineType: this.templateQuery.lineType,
    //       time_second: -1800,
    //       start: 0,
    //       end: 0,
    //       title: '',
    //       unit: '',
    //       data: []
    //     }
    //     if (this.$root.$validate.isEmpty_reset(data)) {
    //       this.noDataTip = true
    //       return
    //     }
    //     data.forEach(item => {
    //       params.data.push(item)
    //     })
    //     this.$root.$httpRequestEntrance.httpRequestEntrance(
    //       'POST',this.$root.apiCenter.metricConfigView.api, params,
    //       responseData => {
    //         responseData.yaxis.unit = this.panalUnit
    //         readyToDraw(this,responseData, 1, { eye: false, chartType: this.templateQuery.chartType, clear: true, params: params })
    //       }
    //     )
    //   },
    //   deep: true,
    //   immediate: true
    // },
    'templateQuery.endpoint': async function (val) {
      if (val && this.options.length > 0) {
        const find = this.options.find(item => item.option_value === val)
        if (find) {
            this.endpointType = find.type
          }
      }
    },
    // // 解决emplateQuery.endpoint值改变，但是接口查询this.options还未获取到问题
    // options: {
    //   handler(val) {
    //     if (val && val.length > 0) {
    //       const find = this.options.find(item => item.option_value === this.templateQuery.endpoint)
    //       if (find) {
    //         this.endpointType = find.type
    //       }
    //     }
    //   },
    //   deep: true
    // },
    templateQuery: {
      handler(val) {
        if (val.metric && typeof(val.metric) === 'string') {
          this.templateQuery.metric = [val.metric]
        }
      }
    },
    deep: true
  },
  created() {
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  mounted() {
    this.initChart()
  },
  methods: {
    requestAgain () {
      this.noDataTip = false
        let params = {
          aggregate: this.templateQuery.aggregate || 'none',
          agg_step: this.templateQuery.agg_step || 60,
          lineType: this.templateQuery.lineType,
          time_second: -1800,
          start: 0,
          end: 0,
          title: '',
          unit: '',
          data: []
        }
        if (this.$root.$validate.isEmpty_reset(this.chartQueryList)) {
          this.noDataTip = true
          return
        }
        if (this.chartQueryList.length === 0) {
          this.noDataTip = true
          return
        }
        this.chartQueryList.forEach(item => {
          params.data.push(item)
        })
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',this.$root.apiCenter.metricConfigView.api, params,
          responseData => {
            responseData.yaxis.unit = this.panalUnit
            readyToDraw(this,responseData, 1, { eye: false, chartType: this.templateQuery.chartType, clear: true, params: params })
          }
        )
    },
    async editQueryParams (queryParams, queryIndex) {
      await this.bb(queryParams, queryIndex)
    },
    async bb (queryParams, queryIndex) {
      this.metricDefaultColor = []
      this.metricDefaultColor.push({
        metric: queryParams.metric,
        defaultColor: queryParams.defaultColor || '',
      })
      this.fixSelect = false
      const search = queryParams.endpointName || '.'
      this.editIndex = queryIndex
      this.templateQuery = {
        ...queryParams
      }
      let params = {
        search: search,
        page: 1,
        size: 10000
      }
      this.oldMetricToColor = JSON.parse(JSON.stringify(this.templateQuery.metricToColor || []))
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        this.$root.apiCenter.resourceSearch.api,
        params,
        responseData => {
          this.options = responseData
          const find = this.options.find(item => item.option_value === queryParams.endpoint)
          if (find) {
            this.endpointType = find.type
          }
          // id为-1展示类型
          if (find && find.id === -1) {
            this.showRecursiveType = true
            let params = {
              guid: find.option_value
            }
            this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.recursiveType, params, responseData => {
              this.templateQuery.endpoint_type = responseData[0]
              this.recursiveTypeOptions = responseData
              this.metricSelectOpen(queryParams.metric)
            })
          } else {
            this.showRecursiveType = false
            this.metricSelectOpen(queryParams.metric)
          }
        }
      )
    },
    changeMetric (val) {
      this.templateQuery.metricToColor = []
      if (!val || val.length === 0) return

      val.forEach(v => {
        const findIndex = this.metricDefaultColor.findIndex(m=> m.metric === v.value)
        if (findIndex === -1) {
          this.metricDefaultColor.push({
            metric: v.value,
            defaultColor: ''
          })
        }
      })

      let tmp = JSON.parse(JSON.stringify(this.templateQuery))
      if (tmp.endpoint_type !== '') {
        tmp.app_object = tmp.endpoint
      }
      tmp.aggregate = 'none'
      tmp.agg_step = 60
      tmp.chartType = 'line'
      const tmpQuery = JSON.parse(JSON.stringify(tmp))
      let params = {
        aggregate: 'none',
        agg_step: 60,
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: tmpQuery.metric.map(m => {
          return {
            ...tmpQuery,
            metric: m
          }
        })
      }

      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigView.api, params,
        responseData => {
          this.templateQuery.metricToColor = responseData.legend.map(r => {
            const findColor = this.oldMetricToColor.find(o => o.metric === r)
            let color = ''
            if (findColor) {
              color = findColor.color
            }
            return {
              metric: r,
              color: color
            }
          })
        }
      )
    },
    // 选择对象
    selectEndpoint (val) {
      this.showRecursiveType = false
      this.templateQuery.endpoint_type = ''
      this.templateQuery.metricToColor = []
      this.templateQuery.metric = []
      const find = this.options.find(item => item.option_value === val)
      if (find && find.id === -1) {
        this.showRecursiveType = true
        let params = {
          guid: find.option_value
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.recursiveType, params, responseData => {
          this.templateQuery.endpoint_type = responseData[0]
          this.recursiveTypeOptions = responseData
        }
      )}
    },
    clearParams () {
      this.templateQuery = {
        endpoint: '',
        metric: '',
        chartType: '',
        lineType: 1,
        aggregate: '',
        agg_step: 60,
        endpoint_type: '',
        app_object: '',
        metricToColor: []
      }
    },
    initChart () {
      let params = {
        templateData: this.parentRouteData,
        panal: this.activeGridConfig
      }
      this.oriParams = params
      this.chartQueryList = []
      this.clearParams()
      if (!this.$root.$validate.isEmpty_reset(params.templateData.cfg)) {
        this.panalTitle = params.panal.i
        this.getEndpointList('.')
        this.viewData = JSON.parse(params.templateData.cfg)
        this.viewData.forEach((itemx, index) => {
          if (itemx.viewConfig.id === params.panal.id) {
            this.templateQuery.chartType = itemx.chartType
            this.templateQuery.lineType = itemx.lineType || 0
            this.templateQuery.aggregate = itemx.aggregate || 'sum'
            this.templateQuery.agg_step = itemx.agg_step || 60
            this.panalIndex = index
            this.panalData = itemx
            this.rsyncQuick()
            this.initPanal()
            return
          }
        })
      }
    },
    switchChartType () {
      if (this.chartQueryList.length === 0) {
        return
      }
      this.updateQuick()
      let params = {
          aggregate: this.templateQuery.aggregate || 'none',
          agg_step: this.templateQuery.agg_step || 60,
          lineType: this.templateQuery.lineType,
          time_second: -1800,
          start: 0,
          end: 0,
          title: '',
          unit: '',
          data: []
        }
      this.chartQueryList.forEach(item => {
        params.data.push({
          endpoint: item.endpoint,
          metric: item.metric,
          app_object: item.app_object,
          endpoint_type: item.endpoint_type,
          metricToColor: item.metricToColor
        })
      })
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigView.api, params,
        responseData => {
          responseData.yaxis.unit = this.panalUnit
          readyToDraw(this,responseData, 1, { eye: false, chartType: this.templateQuery.chartType, params: params})
        }
      )
    },
    updateQuick () {
      this.rsyncQuick()
    },
    rsyncQuick () {
      let passed = false
      for (let i = 0; i < this.quickOptions.length; i++) {
        const item = this.quickOptions[i]
        const keys = Object.keys(item.value)
        passed = keys.every(p => item.value[p] === this.templateQuery[p])

        if (passed) {
          this.quickQueryValue = item.label
          break
        }
      }
      !passed && (this.quickQueryValue = this.$t('other'))
    },
    syncCondition (val) {
      const selected = this.quickOptions.find(item => item.label === val)
      if (selected) {
        const { aggregate, agg_step, chartType, lineType } = selected.value

        this.templateQuery.aggregate = aggregate || ''
        this.templateQuery.agg_step = agg_step || 60
        this.templateQuery.chartType = chartType || ''
        this.templateQuery.lineType = lineType || 1

        this.switchChartType()
      }
    },
    initPanal() {
      this.panalUnit = this.panalData.panalUnit
      let params = {
        aggregate: this.templateQuery.aggregate || 'none',
        agg_step: this.templateQuery.agg_step || 60,
        lineType: this.templateQuery.lineType,
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: []
      }
      this.noDataTip = false
      if (this.$root.$validate.isEmpty_reset(this.panalData.query)) {
        return
      }
      this.initQueryList(this.panalData.query)
      this.panalData.query.forEach(item => {
        params.data.push(
          item
        )
      })
      if (params !== []) {
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',this.$root.apiCenter.metricConfigView.api, params,
          responseData => {
            responseData.yaxis.unit = this.panalUnit
            readyToDraw(this,responseData, 1, { eye: false, lineBarSwitch: true, chartType: this.templateQuery.chartType, params: params })
          }
        )
      }
    },
    initQueryList(query) {
      this.chartQueryList = query
    },
    getEndpointList(query, flag) {
      if (flag === false) return // 下拉框收缩不调用
      if (this.templateQuery.endpoint) return
      let params = {
        search: query,
        page: 1,
        size: 10000
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        this.$root.apiCenter.resourceSearch.api,
        params,
        responseData => {
          this.options = responseData
        }
      )
    },
    handleRemoteEndpoint: lodash.debounce(function(val) {
      this.getEndpointList(val || '.')
    }, 500),
    // 指标下拉查询
    metricSelectOpen(metric) {
      if (this.$root.$validate.isEmpty_reset(metric)) {
        this.$Message.warning(
          this.$t("tableKey.s_metric") + this.$t("tips.required")
        )
      } else {
        let params = { monitorType: this.showRecursiveType ? this.templateQuery.endpoint_type : this.endpointType, serviceGroup: this.showRecursiveType ? this.templateQuery.endpoint : '' }
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'GET',
          this.$root.apiCenter.metricList.api,
          params,
          responseData => {
            this.metricList = responseData
            this.fixSelect = true
          }
        )
      }
    },
    addQuery() {
      if (this.templateQuery.endpoint === '' ||
        this.templateQuery.metric === '' ||
        this.templateQuery.endpoint === undefined ||
        this.templateQuery.metric === undefined ||
        (Array.isArray(this.templateQuery.metric) && this.templateQuery.metric.length === 0))
      {
        this.$Message.warning(this.$t('m_tip_for_save'))
        return
      }
      let tmp = JSON.parse(JSON.stringify(this.templateQuery))
      if (tmp.endpoint_type !== '') {
        tmp.app_object = tmp.endpoint
      }
      const find = this.options.find(item => item.option_value === tmp.endpoint)
      tmp.endpointName = ''
      if (find) {
        tmp.endpointName = find.option_text
      }

      const tmpQuery = JSON.parse(JSON.stringify(tmp))
      let params = tmpQuery.metric.map(m => {
        return {
          ...tmpQuery,
          metric: m,
          metricToColor: tmpQuery.metricToColor.filter(x => x.metric.startsWith(m)),
          defaultColor: this.metricDefaultColor.find(x => x.metric.startsWith(m)).defaultColor || ''
        }
      })
      // if (this.editIndex !== -1) {
      //   this.chartQueryList[this.editIndex ] = tmp
      // } else {
      //   this.chartQueryList.push(tmp)
      // }
      if (this.editIndex !== -1) {
        // this.removeQuery(this.editIndex)
        this.chartQueryList.splice(this.editIndex, 1)
      } else {
        let illegalIndex = []
        params.forEach((p, pIndex) => {
          const findIndex = this.chartQueryList.findIndex(query => query.endpointName === p.endpointName && query.metric === p.metric)
          if (findIndex !== -1) {
            illegalIndex.push(pIndex)
          }
        })
        if (illegalIndex.length > 0) {
          this.$Message.warning(
            this.$t("m_same_exists")
          )
          return
        }
      }

      this.chartQueryList = this.chartQueryList.concat(params)
      this.requestAgain()
      this.editIndex = -1
      this.metricDefaultColor = []
      this.templateQuery = {
        endpoint: '',
        metric: '',
        chartType: tmp.chartType,
        lineType: tmp.lineType,
        aggregate: tmp.aggregate,
        agg_step: tmp.agg_step,
        endpoint_type: '',
        app_object: '',
        metricToColor: []
      }
      this.options = []
      this.metricList = []
      this.$parent.showChartConfig = false
    },
    removeQuery(queryIndex) {
      this.chartQueryList.splice(queryIndex, 1)
      // this.clearParams()
      this.requestAgain()
    },
    saveConfig() {
      this.pp()
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        this.$root.apiCenter.template.save,
        this.params,
        () => {
          this.$Message.success(this.$t("tips.success"))
          this.$parent.$parent.showChartConfig = false
          this.$parent.$parent.reloadPanal(this.params)
        }
      )
    },
    pp() {
      let query = []

      this.chartQueryList.forEach(item => {
        query.push(item)
      })
      let panal = this.oriParams.panal
      panal.i = this.panalTitle
      const temp = {
        panalTitle: this.panalTitle,
        panalUnit: this.panalUnit,
        chartType: this.templateQuery.chartType,
        lineType: this.templateQuery.lineType,
        aggregate: this.templateQuery.aggregate,
        agg_step: this.templateQuery.agg_step,
        query: query,
        viewConfig: panal
      }

      if (this.panalIndex !== null) {
        this.viewData[this.panalIndex] = temp
      } else {
        this.viewData.push(temp)
      }
      let params = {
        name: this.oriParams.templateData.name,
        id: this.oriParams.templateData.id,
        cfg: JSON.stringify(this.viewData)
      }
      this.params = params
    },
    goback() {
      this.$parent.$parent.showChartConfig = false
    }
  },
  components: {
    TagShow
  }
};
</script>

<style scoped lang="less">
li {
  list-style: none;
}
.zone {
  width: 1100px;
  margin: 0 auto;
  background: @gray-f;
  border-radius: 4px;
}
.zone-chart {
  margin-top: 16px;
  margin-bottom: 16px;
}
.zone-chart-title {
  padding: 20px 40%;
  font-size: 14px;
}
.zone-config {
  padding: 8px;
}
.echart {
  height: 300px;
  width: 1100px;
}

.zone-config-operation {
  margin: 24px;
  margin-top: 0;
}
.fa-plus-square-o {
  padding-left: 4px;
}
.zone-config-operation-general {
  margin-top: 24px;
}

.echart-no-data-tip {
  text-align: center;
  vertical-align: middle;
  display: table-cell;
}
.tool-save {
  text-align: right;
  padding: 4px 64px;
}

.tag-display {
  margin: 4px;
}
.tag-display /deep/ .ivu-tag-primary {
  display: table;
}
.zone-config {
  .flex {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-left: 24px;
    
    .left {
      text-align: left;
    }
  }
}
</style>

<style scoped lang="less">
.condition {
  margin: 2px;
  display: inline-block;
}
.condition-title {
  background: @gray-d;
  width: 110px;
  text-align: center;
  vertical-align: middle;
  margin-right: 8px;
  padding: 6px;
}
.condition-zone {
  border: 1px solid @blue-2;
  padding: 4px;
  margin: 4px;
}
.loading-zone {
  display: inline-block;
  width: 100%;
  position: relative;
  height: 200px;
  margin: 4px;
  border: 1px solid @blue-2;
}
.select-option /deep/ .ivu-select-dropdown-list {
  text-align: left;
}
</style>

