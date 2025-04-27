<template>
  <span>
    <Button
      v-if="isShowExportBtn"
      class="btn-upload"
      @click="exportHandler"
    >
      <img src="@/styles/icon/DownloadOutlined.png" class="upload-icon" />
      {{ $t('m_export') }}
    </Button>
    <div style="display: inline-block;margin-bottom: 3px;" v-if="isShowImportBtn">
      <Upload
        :action="uploadUrl"
        :show-upload-list="false"
        :max-size="1000"
        with-credentials
        :headers="{'Authorization': token}"
        :on-success="uploadSucess"
        :on-error="uploadFailed"
        :before-upload="handleBeforeUpload"
      >
        <Button class="btn-upload">
          <img src="@/styles/icon/UploadOutlined.png" class="upload-icon" />
          {{ $t('m_import') }}
        </Button>
      </Upload>
    </div>
  </span>
</template>
<script>
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import axios from 'axios'
export default {
  name: '',
  data() {
    return {
      token: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
    }
  },
  props: {
    isShowExportBtn: {
      type: Boolean,
      required: false,
      default: false
    },
    exportUrl: {
      type: String,
      required: false,
      default: ''
    },
    exportMethod: {
      type: String,
      required: false,
      default: 'GET'
    },
    exportData: {
      type: Array,
      required: false,
      // eslint-disable-next-line vue/require-valid-default-prop
      default: []
    },
    validateExportDataEmpty: {
      type: Boolean,
      required: false,
      default: false
    },
    isShowImportBtn: {
      type: Boolean,
      required: false,
      default: false
    },
    uploadUrl: {
      type: String,
      required: false,
      default: ''
    }
  },
  mounted() {
    this.MODALHEIGHT = document.body.scrollHeight - 300
    this.token = this.returnLatestToken()
  },
  methods: {
    async exportHandler() {
      if (this.validateExportDataEmpty && this.exportData.length === 0) {
        this.$Message.warning(this.$t('m_select_data_tip'))
        return
      }
      this.token = await this.refreshToken()
      axios({
        method: this.exportMethod,
        url: this.exportUrl,
        data: {
          guidList: this.exportData // 后台会固定取guidList
        },
        headers: {
          Authorization: this.token
        }
      }).then(response => {
        if (response.status < 400) {
          const content = JSON.stringify(response.data)
          const fileName = `${response.headers['content-disposition'].split(';')[1].trim().split('=')[1]}`
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
    uploadSucess(res) {
      if (res.status === 'ERROR') {
        this.$Message.error(res.message)
        return
      }
      this.$Message.success(this.$t('m_tips_success'))
      this.$emit('successCallBack')
    },
    uploadFailed(file) {
      this.$Message.warning(file.message)
    },
    async refreshToken() {
      await this.request('GET', '/monitor/api/v1/user/role/list?page=1&size=1', '')
      const token = this.returnLatestToken()
      return new Promise(resolve => {
        resolve(token)
      })
    },
    returnLatestToken() {
      return (window.Request ? 'Bearer ' + getPlatFormToken() : getToken()) || null
    },
    async handleBeforeUpload() {
      this.token = await this.refreshToken()
      return true
    }
  }
}
</script>
<style  lang="less" scoped>
.btn-img {
  width: 16px;
  vertical-align: middle;
}

.btn-left {
  margin-left: 8px;
}
</style>
