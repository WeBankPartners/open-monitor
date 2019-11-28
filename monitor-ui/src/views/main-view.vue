<template>
  <div class="page" id="mainView">
    <Title :title="$t('menu.endpointView')"></Title>
    <Search ref="search" />
    <Charts v-if="showCharts" :charts='charts' ref="parentCharts" />
  </div>
</template>
<script>
import Search from '@/components/components/search'
import Charts from '@/components/components/charts'
export default {
  name: 'main-view',
  data() {
    return {
      showCharts: false,
      charts: {
        chartsConfig: []
      }
    }
  },
  mounted() {
    this.$refs.search.getChartsConfig()
  },
  methods: {
    manageCharts (chartsConfig, params) {
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
      this.$refs.parentCharts.refreshCharts()
    }
  },
  components: {
    Search,
    Charts
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn-jump {
  margin-left: 10px;
}
</style>
