<template>
  <div class="monitor-resource-level">
    <div class="content-seatch">
      <!-- <i class="fa fa-refresh" aria-hidden="true" @click="getAllResource(false)" style="margin-right:16px"></i> -->
      <RadioGroup
        v-model="searchParams.type"
        @on-change="handleTypeChange"
        type="button"
        button-style="solid"
        style="margin-right: 5px"
      >
        <Radio label="group">{{ $t('m_field_resourceLevel') }}</Radio>
        <Radio label="endpoint">{{ $t('m_tableKey_endpoint') }}</Radio>
      </RadioGroup>
      <Input
        v-if="searchParams.type === 'group'"
        v-model="searchParams.name"
        @on-change="getAllResource(true)"
        :placeholder="$t('m_resourceLevel_level_search_name')"
        style="width: 300px;margin-right:8px"
      />
      <Select
        v-if="searchParams.type === 'endpoint'"
        v-model="searchParams.endpoint"
        style="width: 300px;margin-right:8px"
        filterable
        clearable
        ref="selectObject"
        @on-change="clearObject"
        :placeholder="$t('m_resourceLevel_level_search_endpoint')"
        :remote-method="getAllObject"
      >
        <Option v-for="item in allObject" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
      </Select>
      <Button type="success" class='add-content-item' @click="addPanel">{{ $t('m_add') }}</Button>
    </div>
    <recursive class='recursive-content' :recursiveViewConfig="resourceRecursive"></recursive>
    <!-- <template v-if="extend">
      <recursive :recursiveViewConfig="resourceRecursive"></recursive>
    </template>
    <template v-else>
      <template v-for="(rr, index) in resourceRecursive">
        <div :key="index">
          <div class="levelClass" @click="activeLevel(rr.guid)">
            <i v-if="!inShowLevel(rr.guid)" class="fa fa-angle-double-down" aria-hidden="true"></i>
            <i v-else class="fa fa-angle-double-up" aria-hidden="true"></i>
            {{rr.display_name}}
            <TagShow :tagName='rr.type' />
          </div>
          <div v-if="inShowLevel(rr.guid)">
            <recursive :recursiveViewConfig="[rr]"></recursive>
          </div>
        </div>
      </template>
    </template> -->
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
        type: 'group',
        name: '',
        endpoint: ''
      },
      extend: false,
      allObject: [],
      resourceRecursive: [],
      activedLevel: [],
      modelConfig: {
        modalId: 'add_panel_Modal',
        modalTitle: 'm_button_add',
        isAdd: true,
        config: [
          {
            label: 'm_field_guid',
            value: 'guid',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_field_displayName',
            value: 'display_name',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          },
          {
            label: 'm_field_type',
            value: 'type',
            placeholder: 'm_tips_required',
            v_validate: 'required:true',
            disabled: false,
            type: 'text'
          }
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
    disabledSearchBtn() {
      return !this.searchParams.name && !this.searchParams.endpoint
    }
  },
  created() {
    this.$root.$eventBus.$on('updateResource', () => {
      this.getAllResource()
    })
  },
  mounted() {
    this.getAllResource()
    this.getAllObject()
  },
  methods: {
    handleTypeChange() {
      this.searchParams.name = ''
      this.searchParams.endpoint = ''
      this.getAllResource()
    },
    clearObject() {
      this.getAllObject()
      this.getAllResource(true)
    },
    getAllObject(query='.') {
      const params = {
        search: query
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/dashboard/search', params, responseData => {
        this.allObject = []
        responseData.forEach(item => {
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
    activeLevel(guid) {
      if (this.activedLevel.includes(guid)) {
        const index = this.activedLevel.findIndex(item => item === guid)
        this.activedLevel.splice(index, 1)
      } else {
        this.activedLevel.push(guid)
      }
    },
    inShowLevel(guid) {
      return this.activedLevel.includes(guid) || false
    },
    getAllResource(extend = false) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getAll, this.searchParams, responseData => {
        this.resourceRecursive = responseData
        this.extend = extend
      })
    },
    addPanel() {
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_panel_Modal').modal('show')
    },
    addPost() {
      const params = this.modelConfig.addRow
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/panel/add', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
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
.recursive-content {
  max-height: ~'calc(100vh - 200px)';
  overflow-y: auto;
}
.content-seatch {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  align-items: center;
  margin: 15px 0;
  .add-content-item {
    margin-left: auto;
  }
}
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
<style lang="less">
.monitor-resource-level {
  .ivu-radio-group-button .ivu-radio-wrapper-checked {
    background: #2d8cf0;
    color: #fff;
  }
}

</style>
