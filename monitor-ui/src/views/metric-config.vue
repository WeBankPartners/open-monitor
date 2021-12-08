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
            <Select filterable clearable v-model="endpointType" @on-clear="clearEndpointType" @on-change="changeEndpointType" style="width:300px">
              <Option v-for="type in endpointTypeOptions" :value="type" :key="type">{{ type }}</Option>
            </Select>
          </div>
          <div>
            <button class="btn btn-sm btn-confirm-f" @click="configMetric">{{$t('button.search')}}</button>
          </div>
        </div>
        <!-- 操作区 -->
        <div v-if="showConfigTab || isAddMetric">
          <Tabs value="name1">
            <TabPane :label="$t('m_acquisition_configuration')" name="name1">
              <div >
                <Form :label-width="80">
                  <FormItem :label="$t('field.metric')">
                    <Select v-model="metricId" filterable clearable @on-open-change="getMetricOptions" @on-change="changeMetricOptions" ref="metricSelect" :disabled="!endpointType">
                      <Button type="success" style="width:92%;background-color:#19be6b" @click="addMetric" size="small">
                        <Icon type="ios-add" size="24"></Icon>
                      </Button>
                      <Option v-for="metric in metricOptions" :value="metric.id" :key="metric.id">{{ metric.metric }}<span style="float:right">
                          <Button icon="ios-trash" type="error" @click="deleteMetric(metric)" size="small" style="background-color:#ed4014"></Button>
                        </span>
                      </Option>
                    </Select>
                  </FormItem>
                  <FormItem :label="$t('tableKey.name')">
                    <Input v-model="metricConfigData.metric"></Input>
                  </FormItem>
                  <FormItem :label="$t('m_endpoint')">
                    <Select filterable clearable v-model="endpoint" @on-open-change="getEndpointForAcquisitionConfiguration" @on-change="collectedMetric = ''">
                      <Option v-for="item in endpointOptions" :value="item.id" :key="item.id">{{ item.guid }}</Option>
                    </Select>
                  </FormItem>

                  <FormItem :label="$t('m_collected_data')">
                    <Select filterable clearable v-model="collectedMetric" class="select-dropdown" @on-open-change="getCollectedMetric" @on-change="changeCollectedMetric">
                      <Option 
                        style="white-space: normal;"
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
                    <div :id="acquisitionConfigurationElId" class="echart"></div>
                  </div>
                </div>
              </div>
              
              <div style="text-align: right;margin-top:24px">
                <button :disabled="metricConfigData.prom_ql === ''" class="btn btn-sm btn-cancel-f" @click="preview('acquisitionConfiguration')">{{$t('m_preview')}}</button>
                <button class="btn btn-sm btn-confirm-f" @click="saveMetric">{{$t('button.saveConfig')}}</button>
              </div>
            </TabPane>
            <TabPane :label="$t('m_display_group')" name="name2" v-if="!isAddMetric">
              <div>
                <div style="margin-left: 16px;margin-right:16px;display:inline-block;">
                  <span style="font-size: 14px;">
                    {{$t('m_display_group')}}:
                  </span>
                  <Select filterable clearable v-model="selectdPanel" style="width:300px" @on-open-change="getPanelinfo" @on-change="changePanel" ref="selectdPanel">
                    <Button type="success" style="width:92%;background-color:#19be6b" @click="addPanel('panel')" size="small">
                      <Icon type="ios-add" size="24"></Icon>
                    </Button>
                    <Option v-for="panel in panelOptions" :value="panel.chart_group" :key="panel.chart_group">{{ panel.title }}<span style="float:right">
                        <Button icon="ios-trash" @click="deletePanel(panel)" type="error" size="small" style="background-color:#ed4014"></Button>
                      </span>
                      <span style="float:right">
                        <Button icon="ios-create-outline" @click="editPanel('panel', panel)" type="primary" size="small" style="background-color:#0080FF"></Button>
                      </span>
                    </Option>
                  </Select>
                </div>
                <div style="margin-left: 40px;margin-right:16px;display:inline-block;">
                  <span style="font-size: 14px;">
                    {{$t('m_graph')}}:
                  </span>
                  <Select v-model="selectdGraph" filterable clearable style="width:300px" @on-open-change="getGraphInfo" @on-change="changeGraph" :disabled="!selectdPanel" ref="selectdGraph">
                    <Button type="success" style="width:92%;background-color:#19be6b" @click="addGraph()" size="small">
                      <Icon type="ios-add" size="24"></Icon>
                    </Button>
                    <Option v-for="graph in graphOptions" :value="graph.id" :key="graph.id">{{ graph.title }}<span style="float:right">
                        <Button icon="ios-trash" type="error" @click="deleteGraph(graph)" size="small" style="background-color:#ed4014"></Button>
                      </span>
                    </Option>
                  </Select>
                </div>
                <div v-if="showGraphConfig" style="margin-top:48px">
                  <Form :label-width="80" inline>
                    <div>
                      <FormItem :label="$t('m_graph_name')">
                        <Input v-model="graphConfig.graphName" style="width: 300px"></Input>
                      </FormItem>
                      <FormItem :label="$t('field.unit')">
                        <Input v-model="graphConfig.unit"  style="width: 300px" />
                      </FormItem>
                    </div>
                    <div>
                      <FormItem :label="$t('field.legend')">
                        <Select filterable clearable allow-create v-model="graphConfig.legend" @on-create="handleCreateLegend" :placeholder="$t('m_legend_tips')" style="width: 300px">
                          <Option v-for="item in legendOptions" :value="item.value" :key="item.value">{{ item.label }}</Option>
                        </Select>
                      </FormItem>
                      <FormItem :label="$t('field.metric')">
                        <Select v-model="graphConfig.metric" filterable multiple clearable style="width:300px">
                          <Option v-for="(metric, index) in metricOptions" :value="metric.metric" :key="metric.metric + index">{{ metric.metric }}</Option>
                        </Select>
                      </FormItem>
                    </div>
                  </Form>
                  <div v-if="isRequestChartData" class="metric-section" style="margin-top:24px">
                    <div>
                      <div :id="displayGroupElId" class="echart"></div>
                    </div>
                  </div>
                  <div style="text-align: right;margin-top:24px">
                    <button :disabled="graphConfig.metric.length === 0" class="btn btn-sm btn-cancel-f" @click="preview('displayGroup')">{{$t('m_preview')}}</button>
                    <button class="btn btn-sm btn-confirm-f" @click="saveGraphMetric">{{$t('button.saveConfig')}}</button>
                  </div>
                </div>
              </div>
            </TabPane>
          </Tabs>
          <Modal
            v-model="showEndpointSelect"
            @on-ok="checkEndpoint"
            @on-cancel="metricConfigData.endpoint = ''"
            :title="$t('m_select_endpoint')">
            <Form :label-width="80">
              <FormItem :label="$t('m_endpoint')">
                <Select filterable clearable v-model="metricConfigData.endpoint" style="width:300px">
                  <Option v-for="item in endpointOptions" :value="item.id" :key="item.id">{{ item.guid }}</Option>
                </Select>
              </FormItem>
            </Form>
          </Modal>

          <Modal
            v-model="titleManagement.show"
            @on-ok="saveTitle"
            @on-cancel="titleManagement.title = ''"
            :title="titleManagement.isAdd ? $t('button.add') : $t('button.edit')">
            <Form :label-width="80">
              <FormItem :label="$t('field.title')">
                <Input v-model="titleManagement.title"/>
              </FormItem>
            </Form>
          </Modal>
        </div>
      </div>
    </div>
    <Modal
      v-model="isShowWarning"
      :title="$t('delConfirm.title')"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
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
      isShowWarning: false,
      deleteConfirm: {
        id: '',
        method: ''
      },
      endpointType: '',
      endpointTypeOptions: [],
      metricId: '',
      metricOptions: [],
      collectedMetric: '',
      collectedMetricOptions: [],
      showConfigTab: false,
      endpoint: '',
      metricConfigData: {
        id: null,
        metric: '',
        metric_type: '',
        panel_id: null,
        prom_ql: '',
        endpoint: ''
      },
      displayGroupElId: '',
      acquisitionConfigurationElId: '',
      noDataTip: true,
      isRequestChartData: false,
      previewPosition: '',
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
      chart: {
        group_id: '',
        metric: '',
        title: '',
        unit: '',
      },

      isAddMetric: false,
      legendOptions: [
        {label: '指标名称+对象+标签', value: '$custom_all'},
        {label: '指标名称+标签', value: '$custom_metric'},
        {label: '指标名称', value: '$metric'}
      ],

      selectdPanel: '',
      panelOptions: [],
      selectdGraph: '',
      graphOptions: [],
      graphConfig: {
        id: '',
        graphName: '',
        unit: '',
        metric: [],
        legend: ''
      },
      showGraphConfig: false, // 指标配置区

      titleManagement: {
        show:false,
        isAdd: true,
        type: '',
        id: '',
        title: '',
      }
    }
  },
  mounted() {
    this.getEndpointType()
  },
  methods: {
    clearEndpointType () {
      this.showConfigTab = false
      this.metricId = ''
      this.metricConfigData = {
        id: null,
        metric: '',
        metric_type: '',
        panel_id: null,
        prom_ql: '',
        endpoint: ''
      }
      this.collectedMetric = ''
      this.isRequestChartData = false
    },
    handleCreateLegend (val) {
      const exist = this.legendOptions.find(el => el.value === val)
      if (!exist) {
        this.legendOptions.push({label: val, value: val})
      }
    },
    changeEndpointType () {
      this.metricId = ''
    },
    changeMetricOptions (val) {
      if (val !== '') {
        const findMetricConfig = this.metricOptions.find(m => m.id === this.metricId)
        this.metricConfigData = {
          ...findMetricConfig
        }
      }
    },
    ok () {
      this[this.deleteConfirm.method](this.deleteConfirm.id)
    },
    cancel () {
      this.isShowWarning = false
    },
    deleteMetric (metric) {
      this.deleteConfirm.id = metric.id
      this.deleteConfirm.method = 'removeMetric'
      this.isShowWarning = true
    },
    removeMetric (id) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', this.$root.apiCenter.metricManagement + '?id=' + id, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.metricId = ''
        this.showConfigTab = false
      })
    },
    changePanel () {
      this.showGraphConfig =false
      this.selectdGraph = ''
      this.changeGraph('')
    },
    changeGraph (val) {
      if (val) {
        this.editGraph(val)
      } else {
        this.showGraphConfig = false
      }
    },
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
    addPanel (type) {
      this.titleManagement.show = true
      this.titleManagement.isAdd = true
      this.titleManagement.type = type
      this.titleManagement.title = ''
      this.$refs.selectdPanel.visible = false
    },
    editPanel (type, item) {
      this.titleManagement.show = true
      this.titleManagement.isAdd = false
      this.titleManagement.type = type
      this.titleManagement.title = item.title
      this.titleManagement.id = item.id
    },
    deletePanel (item) {
      this.deleteConfirm.id = item.id
      this.deleteConfirm.method = 'removePanel'
      this.isShowWarning = true
    },
    removePanel (id) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', this.$root.apiCenter.addPanel + '?ids=' + id, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.selectdPanel = ''
        this.$root.$eventBus.$emit('hideConfirmModal')
      })
    },
    addGraph () {
      this.selectdGraph = ''
      this.$nextTick( () => {
        this.graphConfig.id = ''
        this.graphConfig.graphName = ''
        this.graphConfig.unit = ''
        this.graphConfig.metric = []
        this.graphConfig.legend = ''
        this.$refs.selectdGraph.visible = false
        this.showGraphConfig = true
      })
    },
    editGraph (id) {
      this.showGraphConfig = true
      const graph = this.graphOptions.find(el => el.id === id)
      this.graphConfig.id = id
      this.graphConfig.graphName = graph.title
      this.graphConfig.unit = graph.unit
      this.graphConfig.metric = graph.metric.split('^').filter(item => item !== '')
      this.graphConfig.legend = graph.legend
      const exist = this.legendOptions.find(el => el.value === graph.legend)
      if (!exist) {
        this.legendOptions.push({label: graph.legend, value: graph.legend})
      }
    },
    deleteGraph (item) {
      this.deleteConfirm.id = item.id
      this.deleteConfirm.method = 'removeGraph'
      this.isShowWarning = true
    },
    removeGraph (id) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', this.$root.apiCenter.getGraph + '?ids=' + id, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.selectMetrc = ''
      })
    },
    saveTitle () {
      if (this.titleManagement.type === 'panel') {
        if (this.titleManagement.isAdd) {
          let params = {
            title: this.titleManagement.title
          }
          this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.addPanel + '/' + this.endpointType, [params], () => {
            this.$Message.success(this.$t('tips.success'))
          })
        } else {
          let params = {
            id: this.titleManagement.id,
            title: this.titleManagement.title
          }
          this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', this.$root.apiCenter.addPanel, [params], () => {
            this.$Message.success(this.$t('tips.success'))
          })
        }
        this.selectdPanel = ''
      }
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
    saveGraphMetric () {
      let params = {
        metric: this.graphConfig.metric.join('^'),
        title: this.graphConfig.graphName,
        unit: this.graphConfig.unit,
        legend: this.graphConfig.legend,
        group_id: this.selectdPanel,
        id: this.selectdGraph
      }
      if (this.selectdGraph) {
        this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', this.$root.apiCenter.getGraph, [params], () => {
          this.$Message.success(this.$t('tips.success'))
        })
      } else {
        delete params.id
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.getGraph, [params], () => {
          this.$Message.success(this.$t('tips.success'))
        })
      }
    },
    getGraphInfo () {
      const params = {
        groupId: this.selectdPanel
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getGraph, params, responseData => {
        this.graphOptions = responseData
      }, {isNeedloading: false})
    },
    getPanelinfo () {
      const params = {
        endpointType: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.panelInfo, params, responseData => {
        this.panelOptions = responseData
      }, {isNeedloading: false})
    },
    async getChartData () {
      this.isRequestChartData = true
      generateUuid().then((elId)=>{
        this.acquisitionConfigurationElId =  `id_${elId}`
      })
      let params = {
        aggregate: 'none',
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: []
      }
      const find = this.endpointOptions.find(e => e.id === this.metricConfigData.endpoint)
      params.data = [{
        endpoint: find.guid,
        prom_ql: this.metricConfigData.prom_ql,
        metric: this.metricConfigData.prom_ql === '' ? this.metricConfigData.metric : ''
      }]
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {eye: false,clear:true, zoomCallback: true}
        readyToDraw(this,responseData, 1, chartConfig, this.acquisitionConfigurationElId)
      }, {isNeedloading: false})
    },
    async getDispalyChartData () {
      generateUuid().then((elId)=>{
        this.displayGroupElId =  `id_${elId}`
      })
      let params = {
        aggregate: 'none',
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: []
      }
      const find = this.endpointOptions.find(e => e.id === this.metricConfigData.endpoint)
      this.graphConfig.metric.forEach(metric => {
        params.data.push({
          endpoint: find.guid,
          prom_ql: '',
          metric: metric
        })
      })
      this.isRequestChartData = true
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {eye: false,clear:true, zoomCallback: true}
        readyToDraw(this,responseData, 1, chartConfig, this.displayGroupElId)
      }, {isNeedloading: false})
    },
    checkEndpoint () {
      if (this.previewPosition === 'acquisitionConfiguration') {
        this.getChartData()
      } else {
        this.getDispalyChartData()
      }
    },
    getEndpoint () {
      this.metricConfigData.endpoint = ''
      const params = {
        type: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.getEndpoint, params, responseData => {
        this.endpointOptions = responseData
        this.metricConfigData.endpoint = this.endpoint
        this.showEndpointSelect = true
      })
    },
    getEndpointForAcquisitionConfiguration () {
      const params = {
        type: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.getEndpoint, params, responseData => {
        this.endpointOptions = responseData
      })
    },
    preview (previewPosition) {
      this.previewPosition = previewPosition
      this.getEndpoint()
    },
    saveMetric () {
      const type = this.metricConfigData.id === null ? 'POST' : 'PUT'
      this.metricConfigData.metric_type = this.endpointType
      this.$root.$httpRequestEntrance.httpRequestEntrance(type, this.$root.apiCenter.metricManagement, [this.metricConfigData], () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.metricManagement, {endpointType: this.endpointType}, (res) => {
          this.metricOptions = res
          const newMetric = res.find(el => el.metric === this.metricConfigData.metric)
          if (newMetric) {
            this.metricId = newMetric.id
            this.configMetric()
          }
        })
      })
    },
    getEndpointType () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, '', (responseData) => {
        this.endpointTypeOptions = responseData
      }, {isNeedloading: false})
    },
    getMetricOptions () {
      const params = {
        endpointType: this.endpointType
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.getMetricByEndpointType, params, responseData => {
        this.metricOptions = responseData
      }, {isNeedloading: false})
    },
    changeCollectedMetric (val) {
      if (val) {
        this.metricConfigData.prom_ql += ' ' + val
      }
    },
    configMetric () {
      this.isAddMetric = false
      this.endpoint = ''
      this.collectedMetric = ''
      this.isRequestChartData = false
      this.selectdPanel = ''
      this.selectdGraph = ''
      this.changeGraph('')
      this.getMetricOptions()
      this.showConfigTab = true
    },
    getCollectedMetric () {
      const params = {
        endpoint_type: this.endpointType,
        id: this.endpoint
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getMetricOptions, params, (responseData) => {
        this.collectedMetricOptions = responseData
      }, {isNeedloading: false})
    }
  }
}
</script>
<style lang="less">
.select-dropdown /deep/ .ivu-select-dropdown {
  max-height: 100px !important;
}
</style>
<style scoped lang="less">
  .ivu-form-item {
    margin-bottom: 8px;
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

  .metric-section {
    width: 1100px;
    // margin: 0 auto;
  }
  .echart {
    height: 400px;
    width: 1100px;
    background: #f5f7f9;
  }
  .echart-no-data-tip {
    text-align: center;
    vertical-align: middle;
    display: table-cell;
  }
</style>