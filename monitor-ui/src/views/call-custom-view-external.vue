<template>
  <div>
    <CallCustomViewExternalPanel ref="callCustomViewExternalPanelRef"  :id="viewId"></CallCustomViewExternalPanel>
  </div>
</template>

<script>
import CallCustomViewExternalPanel from '@/views/call-custom-view-external-panel'
import { setLocalstorage } from '@/assets/js/localStorage.js'
export default {
  name: '',
  data() {
    return {
      viewId: '',
      token: ''
    }
  },
  mounted() {
    const query = this.$route.query
    this.paramsCheck(query)
    // this.refreshToken()
    this.$refs.callCustomViewExternalPanelRef.getDashData(this.viewId)
  },
  methods: {
    paramsCheck (query) {
      this.viewId = Number(query.viewId)
      // this.token = query.token
    },
    refreshToken () {
      fetch('/auth/v1/api/token', {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer ' + this.token,
          // 其他自定义 header
        }
      })
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not ok')
        }
        return response.json() // 解析 JSON 格式的响应数据
      })
      .then(data => {
        // 处理获取到的数据
        if (data.status === 'OK') {
          setLocalstorage(data.data)
          this.$refs.callCustomViewExternalPanelRef.getDashData(this.viewId)
        }
      })
      .catch(error => {
        // 捕获异常
        console.error('There was a problem with the fetch operation:', error)
      })
    }
  },
  components: {
    CallCustomViewExternalPanel
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
