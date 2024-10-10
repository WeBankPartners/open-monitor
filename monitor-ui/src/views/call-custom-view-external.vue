<template>
  <div>
    <view-config
      v-if="viewId"
      permissionType='view'
      :boardId="viewId"
      pageType="link"
      @alarmStatueChange='onAlarmListChange'
    />
  </div>
</template>

<script>
import ViewConfig from '@/views/custom-view/view-config'
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
  },
  methods: {
    onAlarmListChange(status) {
      if (status) {
        this.$nextTick(() => {
          const alarmList = document.querySelector('.alarm-list')
          alarmList.style.height = 'calc(100vh - 365px)'
        })
      }
    },
    paramsCheck(query) {
      this.viewId = Number(query.viewId)
    },
    refreshToken() {
      fetch('/auth/v1/api/token', {
        method: 'GET',
        headers: {
          Authorization: 'Bearer ' + this.token,
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
          }
        })
        .catch(error => {
        // 捕获异常
          console.error('There was a problem with the fetch operation:', error)
        })
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
