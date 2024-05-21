<template>
  <div class=" ">
    <section>
      <ul class="search-ul">
        <li class="search-li">
          <Select v-model="type" style="width:100px" @on-change="typeChange">
            <Option v-for="item in typeList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <Select
            style="width:300px;"
            v-model="targetId"
            filterable
            clearable 
            remote
            ref="select"
            :remote-method="getTargrtList"
            @on-change="search"
            >
            <Option v-for="(option, index) in targetOptions" :value="option.option_value" :key="index">
              <TagShow :tagName="option.type" :index="index"></TagShow> 
              {{option.option_text}}
            </Option>
          </Select>
        </li>
        <li class="search-li">
          <button type="button" class="btn btn-sm btn-confirm-f"
          :disabled="targetId === ''"
          @click="search">
            <i class="fa fa-search" ></i>
            {{$t('button.search')}}
          </button>
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
            <!-- <div style="display: inline-block;margin-bottom: 1px;vertical-align: bottom;line-height: 32px;"> 
              <Upload 
              :action="uploadUrl" 
              :show-upload-list="false"
              :max-size="1000"
              with-credentials
              :headers="{'Authorization': token}"
              :on-success="uploadSucess"
              :on-error="uploadFailed">
                <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}</Button>
              </Upload>
            </div>
            <button type="button" class="btn-cancel-f" @click="exportThreshold">{{$t("m_export")}}</button> -->
          </template>
        </li>
      </ul>
    </section> 
    <section v-show="showTargetManagement" style="margin-top: 16px;">
      <template v-for="(itemType, index) in thresholdTypes">
        <thresholdDetail 
          ref='thresholdDetail'
          v-if="type === itemType"
          :key=index
          :type=type
        >
        </thresholdDetail>
      </template>
    </section>
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
      thresholdTypes: ['group', 'endpoint', 'service']
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
      this.$Message.warning(file.message)
    },
    typeChange () {
      this.clearTargrt()
      this.getTargrtList()
    },
    getTargrtList () {
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
.search-li {
    display: inline-block;
  }
  .search-ul>li:not(:first-child) {
    padding-left: 10px;
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
</style>
