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
      <div class="tool-save">
        <div class="condition">
          <Select filterable v-model="templateQuery.aggregate" @on-change="switchChartType">
            <Option
              v-for="(type) in ['min', 'max', 'avg', 'p95', 'sum', 'none']"
              :value="type"
              :key="type"
            >{{type}}</Option>
          </Select>
        </div>
        <div class="condition">
          <Select filterable clearable v-model="templateQuery.chartType" @on-change="switchChartType">
            <Option
              v-for="(option, index) in chartTypeOption"
              :value="option.value"
              :key="index"
            >{{option.label}}</Option>
          </Select>
        </div>
        <button class="btn btn-sm btn-confirm-f" @click="saveConfig">{{$t('button.saveConfig')}}</button>
        <button class="btn btn-sm btn-cancel-f" @click="goback()">{{$t('button.cancel')}}</button>
      </div>
      <div>
        <section class="zone-config-operation">
            <div class="tag-display">
              <Tag
                v-for="(query, queryIndex) in chartQueryList"
                color="primary"
                type="border"
                :key="queryIndex"
                :name="queryIndex"
                closable
                @on-close="removeQuery(queryIndex)"
              >{{$t('field.endpoint')}}：{{query.endpoint}}; {{$t('field.metric')}}：{{query.metric}}</Tag>
            </div>
            <div class="condition-zone">
              <ul>
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('field.endpoint')}}</div>
                  <div class="condition">
                    <Select
                      style="width:300px"
                      v-model="templateQuery.endpoint"
                      filterable
                      clearable
                      remote
                      :placeholder="$t('requestMoreData')"
                      @on-open-change="getEndpointList('.')"
                      @on-change="selectEndpoint"
                      :remote-method="getEndpointList"
                    >
                      <Option
                        v-for="(option, index) in options"
                        :value="option.option_value"
                        :key="index"
                      >
                        <TagShow :tagName="option.option_type_name" :index="index"></TagShow>{{option.option_text}}</Option>
                      <Option value="moreTips" disabled>{{$t('tips.requestMoreData')}}</Option>
                    </Select>
                  </div>
                </li>
                <li v-if="showRecursiveType">
                  <div class="condition condition-title c-black-gray">{{$t('field.type')}}</div>
                  <div class="condition">
                    <Select
                      v-model="templateQuery.endpoint_type"
                      style="width:300px"
                      filterable
                      clearable
                    >
                      <Option
                        v-for="(item,index) in recursiveTypeOptions"
                        :value="item"
                        :key="item+index"
                      >{{ item }}</Option>
                    </Select>
                  </div>
                </li>
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('field.metric')}}</div>
                  <div class="condition">
                    <Select
                      v-model="templateQuery.metric"
                      style="width:300px"
                      filterable
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
                <li v-if="templateQuery.metricToColor.length >0">
                  <div class="condition condition-title" style="vertical-align: top;">{{$t('个性化配置')}}</div>
                  <div class="condition">
                    <template v-for="mc in templateQuery.metricToColor">
                      <div :key="mc.metric">
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
                <li>
                  <div class="condition condition-title">{{$t('field.unit')}}</div>
                  <div class="condition">
                    <Input v-model="panalUnit" placeholder="" style="width: 300px" />
                  </div>
                  <button class="btn btn-cancel-f" @click="addQuery()">{{$t('button.addConfig')}}</button>
                </li>
              </ul>
            </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script>
import { generateUuid } from "@/assets/js/utils"
import { readyToDraw } from "@/assets/config/chart-rely"
import TagShow from '@/components/Tag-show.vue'
export default {
  name: "",
  data() {
    return {
      viewData: null,
      panalIndex: null,
      panalData: null,

      elId: null,
      noDataTip: false,
      endpointType: null,
      templateQuery: {
        endpoint: '',
        metric: '',
        chartType: '',
        aggregate: '',
        endpoint_type: '',
        app_object: '',
        metricToColor: []
      },
      chartTypeOption: [
        {label: '线性图', value: 'line'},
        {label: '柱状图', value: 'bar'}
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
      params: '' // 保存增加及返回时参数，返回时直接取该值
    }
  },
  watch: {
    chartQueryList: {
      handler(data) {
        this.noDataTip = false
        let params = {
          aggregate: this.templateQuery.aggregate || 'none',
          time_second: -1800,
          start: 0,
          end: 0,
          title: '',
          unit: '',
          data: []
        }
        if (this.$root.$validate.isEmpty_reset(data)) {
          this.noDataTip = true
          return
        }
        data.forEach(item => {
          params.data.push(item)
        })
        console.log(params)
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',this.$root.apiCenter.metricConfigView.api, params,
          responseData => {
            responseData.yaxis.unit = this.panalUnit
            readyToDraw(this,responseData, 1, { eye: false, chartType: this.templateQuery.chartType, clear: true, params: params })
          }
        )
      },
      deep: true
      // immediate: true
    },
    'templateQuery.endpoint': function (val) {
      if (val) {
        this.endpointType = this.options.find(item => item.option_value === val).type  
      }
    }
  },
  created() {
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  mounted() {
  },
  methods: {
    changeMetric (val) {
      this.templateQuery.metricToColor = []
      if (!val) return 
      let tmp = JSON.parse(JSON.stringify(this.templateQuery))
      tmp.aggregate = 'none'
      tmp.chartType = 'line'
      let params = {
        aggregate: 'none',
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: [tmp]
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigView.api, params,
        responseData => {
          this.templateQuery.metricToColor = responseData.legend.map(r => {
            return {
              metric: r,
              color: ''
            }
          })
        }
      )
    },
    selectEndpoint (val) {
      this.showRecursiveType = false
      this.templateQuery.endpoint_type = ''
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
    initChart (params) {
      this.oriParams = params
      this.chartQueryList = []
      if (!this.$root.$validate.isEmpty_reset(params.templateData.cfg)) {
        this.getEndpointList()
        this.viewData = JSON.parse(params.templateData.cfg)
        this.viewData.forEach((itemx, index) => {
          if (itemx.viewConfig.id === params.panal.id) {
            this.templateQuery.chartType = itemx.chartType
            this.templateQuery.aggregate = itemx.aggregate || 'none'
            this.panalIndex = index
            this.panalData = itemx
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
      let params = {
          aggregate: this.templateQuery.aggregate || 'none',
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
          endpoint_type: item.endpoint_type
        })
      })
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigView.api, params,
        responseData => {
          responseData.yaxis.unit = this.panalUnit
          readyToDraw(this,responseData, 1, { eye: false, chartType: this.templateQuery.chartType})
        }
      )
    },
    initPanal() {
      this.panalTitle = this.panalData.panalTitle
      this.panalUnit = this.panalData.panalUnit
      let params = {
        aggregate: this.templateQuery.aggregate || 'none',
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
            readyToDraw(this,responseData, 1, { eye: false, lineBarSwitch: true, chartType: this.templateQuery.chartType})
          }
        )
      }
    },
    initQueryList(query) {
      this.chartQueryList = query
    },
    getEndpointList(query='.') {
      let params = {
        search: query,
        page: 1,
        size: 1000
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
    metricSelectOpen(metric) {
      if (this.$root.$validate.isEmpty_reset(metric)) {
        this.$Message.warning(
          this.$t("tableKey.s_metric") + this.$t("tips.required")
        )
      } else {
        let params = { type: this.showRecursiveType ? this.templateQuery.endpoint_type : this.endpointType }
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'GET',
          this.$root.apiCenter.metricList.api,
          params,
          responseData => {
            this.metricList = responseData
          }
        )
      }
    },
    addQuery() {
      if (this.templateQuery.endpoint === '' || this.templateQuery.metric === '' || this.templateQuery.endpoint === undefined || this.templateQuery.metric === undefined) {
        this.$Message.warning("配置完整方可保存！")
        return
      }
      let tmp = JSON.parse(JSON.stringify(this.templateQuery))
      if (tmp.endpoint_type !== '') {
        tmp.app_object = tmp.endpoint
      }
      this.chartQueryList.push(tmp)
      this.templateQuery = {
        endpoint: '',
        metric: '',
        chartType: tmp.chartType,
        aggregate: tmp.aggregate,
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
        aggregate: this.templateQuery.aggregate,
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
</style>

