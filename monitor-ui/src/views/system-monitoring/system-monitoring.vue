<template>
  <div class=" ">
    <header>
      <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
        <div class="header-name">
          <span>{{$t('m_tableKey_systemName')}}:</span>
          <span> {{sysConfig.systemName}}</span>
        </div>
        <div class="header-tools">
          <Select filterable v-model="sysConfig.metricMulti" multiple style="width:200px" @on-change="getMetric">
            <Option v-for="item in metricLabelList" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
      </div>
    </header>
    <grid-layout
      :layout.sync="layoutData"
      :col-num="12"
      :row-height="30"
      :is-draggable="false"
      :is-resizable="false"
      :is-mirrored="false"
      :vertical-compact="true"
      :use-css-transforms="true"
    >
      <grid-item v-for="(item,index) in layoutData"
                 class="c-dark"
                 :x="item.x"
                 :y="item.y"
                 :w="item.w"
                 :h="item.h"
                 :i="item.i"
                 :key="index"
                 @resize="resizeEvent"
                 @resized="resizeEvent"
      >

        <div class="c-dark" style="display:flex;justify-content:flex-end;padding:0 32px;">
          <div class="header-grid header-grid-name">
            <span>{{item.i}}</span>
          </div>
          <div class="header-grid header-grid-tools">
            <Tooltip :content="$t('m_placeholder_viewChart')" placement="top">
              <i class="fa fa-eye" aria-hidden="true" @click="gridPlus(item)"></i>
            </Tooltip>
          </div>
        </div>
        <div class="">
          <section class="metric-section">
            <div v-if="!noDataTip">
              <div :id="item.id" class="echart"></div>
            </div>
            <div v-else class="echart echart-no-data-tip">
              <span>~~~No Data!~~~</span>
            </div>
          </section>
        </div>
      </grid-item>
    </grid-layout>
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
import VueGridLayout from 'vue-grid-layout'
import ChartLinesModal from '@/components/chart-lines-modal'
import {resizeEvent} from '@/assets/js/gridUtils.ts'
import {readyToDraw} from '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      viewData: [],
      layoutData: [
      ],
      noDataTip: false,
      sysConfig: {
        systemName: 'test',
        ips: ['192.168.0.16','192.168.0.5'],
        // metricMulti:['cpu.used.percent','mem.used.percent','load.1min'],
        metricMulti: ['cpu.used.percent'],
        endpointList: [],
      },
      metricLabelList: [
        {
          value: 'cpu.used.percent',
          label: 'cpu.used.percent'
        },
        {
          value: 'mem.used.percent',
          label: 'mem.used.percent'
        },
        {
          value: 'load.1min',
          label: 'load.1min'
        },
        {
          value: 'disk.read.bytes',
          label: 'disk.read.bytes'
        },
        {
          value: 'disk.write.bytes',
          label: 'disk.write.bytes'
        },
        {
          value: 'disk.iops',
          label: 'disk.iops'
        },
        {
          value: 'net.if.out.bytes',
          label: 'net.if.out.bytes'
        }
      ],
      array1: [],
      isLineSelectModalShow: false,
      setChartConfigId: '',
      chartInstance: null
    }
  },
  mounted() {
    // systemMonitoring?systemName=test&ips=192.168.0.16,192.168.0.5
    this.sysConfig.systemName = this.$route.query.systemName
    this.sysConfig.ips = this.$route.query.ips.split(',')
    window['view-config-selected-line-data'] = {}
    this.$on('editShowLines', this.handleEditShowLines)
    this.getMetric()
  },
  methods: {
    getMetric() {
      let url = '/monitor/api/v1//dashboard/custom/endpoint/get?'
      const ipManage = this.sysConfig.ips.map(item => 'ip=' + item)
      url += ipManage.join('&')
      this.$httpRequestEntrance.httpRequestEntrance('GET',url, '', responseData => {
        this.sysConfig.endpointList = []
        responseData.forEach(i => {
          this.sysConfig.endpointList.push(i.guid)
        })
        this.initData()
      })
    },
    initData() {
      this.viewData = []
      this.layoutData = []
      const res = []
      const num = this.sysConfig.metricMulti.length
      this.array1 = []
      for (let i=0; i<num; i++) {
        const key = ((new Date()).valueOf())
          .toString()
          .substring(10)
        const xx = {
          x: i%2*6,
          y: Math.ceil((i/2+0.000001)-1)*7,
          w: 6,
          h: 7,
          i: '',
          id: `id_${key}${i}`,
          moved: false
        }
        this.array1.push(xx)
      }
      this.sysConfig.metricMulti.forEach((metric, index) => {
        const singleChart = {}
        singleChart.panalTitle = metric
        singleChart.query = []
        for (const endpoint of this.sysConfig.endpointList) {
          const condition = {}
          condition.endpoint = endpoint
          condition.metricLabel = metric
          singleChart.query.push(condition)
        }
        this.array1[index].i = metric
        singleChart.viewConfig = this.array1[index]
        res.push(singleChart)
      })
      this.viewData = res
      this.initPanals()
    },
    initPanals() {
      this.viewData.forEach((item,viewIndex) => {
        this.layoutData.push(item.viewConfig)
        this.requestChart(item.viewConfig.id, item.query,viewIndex)
      })
    },
    requestChart(id, query,viewIndex) {
      const params = []
      query.forEach(item => {
        params.push({
          endpoint: item.endpoint,
          metric: item.metricLabel,
          prom_ql: item.metric,
          time: '-1800'
        })
      })
      this.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        this.elId = id
        responseData.chartId = id
        const chartConfig = {
          eye: false,
          dataZoom: false,
          clear: true,
          lineBarSwitch: true,
          chartId: id,
          canEditShowLines: true
        }
        this.chartInstance = readyToDraw(this,responseData, viewIndex, chartConfig)
        if (this.chartInstance) {
          this.chartInstance.on('legendselectchanged', params => {
            window['view-config-selected-line-data'][id] = cloneDeep(params.selected)
          })
        }
      })
    },
    resizeEvent(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
    },
    gridPlus(item) {
      this.viewData.forEach(vd => {
        if (item.id === vd.viewConfig.id) {
          this.$router.push({
            name: 'sysViewChart',
            params: {
              templateData: vd,
              parentData: this.sysConfig
            }
          })
        }
      })
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
      this.getChartData()
    },
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
    ChartLinesModal
  },
}
</script>

<style scoped lang="less">
  .header-grid {
    flex-grow: 1;
    text-align: end;
    line-height: 32px;
    i {
      margin: 0 4px;
      cursor: pointer;
    }
  }
</style>
<style scoped lang="less">
.vue-grid-item {
  border-radius: 4px;
}
.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}
.echart-no-data-tip {
  text-align: center;
}
</style>
