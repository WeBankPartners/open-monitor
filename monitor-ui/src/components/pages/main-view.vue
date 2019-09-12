<template>
  <div class="page" id="mainView">
    <Title title="监控视图"></Title>
    <Search ref="search" />
    <Charts :charts='charts' ref="child1" />
  </div>
</template>

<script>
import Title from '@/components/components/Title'
import Search from '@/components/components/Search'
import Charts from '@/components/components/Charts'
export default {
  name: 'main-view',
  data() {
    return {
      charts: {
        chartsConfig: []
      }
    }
  },
  mounted() {
    let DOMX = document.getElementById("mainView")
    DOMX.addEventListener('click', ()=>{
      this.$refs.child1.hiddenDetailChart()
    })
    this.$refs.search.getChartsConfig()
  },
  methods: {
    manageCharts (chartsConfig, params) {
      this.charts.chartsConfig = []
      chartsConfig.forEach(item => {
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
      this.$refs.child1.refreshCharts(chartsConfig[0].title + '_')
    }
  },
  components: {
    Title,
    Search,
    Charts
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
