<template>
  <div class=" ">
    <ul>
      <li v-for="(item, itemIndex) in recursiveViewConfig" class="tree-border" :key="itemIndex">
        <div @click="hide(itemIndex)" class="tree-title" :style="stylePadding">
          <div style="display:flex;justify-content: space-between;">
            <div>
              <h5>{{item.display_name}}</h5>
            </div>
            <div>
              <button class="btn-cancle-f btn-small" @click="associatedRole(item)">配置关联角色</button>
              <button class="btn-cancle-f btn-small" @click="associatedObject(item)">配置关联对象</button>
              <button class="btn-cancle-f btn-small" @click="alarmCallback(item)">告警回调</button>
              <i class="fa fa-plus" aria-hidden="true" @click="addPanel(item)"> </i>
              <i class="fa fa-pencil" @click="editPanal(item)" aria-hidden="true"></i>
              <i class="fa fa-trash-o" @click="deletePanal(item)" aria-hidden="true"></i>
            </div>
          </div>
        </div>
        <transition name="fade">
          
          <!-- v-show="item._isShow" -->
          <div>
            <recursive
            :increment="count"
            v-if="item.children"
            :recursiveViewConfig="item.children"></recursive>
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
      title="节点信息">
      <Form :model="currentData" label-position="left" :label-width="60">
        <FormItem :label="$t('field.guid')">
            <Input v-model="currentData.guid" :disabled="!isAdd"></Input>
        </FormItem>
        <FormItem :label="$t('field.displayName')">
            <Input v-model="currentData.display_name"></Input>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancle-f" @click="isEditPanal = false">取消</button>
        <button class="btn-confirm-f" @click="savePanal">保存</button>
      </div>
    </Modal>
    <!-- 关联角色 -->
    <Modal
      label-colon
      v-model="isAssociatedRole"
      title="关联角色">
      <Form :model="currentData" label-position="left" :label-width="60">
        <FormItem label="角色">
          <Select v-model="selectedRole" multiple>
            <Option v-for="item in allRole" :value="item.value" :key="item.value">{{ item.name }}</Option>
          </Select>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancle-f" @click="isAssociatedRole = false">取消</button>
        <button class="btn-confirm-f" @click="saveAssociatedRole">保存</button>
      </div>
    </Modal>

    <!-- 关联对象 -->
    <Modal
      label-colon
      v-model="isAssociatedObject"
      title="关联对象">
      <Form :model="currentData" label-position="left" :label-width="60">
        <FormItem label="对象">
          <Select v-model="selectedObject" multiple>
            <Option v-for="item in allObject" :value="item.option_value" :key="item.option_value">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancle-f" @click="isAssociatedObject = false">取消</button>
        <button class="btn-confirm-f" @click="saveAssociatedObject">保存</button>
      </div>
    </Modal>
    <!-- 告警回调 -->
    <Modal
      label-colon
      v-model="isAlarmCallback"
      title="告警回调">
      <Form label-position="left" :label-width="80">
        <FormItem label="告警异常">
          <Select v-model="selectedFiring">
            <Option v-for="item in allFiring" :value="item.option_text" :key="item.option_text+'ab'">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
        <FormItem label="告警恢复">
          <Select v-model="selectedRecover">
            <Option v-for="item in allRecover" :value="item.option_text" :key="item.option_text+'cd'">{{ item.option_text }}</Option>
          </Select>
        </FormItem>
      </Form>
      <div slot="footer">
        <button class="btn-cancle-f" @click="isAlarmCallback = false">取消</button>
        <button class="btn-confirm-f" @click="saveAlarmCallback">保存</button>
      </div>
    </Modal>
  </div>
</template>

<script>
import { EventBus } from "@/assets/js/event-bus.js"
export default {
  name: 'recursive',
  data() {
    return {
      isEditPanal: false,
      isAdd: true,
      parentPanal: null,
      currentData: {
        guid: null,
        display_name: null
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
      alarmCallbackata: null
    }
  },
  props:{
    recursiveViewConfig:{
      type: Array
    },
    increment:{
      type:Number,
      default:0
    }
  },
  computed:{
    count () {
      var c = this.increment
      return ++c
    },
    stylePadding(){
      return {
        'padding-left':this.count * 16 + 'px'
      }
    }
  },
  created () {
    // if (!this.recursiveViewConfig) {
    //   return
    // }
    // this.recursiveViewConfig.map((_) =>{
    //   _._isShow = true
    // }) 
  },
  methods: {
    hide (index) {
      this.recursiveViewConfig[index]._isShow = !this.recursiveViewConfig[index]._isShow
      this.$set(this.recursiveViewConfig, index, this.recursiveViewConfig[index])
    },
    getAllResource () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceLevel.getAll, '', (responseData) => {
        this.resourceRecursive = responseData
      })
    },
    addPanel (panalData) {
      this.parentPanal = panalData.guid
      this.currentData = {
        guid: null,
        display_name: null
      }
      this.isAdd = true
      this.isEditPanal = true
    },
    deletePanal (panalData) {
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/org/panel/delete', params, () => {
        this.$Message.success(this.$t('tips.success'))
        EventBus.$emit("updateResource", '')
      })
    },
    editPanal (panalData) {
      this.isAdd = false
      this.currentData = panalData
      this.isEditPanal = true
    },
    savePanal () {
      let params = JSON.parse(JSON.stringify(this.currentData))
      let api = 'alarm/org/panel/edit'
      if (this.isAdd) {
        api = 'alarm/org/panel/add'
        params.parent = this.parentPanal
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.isEditPanal = false
        EventBus.$emit("updateResource", '')
      })
    },
    associatedRole (panalData) {
      this.parentPanal = panalData.guid
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/org/role/get', params, (responseData) => {
        this.selectedRole = []
        responseData.forEach((_) => {
          this.selectedRole.push(_.id)
        })
        this.getAllRole()
      })
    },
    getAllRole () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'user/role/list?page=1&size=1000', '', (responseData) => {
        this.allRole = responseData.data.map((_) => {
          return {
            ..._,
            value: _.id
          }
        })
        this.isAssociatedRole = true
      })
    },
    saveAssociatedRole () {
      let params = {
        "guid": this.parentPanal,
        "role_id": this.selectedRole
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/org/role/update', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.isAssociatedRole = false
      })
    },

    associatedObject (panalData) {
      this.parentPanal = panalData.guid
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/org/endpoint/get', params, (responseData) => {
        this.selectedObject = []
        responseData.forEach((_) => {
          this.selectedObject.push(_.option_value)
        })
        this.getAllObject()
      })
    },
    getAllObject () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'dashboard/search?search=.', '', (responseData) => {
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
    saveAssociatedObject () {
      let params = {
        "guid": this.parentPanal,
        "endpoint": this.selectedObject
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/org/endpoint/update', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.isAssociatedObject = false
      })
    },

    alarmCallback (panalData) {
      this.parentPanal = panalData.guid
      const params = {
        guid: panalData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/org/callback/get', params, (responseData) => {
        this.alarmCallbackata = responseData
        this.selectedFiring = responseData.firing_callback.find((_) => {
          return _.active === true
        }).option_text
        this.allFiring = responseData.firing_callback

        this.selectedRecover = responseData.recover_callback.find((_) => {
          return _.active === true
        }).option_text
        this.allRecover = responseData.recover_callback
        this.isAlarmCallback = true
      })
    },
    saveAlarmCallback () {
      const selectedFiring_choiced = this.alarmCallbackata.firing_callback.find((_) => {
        return _.option_text === this.selectedFiring
      })

      const selectedRecover_choiced = this.alarmCallbackata.recover_callback.find((_) => {
        return _.option_text === this.selectedRecover
      })
      let params = {
        "guid": this.parentPanal,
        firing_callback_name: selectedFiring_choiced.option_text,
        firing_callback_key: selectedFiring_choiced.option_value,
        recover_callback_name: selectedRecover_choiced.option_text,
        recover_callback_key: selectedRecover_choiced.option_value
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/org/callback/update', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.isAlarmCallback = false
      })
    }
  },
  components: {
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
</style>
