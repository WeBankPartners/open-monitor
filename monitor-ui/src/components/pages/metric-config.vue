<template>
  <div class="text-align:center; ">
    <div style="margin-bottom:24px;">
        <Select v-model="metricSelected" multiple style="width:260px" :label-in-value="true" 
            @on-change="selectMetric" placeholder="请选择监控指标">
            <Option v-for="item in metricList" :value="item.prom_ql" :key="item.metric">{{ item.metric }}</Option>
        </Select>
        <Select v-model="timeTnterval" style="width:80px;margin: 0 8px;">
          <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
        
        <button class="btn btn-sm btn-confirm-f" @click="requestChart">查询</button>
        <button class="btn btn-sm btn-cancle-f" @click="addMetric">新增指标</button>
        <button class="btn btn-sm btn-cancle-f" @click="saveConfig">保存</button>

    </div>
    <section class="metric-section">
      <ul>
        <template v-for="(metricItem, metricIndex) in totalMetric">
          <li :key="metricIndex" class="metric-display">
            <Tag color="primary" type="border" closable @on-close="delMetric(metricItem)">指标名称：{{metricItem.label}}</Tag>
            <div>
               <textarea v-model="metricItem.value" class="textareaSty"></textarea> 
            </div>
          </li>
        </template>
      </ul>
    </section>
    <section v-if="isRequestChartData" class="metric-section">
      <div v-if="!noDataTip">
        <div :id="elId" class="echart"></div>
      </div>
      <div v-else class="echart echart-no-data-tip">
        <span>~~~暂无数据~~~</span>
      </div>
    </section>
  </div>
</template>

<script>
import {dataPick} from '@/assets/config/common-config'

import {generateUuid} from '@/assets/js/utils'

// 引入 ECharts 主模块
import {drawChart} from  '@/assets/config/chart-rely'

export default {
  name: '',
  data() {
    return {
     elId: '',
     isRequestChartData: false,
     noDataTip: false,

     metricSelected: [],
     metricSelectedOptions: [],
     metricList: [],

     timeTnterval: -1800,
     dataPick: dataPick,

     editMetric: []
    }
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`
    })
  },
  computed: {
    totalMetric: function () {
      return this.metricSelectedOptions.concat(this.editMetric)
    } 
  },
  mounted (){
    this.obtainMetricList()
  },
  methods: {
    addMetric() {
     this.editMetric.push({label: `default${(new Date()).valueOf()}`, value: ''})
    },
    delMetric (metric) {
      if (metric.label.indexOf('default') > -1) {
       this.editMetric =  this.editMetric.filter((item)=>{
           return item.label !== metric.label
        })
      } else {
        this.metricSelectedOptions = this.metricSelectedOptions.filter((item)=>{
           return item.label !== metric.label
        })
        this.metricSelected = this.metricSelected.filter((item)=>{
           return item !== metric.value
        })
      }
    },
    selectMetric (option) {
      this.metricSelectedOptions = option
    },
    obtainMetricList () {
      let params = {type: this.$store.state.ip.type}
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricList.api, params, responseData => {
        this.metricList = responseData
      })
    },
    requestChart () {
      this.noDataTip = false
      if (this.$validate.isEmpty_reset(this.totalMetric)) {
        this.$Message.warning('请先设置监控指标!')
        this.noDataTip = true
        return
      }
      let params = []
      this.totalMetric.forEach((item) => {
        params.push(JSON.stringify({
          endpoint: this.$store.state.ip.value.split(':')[0],
          prom_ql: item.value,
          time: this.timeTnterval + ''
        })) 
      })
      this.isRequestChartData = true
      
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
    saveConfig () {
      this.$Message.info('尚未开放！')
    }
  },
  components: {
  },
}
</script>

<style scoped lang="less">
  textarea:focus {
    outline: none;
  }
  .metric-display {
    margin: 16px 0;
  }
  .textareaSty{
    display: inline-block;
    vertical-align: top;
    width: 100%;
    border-radius: 4px;
    border-color: #dddee1;
    height: 100px;
    padding: 3px;
  }
  
  .metric-section {
    width: 750px;
    margin: 0 auto;
  }
  .echart {
    height:400px;
    width:750px;
    background: #f5f7f9;
  }
  .echart-no-data-tip {
    text-align: center;
    vertical-align: middle;
    display: table-cell;
  }

  
</style>
