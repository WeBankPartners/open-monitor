<template>
  <div class="prometheus-logs-container">
    <div class="search-form">
      <Form :model="searchForm" :label-width="0" inline>
        <FormItem label="">
          <Input
            v-model="searchForm.query"
            placeholder="请输入查询参数"
            style="width: 800px"
            clearable
          />
        </FormItem>
        <FormItem label="">
          <RadioGroup v-model="timeMode" @on-change="onTimeModeChange">
            <Radio label="point">时间点</Radio>
            <Radio label="range">时间范围</Radio>
          </RadioGroup>
        </FormItem>
        <FormItem v-if="timeMode === 'point'" label="">
          <DatePicker
            v-model="searchForm.time"
            type="date"
            placeholder="请选择日期"
            style="width: 200px"
            format="yyyy-MM-dd"
            @on-change="onTimeChange"
          />
        </FormItem>
        <template v-if="timeMode === 'range'">
          <FormItem label="">
            <DatePicker
              v-model="searchForm.start"
              type="date"
              placeholder="请选择开始日期"
              style="width: 200px"
              format="yyyy-MM-dd"
              @on-change="onStartTimeChange"
            />
          </FormItem>
          <FormItem label="">
            <DatePicker
              v-model="searchForm.end"
              type="date"
              placeholder="请选择结束日期"
              style="width: 200px"
              format="yyyy-MM-dd"
              @on-change="onEndTimeChange"
            />
          </FormItem>
        </template>
        <FormItem>
          <Button type="primary" @click="searchLogs" :loading="loading">
            <Icon type="ios-search" />
            搜索
          </Button>
          <Button @click="resetForm" style="margin-left: 8px">
            <Icon type="ios-refresh" />
            重置
          </Button>
        </FormItem>
      </Form>
    </div>

    <div class="log-results" v-if="logData.length > 0">
      <div class="log-list">
        <div
          v-for="(log, index) in logData"
          :key="index"
          class="log-item"
        >
          <div class="log-content">
            <div class="json-viewer">
              <div class="json-header">
                <!-- <span class="json-title">JSON数据</span> -->
                <!-- <Button size="small" @click="copyJson(log.metric)" type="text">
                  <Icon type="ios-copy" />
                  复制
                </Button> -->
              </div>
              <pre class="json-content" v-html="formatJsonWithSyntaxHighlight(log.metric)"></pre>
            </div>
          </div>
        </div>
      </div>
      <!-- 分页组件 -->
      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <Page
          :current="pagination.current"
          :total="pagination.total"
          :page-size="pagination.pageSize"
          :page-size-opts="[50, 100, 200, 500]"
          :show-total="pagination.showTotal"
          :show-sizer="true"
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
  </div>
</template>

<script>
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
        start: '',
        end: ''
      },
      logData: [],
      allLogData: [], // 存储所有原始数据
      pagination: {
        current: 1,
        pageSize: 100,
        total: 0,
        showTotal: (total, range) => `共 ${total} 条记录，当前显示第 ${range[0]}-${range[1]} 条`
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  methods: {
    onTimeModeChange() {
      // 切换时间模式时清空时间相关字段
      this.searchForm.time = ''
      this.searchForm.start = ''
      this.searchForm.end = ''
    },
    onTimeChange(time) {
      this.searchForm.time = time
    },
    onStartTimeChange(time) {
      this.searchForm.start = time
    },
    onEndTimeChange(time) {
      this.searchForm.end = time
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
            params.time = new Date(this.searchForm.time).getTime() / 1000
          }
        } else if (this.timeMode === 'range') {
          // 时间范围模式：传递start和end字段
          if (this.searchForm.start) {
            params.start = new Date(this.searchForm.start).getTime() / 1000
          }
          if (this.searchForm.end) {
            params.end = new Date(this.searchForm.end).getTime() / 1000
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
        start: '',
        end: ''
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
    formatTimestamp(timestamp) {
      if (!timestamp) {
        return ''
      }
      const date = new Date(timestamp * 1000)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    },
    formatJson(data) {
      if (typeof data === 'string') {
        try {
          return JSON.stringify(JSON.parse(data), null, 2)
        } catch (e) {
          return data
        }
      }
      return JSON.stringify(data, null, 2)
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
    },
    copyJson(data) {
      const jsonString = this.formatJson(data)
      if (navigator.clipboard) {
        navigator.clipboard.writeText(jsonString).then(() => {
          this.$Message.success('复制成功')
        })
          .catch(() => {
            this.fallbackCopyTextToClipboard(jsonString)
          })
      } else {
        this.fallbackCopyTextToClipboard(jsonString)
      }
    },
    fallbackCopyTextToClipboard(text) {
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.top = '0'
      textArea.style.left = '0'
      textArea.style.position = 'fixed'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      try {
        document.execCommand('copy')
        this.$Message.success('复制成功')
      } catch (err) {
        this.$Message.error('复制失败')
      }
      document.body.removeChild(textArea)
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

.log-results {
  background: #fff;
  border-radius: 6px;
  border: 1px solid #e9ecef;
  overflow: hidden;
  margin-top: -10px;
}

.result-header {
  background: #f8f9fa;
  padding: 15px 20px;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.result-header h3 {
  margin: 0;
  color: #495057;
  font-size: 16px;
  font-weight: 600;
}

.result-count {
  color: #6c757d;
  font-size: 14px;
}

 .log-list {
   height: calc(100vh - 210px);
   overflow-y: auto;
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

.log-item {
  border-bottom: 1px solid #f1f3f4;
  transition: background-color 0.2s;
}

.log-item:hover {
  background-color: #f8f9fa;
}

.log-item:last-child {
  border-bottom: none;
}

.log-header {
  padding: 12px 20px;
  background: #f8f9fa;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e9ecef;
}

.log-index {
  font-weight: 600;
  color: #495057;
  font-size: 14px;
}

.log-timestamp {
  color: #6c757d;
  font-size: 12px;
  font-family: 'Courier New', monospace;
}

.log-content {
  padding: 15px 20px;
}

.json-viewer {
  border: 1px solid #e9ecef;
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
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
