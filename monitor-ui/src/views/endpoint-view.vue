<template>
  <div class="page" id="endpointView">
    <Title :title="$t('menu.endpointView')"></Title>
    <Search ref="search" />
    <Charts v-if="showCharts" :charts='charts' ref="parentCharts" />
    <div v-if="recursiveViewConfig.length && showRecursive">
      <recursive :recursiveViewConfig="recursiveViewConfig" :params="params"></recursive>
    </div>
    <transition name="slide-fade">
      <div v-show="showMaxChart">
        <MaxChart ref="maxChart"></MaxChart>
      </div>
    </transition>
  </div>
</template>
<script>
import { EventBus } from "@/assets/js/event-bus.js"
import Search from '@/components/search'
import Charts from '@/components/charts'
import recursive from '@/views/recursive-view/recursive'
import MaxChart from '@/components/max-chart'
export default {
  name: 'endpoint-view',
  data() {
    return {
      showCharts: false,
      showRecursive: false,
      params: null,
      charts: {
        chartsConfig: []
      },
      recursiveViewConfig: [],
      showMaxChart: false
    }
  },
  created () {
    EventBus.$on("aaa", data => {
      console.log(111)
      console.log(data)
      this.hello(data)
    })
  },
  mounted() {
    this.$refs.search.getChartsConfig() 
  },
  methods: {
    manageCharts (chartsConfig, params) {
      if (params.sys) {
        this.params = params
        this.showCharts = false
        this.recursiveView(params)
        return
      }
      this.charts.chartsConfig = []
      chartsConfig.forEach(item => {
        item.autoRefresh = params.autoRefresh
        let chart = {
          tabTape: {
            label: item.title,
            name: item.title + '_',
          },
          btns: item.tags.option,
          tagsUrl: item.tags.url,
          charts: item.charts,
          params: params
        }
        this.charts.chartsConfig.push(chart)
      })
      this.showCharts = true
      this.showRecursive = false
      this.$refs.parentCharts.refreshCharts()
    },
    recursiveView (params) {
      this.recursiveViewConfig = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.recursive.api, params, responseData => {
        this.showRecursive = true
        this.recursiveViewConfig = [responseData]
      })
    },
    hello (data) {
      this.showMaxChart = true
      this.$refs.maxChart.getChartConfig(data)
      return
    }
      
  },
  components: {
    Search,
    Charts,
    recursive,
    MaxChart
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn-jump {
  margin-left: 10px;
}
</style>
