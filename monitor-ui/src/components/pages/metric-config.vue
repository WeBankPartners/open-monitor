<template>
  <div class="text-align:center; ">
    <Title title="视图配置"></Title>
    <div style="margin-bottom:24px;">
      <Notice :noticeConfig='noticeConfig'> </Notice>
      <Searchinput :parentConfig="searchInputConfig" ref="choicedIP"></Searchinput> 
      <Select v-model="metricSelected" filterable multiple style="width:260px" :label-in-value="true" 
          @on-change="selectMetric" @on-open-change="metricSelectOpen" placeholder="请选择监控指标">
          <Option v-for="item in metricList" :value="item.id + '^^' + item.prom_ql" :key="item.metric">{{item.metric}}</Option>
      </Select>
      <Select v-model="timeTnterval" style="width:80px;margin: 0 8px;">
        <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
      </Select>
      
      <button class="btn btn-sm btn-confirm-f" :disabled="$store.state.ip.value === ''" @click="requestChart">查询</button>
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
            <Select v-model="originalMetricList[metricItem.key].model" style="width:300px" size="small" @on-change="selectOriginalMetric(metricItem)">
              <Option v-for="item in originalMetricList[metricItem.key].list" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
            </Select>
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
import Notice from '@/components/components/notice'
import Searchinput from '../components/Search-input'
import {dataPick} from '@/assets/config/common-config'

import {generateUuid} from '@/assets/js/utils'

// 引入 ECharts 主模块
import {drawChart} from  '@/assets/config/chart-rely'

export default {
  name: '',
  data() {
    return {
      noticeConfig: {
        type: 'info',
        contents: [
          {
            tip: '1、请先搜索主机作为输入源并选择监控指标；'
          },
          {
            tip: '2、点击查询可查看单钱对象及指标下监控视图；'
          },
          {
            tip: '3、使用新增指标增减指标项，并在点击 保存修改 后面将配置保存；'
          },
        ]
      },
      searchInputConfig: {
        poptipWidth: 300,
        placeholder: '请输入主机名或IP地址，可模糊匹配',
        inputStyle: "width:300px;",
        // api: '/dashboard/search'
        api: this.apiCenter.resourceSearch.api
      },
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
      },
      
      originalList: [],
      originalMetricList: {}
    }
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`
    })
  },
  watch: {
    changeIP () {
      this.metricSelected = []
      this.metricSelectedOptions = []
      this.metricList = []
      this.noDataTip = true
    }
  },
  computed: {
    changeIP() {
      console.log(12)
      return this.$store.state.ip.value
    },
    totalMetric: function () {
      return this.metricSelectedOptions.concat(this.editMetric)
    } 
  },
  mounted (){
    this.getOriginalMetricList()
  },
  methods: {
    addMetric() {
     let key = ((new Date()).valueOf()).toString().substring(10)
     this.editMetric.push({label: `default${key}`, value: '',key: `add_${key}`})
      let o_metric = {
        list: '',
        model: ''
      }
      o_metric.list = this.originalList
      this.originalMetricList[`add_${key}`] = o_metric
    },
    selectOriginalMetric(item) {
      if (item.key.indexOf('add_')> -1) {
        for (let i=0; i< this.editMetric.length; i++) {
          if (item.key === this.editMetric[i].key) 
          this.editMetric[i].value = this.editMetric[i].value + this.originalMetricList[item.key].model
        }
      } else {
        for (let i=0; i< this.metricSelectedOptions.length; i++) {
          if (item.key === this.metricSelectedOptions[i].key) 
          this.metricSelectedOptions[i].value = this.metricSelectedOptions[i].value + this.originalMetricList[item.key].model
        }
      }
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
          label: item.label,
          key: 'origin_' + parseInt(item.value.split('^^')[0])
        })

        let o_metric = {
          list: this.originalList,
          model: ''
        }
        this.originalMetricList['origin_' + parseInt(item.value.split('^^')[0])] = o_metric
      })
    },
    metricSelectOpen (flag) {
      if (flag) {
        if (this.$store.state.ip.value !== '') {
          this.obtainMetricList(this.$store.state.ip.type)
        } else {
          this.metricSelected = []
          this.metricSelectedOptions = []
          this.metricList = []
          this.$Message.warning('请先选择主机名或IP地址！')
        }
      }
    },
    obtainMetricList (type) {
      let params = {type: type}
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
      var requestFlag = true
      this.totalMetric.forEach((item) => {
        if (!item.value.trim()) {
          this.$Message.warning('指标表达式不能为空!')
          this.noDataTip = true
          requestFlag = false 
        }
        params.push(JSON.stringify({
          endpoint: this.$store.state.ip.value.split(':')[0],
          prom_ql: item.value,
          time: this.timeTnterval + ''
        })) 
      })
      if (!requestFlag) {
        return
      }
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
    },
    getOriginalMetricList() {
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.originMetricList.api, {id:2}, responseData => {
        this.originalList = responseData
      })
    }
  },
  components: {
    Notice,
    Searchinput
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
