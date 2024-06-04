<template>
  <div class=" ">
    <div style="display: flex;justify-content: space-between;margin-bottom: 8px">
      <div>
        <Select v-model="type" style="width:100px" @on-change="typeChange">
          <Option v-for="item in typeList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
        </Select>
        <Select
          style="width:300px;margin-left: 12px;"
          v-model="targetId"
          filterable
          clearable 
          remote
          ref="select"
          :remote-method="getTargrtList"
          @on-change="search"
          @on-clear="typeChange"
          >
          <Option v-for="(option, index) in targetOptions" :value="option.option_value" :key="index">
            <TagShow :list="targetOptions" name="type" :tagName="option.type" :index="index"></TagShow> 
            {{option.option_text}}
          </Option>
        </Select>
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
              :on-error="uploadFailed">
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
          <span v-if="type==='service'">{{ $t('field.resourceLevel') }}</span>
          <span v-if="type==='group'">{{ $t('field.group') }}</span>
          <span v-if="type==='endpoint'">{{ $t('field.endpoint') }}</span>
        </Alert>
      </div>
      <div v-if="targetId&&dataEmptyTip">
        <Alert type="error">
          <span v-if="type==='service'">{{ $t('m_empty_data_recrisive') }}</span>
          <span v-if="type==='endpoint'">{{ $t('m_empty_data_endpoint') }}</span>
        </Alert>
      </div>
    </div>
    <div v-show="showTargetManagement" class="table-zone">
      <template v-for="(itemType, index) in thresholdTypes">
        <thresholdDetail 
          ref='thresholdDetail'
          v-if="type === itemType"
          :key=index
          :type=type
          @feedbackInfo="feedbackInfo"
        >
        </thresholdDetail>
      </template>
    </div>
  </div>
</template>

<script>
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
        {label: 'field.resourceLevel', value: 'service'},
        {label: 'field.group', value: 'group'},
        {label: 'field.endpoint', value: 'endpoint'}
      ],
      targetId: '',
      targetOptions: [],
      showTargetManagement: false,
      thresholdTypes: ['group', 'endpoint', 'service'],
      dataEmptyTip: false
    }
  },
  computed: {
    uploadUrl: function() {
      return baseURL_config + `/monitor/api/v2/alarm/strategy/import/${this.type}/${this.targetId}`
    }
  },
  async mounted () {
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    this.getTargrtList()
  },
  beforeDestroy () {
    this.$root.$store.commit('changeTableExtendActive', -1)
  },
  methods: {
    feedbackInfo (val) {
      this.dataEmptyTip = val
    },
    exportThreshold () {
      const api = `/monitor/api/v2/alarm/strategy/export/${this.type}/${this.targetId}`
      axios({
        method: 'GET',
        url: api,
        headers: {
          'Authorization': this.token
        }
      }).then((response) => {
        if (response.status < 400) {
          let content = JSON.stringify(response.data)
        let fileName = `threshold_${new Date().format('yyyyMMddhhmmss')}.json`
        let blob = new Blob([content])
        if('msSaveOrOpenBlob' in navigator){
          // Microsoft Edge and Microsoft Internet Explorer 10-11
          window.navigator.msSaveOrOpenBlob(blob, fileName)
        } else {
          if ('download' in document.createElement('a')) { // 非IE下载
            let elink = document.createElement('a')
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
        this.$Message.warning(this.$t('tips.failed'))
      });
    },
    uploadSucess () {
      this.$Message.success(this.$t('tips.success'))
    },
    uploadFailed (error, file) {
      this.$Message.warning({
          content: file.message,
          duration: 5
      })
    },
    typeChange () {
      this.clearTargrt()
      this.getTargrtList()
    },
    getTargrtList () {
      this.$refs.select.queryProp = ''
      const api = `/monitor/api/v2/alarm/strategy/search?type=${this.type}&search=`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.targetOptions = responseData
      }, {isNeedloading:false})
    },
    clearTargrt () {
      this.targetOptions = []
      this.targetId = ''
      this.showTargetManagement = false
      this.$refs.select.query = ''
    },
    search () {
      if (this.targetId) {
        this.showTargetManagement = true
        const find = this.targetOptions.find(item => item.option_value === this.targetId)
        this.$refs.thresholdDetail[0].setMonitorType(find.type);
        this.$refs.thresholdDetail[0].getDetail(this.targetId);
      }
    }
  },
  components: {
    TagShow,
    thresholdDetail
  },
}
</script>
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
    overflow: auto;
    height: ~"calc(100vh - 180px)";
  }
</style>
