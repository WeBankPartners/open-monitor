<template>
  <div class="classic-table">
    <div class='classic-table-detail'>
      <Table :columns="columns" :border="true" size="small" :data="tableData"></Table>
    </div>
    <slot name="pagination"></slot>
    <Modal
      v-model="isShowWarning"
      :title="$t('m_closeConfirm_title')"
      :mask-closable="false"
      @on-ok="ok"
      @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_closeConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
const alarmLevelMap = {
  low: {
    label: 'm_low',
    buttonType: 'green'
  },
  medium: {
    label: 'm_medium',
    buttonType: 'gold'
  },
  high: {
    label: 'm_high',
    buttonType: 'red'
  }
}
import isEmpty from 'lodash/isEmpty'
export default {
  name: '',
  data() {
    return {
      columns: [
        {
          title: this.$t('m_alarmName'),
          key: 'alarm_name',
          minWidth: 160,
          render: (h, params) => (
            <BaseEllipsis content={params.row.alarm_name}></BaseEllipsis>
          )
        },
        {
          title: this.$t('m_menu_configuration'),
          key: 'strategyGroupsInfo',
          render: (h, params) => (
            <BaseEllipsis content={params.row.strategyGroupsInfo}></BaseEllipsis>
          ),
          minWidth: 120
        },
        {
          title: this.$t('m_field_endpoint'),
          key: 'endpoint',
          minWidth: 120
        },
        {
          title: this.$t('m_tableKey_content'),
          key: 'content',
          minWidth: 160,
          render: (h, params) => (
            <BaseEllipsis content={params.row.content}></BaseEllipsis>
          )
        },
        {
          title: this.$t('m_log'),
          key: 'log',
          minWidth: 160,
          render: (h, params) => (
            <BaseEllipsis content={params.row.log}></BaseEllipsis>
          )
        },
        {
          title: this.$t('m_tableKey_s_priority'),
          key: 's_priority',
          render: (h, params) => (
            <Tag color={alarmLevelMap[params.row.s_priority].buttonType}>{this.$t(alarmLevelMap[params.row.s_priority].label)}</Tag>
          ),
          width: 80
        },
        {
          title: this.$t('m_field_metric'),
          key: 'alarm_metric_list_join',
          width: 120
        },
        {
          title: this.$t('m_field_threshold'),
          key: 'alarm_detail',
          minWidth: 200,
          render: (h, params) => (
            <BaseEllipsis content={params.row.alarm_detail}></BaseEllipsis>
          )
        },
        {
          title: this.$t('m_tableKey_start'),
          key: 'start_string',
          width: 170
        },
        {
          title: this.$t('m_remark'),
          key: 'custom_message',
          render: (h, params) => (
            <div>{params.row.custom_message || '-'}</div>
          ),
          minWidth: 160
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 160,
          align: 'center',
          fixed: 'right',
          render: (h, params) => (
            <div style="text-align: left; cursor: pointer;display: inline-flex;">
              <Tooltip content={this.$t('m_button_view')} placement="top" transfer={true}>
                <Button
                  size="small"
                  type="primary"
                  onClick={() => this.goToEndpointView(params.row)}
                  style="margin-right:5px;"
                >
                  <Icon type="ios-stats" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip content={this.$t('m_close')} placement="top" transfer={true}>
                <Button
                  size="small"
                  type="error"
                  onClick={() => this.deleteConfirmModal(params.row)}
                  style="margin-right:5px;"
                >
                  <Icon type="md-eye-off" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip content={this.$t('m_remark')} placement="top" transfer={true}>
                <Button
                  size="small"
                  type="warning"
                  onClick={() => this.remarkModal(params.row)}
                  style="margin-right:5px;"
                >
                  <Icon type="md-pricetags" size="16"></Icon>
                </Button>
              </Tooltip>
            </div>
          )
        }
      ],
      tableData: [],
      isShowWarning: false,
      selectedData: '',
      strategyNameMaps: {
        endpointGroup: 'm_base_group',
        serviceGroup: 'm_field_resourceLevel'
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  mounted(){
  },
  methods: {
    changeResultData(dataList) {
      if (dataList && !isEmpty(dataList)) {
        dataList.forEach(item => {
          item.strategyGroupsInfo = '-'
          item.alarm_metric_list_join = '-'
          if (!isEmpty(item.strategy_groups)) {
            item.strategyGroupsInfo = item.strategy_groups.reduce((res, cur) => res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '<br/> ', '')
          }

          if (!isEmpty(item.alarm_metric_list)) {
            item.alarm_metric_list_join = item.alarm_metric_list.join(',')
          }
        })
      }
      return dataList
    },
    getAlarm(resultData) {
      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.tableData = this.changeResultData(resultData)
    },
    goToEndpointView(alarmItem) {
      const endpointObject = {
        option_name: alarmItem.endpoint,
        option_value: alarmItem.endpoint,
        type: alarmItem.endpoint.split('_').slice(-1)[0]
      }
      localStorage.setItem('jumpCallData', JSON.stringify(endpointObject))
      const news = this.$router.resolve({name: 'endpointView'})
      window.open(news.href, '_blank')
    },
    remarkModal(item) {
      this.$parent.remarkModal(item)
    },
    deleteConfirmModal(rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok() {
      this.removeAlarm(this.selectedData)
    },
    cancel() {
      this.isShowWarning = false
    },
    removeAlarm(alarmItem) {
      const params = {
        id: alarmItem.id,
        custom: true
      }
      if (!alarmItem.is_custom) {
        params.custom = false
      }
      this.request('POST', this.apiCenter.alarmManagement.close.api, params, () => {
        this.$parent.getAlarm()
      })
    }
  },
  components: {}
}
</script>

<style scoped lang="less">
.classic-table-detail {
  max-height: ~'calc(100vh - 320px)';
  overflow-y: scroll;
}
</style>
