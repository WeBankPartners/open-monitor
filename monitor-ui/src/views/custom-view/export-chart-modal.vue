<template>
  <Modal
    v-model="isExportModalShow"
    :title="$t('m_select_export_chart')"
    @on-visible-change="onVisibleChange"
  >
    <template slot='footer'>
      <div>
        <Button
          type="default"
          @click="onExportModalClose"
        >{{ $t('m_button_cancel') }}</Button>
        <Button
          @click="onExportChartConfirm"
          :disabled="checkedNodeIdList.length === 0"
          type="primary"
        >
          {{ $t('m_save') }}
        </Button>
      </div>
    </template>
    <Tree :data="exportChartList"
          show-checkbox
          multiple
          @on-check-change="onTreeCheckChange"
    />
  </Modal>
</template>

<script>
import cloneDeep from 'lodash/cloneDeep'
import remove from 'lodash/remove'
import find from 'lodash/find'
import axios from 'axios'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'

export default {
  name: '',
  props: {
    pannelId: {
      type: Number
    },
    panalName: {
      type: String,
      default: ''
    },
    isModalShow: {
      type: Boolean,
      default: false
    }
  },
  watch: {
    isModalShow(val) {
      if (val) {
        this.getTreeList()
      } else {
        this.isExportModalShow = false
      }
    }
  },
  data() {
    return {
      isExportModalShow: false,
      checkedNodeIdList: [],
      allNodeIdList: [],
      exportChartList: [],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance
    }
  },
  methods: {
    getTreeList() {
      this.request('GET', '/monitor/api/v2/dashboard/custom', {
        id: this.pannelId
      }, res => {
        this.processChartList(res.charts || [])
      })
    },
    processChartList(chartList = []) {
      this.allNodeIdList = []
      const children = []
      for (let i=0; i < chartList.length; i++) {
        const item = chartList[i]
        this.allNodeIdList.push(item.id)
        const resObj = find(children, {
          title: item.group
        })
        if (resObj) {
          resObj.children = resObj.children || []
          resObj.children.push({
            title: item.name,
            id: item.id
          })
        } else {
          children.push({
            title: item.group,
            expand: true,
            children: [
              {
                title: item.name,
                id: item.id
              }
            ]
          })
        }
      }
      const unclassifiedChart = find(children, {
        title: ''
      })
      if (unclassifiedChart) {
        const tempChart = cloneDeep(unclassifiedChart)
        tempChart.title = this.$t('m_unclassified')
        remove(children, val => val.title === '')
        children.push(tempChart)
      }
      this.exportChartList = [
        {
          title: this.panalName,
          expand: true,
          checked: true,
          children
        }
      ]
      this.isExportModalShow = true
      this.checkedNodeIdList = cloneDeep(this.allNodeIdList)
    },
    getAuthorization() {
      if (localStorage.getItem('monitor-accessToken')) {
        return 'Bearer ' + localStorage.getItem('monitor-accessToken')
      }
      return (window.request ? 'Bearer ' + getPlatFormToken() : getToken()) || null

    },
    onExportChartConfirm() {
      const params = {
        id: this.pannelId,
        chartIds: this.checkedNodeIdList
      }
      const api = '/monitor/api/v2/dashboard/custom/export'

      const headers = {
        'X-Auth-Token': getToken() || null,
        Authorization: this.getAuthorization()
      }
      axios.post(api, params, {
        headers
      }).then(response => {
        if (response.status < 400) {
          this.checkedNodeIdList = []
          this.closeModal()
          const content = JSON.stringify(response.data)
          const fileName = this.panalName + '.json'
          const blob = new Blob([content])
          if ('msSaveOrOpenBlob' in navigator){
            // Microsoft Edge and Microsoft Internet Explorer 10-11
            window.navigator.msSaveOrOpenBlob(blob, fileName)
          } else {
            if ('download' in document.createElement('a')) { // 非IE下载
              const elink = document.createElement('a')
              elink.download = fileName
              elink.style.display = 'none'
              elink.href = URL.createObjectURL(blob)
              document.body.appendChild(elink)
              elink.click()
              URL.revokeObjectURL(elink.href) // 释放URL 对象
              document.body.removeChild(elink)
            } else { // IE10+下载
              navigator.msSaveOrOpenBlob(blob, fileName)
            }
          }
        }
      })
        .catch(() => {
          this.$Message.warning(this.$t('m_tips_failed'))
        })
    },
    onExportModalClose() {
      this.checkedNodeIdList = []
      this.closeModal()
    },
    changeTreeChecked(arr, val) {
      for (let i=0; i<arr.length; i++) {
        const item = arr[i]
        item.checked = val
        if (item.children && item.children.length > 0) {
          this.changeTreeChecked(item.children, val)
        }
      }
    },
    onVisibleChange(state) {
      if (!state) {
        this.closeModal()
      }
    },
    closeModal() {
      this.$emit('close')
    },
    onTreeCheckChange(nodeList) {
      this.checkedNodeIdList = []
      nodeList.forEach(node => {
        if (node.id) {
          this.checkedNodeIdList.push(node.id)
        }
      })
    }
  }
}
</script>
