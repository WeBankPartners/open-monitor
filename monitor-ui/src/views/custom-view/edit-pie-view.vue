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
        <div class="condition" style="text-align: left">
          <Tooltip :content="$t('m_field_type')" :delay="1000">
            <Select v-model="templateQuery.pie_metric_type"
              style="width:160px"
               @on-change="switchType"
            >
              <Option value="tag">标签</Option>
              <Option value="value">值</Option>
            </Select>
          </Tooltip>
        </div>
        <button class="btn btn-sm btn-confirm-f" @click="saveConfig">{{$t('m_button_saveConfig')}}</button>
        <!-- <button class="btn btn-sm btn-cancel-f" @click="goback()">{{$t('m_button_back')}}</button> -->
        <button class="btn btn-sm btn-cancel-f" @click="goback()">{{$t('m_button_cancel')}}</button>
      </div>
      <div>
        <section class="zone-config-operation">
          <div class="tag-display"  v-for="(query, queryIndex) in chartQueryList" :key="queryIndex">
            <Tag
              color="primary"
              type="border"
              :name="queryIndex"
              closable
              @click.native="test(query, queryIndex)"
              @on-close="removeQuery(queryIndex)"
            >{{$t('m_field_endpoint')}}：{{query.endpoint}}; {{$t('field.metric')}}：{{query.metric}}</Tag>
          </div>
          <div class="condition-zone">
            <ul>
              <li>
                <div class="condition condition-title c-black-gray">{{$t('m_field_endpoint')}}</div>
                <div class="condition">
                  <Select
                    style="width:300px"
                    v-model="templateQuery.endpoint"
                    filterable
                    clearable
                    remote
                    @on-change="changeEndpoint"
                    :placeholder="$t('requestMoreData')"
                    @on-open-change="getEndpointList('.')"
                    :remote-method="getEndpointList"
                  >
                    <Option
                      v-for="(option, index) in options"
                      :value="option.option_value"
                      :key="index"
                    >
                      <TagShow :list="options" name="option_type_name" :tagName="option.option_type_name" :index="index"></TagShow>{{option.option_text}}</Option>
                    <Option value="moreTips" disabled>{{$t('m_tips_requestMoreData')}}</Option>
                  </Select>
                </div>
              </li>
              <li v-if="showRecursiveType">
                <div class="condition condition-title c-black-gray">{{$t('m_field_type')}}</div>
                  <div class="condition">
                    <Select
                      v-model="templateQuery.app_object_endpoint_type"
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
                    filterable
                    clearable
                    style="width:300px"
                    :label-in-value="true"
                    @on-change="v=>{ setMetric(v)}"
                    @on-open-change="metricSelectOpen(templateQuery.endpoint)"
                  >
                    <Option
                      v-for="(item,index) in metricList"
                      :value="item.metric"
                      :key="item.prom_ql+index"
                    >{{ item.metric }}</Option>
                  </Select>
                </div>
                <button class="btn btn-cancel-f" @click="addQuery">{{$t('m_button_addConfig')}}</button>
              </li>
            </ul>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script>
// import { generateUuid } from "@/assets/js/utils"
import { drawPieChart} from "@/assets/config/chart-rely"
import TagShow from '@/components/Tag-show.vue'
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
    },
    panel_group_list: {
      type: Object,
      default: () => {}
    }
  },
  data() {
    return {
      viewData: [],
      panalIndex: null,
      panalData: null,

      elId: null,
      noDataTip: false,
      endpointType: null,
      chartQueryList: [],
      templateQuery: {
        endpoint: '',
        metricLabel: '',
        metric: '',
        app_object_endpoint_type: '',
        app_object: '',
        pie_metric_type: 'tag'
      },

      options: [],
      metricList: [],

      panalTitle: "Default title",

      oriParams: null,
      params: '', // 保存增加及返回时参数，返回时直接取该值
      editIndex: -1,
      showRecursiveType: false,
      recursiveTypeOptions: []
    }
  },
  created() {

  },
  mounted() {
    this.initChart()
  },
  methods: {
    switchType (val) {
      if (this.chartQueryList.length === 0) {
        return
      }
      this.chartQueryList.forEach(item => {
        item.pie_metric_type = val
      })
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigPieView.api, this.chartQueryList,
        responseData => {
          drawPieChart(this, responseData)
        }
      )
    },
    async test (a, b) {
      this.editIndex = b
      this.templateQuery = {
        ...a
      }
      let params = {
        search: '.',
        page: 1,
        size: 10000
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        this.$root.apiCenter.resourceSearch.api,
        params,
        responseData => {
          this.options = responseData
          const find = this.options.find(item => item.option_value === a.endpoint)
          if (find) {
            this.endpointType = find.type
          }
          if (find && find.id === -1) {
            this.showRecursiveType = true
            let params = {
              guid: find.option_value
            }
            this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.recursiveType, params, responseData => {
              this.templateQuery.app_object_endpoint_type = responseData[0]
              this.recursiveTypeOptions = responseData
            })
          } else {
            this.showRecursiveType = false
          }
          this.metricSelectOpen(a.metric)
        }
      )
    },
    changeEndpoint (val) {
      if (val) {
        this.endpointType = this.options.find(item => item.option_value === val).type
        this.showRecursiveType = false
        this.templateQuery.app_object_endpoint_type = ''
        const find = this.options.find(item => item.option_value === val)
        if (find && find.id === -1) {
          this.showRecursiveType = true
          let params = {
            guid: find.option_value
          }
          this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.recursiveType, params, responseData => {
            this.templateQuery.app_object_endpoint_type = responseData[0]
            this.recursiveTypeOptions = responseData
          }
        )}
      }
    },
    removeQuery(queryIndex) {
      this.chartQueryList.splice(queryIndex, 1)
      this.clearParams()
      this.finalPaint()
    },
    addQuery () {
      if (this.templateQuery.endpoint === '' || this.templateQuery.metric === '' || this.templateQuery.endpoint === undefined || this.templateQuery.metric === undefined) {
        this.$Message.warning(this.$t('m_tip_for_save'))
        return
      }
      let tmp = JSON.parse(JSON.stringify(this.templateQuery))
      if (tmp.endpoint_type !== '') {
        tmp.app_object = tmp.endpoint
      }
      if (this.editIndex !== -1) {
        this.chartQueryList[this.editIndex ] = tmp
      } else {
        this.chartQueryList.push(tmp)
      }
      this.clearParams()
      this.finalPaint()
    },
    finalPaint () {
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigPieView.api, this.chartQueryList,
        responseData => {
          drawPieChart(this, responseData)
      })
    },
    clearParams () {
      const pie_metric_type = this.templateQuery.pie_metric_type
      this.templateQuery = {
        endpoint: '',
        metricLabel: '',
        metric: '',
        app_object_endpoint_type: '',
        app_object: '',
        pie_metric_type: pie_metric_type
      }
    },
    async initChart () {
      let params = {
        templateData: this.parentRouteData,
        panal: this.activeGridConfig
      }
      this.oriParams = params
      this.elId = params.panal.id + '1'
      if (!this.$root.$validate.isEmpty_reset(params.templateData.cfg)) {
        this.getEndpointList()
        this.viewData = JSON.parse(params.templateData.cfg)
        this.viewData.forEach((itemx, index) => {
          if (itemx.viewConfig.id === params.panal.id) {
            this.panalIndex = index
            this.panalData = itemx
            this.initPanal()
            return
          }
        })
      }
    },
    initPanal() {
      this.panalTitle = this.panalData.panalTitle
      this.panalUnit = this.panalData.panalUnit
      
      this.noDataTip = false
      if (this.$root.$validate.isEmpty_reset(this.panalData.query)) {
        return
      }
      let params =  this.panalData.query
      this.chartQueryList = this.panalData.query
      // this.metricSelectOpen(metric)

      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigPieView.api, params,
        responseData => {
          drawPieChart(this, responseData)
      })
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
          this.$t("m_tableKey_s_metric") + this.$t("m_tips_required")
        )
      } else {
        let params = { monitorType: this.showRecursiveType ? this.templateQuery.app_object_endpoint_type : this.endpointType, serviceGroup: this.showRecursiveType ? this.templateQuery.endpoint : '' }
        // let params = { endpointType: this.endpointType }
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'GET',
          this.$root.apiCenter.getMetricByEndpointType,
          params,
          responseData => {
            this.metricList = responseData
          }
        )
      }
    },
    saveConfig() {
      this.pp()
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        this.$root.apiCenter.template.save,
        this.params,
        () => {
          this.$Message.success(this.$t("m_tips_success"))
          this.$parent.$parent.showChartConfig = false
          this.$parent.$parent.reloadPanal(this.params)
        }
      )
    },
    setMetric(value) {
      if (!this.$root.$validate.isEmpty_reset(value)) {
        this.templateQuery.metricLabel = value.label
      }
    },
    pp() {
      let query = []
      this.chartQueryList.forEach(item => {
        item.chartType = 'pie'
        query.push(item)
      })
      let panal = this.oriParams.panal
      panal.i = this.panalTitle
      const temp = {
        panalTitle: this.panalTitle,
        panalUnit: this.panalUnit,
        chartType: 'pie',
        query: query,
        viewConfig: panal
      }
      this.viewData[this.panalIndex] = temp
      let params = {
        name: this.oriParams.templateData.name,
        id: this.oriParams.templateData.id,
        panel_group_list: this.panel_group_list || [],
        cfg: JSON.stringify(this.viewData)
      }
      this.params = params
    },
    goback() {
      // if (!this.params) {
      //   this.pp()
      // }
      // this.$router.push({ name: "viewConfig", params: this.params })
      this.$parent.$parent.showChartConfig = false
    }
  },
  components: {
    TagShow
  }
}
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

