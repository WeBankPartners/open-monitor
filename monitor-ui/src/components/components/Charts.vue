<template>
  <div class="charts charts-click">
    <Tabs :value="activeTab" @on-click="changeTab"> 
      <template v-for="(chartItem, chartIndex) in charts.chartsConfig">
        <TabPane :label="chartItem.tabTape.label" :name="chartItem.tabTape.name" :key="chartIndex">
          <template v-if="btns.length">
            <div class="btn-content">
              <template v-for="(btnItem,btnIndex) in btns">
                <div class="btn-block" :class="{btnActive: btnItem.isActive}" :key='btnIndex' @click="pitchOnBtn(btnItem,btnIndex)">
                  <span>{{btnItem.option_text}}</span>
                </div>
              </template>
            </div>
          </template>
          <template v-for="(chartItemx,chartIndexx) in activeCharts">
              <SingleChart @sendConfig="receiveConfig" :chartItemx="chartItemx" :key="chartIndexx" :params="params"> </SingleChart>
          </template>
        </TabPane>
      </template>  
    </Tabs>
    <transition name="slide-fade">
      <div v-show="showMaxChart">
        <MaxChart ref="maxChart"></MaxChart>
      </div>
    </transition>
  </div>
</template>

<script>
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
      xx: false,
      showMaxChart: false,
    }
  },
  props: {
    charts: Object
  },
  mounted () {
    if (this.charts.chartsConfig.length !== 0) {
      this.activeCharts = this.charts.chartsConfig[0].charts
    }
  },
  methods: {
    refreshCharts (activeTab=this.activeTab) {
      this.changeTab(activeTab)
    },
    changeTab (name) {
      this.params = this.charts.chartsConfig[0].params
      this.activeTab = name
      this.activeCharts = []
      this.btns = []
      this.charts.chartsConfig.forEach((item) => {
        if (item.tabTape.name === name) {
          this.btns = item.btns
          this.tagsUrl = item.tagsUrl
          this.btns.forEach(element => {
            element.isActive = false
          });      
          this.$nextTick(() => {
            this.activeCharts = item.charts
          })
        }
      })
    },
    pitchOnBtn(btnItem,btnIndex) {
      this.btns.forEach(element => {
        element.isActive = false
      })
      btnItem.isActive = true
      this.params.tagParam = btnItem.option_value
      this.$set(this.btns,btnIndex,btnItem)

      this.$httpRequestEntrance.httpRequestEntrance('GET',this.tagsUrl +  btnItem.option_value, '', responseData => {
        this.activeCharts.forEach((element,index) => {
           element.metric = responseData[index].metric
        })
        this.refreshCharts()
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

  .btn-block {
    margin-left: -1px;
    margin-bottom: 2px;
    line-height: 30px;
    min-width: 50px;
    padding: 0 4px;
    text-align: center;
    display: inline-block;
    background: white;
    border: 1px solid @blue-2;
    color: @blue-2;
    cursor: pointer;
  }

  .btn-content {
  padding: 2px;
  }
  .btnActive {
    background: @gray-f;
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
