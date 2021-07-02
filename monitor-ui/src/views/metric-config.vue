<template>
  <div class="">
    <Title :title="$t('title.metricConfiguration')"></Title>
    <div style="margin-bottom:24px;">
      <!-- 信息提示去 -->
      <div class="page-notice" :class="'page-notice-'+noticeConfig.type">
        <template v-for="(noticeItem, noticeIndex) in noticeConfig.contents">
          <p :key="noticeIndex">{{$t(noticeItem.tip)}}</p>
        </template>      
      </div>
      <div style="width:1100px;margin: 0 auto;margin-top:24px">
         <!-- 条件选择去 -->
        <div style="display:flex;margin-bottom:24px">
          <div style="margin-right:16px">
            <span style="font-size: 14px;">
              {{$t('field.type')}}:
            </span>
            <Select v-model="endpointType" style="width:300px">
              <Option v-for="type in endpointTypeOptions" :value="type" :key="type">{{ type }}</Option>
            </Select>
          </div>
          <div style="margin-right:16px">
            <span style="font-size: 14px;">
              {{$t('field.metric')}}:
            </span>
            <Select v-model="metricId" filterable style="width:300px" @on-open-change="getMetricOptions" ref="metricSelect" :disabled="!endpointType">
              <Button type="success" style="width:92%;background-color:#19be6b" @click="addMetric" size="small">
                <Icon type="ios-add" size="24"></Icon>
              </Button>
              <Option v-for="metric in metricOptions" :value="metric.id" :key="metric.id">{{ metric.metric }}
                <span style="float:right">
                  <Button icon="ios-trash" type="error" size="small" style="background-color:#ed4014"></Button>
                </span>
              </Option>
            </Select>
          </div>
          <div>
            <button class="btn btn-sm btn-confirm-f" :disabled="!metricId" @click="configMetric">{{$t('button.search')}}</button>
          </div>
        </div>
        <!-- 操作区 -->
        <div v-if="showConfigTab || isAddMetric">
          <Tabs value="name1">
            <TabPane :label="$t('field.metric')" name="name1">
              <div style="max-height:600px;overflow-y:auto">
                <Form :label-width="80">
                  <FormItem :label="$t('tableKey.name')">
                    <Input v-model="metricConfigData.metric"></Input>
                  </FormItem>
                  <FormItem :label="$t('m_collected_data')">
                    <Select v-model="collectedMetric" @on-change="changeCollectedMetric">
                      <Option 
                        style="width:300px;"
                        v-for="item in collectedMetricOptions" 
                        :value="item.option_value" 
                        :key="item.option_value">
                        {{ item.option_text }}
                      </Option>
                    </Select>
                  </FormItem>
                  <FormItem :label="$t('field.metric')">
                    <Input v-model="metricConfigData.prom_ql" type="textarea" :rows="6" />
                  </FormItem>
                </Form>

                <div v-if="isRequestChartData" class="metric-section">
                  <div>
                    <div :id="elId" class="echart"></div>
                  </div>
                  <!-- <div v-else class="echart echart-no-data-tip">
                    <span>~~~No Data!~~~</span>
                  </div> -->
                </div>
              </div>
              
              <div style="text-align: right;margin-top:24px">
                <button class="btn btn-sm btn-cancel-f" @click="preview">{{$t('m_preview')}}</button>
                <button class="btn btn-sm btn-confirm-f" @click="saveMetric">{{$t('button.saveEdit')}}</button>
              </div>
            </TabPane>
            <TabPane label="Panel" name="name2" v-if="!isAddMetric">
              <div>
                <div class="marginbottom params-each" style="margin-bottom:8px">
                  <label class="col-md-2 label-name">{{$t('tableKey.s_metric')}}:</label>
                    <Input v-model="chart.title" :placeholder="$t('tableKey.s_metric')" style="width: 300px" />
                </div>
                <div class="marginbottom params-each" style="margin-bottom:8px">
                  <label class="col-md-2 label-name">{{$t('field.unit')}}:</label>
                    <Input v-model="chart.unit" :placeholder="$t('field.unit')" style="width: 300px" />
                </div>
                <RadioGroup v-model="test"  @on-change="checkRadio">
                  <template v-for="(group, gIndex) in panelInfo.panel_list">
                    <div :key="group.panel_title">
                      <Radio :label="gIndex">{{group.panel_title}}</Radio>
                      <template v-for="(chart, cIndex) in group.charts">
                        <div :key="chart.metric" style="padding-left:20px">
                          <Radio :label="gIndex+','+cIndex">{{chart.metric}}</Radio>
                        </div>
                      </template>
                    </div>
                  </template>
                </RadioGroup>
                <div style="text-align: right;margin-top:24px">
                  <button class="btn btn-sm btn-confirm-f" @click="savePanel">{{$t('button.saveEdit')}}</button>
                </div>
              </div>
            </TabPane>
          </Tabs>
          <Modal
            v-model="showEndpointSelect"
            @on-ok="checkEndpoint"
            @on-cancel="metricConfigData.endpoint = ''"
            :title="$t('m_select_host')">
            <Form :label-width="80">
              <FormItem :label="$t('m_host')">
                <Select v-model="metricConfigData.endpoint" style="width:300px">
                  <Option v-for="item in endpointOptions" :value="item.guid" :key="item.guid">{{ item.guid }}</Option>
                </Select>
              </FormItem>
            </Form>
          </Modal>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import {generateUuid} from '@/assets/js/utils'
// 引入 ECharts 主模块
import {readyToDraw} from '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      endpointType: '',
      endpointTypeOptions: [],
      metricId: '',
      metricOptions: [],
      collectedMetric: '',
      collectedMetricOptions: [],
      showConfigTab: false,
      metricConfigData: {
        id: null,
        metric: '',
        metric_type: '',
        panel_id: null,
        prom_ql: '',
        endpoint: ''
      },
      elId: '',
      noDataTip: true,
      isRequestChartData: false,
      showEndpointSelect: false,
      endpointOptions: [],
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
      panelInfo: {
        active_chart: {},
        panel_list: []
      },
      chart: {
        group_id: '',
        metric: '',
        title: '',
        unit: '',
      },
      test: '',

      isAddMetric: false
    }
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`
    })
  },
  mounted() {
    this.getEndpointType()
  },
  methods: {
    addMetric () {
      this.$refs.metricSelect.visible = false
      this.metricId = ''
      this.metricConfigData = {
        id: null,
        metric: '',
        panel_id: null,
        prom_ql: '',
        endpoint: ''
      }
      this.isAddMetric = true
    },
    savePanel () {
      const findMetricConfig = this.metricOptions.find(m => m.id === this.metricId)
      let params = {
        panel_id: this.chart.group_id,
        chart: this.chart,
        metric: findMetricConfig.metric,
        metric_type: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.savePanel, [params], () => {
        this.$Message.success(this.$t('tips.success'))
      })
    },
    canclePanelConfig () {
      this.chart.group_id = ''
      this.chart.metric = ''
      this.chart.title = ''
      this.chart.unit = ''
      this.test = ''
    },
    checkRadio (item) {
      item += ''
      const num = item.split(',')
      let tmp = this.panelInfo.panel_list
      num.forEach(t => {
        tmp = tmp[Number(t)] || tmp.charts[Number(t)]
        this.chart.group_id = tmp.group_id || this.chart.group_id
      })
      if (num.length >1) {
        this.chart.title = tmp.title
        this.chart.unit = tmp.unit
        this.chart.metric = tmp.metric
      } else {
        this.chart.title = ''
        this.chart.unit = ''
        this.chart.metric = this.panelInfo.active_chart.metric
      }
    },
    getPanelinfo () {
      const findMetricConfig = this.metricOptions.find(m => m.id === this.metricId)
      const params ={
        type: this.endpointType,
        metric: findMetricConfig.metric
      }
      this.canclePanelConfig()
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.panelInfo, params, responseData => {
        this.panelInfo = responseData
        this.chart = JSON.parse(JSON.stringify(this.panelInfo.active_chart))
        this.chart.group_id = responseData.panel_group_id
      })
    },
    async getChartData () {
      this.isRequestChartData = true
      const params = [{
        endpoint: this.metricConfigData.endpoint,
        prom_ql: this.metricConfigData.prom_ql,
        metric: this.metricConfigData.metric,
        time: '-1800'
      }]
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {eye: false,clear:true, zoomCallback: true}
        readyToDraw(this,responseData, 1, chartConfig)
      })
    },
    checkEndpoint () {
      this.getChartData()
    },
    getEndpoint () {
      this.metricConfigData.endpoint = ''
      const params = {
        type: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.getEndpoint, params, responseData => {
        this.endpointOptions = responseData
        this.showEndpointSelect = true
      })
    },
    preview () {
      this.getEndpoint()
    },
    saveMetric () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.saveMetric, [this.metricConfigData], () => {
        this.$Message.success(this.$t('tips.success'))
      })
    },
    getEndpointType () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, '', (responseData) => {
        this.endpointTypeOptions = responseData
      }, {isNeedloading: false})
    },
    getMetricOptions () {
      const params = {
        type: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.metricList.api, params, responseData => {
        this.metricOptions = responseData
      }, {isNeedloading: false})
    },
    changeCollectedMetric (val) {
      this.metricConfigData.prom_ql += val
    },
    configMetric () {
      this.isAddMetric = false
      this.isRequestChartData = false
      this.getPanelinfo()
      const findMetricConfig = this.metricOptions.find(m => m.id === this.metricId)
      this.getCollectedMetric()
      this.metricConfigData = {
        ...findMetricConfig
      }
      this.showConfigTab = true
    },
    getCollectedMetric () {
      const params = {
        endpoint_type: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getMetricOptions, params, (responseData) => {
        this.collectedMetricOptions = responseData
      }, {isNeedloading: false})
    }
  }
}
</script>
<style scoped lang="less">
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

  .metric-section {
    width: 1000px;
    margin: 0 auto;
  }
  .echart {
    height: 400px;
    width: 1000px;
    margin-left: 40px;
    background: #f5f7f9;
  }
  .echart-no-data-tip {
    text-align: center;
    vertical-align: middle;
    display: table-cell;
  }
</style>