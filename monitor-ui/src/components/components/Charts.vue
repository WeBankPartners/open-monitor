<template>
  <div class="charts charts-click">
    <Tabs :value="activeTab" @on-click="changeTab"> 
      <template v-for="(chartItem, chartIndex) in charts.chartsConfig">
        <TabPane :label="chartItem.tabTape.label" :name="chartItem.tabTape.name" :key="chartIndex">
        </TabPane>
      </template>  
    </Tabs>
    <!-- <section v-if="showViewMetricConfig">
      <MetricConfig></MetricConfig>  
    </section> -->
    <section>
      <template v-if="btns.length">
        <div class="btn-content">
          <RadioGroup v-model="activeBtn" size="small" type="button">
            <template v-for="(btnItem,btnIndex) in btns">
              <Radio :label="btnItem.option_value" :key="btnIndex">{{btnItem.option_text}}</Radio>
            </template>
          </RadioGroup>
        </div>
      </template>
      <template v-for="(chartItemx,chartIndexx) in activeCharts">
          <SingleChart @sendConfig="receiveConfig" :chartItemx="chartItemx" :key="chartIndexx" :params="params"> </SingleChart>
      </template>
    </section>
    
    <transition name="slide-fade">
      <div v-show="showMaxChart">
        <MaxChart ref="maxChart"></MaxChart>
      </div>
    </transition>
  </div>
</template>

<script>
// import MetricConfig from '@/components/pages/metric-config'
import SingleChart from '@/components/components/Single-chart'
import MaxChart from '@/components/components/Max-chart'
export default {
  name: '',
  data() {
    return {
      activeTab:  '',
      activeCharts: {},
      btns: [],
      tagsUrl: '',
      params: {},
      showMaxChart: false,
      activeBtn: '',
      // showViewMetricConfig: false
    }
  },
  props: {
    charts: Object
  },
  watch: {
    activeBtn: function () {
      this.pitchOnBtn()
    }
  },
  mounted () {
    if (this.charts.chartsConfig.length !== 0) {
      this.activeCharts = this.charts.chartsConfig[0].charts
    }
  },
  methods: {
    refreshCharts (activeTab) {
      if (this.activeTab === '') {
        this.activeTab = activeTab
      }
      this.changeTab(this.activeTab)
    },
    changeTab (name) {
      this.params = this.charts.chartsConfig[0].params
      this.activeTab = name
      this.activeCharts = []
      this.btns = []
      this.charts.chartsConfig.forEach((item) => {
        if (item.tabTape.name === name) {
          this.btns = item.btns
          if (this.btns.length !== 0) {
            this.activeBtn = this.btns[0].option_value
          }
          this.tagsUrl = item.tagsUrl     
          this.$nextTick(() => {
            this.activeCharts = item.charts
          })
        }
      })
    },
    pitchOnBtn() {
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.tagsUrl +  this.activeBtn, '', responseData => {
        this.activeCharts.forEach((element,index) => {
           element.metric = responseData[index].metric
        })
        this.activeCharts = []
        this.charts.chartsConfig.forEach((item) => {
          if (item.tabTape.name === this.activeTab) {    
            this.$nextTick(() => {
              this.activeCharts = item.charts
            })
          }
        })

      })
    },
    hiddenDetailChart () {
      // this.showMaxChart = false
    },
    receiveConfig (chartItem) {
      this.showMaxChart = true
      this.$refs.maxChart.getChartConfig(chartItem)
      return
    }
  },
  components: {
    SingleChart,
    MaxChart
  }
}
</script>
<style scoped lang="less">
  .charts {
    padding-top: 20px;
  }

  .btn-content {
  padding: 2px;
  }

  /* 可以设置不同的进入和离开动画 */
/* 设置持续时间和动画函数 */
.slide-fade-enter-active {
  transition: all .3s ease;
}
.slide-fade-leave-active {
  transition: all .3s ease;
}
</style>
