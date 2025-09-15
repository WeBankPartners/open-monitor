<template>
  <div class="prometheus-logs-container">
    <div class="search-form">
      <Form :model="searchForm" :label-width="0">
        <FormItem label="">
          <Input
            v-model="searchForm.query"
            placeholder="请输入查询参数"
            type="textarea"
            :autosize="{minRows: 1, maxRows: 5}"
            clearable
          />
        </FormItem>
        <FormItem label="">
          <div class="time-controls-row">
            <div class="time-controls">
              <RadioGroup v-model="timeMode" @on-change="onTimeModeChange" type="button" button-style="solid">
                <Radio label="point">时间点</Radio>
                <Radio label="range">时间范围</Radio>
              </RadioGroup>
              <div class="date-pickers">
                <DatePicker
                  v-if="timeMode === 'point'"
                  v-model="searchForm.time"
                  type="datetime"
                  placeholder="请选择日期时间"
                  style="width: 220px"
                  format="yyyy-MM-dd HH:mm:ss"
                  @on-change="onTimeChange"
                />
                <template v-if="timeMode === 'range'">
                  <div class="quick-range-select">
                    <Select
                      v-model="quickRangeValue"
                      style="width: 120px"
                      @on-change="onQuickRangeChange"
                    >
                      <Option value="10">10s</Option>
                      <Option value="60">1分钟</Option>
                      <Option value="1800">半小时</Option>
                      <Option value="7200">2小时</Option>
                    </Select>
                  </div>
                  <DatePicker
                    v-model="searchForm.range"
                    type="datetimerange"
                    placeholder="请选择时间范围（最多2小时）"
                    style="width: 460px"
                    format="yyyy-MM-dd HH:mm:ss"
                    @on-change="onRangeChange"
                  />
                </template>
              </div>
            </div>
            <div class="action-buttons">
              <Button type="primary" @click="searchLogs" :loading="loading">
                <Icon type="ios-search" />
                搜索
              </Button>
              <Button @click="resetForm" style="margin-left: 8px">
                <Icon type="ios-refresh" />
                重置
              </Button>
            </div>
          </div>
        </FormItem>
      </Form>
    </div>

    <div class="log-results" v-if="logData.length > 0">
      <Table :columns="tableColumns" :data="logData" border />
      <!-- 分页组件 -->
      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <Page
          :current="pagination.current"
          :total="pagination.total"
          :page-size="pagination.pageSize"
          :page-size-opts="[50, 100, 200, 500]"
          :show-sizer="true"
          :show-total="true"
          :show-elevator="true"
          @on-change="onPageChange"
          @on-page-size-change="onPageSizeChange"
          class="pagination"
        />
      </div>
    </div>

    <div class="no-data" v-else-if="!loading && searched">
      <Icon type="ios-document-outline" size="48" />
      <p>暂无数据</p>
    </div>

    <!-- 二级弹框：显示 values 详情 -->
    <Modal
      v-model="showValuesModal"
      title="数据详情"
      :mask-closable="true"
      :closable="true"
      :footer-hide="true"
      width="700"
    >
      <div class="values-modal-content">
        <pre class="json-content">{{ valuesModalText }}</pre>
      </div>
    </Modal>
  </div>
</template>

<script>
import dayjs from 'dayjs'
export default {
  name: 'PrometheusLogs',
  data() {
    return {
      loading: false,
      searched: false,
      timeMode: 'point', // 时间模式：'point' 时间点，'range' 时间范围
      searchForm: {
        query: '',
        time: '',
        range: []
      },
      logData: [],
      allLogData: [], // 存储所有原始数据
      pagination: {
        current: 1,
        pageSize: 100,
        total: 0
      },
      tableColumns: [
        {
          type: 'expand',
          width: 80,
          render: (h, params) => (
            <div class="expand-row">
              <div class="expand-col-left">
                <div class="json-viewer">
                  <pre
                    class="json-content"
                    domPropsInnerHTML={this.formatJsonWithSyntaxHighlight(params.row.metric)}
                  />
                </div>
              </div>
            </div>
          )
        },
        {
          title: '日志字段',
          key: 'metric',
          ellipsis: true,
          tooltip: true,
          render: (h, params) => (
            <span>{ this.getPreview(params.row.metric) }</span>
          )
        },
        {
          title: '数据预览',
          key: 'values',
          ellipsis: true,
          width: 300,
          tooltip: true,
          render: (h, params) => (
            <span>
              { this.getValuesPreview(this.timeMode === 'point' ? params.row.value : params.row.values) }
              { this.timeMode !== 'point' && <Button type="text" size="small" onClick={() => this.openValuesModal(params.row.values)} style="margin-left: 8px; color: #409EFF;">
                查看
              </Button>}
            </span>
          )
        }
      ],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      showValuesModal: false,
      valuesModalText: '',
      quickRangeValue: ''
    }
  },
  methods: {
    onQuickRangeChange(value) {
      if (!value) {
        return
      }
      const seconds = Number(value)
      if (Number.isNaN(seconds)) {
        return
      }
      this.setQuickRange(seconds)
    },
    setQuickRange(seconds) {
      const end = dayjs()
      const start = end.subtract(seconds, 'second')
      this.searchForm.range = [start.toDate(), end.toDate()]
    },
    openValuesModal(values) {
      this.valuesModalText = this.formatValuesAsText(values)
      this.showValuesModal = true
    },
    getColumnWidth(key) {
      try {
        const col = this.tableColumns.find(c => c.key === key)
        return col && col.width ? Number(col.width) : undefined
      } catch (e) {
        return undefined
      }
    },
    onTimeModeChange() {
      // 切换时间模式时清空时间相关字段
      this.searchForm.time = ''
      this.searchForm.range = []
      // 切换到时间范围时，默认选择并应用 10s
      if (this.timeMode === 'range') {
        this.quickRangeValue = '10'
        this.setQuickRange(10)
      }
    },
    onTimeChange(time) {
      this.searchForm.time = time
    },
    onRangeChange(range) {
      this.searchForm.range = range
    },
    async searchLogs() {
      if (!this.searchForm.query.trim()) {
        this.$Message.warning('请输入查询参数')
        return
      }
      this.loading = true
      this.searched = true
      try {
        const params = {
          query: this.searchForm.query
        }
        // 根据时间模式添加不同的时间参数
        if (this.timeMode === 'point') {
          // 时间点模式：传递time字段
          if (this.searchForm.time) {
            params.time = dayjs(this.searchForm.time).unix()
          }
        } else if (this.timeMode === 'range') {
          // 时间范围模式：传递start和end字段（校验不超过2小时）
          if (Array.isArray(this.searchForm.range) && this.searchForm.range.length === 2) {
            const start = dayjs(this.searchForm.range[0])
            const end = dayjs(this.searchForm.range[1])
            if (!start.isValid() || !end.isValid()) {
              this.$Message.error('请选择合法的时间范围')
              this.loading = false
              return
            }
            const diffHours = end.diff(start, 'hour', true)
            if (diffHours > 2) {
              this.$Message.error('时间范围不能超过2小时')
              this.loading = false
              return
            }
            params.start = start.unix()
            params.end = end.unix()
          }
        }
        this.request('GET', this.apiCenter.getPrometheusLogs, params, response => {
          this.allLogData = response.data.result || []
          this.pagination.total = this.allLogData.length
          this.pagination.current = 1
          this.updateDisplayData()
          this.$Message.success('搜索成功')
        })
      } catch (error) {
        console.error('Search logs error:', error)
        this.$Message.error('搜索失败')
        this.allLogData = []
        this.logData = []
        this.pagination.total = 0
      } finally {
        this.loading = false
      }
    },
    resetForm() {
      this.timeMode = 'point' // 重置为默认的时间点模式
      this.searchForm = {
        query: '',
        time: '',
        range: []
      }
      this.allLogData = []
      this.logData = []
      this.searched = false
      this.pagination.current = 1
      this.pagination.total = 0
    },
    // 更新显示数据（前端分页）
    updateDisplayData() {
      const start = (this.pagination.current - 1) * this.pagination.pageSize
      const end = start + this.pagination.pageSize
      this.logData = this.allLogData.slice(start, end)
    },
    // 分页变化事件
    onPageChange(page) {
      this.pagination.current = page
      this.updateDisplayData()
    },
    // 每页条数变化事件
    onPageSizeChange(pageSize) {
      this.pagination.pageSize = pageSize
      this.pagination.current = 1
      this.updateDisplayData()
    },
    // 将 values 数组格式化为可阅读文本，时间为 YYYY-MM-DD HH:mm:ss
    formatValuesAsText(values) {
      if (!Array.isArray(values) || values.length === 0) {
        // 兼容只返回单点 value 的情况
        return this.formatSingleValueAsText(values)
      }
      try {
        return values
          .map(item => {
            const ts = Array.isArray(item) ? item[0] : undefined
            const val = Array.isArray(item) ? item[1] : ''
            const timeStr = this.formatTimeFromSeconds(ts)
            return `${timeStr}  ${val}`
          })
          .join('\n')
      } catch (e) {
        return ''
      }
    },
    // 预览首条 values 内容（合并时间与数值）
    getValuesPreview(values) {
      if (Array.isArray(values) && values.length > 0) {
        const first = values[0]
        if (Array.isArray(first)) {
          const ts = Array.isArray(first) ? first[0] : undefined
          const val = Array.isArray(first) ? first[1] : ''
          const timeStr = this.formatTimeFromSeconds(ts)
          return `${timeStr}  ${val}`
        }
        // 兼容只有单点 value 的情况
        return this.formatSingleValueAsText(values)
      }
    },
    // 兼容只有单点 value 字段时的展示
    formatSingleValueAsText(valueField) {
      const ts = valueField[0]
      const val = valueField[1]
      const timeStr = this.formatTimeFromSeconds(ts)
      return `${timeStr}  ${val}`
    },
    formatJsonWithSyntaxHighlight(data) {
      let jsonString
      if (typeof data === 'string') {
        try {
          jsonString = JSON.stringify(JSON.parse(data), null, 2)
        } catch (e) {
          return this.escapeHtml(data)
        }
      } else {
        jsonString = JSON.stringify(data, null, 2)
      }
      return this.syntaxHighlight(jsonString)
    },
    getPreview(data) {
      try {
        let text = ''
        if (typeof data === 'string') {
          // If it's already a JSON string, try to compact it
          try {
            const obj = JSON.parse(data)
            text = JSON.stringify(obj)
          } catch (e) {
            text = data
          }
        } else {
          text = JSON.stringify(data)
        }
        if (text.length > 200) {
          return text.slice(0, 200) + '...'
        }
        return text
      } catch (e) {
        return ''
      }
    },
    formatTimeFromSeconds(seconds) {
      if (!seconds && seconds !== 0) {
        return ''
      }
      const value = Number(seconds)
      if (Number.isNaN(value)) {
        return ''
      }
      return dayjs.unix(value).format('YYYY-MM-DD HH:mm:ss')
    },
    syntaxHighlight(json) {
      const escapedJson = this.escapeHtml(json)
      return escapedJson.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, match => {
        let cls = 'json-number'
        if (/^"/.test(match)) {
          if (/:$/.test(match)) {
            cls = 'json-key'
          } else {
            cls = 'json-string'
          }
        } else if (/true|false/.test(match)) {
          cls = 'json-boolean'
        } else if (/null/.test(match)) {
          cls = 'json-null'
        }
        return '<span class="' + cls + '">' + match + '</span>'
      })
    },
    escapeHtml(text) {
      const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;'
      }
      return text.replace(/[&<>"']/g, m => map[m])
    }
  }
}
</script>

<style scoped>
.prometheus-logs-container {
  padding: 10px;
  background: #fff;
  min-height: 100vh;
}

.search-form {
  margin-bottom: 20px;
}

.search-form .ivu-form-item {
  margin-bottom: 16px;
}

.search-form .ivu-input {
  width: 100% !important;
  resize: none;
}

.time-controls-row {
  display: flex;
  justify-content: flex-start;
  align-items: center;
}

.time-controls {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin-right: 20px;
}

.date-pickers {
  display: flex;
  align-items: center;
  margin-left: 10px;
}

.quick-range {
  margin-left: 8px;
}

.quick-range-select {
  margin-right: 8px;
}

.log-results {
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  overflow: hidden;
  margin-top: -10px;
}

.pagination-wrapper {
  padding: 5px 15px;
  text-align: center;
  border-top: 1px solid #f1f3f4;
  background: #fff;
}

.pagination {
  margin: 0;
}

.json-viewer {
  border: 1px solid #e9ecef;
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
}

.expand-row {
  display: flex;
  flex-direction: row;
  gap: 12px;
}

.expand-col-left {
  width: 800px;
}

.expand-col-right {
  width: 280px;
}

.values-modal-content {
  max-height: 70vh;
  overflow: auto;
}

.json-header {
  background: #f8f9fa;
  padding: 10px 15px;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.json-title {
  font-weight: 600;
  color: #495057;
  font-size: 14px;
}

.json-content {
  background: #2d3748;
  color: #e2e8f0;
  padding: 15px;
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 400px;
  overflow-y: auto;
  border: none;
}

.no-data {
  text-align: center;
  padding: 60px 20px;
  color: #6c757d;
}

.no-data p {
  margin-top: 16px;
  font-size: 16px;
}

/* JSON语法高亮样式 */
.json-content .json-key {
  color: #68d391;
  font-weight: 600;
}

.json-content .json-string {
  color: #fbb6ce;
}

.json-content .json-number {
  color: #f6ad55;
}

.json-content .json-boolean {
  color: #63b3ed;
  font-weight: 600;
}

.json-content .json-null {
  color: #a0aec0;
  font-style: italic;
}
</style>
