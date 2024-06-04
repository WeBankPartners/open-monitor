<template>
  <div class="classic-table">
    <Table :columns="columns" :data="tableData"></Table>
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
  const alarmLevelMap = {
    low: {
      label: "m_low",
      buttonType: "blue"
    },
    medium: {
      label: "m_medium",
      buttonType: "warning"
    },
    high: {
      label: "m_high",
      buttonType: "error"
    }
  }
import isEmpty from 'lodash/isEmpty';
export default {
  name: '',
  data () {
    return {
      columns: [
        {
          title: this.$t('m_alarmName'),
          key: 'alarm_name'
        },
        {
          title: this.$t('menu.configuration'),
          key: 'strategyGroupsInfo',
          render: (h, params) => {
            return (
              <div domPropsInnerHTML={params.row.strategyGroupsInfo}></div>
            )
          }
        },
        {
          title: this.$t('field.endpoint'),
          key: 'endpoint'
        },
        {
          title: this.$t('alarmContent'),
          key: 'content',
          ellipsis: true,
          tooltip: true
        },
        {
          title: this.$t('tableKey.s_priority'),
          key: 's_priority',
          width: 90,
          render: (h, params) => {
            return (
              <Tag color={alarmLevelMap[params.row.s_priority].buttonType}>{this.$t(alarmLevelMap[params.row.s_priority].label)}</Tag>
            )
          }
        },
        {
          title: this.$t('field.metric'),
          key: 'alarm_metric_list_join'
        },
        {
          title: this.$t('field.threshold'),
          key: 'alarm_detail',
          width: 300,
          render: (h, params) => {
            return (
              <Tooltip transfer={true} placement="bottom-start" max-width="300">
                <div slot="content">
                  <div domPropsInnerHTML={params.row.alarm_detail}></div>
                </div>
                <div domPropsInnerHTML={params.row.alarm_detail}></div>
              </Tooltip>
            )
          }
        },
        {
          title: this.$t('tableKey.start'),
          key: 'start_string',
          width: 170,
        },
        {
          title: this.$t('m_remark'),
          key: 'custom_message',
          width: 120,
          render: (h, params) => {
            return(
              <div>{params.row.custom_message || '-'}</div>
            )
          }
        },
        {
          title: this.$t('table.action'),
          key: 'action',
          width: 160,
          align: 'left',
          fixed: 'right',
          render: (h, params) => {
            return (
              <div style="text-align: left; cursor: pointer;display: inline-flex;">
              <Tooltip content={this.$t('button.view')} placement="top" transfer={true}>
                  <Button
                    size="small"
                    type="primary"
                    onClick={() => this.goToEndpointView(params.row)}
                    style="margin-right:5px;"
                  >
                    <Icon type="ios-stats" size="16"></Icon>
                  </Button>
                </Tooltip>
                <Tooltip content={this.$t('close')} placement="top" transfer={true}>
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
        }
      ],
      tableData: [],
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
            item.alarm_metric_list_join = item.alarm_metric_list.join(',')
          }
        });
      }
      return dataList
    },
    getAlarm(resultData) {
      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.tableData = this.changeResultData(resultData)
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
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        this.$parent.$parent.$parent.getAlarm()
      })
    }
  },
  components: {}
}
</script>