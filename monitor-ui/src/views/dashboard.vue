<template>
  <div>
    <section v-if="isPlugin">
      <div style="margin: 100px 0;text-align: center;font-size:14px">
        {{$t('m_tips_dashboardEmpty')}}
      </div>
    </section>
    <section v-else>
      <template v-for="(dash, dashIndex) in dataHome">
        <Card style="width: 100%" :key="dashIndex">
          <p slot="title">
            {{dash.name}}
          </p>
          <view-config
            v-if="dash.id"
            permissionType='view'
            :boardId="dash.id"
            pageType="dashboard"
          />
        </Card>
      </template>
    </section>
  </div>
</template>

<script>
import ViewConfig from '@/views/custom-view/view-config'
export default {
  name: '',
  data() {
    return {
      isPlugin: false,
      viewCondition: {
        timeTnterval: -1800,
        dateRange: ['', ''],
        autoRefresh: 10,
      },
      disableTime: false,
      viewData: [],
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ],

      showAlarm: true, // 显示告警信息
      cutsomViewId: null,

      dataHome: []
    }
  },
  mounted() {
    this.getDashboardData()
  },
  methods: {
    datePick(data) {
      this.viewCondition.dateRange = data
      this.disableTime = false
      if (this.viewCondition.dateRange[0] && this.viewCondition.dateRange[1]) {
        if (this.viewCondition.dateRange[0] === this.viewCondition.dateRange[1]) {
          this.viewCondition.dateRange[1] = this.viewCondition.dateRange[1].replace('00:00:00', '23:59:59')
        }
        this.disableTime = true
        this.viewCondition.autoRefresh = 0
      }
      this.initPanals()
    },
    getDashboardData() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.get, '', responseData => {
        this.dataHome = responseData
        if (this.dataHome.length === 0) {
          if (window.request) {
            this.isPlugin = true
          } else {
            this.$router.push({path: 'portal'})
          }
        }
      })
    },
    initPanals() {
      const tmp = []
      this.viewData.forEach(item => {
        const params = {
          aggregate: item.aggregate || 'none',
          agg_step: item.agg_step || 60,
          time_second: this.viewCondition.timeTnterval,
          start: 0,
          end: 0,
          title: '',
          unit: '',
          data: []
        }
        item.query.forEach(_ => {
          params.data.push(_)
        })
        const height = (item.viewConfig.h) * 30
        const _activeCharts = []
        _activeCharts.push({
          style: `height:${height}px;`,
          panalUnit: item.panalUnit,
          elId: item.viewConfig.id,
          chartParams: params,
          chartType: item.chartType
        })
        item.viewConfig._activeCharts = _activeCharts
        tmp.push(item.viewConfig)
      })
      this.layoutData = tmp
    }
  },
  components: {
    ViewConfig
  },
}
</script>

<style scoped lang="less">
.grid-style {
  width: 100%;
  display: inline-block;
}
.alarm-style {
  width: 800px;
  display: inline-block;
}

header {
  margin: 16px 8px;
}
.search-zone {
  display: inline-block;
}
.params-title {
  margin-left: 4px;
  font-size: 13px;
}

.header-grid {
  flex-grow: 1;
  text-align: center;
  line-height: 32px;
  i {
    margin: 0 4px;
    cursor: pointer;
  }
}
.vue-grid-item {
  border-radius: 4px;
}
.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}
</style>
