<template>
  <div class=" ">
      <!-- <header>
        <div style="display:flex;justify-content:space-between">
            <div class="header-name">
              <i class="fa fa-th-large fa-2x" aria-hidden="true"></i>
              <span> 12123123</span>
            </div>
        </div>
      </header> -->
      <div class="zone zone-chart" >
        <div class="zone-chart-title">{{panalTitle}}</div>
        <div v-if="!noDataTip">
          <div :id="elId" class="echart"></div>
        </div>
        <div v-else class="echart echart-no-data-tip">
          <span>~~~暂无数据~~~</span>
        </div>
      </div>
      <div class="zone zone-config" >
        <div class="tool-save" > 
          <button class="btn btn-sm btn-confirm-f" @click="saveConfig">{{$t('button.save')}}</button>
          <button class="btn btn-sm btn-cancle-f" @click="goback()">{{$t('button.back')}}</button>
          <!-- <i class="fa fa-floppy-o fa-16"  aria-hidden="true"></i> -->
        </div>
        <div style="display:flex">
          <section>
            <ul>
              <li>
                <Tooltip :content="$t('placeholder.metricConfiguration')" placement="right">
                  <div class="step-icon" @click="activeStep='chat_query'">
                    <i class="fa fa-line-chart" aria-hidden="true"></i>
                  </div>
                </Tooltip>
              </li>
              <li>
                <div class="step-link"></div>
              </li>
              <li>
                <Tooltip :content="$t('placeholder.globalConfiguration')" placement="right">
                  <div class="step-icon" @click="activeStep='chat_general'">
                    <i class="fa fa-cog" aria-hidden="true"></i>
                  </div>
                </Tooltip>
              </li>
            </ul>
          </section>
          <section class="zone-config-operation">
            <div v-if="activeStep==='chat_query'">
              <div class="tag-display">
                <Tag v-for="(query, queryIndex) in chartQueryList"  color="primary" 
               type="border" :key="queryIndex" :name="queryIndex" closable
               @on-close="removeQuery(query)">{{$t('field.endpoint')}}：{{query.endpoint}};  {{$t('field.metric')}}：{{query.metricLabel}}</Tag>
              </div>
              <div class="condition-zone">
                <ul>
                    <li>
                    <div class="condition condition-del" @click="addQuery">
                      <i class="fa fa-floppy-o fa-14" aria-hidden="true"></i>
                    </div>
                  </li>
                  <li>
                    <div class="condition condition-title">{{$t('field.endpoint')}}</div>
                    <div class="condition">
                      <Select
                        style="width:300px"
                        v-model="templateQuery.endpoint"
                        filterable
                        remote
                        :remote-method="endpointList"
                        >
                        <Option v-for="(option, index) in options" :value="option.option_value" :key="index">{{option.option_text}}</Option>
                      </Select>
                    </div>
                  </li>
                  <li>
                    <div class="condition condition-title">{{$t('field.metric')}}</div>
                    <div class="condition">
                      <Select v-model="templateQuery.metric" style="width:300px" :label-in-value="true" @on-change="v=>{ setMetric(v)}" @on-open-change="metricSelectOpen(templateQuery.endpoint)">
                        <Option v-for="(item,index) in metricList" :value="item.prom_ql" :key="item.prom_ql+index">{{ item.metric }}</Option>
                      </Select>
                    </div>
                  </li>
                </ul>
              </div>
            </div>
            <div v-if="activeStep==='chat_general'" class="zone-config-operation-general">
              <ul>
                <li>
                  <div class="condition condition-title">{{$t('field.title')}}</div>
                  <div class="condition">
                    <Input v-model="panalTitle" placeholder="Enter something..." style="width: 300px" />
                  </div>
                </li>
                <li>
                  <div class="condition condition-title">{{$t('field.unit')}}</div>
                  <div class="condition">
                    <Input v-model="panalUnit" placeholder="Enter something..." style="width: 300px" />
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
import {generateUuid} from '@/assets/js/utils'
import {drawChart} from  '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      viewData: null,
      panalIndex: null,
      panalData: null,

      elId: null,
      noDataTip: false,
      activeStep: 'chat_query',
      templateQuery: {
        endpoint: '',
        metricLabel: '',
        metric: ''
      },
      chartQueryList:[
        // {
        //   endpoint: '',
        //   metricLabel: '',
        //   metric: ''
        // }
      ],
      
      options: [],
      metricList: [],

      panalTitle: 'Default title',
      panalUnit: ''
    }
  },
  watch: {
    chartQueryList: {
      handler (data) {
        this.noDataTip = false
        let params = []
        if (this.$validate.isEmpty_reset(data)) {
          this.noDataTip = true
          return
        }
        data.forEach((item) => {
          params.push(JSON.stringify({
            endpoint: item.endpoint.split(':')[0],
            prom_ql: item.metric,
            metric: item.metricLabel,
            time: '-1800'
          })) 
        })
        this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricConfigView.api, {config: `[${params.join(',')}]`}, responseData => {
          var legend = []
          if (responseData.series.length === 0) {
            this.noDataTip = true
            return
          }
          responseData.series.forEach((item)=>{
            legend.push(item.name)
            item.symbol = 'none'
            item.smooth = true
            item.lineStyle = {
              width: 1
            }
          }) 
          let config = {
            title: responseData.title,
            legend: legend,
            series: responseData.series,
            yaxis: responseData.yaxis,
          }
          drawChart(this, config, {eye: false})
        })
      },
      deep: true
      // immediate: true
    },
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
    })
  },
  mounted() {
    if (this.$validate.isEmpty_reset(this.$route.params)) {
      this.$router.push({path: 'viewConfig'})
    } else {
      if (!this.$validate.isEmpty_reset(this.$route.params.templateData.cfg)) {
        this.viewData = JSON.parse(this.$route.params.templateData.cfg)
        this.viewData.forEach((itemx,index)=>{
          if (itemx.viewConfig.id === this.$route.params.panal.id) {
            this.panalIndex = index
            this.panalData = itemx
            this.initPanal()
            return
          }
        })
      }
    }
  },
  methods: {
    initPanal() {
      this.panalTitle = this.panalData.panalTitle
      this.panalUnit = this.panalData.panalUnit
      let params = []
      this.noDataTip = false
      if (this.$validate.isEmpty_reset(this.panalData.query)) {
        return
      }
      this.initQueryList(this.panalData.query)
      this.panalData.query.forEach((item) => {
        params.push(JSON.stringify({
          endpoint: item.endpoint,
          prom_ql: item.metric,
          metric: item.metricLabel,
          time: '-1800'
        })) 
      })
      if (params !== []) {
        this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricConfigView.api, {config: `[${params.join(',')}]`}, responseData => {
          var legend = []
          if (responseData.series.length === 0) {
            this.noDataTip = true
            return
          }
          responseData.series.forEach((item)=>{
            legend.push(item.name)
            item.symbol = 'none'
            item.smooth = true
            item.lineStyle = {
              width: 1
            }
          }) 
          let config = {
            title: responseData.title,
            legend: legend,
            series: responseData.series,
            yaxis: responseData.yaxis,
          }
          drawChart(this, config, {eye: false})
        })
      }
    },
    initQueryList(query) {
      this.chartQueryList = query
    },
    endpointList(query) {
      let params = {
        search: query,
        page: 1,
        size: 1000
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET', this.apiCenter.resourceSearch.api, params, (responseData) => {
       this.options = responseData
      })
    },
    metricSelectOpen(metric) {
      if (this.$validate.isEmpty_reset(metric)) {
        this.$Message.warning(this.$t('tableKey.s_metric')+this.$t('tips.required'))
      } else {
        let params = {type: metric.split(':')[1]}
        this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricList.api, params, responseData => {
          this.metricList = responseData
        })
      } 
    },
    addQuery () {
      if (this.templateQuery.endpoint === '' || this.templateQuery.metricLabel === '' ) {
        this.$Message.warning('配置完整方可保存！')
        return
      }
      this.chartQueryList.push(this.templateQuery)
      this.templateQuery = { 
        endpoint: '',
        metricLabel: '',
        metric: ''
      }
      this.options = []
      this.metricList = []
    },
    removeQuery (query) {
      this.chartQueryList.splice(query,1)
    },
    saveConfig () {
      let query = []
      this.chartQueryList.forEach((item) => {
        query.push({
          endpoint: item.endpoint.split(':')[0],
          metricLabel: item.metricLabel,
          metric: item.metric
        }) 
      })
      let panal = this.$route.params.panal
      panal.i = this.panalTitle
      const temp = {
        panalTitle: this.panalTitle,
        panalUnit: this.panalUnit,
        query: query,
        viewConfig: panal
      }

      if (this.panalIndex !== null) {
        this.viewData[this.panalIndex] = temp
      } else {
        this.viewData.push(temp)
      }
      let params = {
        name: this.$route.params.templateData.name,
        id: this.$route.params.templateData.id,
        cfg: JSON.stringify(this.viewData)
      }
      this.$httpRequestEntrance.httpRequestEntrance('POST','dashboard/custom/save', params, () => {
        this.$Message.success(this.$t('tips.success'))
      })
    },
    setMetric(value) {
      if (!this.$validate.isEmpty_reset(value)) {
        this.templateQuery.metricLabel = value.label
      }
    },
    goback () {
      this.$router.push({name:'viewConfig',params:this.$route.params.parentData})
    },
  },
  components: {},
}
</script>

<style scoped lang="less">
.zone {
    width:1100px;
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
  position: absolute;
  font-size: 14px;
}
.zone-config {
  padding: 8px;
}
.echart {
    height:300px;
    width:1100px;
}

.step-icon {
    i {
        height: 24px;
        width: 24px;
        font-size: 18px;
        color: @blue-lingt;
    }
    .fa-line-chart {
        margin: 7px 6px;
    }
    .fa-cog {
        margin: 8px;
    }
    width: 36px;
    height: 36px;
    border: 2px solid @blue-lingt;
    border-radius: 18px;
    cursor: pointer;
}
.step-link {
    height:64px;
    border-left:2px solid @blue-lingt;
    margin-left:16px;
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
    width: 900px;
    border: 1px solid @blue-2;
    padding: 4px;
    margin: 4px;
  }
  .condition-del {
    float: right;
  }
</style>

