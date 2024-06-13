<template>
  <div class="width: 700px">
    <section style="margin: 10px 2px 2px" class="c-dark-exclude-color">
      <template v-for="(filterItem, filterIndex) in filtersForShow">
        <Tag color="success" type="border" closable @on-close="clearFiltersForShow" :key="filterIndex">{{filterItem.key}}：{{filterItem.value}}</Tag>
      </template>
    </section>
    <div class="alarm-total">
      <Button type="success" @click="addParams('low')" size="small"><span style="font-size:14px">{{$t('m_low')}}:{{this.low}}</span></Button>
      <Button type="warning" @click="addParams('medium')" size="small"><span style="font-size:14px">{{$t('m_medium')}}:{{this.mid}}</span></Button>
      <Button type="error" @click="addParams('high')" size="small"><span style="font-size:14px">{{$t('m_high')}}:{{this.high}}</span></Button>
    </div>
    <div class="alarm-list">
      <section class="alarm-card-container">
        <alarm-card v-for="(item, alarmIndex) in resultData" @openRemarkModal="remarkModal" :key="alarmIndex" :data="item" :button="true" :hideFilter="true"></alarm-card>
      </section>
      <div style="margin: 4px 0; text-align:right">
        <Page :total="paginationInfo.total" @on-change="pageIndexChange" @on-page-size-change="pageSizeChange" show-elevator show-sizer show-total />
      </div>
    </div>
    <Modal
      v-model="isShowWarning"
      :title="$t('closeConfirm.title')"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('closeConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import AlarmCard from "@/components/alarm-card.vue"
export default {
  name: '',
  components: {
    AlarmCard
  },
  data() {
    return {
      cutsomViewId: null,
      interval: '',

      resultData: [],
      selectedData: '', // 存放选中数据
      isShowWarning: false,
      low: 0,
      mid: 0,
      high: 0,
      modelConfig: {
        modalId: 'remark_Modal',
        modalTitle: 'm_remark',
        saveFunc: 'remarkAlarm',
        isAdd: true,
        config: [
          {label: 'm_remark', value: 'message', placeholder: '', v_validate: '', disabled: false, type: 'textarea'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          id: '',
          message: '',
          is_custom: false
        }
      },
      filtersForShow: [], // 缓存级别过滤
      cacheParams: {
        id: '',
        viewCondition: ''
      },
      paginationInfo: {
        total: 0,
        startIndex: 1,
        pageSize: 10
      }
    }
  },
  mounted () {
    window.addEventListener("visibilitychange", this.isTabActive, true)
  },
  destroyed() {
    this.clearAlarmInterval()
    window.removeEventListener("visibilitychange", this.isTabActive, true)
  },
  methods: {
    isTabActive () {
       if (document.hidden) {
        this.clearAlarmInterval()
      } else {
        if (this.cacheParams.id) {
          this.getAlarm(this.cacheParams.id, this.cacheParams.viewCondition)
        }
      }
    },
    clearAlarmInterval () {
      clearInterval(this.interval)
    },
    getAlarm (id, viewCondition, permission) {
      if (!String(id).length) return
      this.permission = permission
      this.cacheParams.id = id
      this.cacheParams.viewCondition = viewCondition
      this.getAlarmdata(id)
      this.interval = setInterval(()=>{
        this.getAlarmdata(id)
      }, (viewCondition.autoRefresh || 10) * 1000)

    },
    getAlarmdata (id) {
      let params = {
        customDashboardId: id,
        page: this.paginationInfo,
        priority: this.filtersForShow.length === 1 ? this.filtersForShow[0].value : undefined
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/problem/page', params, (responseData) => {
        this.paginationInfo.total = responseData.page.totalRows
        this.paginationInfo.startIndex = responseData.page.startIndex
        this.paginationInfo.pageSize = responseData.page.pageSize
        this.resultData = responseData.data
        this.low = responseData.low
        this.mid = responseData.mid
        this.high = responseData.high
      }, {isNeedloading: false})
    },
    addParams (type) {
      this.filtersForShow = [{
        key: 'priority',
        value: type
      }]
      this.getAlarmdata(this.cacheParams.id)
    },
    clearFiltersForShow () {
      this.filtersForShow = []
      this.getAlarmdata(this.cacheParams.id)
    },
    pageIndexChange(pageIndex) {
      this.paginationInfo.startIndex = pageIndex
      this.getAlarmdata(this.cacheParams.id)
    },
    pageSizeChange(pageSize) {
      this.paginationInfo.startIndex = 1
      this.paginationInfo.pageSize = pageSize
      this.getAlarmdata(this.cacheParams.id)
    },
    goToEndpointView (alarmItem) {
      const endpointObject = {
        option_value: alarmItem.endpoint,
        type: alarmItem.endpoint.split('_').slice(-1)[0]
      }
      localStorage.setItem('jumpCallData', JSON.stringify(endpointObject))
      this.$router.push({path: '/endpointView'})
      // const news = this.$router.resolve({name: 'endpointView'})
      // window.open(news.href, '_blank')
    },
    goToNotify (item) {
      let params = {
        id: item.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.startNotify, params, () => {
        this.$Message.success(this.$t('tips.success'))
      },{isNeedloading: false})
    },
    deleteConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok () {
      this.removeAlarm(this.selectedData)
    },
    cancel () {
      this.isShowWarning = false
    },
    removeAlarm(alarmItem) {
      let params = {
        id: alarmItem.id,
        custom: true
      }
      if (!alarmItem.is_custom) {
        params.custom = false
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.getAlarm(this.cacheParams.id, this.cacheParams.viewCondition)
      })
    },
    remarkModal (item) {
      this.modelConfig.addRow = {
        id: item.id,
        message: item.custom_message,
        is_custom: false
      }
      this.$root.JQ('#remark_Modal').modal('show')
    },
    remarkAlarm () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.remarkAlarm, this.modelConfig.addRow, () => {
        this.$Message.success(this.$t('tips.success'))
        this.getAlarm(this.cacheParams.id, this.cacheParams.viewCondition)
        this.$root.JQ('#remark_Modal').modal('hide')
      })
    }
  }
}
</script>

<style scoped lang="less">
.alarm-item-label {
  width: 70px;
}
.alarm-empty {
  height: ~"calc(100vh - 180px)";
  width: ~"calc(100vw * 0.4)";
  text-align: center;
  padding:50px;
  color: #2d8cf0;
}
.flex-container {
  display: flex;
}
li {
  list-style: none;
} 

label {
  margin-bottom: 0;
  text-align: right;
}
.alarm-total {
  // float: right;
  font-size: 18px;
  margin-bottom: 8px;
}
.alarm-list {
  // height: ~"calc(100vh - 150px)"
  width: 700px;
  overflow-y: auto;
}
.alarm-item {
  border: 1px solid @gray-d;
  margin: 8px;
  padding: 4px;
  border-radius: 4px;
  li {
    padding: 2px;
  }
}
.alarm-item-border-high {
  // border: 1px solid @color-orange-F;
  color: @color-orange-F;
}
.alarm-item-border-medium {
  // border: 1px solid @blue-2;
  color: @blue-2;
}
// .alarm-item-border-low {
//   // border: 1px solid @gray-d;
// }

.alarm-item:hover {
  box-shadow: 0 0 12px @gray-d;
}

.alarm-item /deep/.ivu-icon-ios-close:before {
  content: "\F102";
}

.fa-operate {
  margin: 8px;
  float: right;
  font-size: 16px;
  cursor: pointer;
}

/* 可以设置不同的进入和离开动画 */
/* 设置持续时间和动画函数 */
.slide-fade-enter-active {
  transition: all .3s ease;
}
.slide-fade-leave-active {
  transition: all .8s cubic-bezier(1.0, 0.5, 0.8, 1.0);
}
.slide-fade-enter, .slide-fade-leave-to
/* .slide-fade-leave-active for below version 2.1.8 */ {
  transform: translateX(10px);
  opacity: 0;
}
</style>
