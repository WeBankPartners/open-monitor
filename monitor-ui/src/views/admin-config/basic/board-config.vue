<!--看板配置-->
<template>
  <div class="monitor-group-board">
    <div style="margin-right:16px;display:inline-block;">
      <Form :label-width="100" inline>
        <div>
          <FormItem :label="$t('m_field_type')">
            <Select filterable v-model="monitorType" @on-clear="clearEndpointType" @on-change="changeEndpointType" style="width:300px">
              <Option v-for="type in monitorTypeOptions" :value="type" :label="type" :key="type">{{ type }}</Option>
            </Select>
          </FormItem>
        </div>
        <div>
          <FormItem :label="$t('m_display_group')">
            <Select filterable clearable v-model="selectdPanel" style="width:300px" @on-open-change="getPanelinfo" @on-change="changePanel" ref="selectdPanel">
              <Button type="success" style="width:92%;background-color:#00CB91" @click="addPanel('panel')" size="small">
                <Icon type="ios-add" size="24"></Icon>
              </Button>
              <Option v-for="panel in panelOptions" :value="panel.chart_group" :key="panel.chart_group">{{ panel.title }}<span style="float:right">
                <Button icon="md-trash" @click="deletePanel(panel)" type="error" size="small"></Button>
              </span>
                <span style="float:right">
                  <Button icon="md-create" @click="editPanel('panel', panel)" type="primary" size="small" style="background-color:#0080FF"></Button>
                </span>
              </Option>
            </Select>
          </FormItem>
          <FormItem :label="$t('m_graph')">
            <Select v-model="selectdGraph" filterable clearable style="width:300px" @on-open-change="getGraphInfo" @on-change="changeGraph" :disabled="!selectdPanel" ref="selectdGraph">
              <Button type="success" style="width:92%;background-color:#00CB91" @click="addGraph()" size="small">
                <Icon type="ios-add" size="24"></Icon>
              </Button>
              <Option v-for="graph in graphOptions" :value="graph.id" :key="graph.id">{{ graph.title }}<span style="float:right">
                <Button icon="md-trash" type="error" @click="deleteGraph(graph)" size="small"></Button>
              </span>
              </Option>
            </Select>
          </FormItem>
        </div>
      </Form>
    </div>
    <div v-if="showGraphConfig" style="margin-top: 20px">
      <Form :label-width="100" inline>
        <div>
          <FormItem :label="$t('m_graph_name')">
            <Input v-model="graphConfig.graphName" style="width: 300px"></Input>
          </FormItem>
          <FormItem :label="$t('m_field_unit')">
            <Input v-model="graphConfig.unit"  style="width: 300px" />
          </FormItem>
        </div>
        <div>
          <FormItem :label="$t('m_field_legend')">
            <Select filterable clearable allow-create v-model="graphConfig.legend" @on-create="handleCreateLegend" :placeholder="$t('m_legend_tips')" style="width: 300px">
              <Option v-for="item in legendOptions" :value="item.value" :label="item.label" :key="item.value">{{ item.label }}</Option>
            </Select>
          </FormItem>
          <FormItem :label="$t('m_field_metric')">
            <Select v-model="graphConfig.metric" filterable multiple clearable style="width:300px">
              <Option v-for="(metric, index) in metricOptions" :value="metric.metric" :label="metric.metric" :key="metric.metric + index">{{ metric.metric }}</Option>
            </Select>
          </FormItem>
        </div>
      </Form>
      <div v-if="isRequestChartData" class="metric-section" style="margin-top:24px">
        <div>
          <div :id="displayGroupElId" class="echart"></div>
        </div>
      </div>
      <div style="width:900px;text-align:center;margin-top:24px">
        <Button :disabled="graphConfig.metric.length === 0" @click="preview('displayGroup')">{{$t('m_preview')}}</Button>
        <Button type="primary" @click="saveGraphMetric">{{$t('m_button_saveConfig')}}</Button>
      </div>
    </div>
    <Modal
      v-model="showEndpointSelect"
      :mask-closable="false"
      @on-ok="checkEndpoint"
      @on-cancel="metricConfigData.endpoint = ''"
      :title="$t('m_select_endpoint')"
    >
      <Form :label-width="80">
        <FormItem :label="$t('m_endpoint')">
          <Select filterable clearable v-model="metricConfigData.endpoint" style="width:300px">
            <Option v-for="item in endpointOptions" :value="item.guid" :label="item.guid" :key="item.guid">{{ item.guid }}</Option>
          </Select>
        </FormItem>
      </Form>
    </Modal>
    <Modal
      v-model="titleManagement.show"
      :mask-closable="false"
      @on-ok="saveTitle"
      @on-cancel="titleManagement.title = ''"
      :title="titleManagement.isAdd ? $t('m_button_add') : $t('m_button_edit')"
    >
      <Form :label-width="80">
        <FormItem :label="$t('m_field_title')">
          <Input v-model="titleManagement.title"/>
        </FormItem>
      </Form>
    </Modal>
    <Modal
      v-model="isShowWarning"
      :title="$t('m_delConfirm_title')"
      @on-ok="ok"
      @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
    <ChartLinesModal
      :isLineSelectModalShow="isLineSelectModalShow"
      :chartId="setChartConfigId"
      @modalClose="onLineSelectChangeCancel"
    >
    </ChartLinesModal>
  </div>
</template>
<script>
import {isEmpty, cloneDeep} from 'lodash'
import {generateUuid} from '@/assets/js/utils'
import {readyToDraw} from '@/assets/config/chart-rely'
import ChartLinesModal from '@/components/chart-lines-modal'

export const custom_api_enum = [
  {
    addPanel: 'delete'
  },
  {
    getGraph: 'delete'
  },
  {
    newPanelHost: 'post'
  }
]

export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      deleteConfirm: {
        id: '',
        method: ''
      },
      monitorType: 'host',
      monitorTypeOptions: [],
      metricOptions: [],
      endpoint: '',
      metricConfigData: {
        guid: null,
        metric: '',
        monitor_type: '',
        panel_id: null,
        prom_expr: '',
        endpoint: ''
      },
      displayGroupElId: '',
      isRequestChartData: false,
      showEndpointSelect: false,
      endpointOptions: [],
      legendOptions: [
        {
          label: '指标名称+对象+标签',
          value: '$custom_all'
        },
        {
          label: '指标名称+标签',
          value: '$custom_metric'
        },
        {
          label: '指标名称',
          value: '$metric'
        }
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
        show: false,
        isAdd: true,
        type: '',
        id: '',
        title: '',
      },
      originalMetricsId: '',
      isLineSelectModalShow: false,
      setChartConfigId: '',
      chartInstance: null,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  mounted() {
    this.getMetricOptions()
    this.getEndpointType()
    this.$on('editShowLines', this.handleEditShowLines)
  },
  methods: {
    configMetric() {
      this.clearData()
      this.changeGraph('')
      this.getMetricOptions()
    },
    changeEndpointType() {
      this.configMetric()
    },
    getMetricOptions() {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup
      }
      this.request('GET', this.apiCenter.metricList.api, params, responseData => {
        this.metricOptions = responseData
      }, {isNeedloading: false})
    },
    clearEndpointType() {
      this.metricConfigData = {
        guid: null,
        metric: '',
        monitor_type: '',
        panel_id: null,
        prom_expr: '',
        endpoint: ''
      }
      this.isRequestChartData = false
    },
    handleCreateLegend(val) {
      const exist = this.legendOptions.find(el => el.value === val)
      if (!exist) {
        this.legendOptions.push({
          label: val,
          value: val
        })
      }
    },
    ok() {
      this[this.deleteConfirm.method](this.deleteConfirm.id)
    },
    cancel() {
      this.isShowWarning = false
    },
    changePanel() {
      this.showGraphConfig =false
      this.selectdGraph = ''
      this.changeGraph('')
    },
    changeGraph(val) {
      if (val) {
        this.editGraph(val)
      } else {
        this.showGraphConfig = false
      }
    },
    addPanel(type) {
      this.titleManagement.show = true
      this.titleManagement.isAdd = true
      this.titleManagement.type = type
      this.titleManagement.title = ''
      this.$refs.selectdPanel.visible = false
    },
    editPanel(type, item) {
      this.titleManagement.show = true
      this.titleManagement.isAdd = false
      this.titleManagement.type = type
      this.titleManagement.title = item.title
      this.titleManagement.id = item.id
    },
    deletePanel(item) {
      this.deleteConfirm.id = item.id
      this.deleteConfirm.method = 'removePanel'
      this.isShowWarning = true
    },
    removePanel(id) {
      this.request('DELETE', this.apiCenter.addPanel + '?ids=' + id, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.selectdPanel = ''
        this.$root.$eventBus.$emit('hideConfirmModal')
      })
    },
    addGraph() {
      this.selectdGraph = ''
      this.$nextTick(() => {
        this.graphConfig.id = ''
        this.graphConfig.graphName = ''
        this.graphConfig.unit = ''
        this.graphConfig.metric = []
        this.graphConfig.legend = ''
        this.$refs.selectdGraph.visible = false
        this.showGraphConfig = true
      })
    },
    editGraph(id) {
      this.showGraphConfig = true
      const graph = this.graphOptions.find(el => el.id === id)
      this.graphConfig.id = id
      this.graphConfig.graphName = graph.title
      this.graphConfig.unit = graph.unit
      this.graphConfig.metric = graph.metric.split('^').filter(item => item !== '')
      this.graphConfig.legend = graph.legend
      const exist = this.legendOptions.find(el => el.value === graph.legend)
      if (!exist) {
        this.legendOptions.push({
          label: graph.legend,
          value: graph.legend
        })
      }
    },
    deleteGraph(item) {
      this.deleteConfirm.id = item.id
      this.deleteConfirm.method = 'removeGraph'
      this.isShowWarning = true
    },
    removeGraph(id) {
      this.request('DELETE', this.apiCenter.getGraph + '?ids=' + id, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.selectMetrc = ''
      })
    },
    saveTitle() {
      if (this.titleManagement.type === 'panel') {
        if (this.titleManagement.isAdd) {
          const params = {
            title: this.titleManagement.title,
            service_group: this.serviceGroup
          }
          this.request('POST', this.apiCenter.addPanel + '/' + this.monitorType, [params], () => {
            this.$Message.success(this.$t('m_tips_success'))
          })
        } else {
          const params = {
            id: this.titleManagement.id,
            title: this.titleManagement.title,
            service_group: this.serviceGroup
          }
          this.request('PUT', this.apiCenter.addPanel, [params], () => {
            this.$Message.success(this.$t('m_tips_success'))
          })
        }
        this.selectdPanel = ''
      }
    },
    saveGraphMetric() {
      const params = {
        metric: this.graphConfig.metric.join('^'),
        title: this.graphConfig.graphName,
        unit: this.graphConfig.unit,
        legend: this.graphConfig.legend,
        group_id: this.selectdPanel,
        id: this.selectdGraph
      }
      if (this.selectdGraph) {
        this.request('PUT', this.apiCenter.getGraph, [params], () => {
          this.$Message.success(this.$t('m_tips_success'))
        })
      } else {
        delete params.id
        this.request('POST', this.apiCenter.getGraph, [params], () => {
          this.$Message.success(this.$t('m_tips_success'))
        })
      }
    },
    getGraphInfo() {
      const params = {
        groupId: this.selectdPanel
      }
      this.request('GET', this.apiCenter.getGraph, params, responseData => {
        this.graphOptions = responseData
      }, {isNeedloading: false})
    },
    getPanelinfo() {
      const params = {
        monitorType: this.monitorType,
        serviceGroup: this.serviceGroup
      }
      this.request('GET', this.apiCenter.panelInfo, params, responseData => {
        this.panelOptions = responseData
      }, {isNeedloading: false})
    },
    async getDispalyChartData() {
      const params = {
        aggregate: 'none',
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: []
      }
      const find = this.endpointOptions.find(e => e.guid === this.metricConfigData.endpoint)
      this.graphConfig.metric.forEach(metric => {
        params.data.push({
          endpoint: find.guid,
          prom_expr: '',
          metric
        })
      })
      this.isRequestChartData = true
      this.request('POST',this.apiCenter.metricConfigView.api, params, responseData => {
        // const chartConfig = {eye: false,clear:true, zoomCallback: true}
        const chartConfig = {
          eye: false,
          clear: true,
          lineBarSwitch: true,
          chartId: this.displayGroupElId,
          canEditShowLines: true,
          dataZoom: false
        }
        responseData.chartId = this.displayGroupElId
        this.chartInstance = readyToDraw(this,responseData, 1, chartConfig, this.displayGroupElId)
        if (this.chartInstance) {
          this.chartInstance.on('legendselectchanged', params => {
            window['view-config-selected-line-data'][this.displayGroupElId] = cloneDeep(params.selected)
          })
        }
      }, {isNeedloading: false})
    },
    checkEndpoint() {
      window['view-config-selected-line-data'] = {}
      generateUuid().then(elId => {
        this.displayGroupElId = `id_${elId}`
      })
      this.getDispalyChartData()
    },
    getEndpoint() {
      this.metricConfigData.endpoint = ''
      const params = {
        type: this.monitorType,
        serviceGroup: this.serviceGroup
      }
      this.request('GET',this.apiCenter.getEndpoint, params, responseData => {
        this.endpointOptions = responseData
        this.metricConfigData.endpoint = this.endpoint
        this.showEndpointSelect = true
      })
    },
    preview() {
      this.getEndpoint()
    },
    getEndpointType() {
      this.request('GET', this.apiCenter.getEndpointType, '', responseData => {
        this.monitorTypeOptions = responseData
      }, {isNeedloading: false})
    },
    clearData() {
      this.endpoint = ''
      this.isRequestChartData = false
      this.selectdPanel = ''
      this.selectdGraph = ''

      this.metricId = ''
      this.metricConfigData.metric = ''
      this.metricConfigData.prom_expr = ''
      this.isRequestChartData = false
    },
    handleEditShowLines(config) {
      this.setChartConfigId = config.chartId
      if (isEmpty(window['view-config-selected-line-data'][this.setChartConfigId])) {
        window['view-config-selected-line-data'][this.setChartConfigId] = {}
        config.legend.forEach(one => {
          window['view-config-selected-line-data'][this.setChartConfigId][one] = true
        })
      }
      this.isLineSelectModalShow = true
    },
    onLineSelectChangeCancel() {
      this.isLineSelectModalShow = false
      this.getDispalyChartData()
    },
  },
  components: {
    ChartLinesModal
  }
}
</script>
<style lang="less">
.select-dropdown /deep/ .ivu-select-dropdown {
  max-height: 100px !important;
}
</style>
<style scoped lang="less">
.monitor-group-board {
  padding: 12px;
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
}
</style>
