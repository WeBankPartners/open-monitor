<template>
  <div class="">
    <button class="btn-confirm-f btn-small" @click="addPanel">{{$t('resourceLevel.addPanel')}}</button>
    <i class="fa fa-refresh" aria-hidden="true" @click="getAllResource(false)" style="margin-right:16px"></i>
    <Input v-model="searchParams.name" @on-enter="getAllResource(true)" :placeholder="$t('resourceLevel.level_search_name')" style="width: 300px;margin-right:8px" />
    <span> OR</span>
    <Select
      v-model="searchParams.endpoint"
      class="col-md-2"
      filterable
      clearable
      ref="selectObject"
      @on-change="clearObject"
       :placeholder="$t('resourceLevel.level_search_endpoint')"
      :remote-method="getAllObject"
      >
      <Option v-for="item in allObject" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
    </Select>
    <button type="button" :disabled="disabledSearchBtn" class="btn btn-confirm-f" @click="getAllResource(true)">{{$t('button.search')}}</button>
    
    <template v-if="extend">
      <recursive :recursiveViewConfig="resourceRecursive"></recursive>
    </template>
    <template v-else>
      <template v-for="(rr, index) in resourceRecursive">
        <div :key="index">
          <div class="levelClass" @click="activeLevel(rr.guid)">
            <i v-if="!inShowLevel(rr.guid)" class="fa fa-angle-double-down" aria-hidden="true"></i>
            <i v-else class="fa fa-angle-double-up" aria-hidden="true"></i>
            {{rr.display_name}}
            <Tag type="border" color="primary" style="margin-left:8px">{{rr.type}}</Tag>    
          </div>
          <div v-if="inShowLevel(rr.guid)">
            <recursive :recursiveViewConfig="[rr]"></recursive>
          </div>
        </div>
      </template>
    </template>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import recursive from '@/views/monitor-config/resource-recursive'
export default {
  name: '',
  data() {
    return {
      searchParams: {
        name: '',
        endpoint: ''
      },
      extend: false,
      allObject: [],
      resourceRecursive: [],
      activedLevel: [],
      modelConfig: {
        modalId: 'add_panel_Modal',
        modalTitle: 'button.add',
        isAdd: true,
        config: [
          {label: 'field.guid', value: 'guid', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'field.displayName', value: 'display_name', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {label: 'field.type', value: 'type', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          guid: null,
          display_name: null
        }
      },
      id: null
    }
  },
  computed: {
    disabledSearchBtn: function() {
      return !this.searchParams.name && !this.searchParams.endpoint
    }
  },
  created () {
    this.$root.$eventBus.$on('updateResource', () => {
      this.getAllResource()
    })
  },
  mounted () {
    this.getAllResource()
    this.getAllObject()
  },
  methods: {
    clearObject () {
      this.getAllObject()
      this.getAllResource(true)
    },
    getAllObject (query='.') {
      let params = {
        search: query
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/dashboard/search', params, (responseData) => {
        this.allObject = []
        responseData.forEach((item) => {
            if (item.id !== -1) {
              this.allObject.push({
                ...item,
                value: item.id
              })
            }
          })
        this.isAssociatedObject = true
      })
    },
    activeLevel (guid) {
      if (this.activedLevel.includes(guid)) {
        const index = this.activedLevel.findIndex(item => item === guid)
        this.activedLevel.splice(index, 1)
      } else {
        this.activedLevel.push(guid)
      }
    },
    inShowLevel (guid) {
      return this.activedLevel.includes(guid) || false
    },
    getAllResource (extend = false) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getAll, this.searchParams, (responseData) => {
        this.resourceRecursive = responseData
        this.extend = extend
      })
    },
    addPanel () {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_panel_Modal').modal('show')
    },
    addPost () {
      const params = this.modelConfig.addRow
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/panel/add', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_panel_Modal').modal('hide')
        this.getAllResource()
      })
    }
  },
  components: {
    recursive
  },
}
</script>

<style scoped lang="less">
 .fa {
   margin-left: 4px;
   padding: 4px 6px;
   border-radius: 4px;
 }
 .levelClass {
    font-size: 14px;
    border: 1px solid @gray-d;
    padding: 4px 8px;
    border-radius: 4px;
    margin: 6px;
 }
</style>
