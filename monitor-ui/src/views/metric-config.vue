<template>
  <div class="">
    <Title :title="$t('title.metricConfiguration')"></Title>
    <div style="margin-bottom:24px;">
      <!-- <Notice :noticeConfig='noticeConfig'> </Notice> -->
      <div class="page-notice" :class="'page-notice-'+noticeConfig.type">
        <template v-for="(noticeItem, noticeIndex) in noticeConfig.contents">
          <p :key="noticeIndex">{{$t(noticeItem.tip)}}</p>
        </template>      
      </div>
      <Select
        style="width:300px;"
        v-model="endpoint"
        filterable
        clearable
        remote
        :label-in-value="true" 
        :placeholder="$t('requestMoreData')"
        :remote-method="getEndpointList"
        >
        <Option v-for="(option, index) in endpointList" :value="option.option_value" :key="index">
          <TagShow :tagName="option.option_type_name" :index="index"></TagShow> 
          {{option.option_text}}
        </Option>
        <Option value="moreTips" disabled>{{$t('tips.requestMoreData')}}</Option>
      </Select>
      <Select v-model="metricSelected" filterable multiple style="width:260px" :label-in-value="true" 
          @on-change="selectMetric" @on-open-change="metricSelectOpen" :placeholder="$t('placeholder.metric')">
          <Option v-for="item in metricList" :value="item.id + '^^' + item.prom_ql" :key="item.metric">{{item.metric}}</Option>
      </Select>
      <DatePicker 
        type="datetimerange" 
        :value="dateRange" 
        format="yyyy-MM-dd HH:mm:ss" 
        placement="bottom-start" 
        @on-change="datePick" 
        :placeholder="$t('placeholder.datePicker')" 
        style="width: 320px">
      </DatePicker>
      <Select v-model="timeTnterval" style="width:80px;margin: 0 8px;">
        <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
      </Select>
      
      <button class="btn btn-sm btn-confirm-f" :disabled="btnDisable" @click="getChartData">{{$t('button.search')}}</button>
      <button class="btn btn-sm btn-cancel-f" :disabled="btnDisable" @click="addMetric">{{$t('button.addMetric')}}</button>
      <button class="btn btn-sm btn-cancel-f" :disabled="btnDisable" @click="saveConfig">{{$t('button.saveEdit')}}</button>
    </div>
    <section class="metric-section">
      <ul>
        <template v-for="(metricItem, metricIndex) in totalMetric">
          <li :key="metricIndex" class="metric-display">
            <Tag color="primary" type="border" closable @on-close="delMetric(metricItem)">{{$t('tableKey.s_metric')}}：{{metricItem.label}} 
              <i class="fa fa-pencil" aria-hidden="true" @click="editMetricName(metricItem,metricIndex)"></i>
            </Tag>
            <Select v-model="originalMetricList[metricItem.key].model" style="width:300px" filterable size="small" @on-change="selectOriginalMetric(metricItem)">
              <Option 
                style="width:300px;"
                v-for="item in originalMetricList[metricItem.key].list" 
                :value="item.option_value" 
                :key="item.option_value">
                {{ item.option_text }}
              </Option>
            </Select>
            <div>
               <textarea v-model="metricItem.value" class="textareaSty"></textarea> 
            </div>
          </li>
        </template>
      </ul>
    </section>
    <div v-if="isRequestChartData" class="metric-section">
      <div v-if="!noDataTip">
        <div :id="elId" class="echart"></div>
      </div>
      <div v-else class="echart echart-no-data-tip">
        <span>~~~No Data!~~~</span>
      </div>
    </div>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
// import Notice from '@/components/notice'
import {dataPick} from '@/assets/config/common-config'
import {generateUuid} from '@/assets/js/utils'
import TagShow from '@/components/Tag-show.vue'
// 引入 ECharts 主模块
import {readyToDraw} from  '@/assets/config/chart-rely'

export default {
  name: '',
  data() {
    return {
      endpoint: '',
      endpointObject: {},
      endpointId: '',
      endpointList: [],
      noticeConfig: {
        type: 'info',
        contents: [
          {
            tip: 'metricConfigTipsone'
          },
          {
            tip: 'metricConfigTipstwo'
          },
          {
            tip: 'metricConfigTipsthree'
          },
        ]
      },
      elId: '',
      isRequestChartData: false,
      noDataTip: false,

      metricSelected: [],
      metricSelectedOptions: [],
      metricList: [],

      dateRange: ['', ''],

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
    endpoint: function (val) {
      if (val) {
        this.endpointObject = this.endpointList.find(ep => {
          return ep.option_value === val
        })
      } else {
        this.endpointObject = {}
      }
      this.getOriginalMetricList()
    },
    changeIP () {
      this.metricSelected = []
      this.metricSelectedOptions = []
      this.metricList = []
      this.noDataTip = true
    },
    isShow: function () {
      this.getEndpointList('.')
    }
  },
  computed: {
    btnDisable: function() {
      return this.$root.$validate.isEmpty_reset(this.endpoint)
    },
    changeIP() {
      return this.endpoint
    },
    totalMetric: function () {
      return this.metricSelectedOptions.concat(this.editMetric)
    },
    isShow: function () {
      return !this.$root.$validate.isEmpty_reset(this.endpoint)
    }
  },
  mounted() {
    this.getEndpointList('.')
  },
  methods: {
    datePick (data) {
      this.dateRange = data
      if (this.dateRange[0] && this.dateRange[1]) {
        if (this.dateRange[0] === this.dateRange[1]) {
          this.dateRange[1] = this.dateRange[1].replace('00:00:00', '23:59:59')
        }
      }
    },
    getEndpointList(query) {
      let params = {
        search: query,
        page: 1,
        size: 1000
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.api, params, (responseData) => {
        this.endpointList = []
        responseData.forEach((item) => {
          if (item.id !== -1) {
            this.endpointList.push(item)
          }
        })
      })
    },
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
          if (this.originalMetricList[item.key].model) {
            this.editMetric[i].value = this.editMetric[i].value + this.originalMetricList[item.key].model
          }
          
        }
      } else {
        for (let i=0; i< this.metricSelectedOptions.length; i++) {
          if (item.key === this.metricSelectedOptions[i].key) 
          if (this.originalMetricList[item.key].model) {
            this.metricSelectedOptions[i].value = this.metricSelectedOptions[i].value + this.originalMetricList[item.key].model
          }
        }
      }
    },
    delMetric (metric) {
      this.isRequestChartData = false
      this.editMetric =  this.editMetric.filter((item)=>{
        return item.label !== metric.label
      })
      this.metricSelectedOptions = this.metricSelectedOptions.filter((item)=>{
        return item.label !== metric.label
      })
      this.metricSelected = this.metricSelected.filter((item)=>{
        return item !== `${metric.id}^^${metric.value}`
      })
    },
    selectMetric (option) {
      this.isRequestChartData = false
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
        if (!this.$root.$validate.isEmpty_reset(this.endpoint)) {
          this.obtainMetricList(this.endpointObject.type)
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
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.metricList.api, params, responseData => {
        this.metricList = responseData
      })
    },
    getChartData () {
      this.noDataTip = false
      if (this.$root.$validate.isEmpty_reset(this.totalMetric)) {
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
        params.push({
          endpoint: this.endpointObject.option_value,
          prom_ql: item.value,
          metric: item.label,
          time: this.timeTnterval + '',
          start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0].replace(/-/g, '/'))/1000+'',
          end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1].replace(/-/g, '/'))/1000+''
        })
      })
      if (!requestFlag) {
        return
      }
      this.isRequestChartData = true
      
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {eye: false,clear:true, zoomCallback: true}
        readyToDraw(this,responseData, 1, chartConfig)
      })

    },
    saveConfig () {
      let params = []
      this.totalMetric.forEach((item) => {
        if (item.value === '') {
          return
        }
        let {id:id,label:metric,value:prom_ql} = item
        params.push({id,metric,prom_ql,metric_type: this.endpointObject.type})
      })
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.metricUpdate.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.metricSelected = []
        this.editMetric = []
        this.endpoint = ''
        this.isRequestChartData = false
        this.obtainMetricList()
      })
    },
    editMetricName (metricItem,metricIndex) {
      this.editingMetric = metricIndex
      this.modelConfig.addRow.name = metricItem.label
      this.$root.JQ('#edit_metric_Modal').modal('show')
    },
    addPost (){
      this.totalMetric[this.editingMetric].label = this.modelConfig.addRow.name
      this.$root.JQ('#edit_metric_Modal').modal('hide')
    },
    getOriginalMetricList() {
      if (this.$root.$validate.isEmpty_reset(this.endpointObject)) {
        return
      }
      this.originalList = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.originMetricList.api, {id:this.endpointObject.id}, responseData => {
        this.originalList = responseData
      })
    }
  },
  components: {
    TagShow
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

   // 页面提示信息样式--开始
  .page-notice {
    margin:16px 0;
    padding: 10px;
  }
  .page-notice-info {
    color: #5db558;
    background: #f4fbf5;
    border:1px solid #35b34a;
    border-radius: 4px;
  }
</style>
