<template>
  <div class="page" id="mainView">
    <Title :title="$t('menu.endpointView')"></Title>
    <Search ref="search" />
    <button type="button" v-if="isShow" @click="changeRoute" class="btn btn-sm btn-cancle-f btn-jump">{{$t('button.endpointManagement')}}</button>
    <Charts :charts='charts' ref="parentCharts" />
  </div>
</template>
<script>
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
  computed: {
    isShow: function () {
      if (this.$validate.isEmpty_reset(this.$store.state.ip)) {
        return false 
      } else {
        return true
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
      this.$refs.parentCharts.refreshCharts()
    },
    changeRoute () {
      this.$router.push({name: 'endpointManagement', params: {search: this.$store.state.ip.value.split(':')[0]}})
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
