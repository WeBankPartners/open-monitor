<template>
  <div class>
    <header>
      <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
        <div class="header-name">
        </div>
        <div class="header-tools">
          <Button type="primary" @click="goBack()">{{$t('m_button_back')}}</Button>
        </div>
      </div>
    </header>
    <div class="zone zone-chart c-dark">
      <div class="col-md-12">
        <div class="zone-chart-title">{{panalTitle}}</div>
        <div v-if="!noDataTip">
          <div :id="elId" class="echart"  style="height:80vh"></div>
        </div>
        <div v-else class="echart echart-no-data-tip">
          <span>~~~No Data!~~~</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { generateUuid } from '@/assets/js/utils'
import { readyToDraw } from '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      elId: null,
      noDataTip: false,
      panalTitle: '',
      panalUnit: ''
    }
  },
  created() {
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  mounted() {
    if (this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.$router.push({ path: 'systemMonitoring' })
    } else {
      if (!this.$root.$validate.isEmpty_reset(this.$route.params.templateData)) {
        this.initPanal()
      }
    }
  },
  methods: {
    initPanal() {
      this.panalTitle = this.$route.params.templateData.panalTitle
      const params = []
      this.noDataTip = false
      this.$route.params.templateData.query.forEach(item => {
        params.push(
          {
            endpoint: item.endpoint,
            metric: item.metricLabel,
            time: '-1800'
          }
        )
      })
      if (params !== []) {
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',
          this.$root.apiCenter.metricConfigView.api,
          params,
          responseData => {

            responseData.yaxis.unit = this.panalUnit
            const chartConfig = {
              eye: false,
              clear: true,
              lineBarSwitch: true
            }
            readyToDraw(this,responseData, 1, chartConfig)
          }
        )
      }
    },
    goBack() {
      this.$router.push({
        name: 'systemMonitoring',
        params: this.$route.params.parentData
      })
    }
  },
  components: {}
}
</script>

<style scoped lang="less">
.zone {
  margin: 0 auto;
  background: @gray-f;
  border-radius: 4px;
}
.zone-chart-title {
  padding: 20px 40%;
  font-size: 14px;
}

.echart-no-data-tip {
  text-align: center;
  vertical-align: middle;
  display: table-cell;
}
</style>
