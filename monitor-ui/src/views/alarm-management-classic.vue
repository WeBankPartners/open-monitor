<template>
  <div class="classic-table">
    <PageTable :pageConfig="pageConfig"></PageTable>
    <slot name="pagination"></slot>
    <Modal
      v-model="isShowWarning"
      :title="$t('closeConfirm.title')"
      :mask-closable="false"
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
    {title: 'm_alarmName', value: 'alarm_name', display: true},
    {title: 'menu.configuration', value: 'strategyGroupsInfo', display: true, renderContent: true},
    {title: 'field.endpoint', value: 'endpoint', display: true},
    {title: 'alarmContent', value: 'content', display: true},
    {title: 'tableKey.s_priority', value: 's_priority', display: true},
    {title: 'field.metric', value: 'alarm_metric_list_join', display: true},
    {title: 'field.threshold', value: 'alarm_detail', display: true, renderContent: true},
    {title: 'tableKey.start', value: 'start_string', display: true},
    {title: 'm_remark', value: 'custom_message', display: true},
  ]
  const btn = [
    {btn_name: 'button.view', btn_func: 'goToEndpointView'},
    {btn_name: 'm_remark', btn_func: 'remarkModal'},
    {btn_name: 'close', btn_func: 'deleteConfirmModal'}
  ]

import isEmpty from 'lodash/isEmpty';
export default {
  name: '',
  data () {
    return {
      pageConfig: {
        CRUD: '',
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
      selectedData: '',
      strategyNameMaps: {
        "endpointGroup": "m_base_group",
        "serviceGroup": "field.resourceLevel"
      }
    }
  },
  mounted(){
  },
  methods: {
    changeResultData(dataList) {
      if (dataList && !isEmpty(dataList)) {
        dataList.forEach(item => {
          item.strategyGroupsInfo = '-';
          item.alarm_metric_list_join = '-';
          if (!isEmpty(item.strategy_groups)) {
            item.strategyGroupsInfo = item.strategy_groups.reduce((res, cur)=> {
              return res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '<br/> '
            }, '')
          }

          if (!isEmpty(item.alarm_metric_list)) {
            item.alarm_metric_list_join = item.alarm_metric_list.join(';')
          }
        });
      }
      return dataList
    },
    getAlarm(resultData) {
      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.pageConfig.table.tableData = this.changeResultData(resultData);
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
    remarkModal (item) {
      this.$parent.remarkModal(item)
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
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        this.$parent.$parent.$parent.getAlarm()
      })
    }
  },
  components: {}
}
</script>

<style scoped lang="less">
.classic-table {
  /deep/ .table {
    .th-border-bottom {
      border-top: 0;
    }
  }
}
</style>

<style lang="less">
tr {
  position: relative;
  .td-operation {
    position: absolute;
    top: 0;
    bottom: 0;
    right: 0;
    display: flex;
    align-items: center;
  }

  .c-dark.tdoverflow {
    max-width: 172px!important;
  }
  .render-content {
    overflow: scroll;
  }
}


</style>