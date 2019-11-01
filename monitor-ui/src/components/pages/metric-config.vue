<template>
  <div class="text-align:center; ">
    <Title :title="$t('title.metricConfiguration')"></Title>
    <div style="margin-bottom:24px;">
      <Notice :noticeConfig='noticeConfig'> </Notice>
      <Searchinput :parentConfig="searchInputConfig" ref="choicedIP"></Searchinput> 
      <Select v-model="metricSelected" filterable multiple style="width:260px" :label-in-value="true" 
          @on-change="selectMetric" @on-open-change="metricSelectOpen" :placeholder="$t('placeholder.metric')">
          <Option v-for="item in metricList" :value="item.id + '^^' + item.prom_ql" :key="item.metric">{{item.metric}}</Option>
      </Select>
      <Select v-model="timeTnterval" style="width:80px;margin: 0 8px;">
        <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
      </Select>
      
      <button class="btn btn-sm btn-confirm-f" :disabled="btnDisable" @click="requestChart">{{$t('button.search')}}</button>
      <button class="btn btn-sm btn-cancle-f" :disabled="btnDisable" @click="addMetric">{{$t('button.addMetric')}}</button>
      <button class="btn btn-sm btn-cancle-f" :disabled="btnDisable" @click="saveConfig">{{$t('button.saveEdit')}}</button>
    </div>
    <section class="metric-section">
      <ul>
        <template v-for="(metricItem, metricIndex) in totalMetric">
          <li :key="metricIndex" class="metric-display">
            <Tag color="primary" type="border" closable @on-close="delMetric(metricItem)">{{$t('tableKey.s_metric')}}：{{metricItem.label}} 
              <i class="fa fa-pencil" aria-hidden="true" @click="editMetricName(metricItem,metricIndex)"></i>
            </Tag>
            <Select v-model="originalMetricList[metricItem.key].model" style="width:300px" filterable size="small" @on-change="selectOriginalMetric(metricItem)">
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
            tip: 'tips.metricConfigTips.one'
          },
          {
            tip: 'tips.metricConfigTips.two'
          },
          {
            tip: 'tips.metricConfigTips.three'
          },
        ]
      },
      searchInputConfig: {
        poptipWidth: 300,
        placeholder: 'placeholder.endpointSearch',
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
        modalTitle: 'tableKey.s_metric',
        isAdd: true,
        config: [
          {label: 'tableKey.name', value: 'name', placeholder: 'tips.inputRequired', v_validate: 'required:true|min:2|max:60', disabled: false, type: 'text'},
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
    btnDisable: function() {
      return this.$validate.isEmpty_reset(this.$store.state.ip)
    },
    changeIP() {
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
          this.$Message.warning(this.$t('tableKey.endpoint')+this.$t('tips.required'))
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
        this.$Message.warning(this.$t('tableKey.s_metric')+this.$t('tips.required'))
        this.noDataTip = true
        return
      }
      let params = []
      var requestFlag = true
      this.totalMetric.forEach((item) => {
        if (!item.value.trim()) {
          this.$Message.warning(this.$t('tableKey.expr')+this.$t('tips.required'))
          this.noDataTip = true
          requestFlag = false 
        }
        params.push(JSON.stringify({
          endpoint: this.$store.state.ip.value.split(':')[0],
          prom_ql: item.value,
          metric: item.label,
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
          item.areaStyle = {}
        }) 
        let config = {
          title: responseData.title,
          legend: legend,
          series: responseData.series,
          yaxis: responseData.yaxis,
        }
        drawChart(this, config, {eye: false,clear:true})
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
        this.$Message.success(this.$t('tips.success'))
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
    height: 80px;
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
