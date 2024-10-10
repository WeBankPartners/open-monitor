<template>
  <div>
    <ul>
      <li v-for="(item, itemIndex) in recursiveViewConfig" class="tree-border" :key="itemIndex">
        <div @click="hide(itemIndex)" class="tree-title" :style="[addTag(item.fetch_search), stylePadding]">
          <div style="display:flex;justify-content: space-between;">
            <div>
              <span class="title-style">{{item.display_name}}</span>
              <TagShow :tagName='item.type' />
              <!-- <Tag :color="choiceColor(item.type)" class="tag-width">{{item.type}}</Tag> -->
            </div>
            <div>
              <!-- <button class="btn-cancel-f btn-small" @click="alarmReceivers(item)">{{$t('m_button_receiversConfiguration')}}</button> -->
              <Tooltip placement="top" max-width="400" :content="$t('m_resourceLevel_addAssociatedObject')">
                <Button size="small" type="primary" @click="associatedObject(item)">
                  <Icon type="ios-cube" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" max-width="400" :content="$t('m_button_edit')">
                <Button size="small" type="primary" @click="editPanal(item)">
                  <Icon type="md-create" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" max-width="400" :content="$t('m_resourceLevel_addAssociatedRole')">
                <Button size="small" type="warning" @click="associatedRole(item)">
                  <Icon type="md-person" />
                </Button>
              </Tooltip>

              <Tooltip placement="top" max-width="400" :content="$t('m_resourceLevel_alarmCallback')">
                <Button size="small" type="info" @click="alarmCallback(item)">
                  <Icon type="md-warning" />
                </Button>
              </Tooltip>

              <Tooltip placement="top" max-width="400" :content="$t('m_add')">
                <Button size="small" type="success" @click="addPanel(item)">
                  <Icon type="md-add" />
                </Button>
              </Tooltip>

              <Tooltip placement="top" max-width="400" :content="$t('m_button_remove')">
                <Button size="small" class="mr-2" type="error" @click="deleteConfirmModal(item)">
                  <Icon type="md-trash" />
                </Button>
              </Tooltip>
              <!-- <button class="btn-cancel-f btn-small" @click="associatedRole(item)">{{$t('m_resourceLevel_addAssociatedRole') + '77'}}</button> -->
              <!-- <button class="btn-cancel-f btn-small" @click="associatedObject(item)">{{$t('m_resourceLevel_addAssociatedObject')}}</button> -->
              <!-- <button class="btn-cancel-f btn-small" v-if="isPlugin" @click="alarmCallback(item)">{{$t('m_resourceLevel_alarmCallback')}}</button> -->
              <!-- <i class="fa fa-plus" aria-hidden="true" @click="addPanel(item)"> </i> -->
              <!-- <i class="fa fa-pencil" @click="editPanal(item)" aria-hidden="true"></i> -->
              <!-- <i class="fa fa-trash-o" style="color:red" @click="deleteConfirmModal(item)" aria-hidden="true"></i> -->
            </div>
          </div>
        </div>
        <transition name="fade">

          <!-- v-show="item._isShow" -->
          <div>
            <recursive
              :increment="count"
              v-if="item.children"
              :recursiveViewConfig="item.children"
            ></recursive>
            <div>
            </div>
          </div>
        </transition>
      </li>
    </ul>
    <!-- 节点新增、编辑 -->
    <Modal
      label-colon
      v-model="isEditPanal"
      :mask-closable="false"
      :title="$t('m_resourceLevel_levelMsg')"
    >
      <Form :model="currentData" label-position="left" :label-width="60">
        <FormItem :label="$t('m_field_guid')">
          <Input v-model="currentData.guid" :disabled="!isAdd"></Input>
        </FormItem>
        <FormItem :label="$t('m_field_displayName')">
          <Input v-model="currentData.display_name"></Input>
        </FormItem>
        <FormItem :label="$t('m_field_type')">
          <Input v-model="currentData.type"></Input>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancel-f" @click="isEditPanal = false">{{$t('m_button_cancel')}}</button>
        <button class="btn-confirm-f" @click="savePanal">{{$t('m_button_save')}}</button>
      </div>
    </Modal>
    <!-- 关联角色 -->
    <Modal
      label-colon
      v-model="isAssociatedRole"
      :mask-closable="false"
      :title="$t('m_resourceLevel_associatedRole')"
    >
      <Form :model="currentData" label-position="left" :label-width="60">
        <FormItem :label="$t('m_resourceLevel_role')">
          <Select v-model="selectedRole" multiple filterable>
            <Option v-for="item in allRole" :value="item.value" :key="item.value">{{ item.name }}</Option>
          </Select>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancel-f" @click="isAssociatedRole = false">{{$t('m_button_cancel')}}</button>
        <button class="btn-confirm-f" @click="saveAssociatedRole">{{$t('m_button_save')}}</button>
      </div>
    </Modal>

    <!-- 关联对象 -->
    <Modal
      label-colon
      v-model="isAssociatedObject"
      :mask-closable="false"
      :width="550"
      :title="$t('m_resourceLevel_associatedObject')"
    >
      <Form :model="currentData" label-position="right" label-colon :label-width="100">
        <FormItem :label="$t('m_add_object')">
          <Select
            v-model="addObject"
            filterable
            clearable
            multiple
            style="width:300px"
            :placeholder="$t('requestMoreData')"
            :remote-method="getAllObject"
          >
            <Option v-for="(item, index) in allObject" :value="item.option_value" :label="item.option_text" :key="item.option_value">
              <TagShow :list="allObject" name="type" :tagName="item.type" :index="index"></TagShow>
              {{ item.option_text }}</Option>
          </Select>
          <Button @click="addObjectItem">{{$t('m_button_add')}}</Button>
        </FormItem>
        <FormItem :label="$t('m_selected_object')" style="max-height:500px;overflow:auto">
          <template v-for="(obj, objIndex) in selectedObject">
            <Tooltip :key="objIndex" transfer>
              <Tag
                :key="objIndex"
                type="border"
                closable
                @on-close="removeObj(objIndex)"
                color="primary"
              >
                <span style="color:red">{{obj.type}}:</span>
                {{obj.option_text.length > 40 ? obj.option_text.substring(0,40) + '...' : obj.option_text}}
              </Tag>
              <div slot="content" style="white-space: normal;max-width:200px;word-break: break-all;">
                {{obj.option_text}}
              </div>
            </Tooltip>
          </template>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancel-f" @click="isAssociatedObject = false">{{$t('m_button_cancel')}}</button>
        <button class="btn-confirm-f" @click="saveAssociatedObject">{{$t('m_button_save')}}</button>
      </div>
    </Modal>
    <!-- 告警回调 -->
    <Modal
      label-colon
      v-model="isAlarmCallback"
      :mask-closable="false"
      :title="$t('m_resourceLevel_alarmCallback')"
    >
      <Form label-position="left" :label-width="80">
        <FormItem :label="$t('m_resourceLevel_alarmFiring')">
          <Select v-model="selectedFiring" filterable clearable>
            <Option v-for="item in allFiring" :value="item.option_text" :key="item.option_text + 'ab'">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
        <FormItem :label="$t('m_resourceLevel_alarmRecover')">
          <Select v-model="selectedRecover" filterable clearable>
            <Option v-for="item in allRecover" :value="item.option_text" :key="item.option_text + 'cd'">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancel-f" @click="isAlarmCallback = false">{{$t('m_button_cancel')}}</button>
        <button class="btn-confirm-f" @click="saveAlarmCallback">{{$t('m_button_save')}}</button>
      </div>
    </Modal>
    <!-- 告警接收人 -->
    <Modal
      label-colon
      v-model="isAlarmReceivers"
      :mask-closable="false"
      :title="$t('m_button_receivers')"
    >
      <div>
        <label style="width:110px">{{$t('m_button_receiversSelect')}}:</label>
        <Select v-model="selectRole" multiple filterable style="width:280px">
          <Option v-for="item in roleList" :value="item.id" :key="item.id">
            {{item.display_name}}</Option>
        </Select>
        <button class="btn-cancel-f" @click="addSelectReceivers">{{$t('m_button_add')}}</button>
      </div>
      <div style="margin: 8px 0">
        <label style="width:110px">{{$t('m_button_receiversInput')}}:</label>
        <input
          v-model="inputRole"
          type="text"
          :placeholder="$t('m_button_receiversInputTip')"
          class="form-control search-input c-dark"
        />
        <button class="btn-cancel-f" @click="addInputReceivers">{{$t('m_button_add')}}</button>
      </div>
      <div slot="footer">
        <button class="btn-cancel-f" @click="isAlarmReceivers = false">{{$t('m_button_cancel')}}</button>
        <button class="btn-confirm-f" @click="saveAlarmReceivers">{{$t('m_button_save')}}</button>
      </div>
      <template>
        <Tag
          v-for="(receiver, receiverIndex) in tagInfo"
          :key="receiverIndex"
          type="border"
          color="primary"
          @on-close="closeTag(receiverIndex)"
          closable
        >
          {{receiver.value}}
        </Tag>
      </template>
    </Modal>
    <!-- <Modal
      v-model="isShowWarning"
      :title="$t('m_delConfirm_title')"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal> -->
    <Modal v-model="confirmModal.isShowConfirmModal" width="900">
      <div>
        <Icon :size="28" :color="'#f90'" type="md-help-circle" />
        <span class="confirm-msg">{{ $t('m_delConfirm_title') }}</span>
      </div>
      <div>
        <p style="margin-left: 10px;margin-top: 22px;">{{ $t(this.confirmModal.message) }}</p>
      </div>
      <div slot="footer">
        <span style="color:#ed4014;float: left;text-align:left">
          <Checkbox v-model="confirmModal.check">{{ $t('dangerous_confirm_tip') }}</Checkbox>
        </span>
        <Button @click="cancelConfirmModal">{{$t('m_button_cancel')}}</Button>
        <Button
          @click="getDeleteData"
          :disabled="!confirmModal.check"
          type="warning"
        >{{ $t('m_button_confirm') }}</Button>
      </div>
    </Modal>
    <Modal
      v-model="doubleConfirm.isShow"
      :title="$t('m_delConfirm_title')"
      @on-ok="ok"
      @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:10px">
        <div style="color:#ed4014">{{$t('delete_follow')}}:</div>
        <p v-for="msg in doubleConfirm.warningData" :key="msg">{{msg}}</p>
      </div>
    </Modal>
  </div>
</template>

<script>
import {randomColor} from '@/assets/config/common-config'
import TagShow from '@/components/Tag-show.vue'
export default {
  name: 'recursive',
  data() {
    return {
      isShowWarning: false,
      isEditPanal: false,
      isAdd: true,
      parentPanal: null,
      currentData: {
        guid: null,
        display_name: null,
        type: null
      },

      isAssociatedRole: false,
      selectedRole: [],
      allRole: [],

      isAssociatedObject: false,
      selectedObject: [],
      allObject: [],

      isAlarmCallback: false,
      selectedFiring: '',
      allFiring: [],
      selectedRecover: '',
      allRecover: [],
      alarmCallbackata: null,

      isAlarmReceivers: false,
      currentRecursive: null,
      inputRole: '',
      selectRole: [],
      roleList: [],
      tagInfo: [],
      confirmModal: {
        isShowConfirmModal: false,
        check: false,
        message: 'resource_delete_tip',
      },
      doubleConfirm: {
        isShow: false,
        warningData: []
      },
      addObject: []
    }
  },
  props: {
    recursiveViewConfig: {
      type: Array
    },
    increment: {
      type: Number,
      default: 0
    }
  },
  computed: {
    count() {
      let c = this.increment
      return ++c
    },
    stylePadding(){
      return {
        'padding-left': this.count * 16 + 'px'
      }
    },
    isPlugin() {
      return window.request ? true: false
    }
  },
  created() {
    // if (!this.recursiveViewConfig) {
    //   return
    // }
    // this.recursiveViewConfig.map((_) =>{
    //   _._isShow = true
    // })
  },
  methods: {
    addObjectItem() {
      this.addObject.forEach(obj => {
        const isExist = this.selectedObject.find(s => s.option_value === obj)
        if (!isExist) {
          const find = this.allObject.find(a => a.option_value === obj)
          this.selectedObject.push(find)
        }
      })
      this.addObject = []
    },
    removeObj(index) {
      this.selectedObject.splice(index, 1)
    },
    cancelConfirmModal() {
      this.confirmModal.isShowConfirmModal = false
      this.confirmModal.check = false
    },
    addTag(fetch_search) {
      if (fetch_search) {
        return {background: '#c9dded'}
      }
    },
    alarmReceivers(item) {
      this.currentRecursive = item.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.groupManagement.allRoles.api, '', responseData => {
        this.roleList = responseData.data
        this.getReceivers(item)
      })
    },
    getReceivers(item) {
      this.tagInfo = []
      const params ={
        guid: item.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getReceivers, params, responseData => {
        ['mail', 'phone'].forEach(type => {
          responseData[type].forEach(item => {
            if (!item) {
              return
            }
            this.tagInfo.push({
              id: this.guid(),
              type,
              dispalyName: item,
              value: item
            })
          })
        })
        this.isAlarmReceivers = true
      })

    },
    addSelectReceivers() {
      this.selectRole.forEach(r => {
        const isSelected = this.tagInfo.findIndex(tag => tag.id === r)
        if (isSelected > -1) {
          return
        }
        const role = this.roleList.find(rl => r === rl.id)
        this.tagInfo.push({
          id: role.id,
          type: 'mail',
          dispalyName: role.name,
          value: role.email
        })
      })
      this.selectRole = []
    },
    addInputReceivers() {
      const regx_email = /^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$/gi
      const regs_phone = /^1[345678]\d{9}$/
      const regx_email_res = regx_email.test(this.inputRole)
      const regs_phone_res = regs_phone.test(this.inputRole)
      if (regx_email_res) {
        this.tagInfo.push({
          id: this.guid(),
          type: 'mail',
          dispalyName: this.inputRole,
          value: this.inputRole
        })
        this.inputRole = ''
        return
      }
      if (regs_phone_res) {
        this.tagInfo.push({
          id: this.guid(),
          type: 'phone',
          dispalyName: this.inputRole,
          value: this.inputRole
        })
        this.inputRole = ''
        return
      }
      this.$Message.warning('Wrong Format !')
    },
    closeTag(index) {
      this.tagInfo.splice(index, 1)
    },
    saveAlarmReceivers() {
      const params = {
        guid: this.currentRecursive,
        mail: [],
        phone: []
      }
      for (const tag of this.tagInfo) {
        params[tag.type].push(tag.value)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.resourceLevel.updateReceivers, params, () => {
        this.isAlarmReceivers = false
        this.$Message.success(this.$t('m_tips_success'))
      })
    },
    guid() {
      return 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        const r = Math.random()*16|0, v = c === 'x' ? r : (r&0x3|0x8)
        return v.toString(16)
      })
    },
    choiceColor(type) {
      const cacheColor = this.$root.$store.state.cacheTagColor
      let color = ''
      // eslint-disable-next-line no-prototype-builtins
      if (Object.keys(cacheColor).includes(type)) {
        color = cacheColor[type]
      }
      else {
        color = randomColor[this.count]
        cacheColor[type] = randomColor[this.count]
        this.$root.$store.commit('cacheTagColor', cacheColor)
      }
      return color
    },
    hide(index) {
      this.recursiveViewConfig[index]._isShow = !this.recursiveViewConfig[index]._isShow
      this.$set(this.recursiveViewConfig, index, this.recursiveViewConfig[index])
    },
    getAllResource() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getAll, '', responseData => {
        this.resourceRecursive = responseData
      })
    },
    addPanel(panalData) {
      this.parentPanal = panalData.guid
      this.currentData = {
        guid: null,
        display_name: null,
        type: null
      }
      this.isAdd = true
      this.isEditPanal = true
    },
    deleteConfirmModal(rowData) {
      this.selectedData = rowData
      // this.isShowWarning = true
      this.confirmModal.isShowConfirmModal = true
    },
    getDeleteData() {
      const params = {
        guid: this.selectedData.guid,
        force: 'no'
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/panel/delete', params, res => {
        this.confirmModal.isShowConfirmModal = false
        if (res.length > 0) {
          this.doubleConfirm.isShow= true
          this.doubleConfirm.warningData= res
        }
        else {
          this.$root.$eventBus.$emit('updateResource', '')
        }
      })
    },
    ok() {
      this.deletePanal(this.selectedData)
    },
    cancel() {
      this.doubleConfirm.isShow= false
    },
    // deleteConfirm (panalData) {
    //   this.parentPanal =  panalData.guid
    //   this.$delConfirm({
    //     msg: panalData.guid,
    //     callback: () => {
    //       this.deletePanal(panalData)
    //     }
    //   })
    // },
    deletePanal() {
      const params = {
        guid: this.selectedData.guid,
        force: 'yes'
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/panel/delete', params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.doubleConfirm.isShow= false
        this.$root.$eventBus.$emit('updateResource', '')
      })
    },
    editPanal(panalData) {
      this.isAdd = false
      this.currentData = panalData
      this.isEditPanal = true
    },
    savePanal() {
      const params = JSON.parse(JSON.stringify(this.currentData))
      let api = '/monitor/api/v1/alarm/org/panel/edit'
      if (this.isAdd) {
        api = '/monitor/api/v1/alarm/org/panel/add'
        params.parent = this.parentPanal
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', api, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.isEditPanal = false
        this.$root.$eventBus.$emit('updateResource', '')
      })
    },
    associatedRole(panalData) {
      this.parentPanal = panalData.guid
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/alarm/org/role/get', params, responseData => {
        this.selectedRole = []
        responseData.forEach(_ => {
          this.selectedRole.push(_.id)
        })
        this.getAllRole()
      })
    },
    getAllRole() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/user/role/list?page=1&size=1000', '', responseData => {
        this.allRole = responseData.data.map(_ => ({
          ..._,
          value: _.id
        }))
        this.isAssociatedRole = true
      })
    },
    saveAssociatedRole() {
      const params = {
        guid: this.parentPanal,
        role_id: this.selectedRole
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/role/update', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.isAssociatedRole = false
      })
    },
    associatedObject(panalData) {
      this.parentPanal = panalData.guid
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/alarm/org/endpoint/get', params, responseData => {
        this.selectedObject = responseData
        this.getAllObject()
      })
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
    saveAssociatedObject() {
      const params = {
        guid: this.parentPanal,
        endpoint: this.selectedObject.map(item => item.option_value)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/endpoint/update', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.isAssociatedObject = false
      })
    },
    alarmCallback(panalData) {
      this.parentPanal = panalData.guid
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/alarm/org/callback/get', params, responseData => {
        this.alarmCallbackata = responseData
        this.selectedFiring = responseData.firing_callback.find(_ => _.active === true).option_text
        this.allFiring = responseData.firing_callback

        this.selectedRecover = responseData.recover_callback.find(_ => _.active === true).option_text
        this.allRecover = responseData.recover_callback
        this.isAlarmCallback = true
      })
    },
    saveAlarmCallback() {
      const selectedFiring_choiced = this.alarmCallbackata.firing_callback.find(_ => _.option_text === this.selectedFiring)

      const selectedRecover_choiced = this.alarmCallbackata.recover_callback.find(_ => _.option_text === this.selectedRecover)
      const params = {
        guid: this.parentPanal,
        firing_callback_name: selectedFiring_choiced.option_text,
        firing_callback_key: selectedFiring_choiced.option_value,
        recover_callback_name: selectedRecover_choiced.option_text,
        recover_callback_key: selectedRecover_choiced.option_value
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/org/callback/update', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.isAlarmCallback = false
      })
    }
  },
  components: {
    TagShow
  }
}
</script>

<style scoped lang="less">
  .fa {
   font-size: 16px;
    padding: 6px;
  }

  ul {
    padding: 0;
    margin: 0;
    list-style: none;
  }

  .tree-menu {
    height: 100%;
    padding: 0px 12px;
    border-right: 1px solid #e6e9f0;
  }

  .tree-menu-comm span {
    display: block;
    font-size: 12px;
    position: relative;
  }

  .tree-menu-comm span strong {
    display: block;
    width: 82%;
    position: relative;
    line-height: 22px;
    padding: 2px 0;
    padding-left: 5px;
    color: #161719;
    font-weight: normal;
  }

  .tree-title {
    margin-top: 1px;
    // cursor: pointer;
    color: @blue-2;
  }
  .tree-border {
    border: 1px solid #9966;
    border-right: none;
    padding: 4px 0 4px 4px;
    margin: 4px 0 4px 4px;
  }
  .title-style {
    font-size: 14px;
    font-weight: 500;
    padding-right: 8px;
    vertical-align: middle;
  }
  .ivu-form-item {
    margin-bottom: 0;
  }

  .search-input {
    height: 32px;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    display: inline-block;
    border-radius: 4px;
    width: 280px;
  }

  .search-input:focus {
    outline: 0;
    // border-color: #57a3f3;
  }
  .is-danger {
    color: red;
    margin-bottom: 0px;
  }
</style>
