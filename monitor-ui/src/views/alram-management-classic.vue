<template>
  <div class="">
    <PageTable :pageConfig="pageConfig"></PageTable>
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
  </div>
</template>

<script>
  let tableEle = [
    {title: 'alarmContent', value: 'content', display: true},
    {title: 'field.endpoint', value: 'endpoint', display: true},
    {title: 'tableKey.s_priority', value: 's_priority', display: true},
    {title: 'tableKey.start', value: 'start_string', display: true},
    {title: 'field.metric', value: 's_metric', display: true},
    {title: 'tableKey.tags', value: 'tags', display: true},
    {title: 'tableKey.start_value', value: 'start_value', display: true},
    {title: 'tableKey.threshold', value: 's_cond', display: true},
    {title: 'tableKey.s_last', value: 's_last', display: true}
  ]
  const btn = [
    {btn_name: 'button.view', btn_func: 'goToEndpointView'},
    {btn_name: 'remark', btn_func: 'remarkAlarm'},
    {btn_name: 'close', btn_func: 'deleteConfirmModal'}
  ]
export default {
  name: '',
  data () {
    return {
      pageConfig: {
        CRUD: '',
        researchConfig: {
          input_conditions: [
            {value: 'search', type: 'input', placeholder: 'placeholder.input', style: ''}],
          btn_group: [
            {btn_name: 'button.search', btn_func: 'search', class: 'btn-confirm-f', btn_icon: 'fa fa-search'}
          ],
          filters: {
          }
        },
        table: {
          selection: false,
          tableData: [],
          tableEle: tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btn,
          pagination: this.pagination,
          handleFloat:true,
        },
        pagination: false
      },
      isShowWarning: false,
      selectedData: ''
    }
  },
  mounted(){
    this.getAlarm()
    // this.interval = setInterval(()=>{
    //   this.getAlarm()
    // }, 10000)
    // this.$once('hook:beforeDestroy', () => {
    //   clearInterval(this.interval)
    // })
  },
  methods: {
    getAlarm() {
      let params = {}
      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/problem/query', params, (responseData) => {
        this.pageConfig.table.tableData = responseData.data
        console.log(this.resultData)
      }, {isNeedloading: false})
    },
    goToEndpointView (alarmItem) {
      const endpointObject = {
        option_value: alarmItem.endpoint,
        type: alarmItem.endpoint.split('_').slice(-1)[0]
      }
      localStorage.setItem('jumpCallData', JSON.stringify(endpointObject))
      const news = this.$router.resolve({name: 'endpointView'})
      window.open(news.href, '_blank')
    },
    remarkAlarm () {},
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
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.getAlarm()
      })
    }
  },
  components: {}
}
</script>

<style scoped lang="less">
</style>