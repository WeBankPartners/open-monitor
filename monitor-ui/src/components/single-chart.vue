<template>
  <div class="single-chart" @mouseleave="onMouseLeaveContent">
    <div v-if="!noDataTip" :id="elId" class="echart">
    </div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
      {{chartTitle}}:
      <span>~~~No Data!~~~</span>
    </div>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'

// 引入 ECharts 主模块
import {cloneDeep} from 'lodash'
import {readyToDraw} from '@/assets/config/chart-rely'
// const echarts = require('echarts/lib/echarts');

export default {
  name: '',
  data() {
    return {
      elId: null,
      chartTitle: null,
      noDataTip: false,
      config: '',
      myChart: '',
      interval: '',
      chartInstance: null,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  props: {
    chartInfo: Object,
    params: Object,
    chartIndex: Number,
    refreshNow: {
      type: Boolean,
      default: null
    }
  },
  created(){
    // 外部触发清除刷新
    this.$root.$eventBus.$on('clearSingleChartInterval', () => {
      clearInterval(this.interval)
    })
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  watch: {
    params() {
      this.getChartData()
      if (this.params.autoRefresh > 0) {
        this.interval = setInterval(() => {
          this.getChartData()
        },this.params.autoRefresh*1000)
      }
    },
    refreshNow: {
      handler() {
        this.getChartData()
      }
    },
  },
  mounted() {
    this.getChartData()
    if (this.params.autoRefresh > 0) {
      this.interval = setInterval(() => {
        this.getChartData()
      },this.params.autoRefresh*1000)
    }
  },
  destroyed() {
    clearInterval(this.interval)
    if (this.chartInstance) {
      setTimeout(() => {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
      }, 500)
    }
  },
  methods: {
    getChartData(tmp, start, end) {
      const params = {
        aggregate: 'none',
        time_second: this.params.time,
        start: start ? 0 : this.params.start || 0,
        end: end ? 0 : this.params.end || 0,
        title: this.chartInfo.title,
        unit: '',
        chart_id: this.chartInfo.id,
        compare: {
          compare_first_start: this.params.compare_first_start,
          compare_first_end: this.params.compare_first_end,
          compare_second_start: this.params.compare_second_start,
          compare_second_end: this.params.compare_second_end
        },
        data: []
      }
      if (this.params.sys) {
        this.chartInfo.endpoint.forEach(ep => {
          this.chartInfo.metric.forEach(me => {
            params.data.push({
              endpoint: ep,
              metric: me
            })
          })
        })
      } else {
        params.compare = {
          compare_first_start: this.params.compare_first_start,
          compare_first_end: this.params.compare_first_end,
          compare_second_start: this.params.compare_second_start,
          compare_second_end: this.params.compare_second_end
        }
        params.data.push({
          endpoint: this.chartInfo.endpoint[0],
          metric: this.chartInfo.metric[0],
          compare_first_start: this.params.compare_first_start,
          compare_first_end: this.params.compare_first_end,
          compare_second_start: this.params.compare_second_start,
          compare_second_end: this.params.compare_second_end,
          title: this.chartInfo.title
        })
      }
      window.intervalFrom = 'single-chart'
      this.request('POST', this.apiCenter.metricConfigView.api, params, responseData => {
        const chartConfig = {
          clear: false,
          editTitle: false,
          lineBarSwitch: true,
          canEditShowLines: true,
          dataZoom: true,
          chartId: this.chartInfo.id + ''
        }
        responseData.metric = this.chartInfo.metric[0]
        this.chartInstance = readyToDraw(this,responseData, this.chartIndex, chartConfig)
        this.chartInstance.on('legendselectchanged', params => {
          window['view-config-selected-line-data'][chartConfig.chartId] = cloneDeep(params.selected)
        })
      }, { isNeedloading: false })
    },
    onMouseLeaveContent() {
      if (this.chartInstance) {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
      }
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  .single-chart {
    display: inline-block;
    padding: 5px;
    .echart {
       height: 300px;
       width: 580px;
       border-radius: 4px;
      //  background: @gray-f;
    }
    .echart-no-data-tip {
      text-align: center;
      vertical-align: middle;
      display: table-cell;
    }
  }
</style>
