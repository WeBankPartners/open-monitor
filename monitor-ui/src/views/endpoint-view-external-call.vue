<template>
  <div class="page" id="endpointView">
    <Title :title="$t('m_menu_endpointView')"></Title>
    <SearchExternal ref="search" />
    <Charts v-if="showCharts" :charts='charts' ref="parentCharts" />
    <div v-if="recursiveViewConfig.length && showRecursive">
      <recursive :recursiveViewConfig="recursiveViewConfig" :params="params"></recursive>
    </div>
  </div>
</template>
<script>
import SearchExternal from '@/components/search-external'
import Charts from '@/components/charts'
import recursive from '@/views/recursive-view/recursive'
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
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  mounted() {
    // this.$refs.search.getChartsConfig()
  },
  methods: {
    manageCharts(chartsConfig, params) {
      if (params.sys) {
        this.params = params
        this.showCharts = false
        this.recursiveView(params)
        return
      }
      this.charts.chartsConfig = []
      chartsConfig.forEach(item => {
        item.autoRefresh = params.autoRefresh
        const chart = {
          tabTape: {
            label: item.title,
            name: item.title + '_',
          },
          btns: item.tags.option,
          tagsUrl: item.tags.url,
          charts: item.charts,
          params
        }
        this.charts.chartsConfig.push(chart)
      })
      this.showCharts = true
      this.showRecursive = false
      this.$refs.parentCharts.refreshCharts()
    },
    recursiveView(params) {
      this.recursiveViewConfig = []
      this.request('GET',this.apiCenter.recursive.api, params, responseData => {
        this.showRecursive = true
        this.recursiveViewConfig = [responseData]
      })
    }
  },
  components: {
    SearchExternal,
    Charts,
    recursive
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.page {
  margin: 16px;
}
</style>
