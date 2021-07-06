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
        <button class="btn btn-sm btn-confirm-f" @click="saveConfig">{{$t('button.saveConfig')}}</button>
        <!-- <button class="btn btn-sm btn-cancel-f" @click="goback()">{{$t('button.back')}}</button> -->
        <button class="btn btn-sm btn-cancel-f" @click="goback()">{{$t('button.cancel')}}</button>
      </div>
      <div>
        <section class="zone-config-operation">
            <div class="condition-zone">
              <ul>
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('field.endpoint')}}</div>
                  <div class="condition">
                    <Select
                      style="width:300px"
                      v-model="templateQuery.endpoint"
                      filterable
                      remote
                      :placeholder="$t('requestMoreData')"
                      @on-open-change="getEndpointList('.')"
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
                <li>
                  <div class="condition condition-title c-black-gray">{{$t('field.metric')}}</div>
                  <div class="condition">
                    <Select
                      v-model="templateQuery.metric"
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
                </li>
                <li>
                  <div class="condition condition-title">{{$t('field.title')}}</div>
                  <div class="condition">
                    <Input
                      v-model="panalTitle"
                      placeholder=""
                      style="width: 300px"
                    />
                  </div>
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
  data() {
    return {
      viewData: [],
      panalIndex: null,
      panalData: null,

      elId: null,
      noDataTip: false,
      endpointType: null,
      templateQuery: {
        endpoint: "",
        metricLabel: "",
        metric: ""
      },

      options: [],
      metricList: [],

      panalTitle: "Default title",

      oriParams: null,
      params: '' // 保存增加及返回时参数，返回时直接取该值
    }
  },
  watch: {
    templateQuery: {
      handler (val) {
        if (
          val.endpoint === "" ||
          val.metricLabel === ""
        ) {
          return
        }
        const params = {
          endpoint: val.endpoint,
          metric: val.metricLabel, 
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',this.$root.apiCenter.metricConfigPieView.api, params,
          responseData => {
            drawPieChart(this, responseData)
          }
        )
      },
      deep: true
    },
    'templateQuery.endpoint': function (val) {
      this.endpointType = this.options.find(item => item.option_value === val).type
    }
  },
  created() {

  },
  mounted() {
    // if (this.$root.$validate.isEmpty_reset(this.$route.params)) {
    //   this.$router.push({ path: "viewConfig" })
    // } else {
    //   if (!this.$root.$validate.isEmpty_reset(this.$route.params.templateData.cfg)) {
    //     this.elId = this.$route.params.panal.id
    //     this.getEndpointList()
    //     this.viewData = JSON.parse(this.$route.params.templateData.cfg)
    //     this.viewData.forEach((itemx, index) => {
    //       if (itemx.viewConfig.id === this.$route.params.panal.id) {
    //         this.panalIndex = index
    //         this.panalData = itemx
    //         this.initPanal()
    //         return
    //       }
    //     })
    //   }
    // }
  },
  methods: {
    initChart (params) {
      this.oriParams = params
      if (!this.$root.$validate.isEmpty_reset(params.templateData.cfg)) {
        this.elId = params.panal.id
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

      let {endpoint, metric} =  this.panalData.query[0]
      this.templateQuery.endpoint = endpoint
      this.templateQuery.metric = metric
      this.templateQuery.metricLabel = metric
      this.metricSelectOpen(metric)
      let params = {
        endpoint,
        metric
      }
      if (params !== {}) {
        this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigPieView.api, params,
        responseData => {
          drawPieChart(this, responseData)
        }
      )
      }
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
          this.options = []
          responseData.forEach((item) => {
            if (item.id !== -1) {
              this.options.push(item)
            }
          })
        }
      )
    },
    metricSelectOpen(metric) {
      if (this.$root.$validate.isEmpty_reset(metric)) {
        this.$Message.warning(
          this.$t("tableKey.s_metric") + this.$t("tips.required")
        )
      } else {
        let params = { type: this.endpointType }
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
    setMetric(value) {
      if (!this.$root.$validate.isEmpty_reset(value)) {
        this.templateQuery.metricLabel = value.label
      }
    },
    pp() {
      let panal = this.oriParams.panal
      panal.i = this.panalTitle
      const temp = {
        panalTitle: this.panalTitle,
        panalUnit: this.panalUnit,
        chartType: 'pie',
        query: [{
          chartType: 'pie',
          endpoint: this.templateQuery.endpoint,
          metric: this.templateQuery.metricLabel,  
        }],
        viewConfig: panal
      }
      this.viewData[this.panalIndex] = temp
      let params = {
        name: this.oriParams.templateData.name,
        id: this.oriParams.templateData.id,
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

