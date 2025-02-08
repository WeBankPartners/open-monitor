<template>
  <div class="log-management">
    <div class='log-management-top'>
      <RadioGroup
        v-model="type"
        type="button"
        button-style="solid"
        @on-change="typeChange"
        style="margin-right: 5px"
      >
        <Radio label="group">{{ $t('m_field_resourceLevel') }}</Radio>
        <Radio label="endpoint">{{ $t('m_tableKey_endpoint') }}</Radio>
      </RadioGroup>
      <Select
        :key='selectKey'
        style="width:300px;"
        v-model="targetId"
        filterable
        clearable
        remote
        ref="select"
        @on-clear="onTargetIdClear"
        @on-change="onFilterChange"
        @on-open-change="onSelectOpenChange"
      >
        <Option v-for="(option, index) in targetOptions" :value="option.guid" :label="option.display_name" :key="index">
          <TagShow :list="targetOptions" name="type" :tagName="option.type" :index="index"></TagShow>
          {{option.display_name}}
        </Option>
      </Select>
      <Input
        v-model="alarmName"
        @on-change="onFilterChange"
        style="width:250px;margin-left:12px;"
        clearable
        :placeholder="$t('m_placeholder_input') + $t('m_alarmName')"
      />
      <div class='upload-content mr-4'>
        <Button
          class="btn-upload"
          v-if="typeMap[type] === 'service'"
          @click="exportData"
        >
          <img src="@/styles/icon/DownloadOutlined.svg" class="upload-icon" />
          {{ $t('m_export') }}
        </Button>
        <Upload
          v-if="typeMap[type] === 'service'"
          :action="uploadUrl"
          :show-upload-list="false"
          :max-size="1000"
          with-credentials
          :headers="{'Authorization': token}"
          :on-success="uploadSucess"
          :on-error="uploadFailed"
        >
          <!-- <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}</Button> -->
          <Button class="btn-upload">
            <img src="@/styles/icon/UploadOutlined.svg" class="upload-icon" />
            {{ $t('m_import') }}
          </Button>
        </Upload>
      </div>
    </div>
    <div v-if="!targetId" style="margin: 10px 0">
      <Alert type="error">
        <span>{{ $t('m_empty_tip_1') }}</span>
        <span v-if="type === 'group'">{{ $t('m_field_resourceLevel') }}</span>
        <span v-if="type === 'endpoint'">{{ $t('m_field_endpoint') }}</span>
      </Alert>
    </div>
    <div v-if="targetId && isDataEmpty" style="margin: 10px 0">
      <Alert type="error">
        <span>{{ $t('m_noData') }}</span>
      </Alert>
    </div>
    <section v-show="showTargetManagement" class='key-word-content'>
      <keywordContent
        ref='keywordContent'
        :keywordType="typeMap[type]"
        @feedbackInfo="onFeedbackInfo"
      >
      </keywordContent>
    </section>
  </div>
</template>

<script>
import axios from 'axios'
import {debounce} from 'lodash'
import {baseURL_config} from '@/assets/js/baseURL'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import keywordContent from './keyword-content.vue'
import TagShow from '@/components/Tag-show.vue'
export default {
  name: '',
  data() {
    return {
      type: 'group',
      targetId: '',
      targetOptions: [],
      showTargetManagement: false,
      typeMap: {
        group: 'service',
        endpoint: 'endpoint'
      },
      alarmName: '',
      token: '',
      isDataEmpty: false,
      selectKey: ''
    }
  },
  async mounted() {
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    this.getTargrtList()
  },
  computed: {
    uploadUrl() {
      return baseURL_config + `${this.$root.apiCenter.bussinessMonitorImport}?serviceGroup=${this.targetId}`
    }
  },
  beforeDestroy() {
    this.$root.$store.commit('changeTableExtendActive', -1)
  },
  methods: {
    typeChange() {
      this.getTargrtList()
      this.selectKey = +new Date() + ''
    },
    getTargrtList(type = 'init') {
      const api = this.$root.apiCenter.getTargetByEndpoint + '/' + this.type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
        this.targetOptions = responseData
        if (type === 'init') {
          this.targetId = this.targetOptions[0].guid
          this.search()
        }
      }, {isNeedloading: false})
    },
    search() {
      if (this.targetId) {
        this.showTargetManagement = true
        this.$refs.keywordContent.getDetail(this.targetId, this.alarmName)
      }
    },
    onTargetIdClear() {
      this.$refs.keywordContent.getDetail('', this.alarmName)
    },
    onFilterChange: debounce(function () {
      this.search()
    }, 300),
    exportData() {
      const api = `${this.$root.apiCenter.bussinessMonitorExport}?serviceGroup=${this.targetId}`
      axios({
        method: 'GET',
        url: api,
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
      this.search()
    },
    uploadFailed(file) {
      this.$Message.warning(file.message)
    },
    onSelectOpenChange(open) {
      if (open) {
        this.getTargrtList('')
      }
    },
    onFeedbackInfo(allData) {
      this.isDataEmpty = allData.length === 0
    }
  },
  components: {
    TagShow,
    keywordContent
  },
}
</script>

<style lang="less">
.log-management {
  .ivu-radio-group-button .ivu-radio-wrapper-checked {
    background: #5384FF;
    color: #fff;
  }
}
</style>
<style scoped lang="less">
.log-management-top {
  display: flex;
  align-items: center;
  .btn-img {
    width: 16px;
    vertical-align: middle;
  }
}
.is-danger {
  color: red;
  margin-bottom: 0px;
}
.search-input {
  height: 32px;
  padding: 4px 7px;
  font-size: 12px;
  border: 1px solid #dcdee2;
  border-radius: 4px;
  width: 230px;
}

.section-table-tip {
  margin: 24px 20px 0;
}
.search-input:focus {
  outline: 0;
  border-color: #57a3f3;
}

.search-input-content {
  display: inline-block;
  vertical-align: middle;
}
.tag-width {
  cursor: auto;
  width: 55px;
  text-align: center;
}
</style>

<style scoped lang='less'>
.upload-content {
  display: flex;
  margin-left: auto
}
.key-word-content {
  margin-top: 16px;
  max-height: ~'calc(100vh - 170px)';
  overflow-y: auto;
}

</style>
