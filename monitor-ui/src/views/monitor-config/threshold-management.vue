<template>
  <div class="threshold-management">
    <div style="display: flex;justify-content: space-between;margin-bottom: 8px">
      <div style="display:flex;align-items:center;">
        <RadioGroup
          v-model="type"
          type="button"
          button-style="solid"
          @on-change="typeChange"
          style="margin-right: 5px"
        >
          <Radio v-for="item in typeList" :label="item.value" :key="item.value">{{ $t(item.label) }}</Radio>
        </RadioGroup>
        <Select
          :key='selectKey'
          style="width:250px;margin-left:12px;"
          v-model="targetId"
          filterable
          clearable
          ref="select"
          @on-clear="onTargetIdClear"
          @on-change="searchTableDetail"
        >
          <Option v-for="(option, index) in targetOptions"
                  :value="option.option_value"
                  :label="option.option_text"
                  :key="index"
          >
            <TagShow :list="targetOptions" name="type" :tagName="option.type" :index="index"></TagShow>
            {{option.option_text}}
          </Option>
        </Select>
        <!--告警名称-->
        <Input
          v-model="alarmName"
          @on-change="fetchDetailData"
          style="width:250px;margin-left:12px;"
          clearable
          :placeholder="$t('m_alarmName')"
        />
        <div>
          <span style="margin: 0 10px;">仅展示用户创建</span>
          <i-switch
            v-model="onlyShowCreated"
            true-value="me"
            false-value="all"
            size="default"
            @on-change="fetchDetailData"
          />
        </div>
      </div>
      <div>
        <template v-if="type !== 'endpoint' && targetId">
          <Button
            type="info"
            class="btn-left"
            @click="exportThreshold"
          >
            <img src="../../assets/img/export.png" class="btn-img" alt="" />
            {{ $t('m_export') }}
          </Button>
          <div style="display: inline-block;margin-bottom: 3px;">
            <Upload
              :action="uploadUrl"
              :show-upload-list="false"
              :max-size="1000"
              with-credentials
              :headers="{'Authorization': token}"
              :on-success="uploadSucess"
              :on-error="uploadFailed"
            >
              <Button type="primary" class="btn-left">
                <img src="../../assets/img/import.png" class="btn-img" alt="" />
                {{ $t('m_import') }}
              </Button>
            </Upload>
          </div>
        </template>
      </div>
    </div>
    <div>
      <div v-if="!targetId">
        <Alert type="error">
          <span>{{ $t('m_empty_tip_1') }}</span>
          <span v-if="type === 'service'">{{ $t('m_field_resourceLevel') }}</span>
          <span v-if="type === 'group'">{{ $t('m_field_group') }}</span>
          <span v-if="type === 'endpoint'">{{ $t('m_field_endpoint') }}</span>
        </Alert>
      </div>
      <div v-if="targetId && dataEmptyTip">
        <Alert type="error">
          <span v-if="type === 'service'">{{ $t('m_empty_data_recrisive') }}</span>
          <span v-if="type === 'endpoint'">{{ $t('m_empty_data_endpoint') }}</span>
        </Alert>
      </div>
    </div>
    <div v-show="showTargetManagement" class="table-zone">
      <thresholdDetail
        ref='thresholdDetail'
        :type="type"
        :alarmName="alarmName"
        :onlyShowCreated="onlyShowCreated"
        @feedbackInfo="feedbackInfo"
      >
      </thresholdDetail>
    </div>
  </div>
</template>

<script>
import isEmpty from 'lodash/isEmpty'
import debounce from 'lodash/debounce'
import cloneDeep from 'lodash/cloneDeep'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import thresholdDetail from './config-detail.vue'
import TagShow from '@/components/Tag-show.vue'
import {baseURL_config} from '@/assets/js/baseURL'
import axios from 'axios'
export default {
  name: '',
  data() {
    return {
      token: null,
      type: 'service',
      typeList: [
        {
          label: 'm_field_resourceLevel',
          value: 'service'
        },
        {
          label: 'm_field_group',
          value: 'group'
        },
        {
          label: 'm_field_endpoint',
          value: 'endpoint'
        }
      ],
      targetId: '',
      targetOptions: [],
      showTargetManagement: false,
      thresholdTypes: ['group', 'endpoint', 'service'],
      dataEmptyTip: false,
      getTargetOptionsSearch: '',
      alarmName: '', // 告警名称
      onlyShowCreated: 'all', // me用户创建 all所有
      selectKey: ''
    }
  },
  computed: {
    uploadUrl() {
      return baseURL_config + `/monitor/api/v2/alarm/strategy/import/${this.type}/${this.targetId}`
    }
  },
  async mounted() {
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    this.getTargetOptionsSearch = ''
    this.initTargetByType()
  },
  beforeDestroy() {
    this.$root.$store.commit('changeTableExtendActive', -1)
  },
  methods: {
    feedbackInfo(val) {
      this.dataEmptyTip = val
    },
    async initTargetByType() {
      await this.getTargetOptions()
      if (!isEmpty(this.targetOptions)) {
        this.targetId = this.targetOptions[0].option_value
        this.searchTableDetail()
      }
    },
    exportThreshold() {
      const api = `/monitor/api/v2/alarm/strategy/export/${this.type}/${this.targetId}`
      axios({
        method: 'GET',
        url: api,
        headers: {
          Authorization: this.token
        }
      }).then(response => {
        if (response.status < 400) {
          const content = JSON.stringify(response.data)
          const fileName = `threshold_${new Date().format('yyyyMMddhhmmss')}.json`
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
      this.searchTableDetail()
    },
    uploadFailed(file) {
      this.$Message.warning({
        content: file.message,
        duration: 5
      })
    },
    typeChange() {
      this.alarmName = ''
      // this.clearTargrt()
      this.initTargetByType()
      this.selectKey = +new Date() + ''
    },
    getTargetOptions() {
      return new Promise(resolve => {
        const api = `/monitor/api/v2/alarm/strategy/search?type=${this.type}&search=${this.getTargetOptionsSearch}`
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
          this.targetOptions = cloneDeep(responseData)
          window.targetOptions = this.targetOptions
          resolve(this.targetOptions)
        }, {isNeedloading: false})
      })
    },
    clearTargrt() {
      // this.targetOptions = []
      // this.targetId = ''
      // this.showTargetManagement = false
      // this.getTargetOptionsSearch = ''
    },
    searchTableDetail() {
      if (this.targetId) {
        this.alarmName = ''
        this.showTargetManagement = true
        const find = this.targetOptions.find(item => item.option_value === this.targetId)
        this.$refs.thresholdDetail.setMonitorType(find.type)
        this.$refs.thresholdDetail.getDetail(this.targetId)
      }
    },
    // debounceGetTargetOptions: debounce(async function () {
    //   const targetItem = find(this.targetOptions, {
    //     option_value: this.targetId
    //   })
    //   if (targetItem && this.getTargetOptionsSearch !== targetItem.option_text) {
    //     return
    //   }
    //   await this.getTargetOptions()
    // }, 400),
    fetchDetailData: debounce(function () {
      this.$refs.thresholdDetail.getDetail(this.targetId)
    }, 300),
    onTargetIdClear() {
      this.$refs.thresholdDetail.getDetail('')
    }
  },
  components: {
    TagShow,
    thresholdDetail
  },
}
</script>
<style lang="less">
.threshold-management {
  .ivu-radio-group-button .ivu-radio-wrapper-checked {
    background: #2d8cf0;
    color: #fff;
  }
}
</style>
<style scoped lang="less">
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
  .btn-img {
    width: 16px;
    vertical-align: middle;
  }
  .btn-left {
    margin-left: 8px;
  }

  .table-zone {
    overflow-y: auto;
    height: ~"calc(100vh - 180px)";
  }
</style>
