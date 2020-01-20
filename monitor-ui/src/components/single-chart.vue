<template>
  <div class="single-chart">
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
import {readyToDraw} from  '@/assets/config/chart-rely'
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
      interval: ''
    }
  },
  props: {
    chartItemx: Object,
    params: Object,
    chartIndex: Number
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
    })
  },
  watch: {
    params: function () {
      this.getchartdata()
      if (this.params.autoRefresh > 0) {
        this.interval = setInterval(()=>{
          this.getchartdata()
        },this.params.autoRefresh*1000)
      }
    }
  },
  mounted() {
    this.getchartdata()
    if (this.params.autoRefresh > 0) {
      this.interval = setInterval(()=>{
        this.getchartdata()
      },this.params.autoRefresh*1000)
    }
  },
  destroyed() {
    clearInterval(this.interval)
  },
  methods: {
    getchartdata () {
      let params = []
      if (this.params.sys) {
        this.params.endpoint.forEach((ep) => {
          this.params.metric.forEach((me) => {
            params.push({
              id: this.chartItemx.id,
              endpoint: ep,
              metric: me,
              time: this.params.time.toString(),
              start: this.params.start + '',
              end: this.params.end + ''
            })
          })
        })
      } else {
        params.push({
          id: this.chartItemx.id,
          endpoint: this.params.endpoint,
          metric: this.chartItemx.metric[0],
          time: this.params.time.toString(),
          start: this.params.start + '',
          end: this.params.end + ''
        })
      }
      this.$httpRequestEntrance.httpRequestEntrance('POST', 'dashboard/newchart', params, responseData => {
        readyToDraw(this,responseData, this.chartIndex)
      })
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
