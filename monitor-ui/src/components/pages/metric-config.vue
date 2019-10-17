<template>
  <div class="text-align:center; ">
    <div style="margin-bottom:24px;">
        <Select v-model="metricSelected" filterable multiple style="width:260px" :label-in-value="true" 
            @on-change="selectMetric" placeholder="请选择监控指标">
            <Option v-for="item in metricList" :value="item.id + '^^' + item.prom_ql" :key="item.metric">{{item.metric}}</Option>
        </Select>
        <Select v-model="timeTnterval" style="width:80px;margin: 0 8px;">
          <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
        
        <button class="btn btn-sm btn-confirm-f" @click="requestChart">查询</button>
        <button class="btn btn-sm btn-cancle-f" @click="addMetric">新增指标</button>
        <button class="btn btn-sm btn-cancle-f" @click="saveConfig">保存修改</button>

    </div>
    <section class="metric-section">
      <ul>
        <template v-for="(metricItem, metricIndex) in totalMetric">
          <li :key="metricIndex" class="metric-display">
            <Tag color="primary" type="border" closable @on-close="delMetric(metricItem)">指标名称：{{metricItem.label}} 
              <i class="fa fa-pencil" aria-hidden="true" @click="editMetricName(metricItem,metricIndex)"></i>
            </Tag>
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
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
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

     editMetric: [],
     editingMetric: null,
     modelConfig: {
        modalId: 'edit_metric_Modal',
        modalTitle: '指标名称',
        isAdd: true,
        config: [
          {label: '名称', value: 'name', placeholder: '必填,2-60字符', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null
        }
     }
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
     this.editMetric.push({label: `default${((new Date()).valueOf()).toString().substring(10)}`, value: ''})
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
           return item !== `${metric.id}^^${metric.value}`
        })
      }
    },
    selectMetric (option) {
      this.metricSelectedOptions = []
      option.forEach((item) => {
        this.metricSelectedOptions.push({
          id: parseInt(item.value.split('^^')[0]),
          value: item.value.split('^^')[1],
          label: item.label
        })
      })
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
      let params = []
      this.totalMetric.forEach((item) => {
        if (item.value === '') {
          this.$Message.warning('sfdsghfd！')
          return
        }
        let {id:id,label:metric,value:prom_ql} = item
        params.push({id,metric,prom_ql,metric_type: this.$store.state.ip.type})
      })
      this.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.metricUpdate.api, params, () => {
        this.$Message.success('新增成功 !')
        this.metricSelected = []
        this.editMetric = []
        this.isRequestChartData = false
        this.obtainMetricList()
      })
    },
    editMetricName (metricItem,metricIndex) {
      this.editingMetric = metricIndex
      this.modelConfig.addRow.name = metricItem.label
      this.JQ('#edit_metric_Modal').modal('show')
    },
    addPost (){
      this.totalMetric[this.editingMetric].label = this.modelConfig.addRow.name
      this.JQ('#edit_metric_Modal').modal('hide')
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
