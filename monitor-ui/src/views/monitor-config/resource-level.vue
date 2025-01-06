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
        clearable
        @on-change="debounceGetAllResource"
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
        @on-query-change="debounceGetAllObject"
      >
        <Option v-for="item in allObject" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
      </Select>
      <Button type="success" class='add-content-item' @click="addPanel">{{ $t('m_add') }}</Button>
    </div>
    <recursive class='recursive-content' :recursiveViewConfig="resourceRecursive"></recursive>
    <Page
      class="table-pagination"
      :total="pagination.total"
      @on-change="(e) => {pagination.page = e; this.getAllResource()}"
      @on-page-size-change="(e) => {pagination.size = e; this.getAllResource()}"
      :current="pagination.page"
      :page-size="pagination.size"
      show-total
      show-sizer
    />
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import {debounce} from 'lodash'
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
      id: null,
      pagination: {
        total: 0,
        size: 10,
        page: 1
      }
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
      this.resetPagination()
      this.searchParams.name = ''
      this.searchParams.endpoint = ''
      this.getAllResource()
    },
    clearObject() {
      this.getAllObject()
      this.resetPagination()
      this.getAllResource(true)
    },
    debounceGetAllObject: debounce(function (tempQuery) {
      const query = tempQuery ? tempQuery : '.'
      this.getAllObject(query)
    }, 500),
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
    debounceGetAllResource: debounce(function () {
      this.resetPagination()
      this.getAllResource()
    }, 500),
    getAllResource(extend = false) {
      const params = {
        ...this.searchParams,
        pageSize: this.pagination.size,
        startIndex: this.pagination.size * (this.pagination.page - 1)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getAll, params, responseData => {
        this.resourceRecursive = responseData.contents || []
        this.pagination.total = responseData.pageInfo.totalRows
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
        this.resetPagination()
        this.getAllResource()
      })
    },
    resetPagination() {
      this.pagination.size = 10
      this.pagination.page = 1
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
 .table-pagination {
  position: fixed;
  right: 10px;
  bottom: 20px;
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
