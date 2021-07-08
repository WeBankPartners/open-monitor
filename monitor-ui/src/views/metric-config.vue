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
            <Select filterable v-model="endpointType" style="width:300px">
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
              <Option v-for="metric in metricOptions" :value="metric.id" :key="metric.id">{{ metric.metric }}<span style="float:right">
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
            <TabPane :label="$t('m_acquisition_configuration')" name="name1">
              <div style="max-height:600px;overflow-y:auto">
                <Form :label-width="80">
                  <FormItem :label="$t('tableKey.name')">
                    <Input v-model="metricConfigData.metric"></Input>
                  </FormItem>
                  <FormItem :label="$t('m_collected_data')">
                    <Select filterable v-model="collectedMetric" @on-change="changeCollectedMetric">
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
                    <div :id="acquisitionConfigurationElId" class="echart"></div>
                  </div>
                </div>
              </div>
              
              <div style="text-align: right;margin-top:24px">
                <button class="btn btn-sm btn-cancel-f" @click="preview('acquisitionConfiguration')">{{$t('m_preview')}}</button>
                <button class="btn btn-sm btn-confirm-f" @click="saveMetric">{{$t('button.saveConfig')}}</button>
              </div>
            </TabPane>
            <TabPane :label="$t('m_display_group')" name="name2" v-if="!isAddMetric">
              <div>
                <div style="margin-right:16px;display:inline-block;">
                  <span style="font-size: 14px;">
                    {{$t('m_display_group')}}:
                  </span>
                  <Select filterable v-model="selectdPanel" style="width:300px" @on-open-change="getPanelinfo" @on-change="changePanel">
                    <Button type="success" style="width:92%;background-color:#19be6b" @click="addPanel('panel')" size="small">
                      <Icon type="ios-add" size="24"></Icon>
                    </Button>
                    <Option v-for="panel in panelOptions" :value="panel.chart_group" :key="panel.chart_group">{{ panel.title }}<span style="float:right">
                        <Button icon="ios-trash" @click="deletePanel('panel', panel)" type="error" size="small" style="background-color:#ed4014"></Button>
                      </span>
                      <span style="float:right">
                        <Button icon="ios-create-outline" @click="editPanel('panel', panel)" type="primary" size="small" style="background-color:#0080FF"></Button>
                      </span>
                    </Option>
                  </Select>
                </div>
                <div style="margin-right:16px;display:inline-block;">
                  <span style="font-size: 14px;">
                    {{$t('m_graph')}}:
                  </span>
                  <Select v-model="selectdGraph" filterable style="width:300px" @on-open-change="getGraphInfo" @on-change="changeGraph" :disabled="!selectdPanel">
                    <Button type="success" style="width:92%;background-color:#19be6b" @click="addMetric" size="small">
                      <Icon type="ios-add" size="24"></Icon>
                    </Button>
                    <Option v-for="graph in graphOptions" :value="graph.metric" :key="graph.metric">{{ graph.title }}<span style="float:right">
                        <Button icon="ios-trash" type="error" size="small" style="background-color:#ed4014"></Button>
                      </span>
                    </Option>
                  </Select>
                </div>
                <Button type="info" style="background-color:#2db7f5" @click="getGraphConfig" size="small">
                  {{$t('m_search')}}
                </Button>

                <div v-if="showGraphConfig" style="margin-top:48px">
                  <Form :label-width="80" inline>
                    <FormItem :label="$t('m_graph_name')">
                      <Input v-model="graphConfig.graphName"></Input>
                    </FormItem>
                    <FormItem :label="$t('field.unit')">
                      <Input v-model="graphConfig.unit"/>
                    </FormItem>
                    <FormItem :label="$t('field.metric')">
                      <Select v-model="graphConfig.metric" filterable style="width:300px" @on-change="selectMetrc">
                        <Option v-for="metric in metricOptions" :value="metric.metric" :key="metric.metric">{{ metric.metric }}</Option>
                      </Select>
                    </FormItem>
                  </Form>
                  <div>
                    <Tag
                      v-for="(graphMetric, index) in graphConfig.graphContainsMetric"
                      :key="graphMetric"
                      type="border"
                      @on-close="removeGraphMetric(index)"
                      closable color="primary">
                      {{graphMetric}}
                    </Tag>
                  </div>
                  <div v-if="isRequestChartData" class="metric-section" style="margin-top:24px">
                    <div>
                      <div :id="displayGroupElId" class="echart"></div>
                    </div>
                  </div>
                  <div style="text-align: right;margin-top:24px">
                    <button class="btn btn-sm btn-cancel-f" @click="preview('displayGroup')">{{$t('m_preview')}}</button>
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
            :title="$t('m_select_host')">
            <Form :label-width="80">
              <FormItem :label="$t('m_host')">
                <Select filterable v-model="metricConfigData.endpoint" style="width:300px">
                  <Option v-for="item in endpointOptions" :value="item.guid" :key="item.guid">{{ item.guid }}</Option>
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

      selectdPanel: '',
      panelOptions: [],
      selectdGraph: '',
      graphOptions: [],
      graphConfig: {
        graphName: '',
        unit: '',
        metric: '',
        graphContainsMetric: []
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
    changePanel () {
      this.showGraphConfig =false
      this.selectdGraph = ''
      this.changeGraph()
    },
    changeGraph () {
      this.showGraphConfig =false
      this.graphConfig.graphName = ''
      this.graphConfig.unit = ''
      this.graphConfig.graphContainsMetric = []
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
    },
    editPanel (type, item) {
      this.titleManagement.show = true
      this.titleManagement.isAdd = false
      this.titleManagement.type = type
      this.titleManagement.title = item.title
      this.titleManagement.id = item.id
    },
    deletePanel (type, item) {
      this.$delConfirm({
        msg: item.title,
        callback: () => {
          this.removePanel(item.id)
        }
      })
    },
    removePanel (id) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', this.$root.apiCenter.addPanel + '?ids=' + id, '', () => {
        this.$Message.success(this.$t('tips.success'))
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
    removeGraphMetric (index) {
      this.graphConfig.graphContainsMetric.splice(index, 1)
    },
    selectMetrc (val) {
      const find = this.graphConfig.graphContainsMetric.find(item => item === val)
      if (!find) {
        this.graphConfig.graphContainsMetric.push(val)
      }
    },
    saveGraphMetric () {
       const graph = this.graphOptions.find(item => item.metric === this.selectdGraph)
      const params = {
        metric: this.graphConfig.graphContainsMetric.join('^'),
        title: this.graphConfig.graphName,
        unit: this.graphConfig.unit,
        group_id: this.selectdPanel,
        id: graph.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', this.$root.apiCenter.getGraph, [params], () => {
        this.$Message.success(this.$t('tips.success'))
      })
    },
    getGraphConfig () {
      this.showGraphConfig = true
      this.graphConfig.graphName = ''
      this.graphConfig.unit = ''
      const graph = this.graphOptions.find(item => item.metric === this.selectdGraph)
      if (graph) {
        this.graphConfig.graphName = graph.title
        this.graphConfig.unit = graph.unit
        this.graphConfig.graphContainsMetric = graph.metric.split('^').filter(item => item !== '')
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
      const params = [{
        endpoint: this.metricConfigData.endpoint,
        prom_ql: this.metricConfigData.prom_ql,
        metric: this.metricConfigData.metric,
        time: '-1800'
      }]
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {eye: false,clear:true, zoomCallback: true}
        readyToDraw(this,responseData, 1, chartConfig, this.acquisitionConfigurationElId)
      })
    },
    async getDispalyChartData () {
      generateUuid().then((elId)=>{
        this.displayGroupElId =  `id_${elId}`
      })
      let params = []
      this.graphConfig.graphContainsMetric.forEach(metric => {
        params.push({
          endpoint: this.metricConfigData.endpoint,
          prom_ql: '',
          metric: metric,
          time: '-1800'
        })
      })
      this.isRequestChartData = true
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {eye: false,clear:true, zoomCallback: true}
        readyToDraw(this,responseData, 1, chartConfig, this.displayGroupElId)
      })
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
        this.showEndpointSelect = true
      })
    },
    preview (previewPosition) {
      this.previewPosition = previewPosition
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
  ivu-form-item {
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
    width: 1000px;
    // margin: 0 auto;
  }
  .echart {
    height: 400px;
    width: 1000px;
    background: #f5f7f9;
  }
  .echart-no-data-tip {
    text-align: center;
    vertical-align: middle;
    display: table-cell;
  }
</style>