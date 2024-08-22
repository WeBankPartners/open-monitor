<template>
  <div class="modal-component">
    <div
      class="modal fade"
      :id="modelConfig.modalId ? modelConfig.modalId : 'add_edit_Modal'"
      aria-hidden="false"
      data-backdrop="static"
      role="dialog"
      aria-labelledby="myModalLabel"
    >
      <div class="modal-dialog" :style="modelConfig.modalStyle" role="document">
        <div class="modal-content c-dark">
          <div class="modal-header c-dark">
            <h4 class="modal-title" v-if="modelConfig.modalTitle">{{$t(modelConfig.modalTitle)}}
              <template v-if="!modelConfig.isAdd && interceptParams()">--
                【<span style="color: red">{{interceptParams()}}</span>】
              </template>
            </h4>
            <h4 class="modal-title" v-if="!modelConfig.modalTitle">{{$t(modelConfig.modalTitle) + '--'}}
              <label v-if="modelConfig.isAdd">
                <span>{{$t('m_button_add')}}</span>
              </label>
              <label v-else>
                <span>{{$t('m_button_edit')}}</span>
                <span style="color: red;">
                  {{interceptParams()}}
                </span>】
              </label>
            </h4>
            <icon type="close-round" data-dismiss="modal"/>
          </div>
          <div class="modal-body">
            <form class="">

              <div v-for="(item, index) in modelConfig.config"  class="params-each" :key="index">
                <label class="col-md-2 label-name" v-if="item.type !== 'slot' && isHide(item.hide)" :class="item.type === 'select' ? 'lable-name-select' : ''">
                  {{$t(item.label)}}:</label>
                <input v-if="item.type === 'text' && isHide(item.hide)"
                       v-model="modelConfig.addRow[item.value]"
                       autocomplete="off"
                       :disabled="modelConfig.isAdd ? false : item.disabled"
                       :placeholder="$t(item.placeholder)"
                       type="text"
                       v-validate="item.v_validate"
                       :name="item.value"
                       :maxlength="item.max"
                       :class="{'red-border': veeErrors.has(item.value)}"
                       class="col-md-7 form-control model-input"
                />
                <input v-if="item.type === 'number' && isHide(item.hide)"
                       v-model.number="modelConfig.addRow[item.value]"
                       autocomplete="off"
                       :disabled="modelConfig.isAdd ? false : item.disabled"
                       :placeholder="$t(item.placeholder)"
                       type="text"
                       :name="item.value"
                       v-validate="item.v_validate"
                       class="col-md-7 form-control model-input"
                />
                <input v-if="item.type === 'password' && isHide(item.hide)"
                       v-model="modelConfig.addRow[item.value]"
                       autocomplete="new-password"
                       :disabled="modelConfig.isAdd ? false : item.disabled"
                       type="password"
                       :placeholder="$t(item.placeholder)"
                       :name="item.value"
                       v-validate="item.v_validate"
                       class="col-md-7 form-control model-input"
                />
                <input v-if="item.type === 'checkbox' && isHide(item.hide)"
                       v-model="modelConfig.addRow[item.value]"
                       autocomplete="off"
                       :type="item.type"
                       :name="item.value"
                       :disabled="modelConfig.isAdd ? false : item.disabled"
                       class="checkbox"
                />
                <i-switch v-if="item.type === 'switch' && isHide(item.hide)"
                          v-model="modelConfig.addRow[item.value]"
                />
                <textarea v-if="item.type === 'textarea' && isHide(item.hide)"
                          v-model="modelConfig.addRow[item.value]"
                          :placeholder="$t(item.placeholder)"
                          v-validate="item.v_validate"
                          :class="item.isError ? 'red-border textareaSty' : 'textareaSty'"
                          :disabled="item.disabled"
                          :name="item.value"
                ></textarea>
                <!-- <v-select v-if="item.type === 'select' && isHide(item.hide)"
                          v-model="modelConfig.v_select_configs[item.value]"
                          :disabled="modelConfig.isAdd ? false : item.disabled"
                          label='name'
                          :name="item.showName? item.showName : 'name'"
                          :multiple="item.multiple? item.multiple : false"
                          class="col-md-7 v-selectss "
                          :options="modelConfig.v_select_configs[item.option]">
                </v-select> -->
                <select
                  v-if="item.type === 'multiSelect' && isHide(item.hide)"
                  :disabled="modelConfig.isAdd ? false : item.disabled"
                  multiple
                  filterable
                  clearable
                  class="col-md-7 v-selectss"
                  v-model="modelConfig.addRow[item.value]"
                >
                  <option
                    v-for="item in modelConfig.v_select_configs[item.option]"
                    :value="item.value"
                    :key="item.value"
                  >{{ item.label }}</option>
                </select>
                <Select
                  v-if="item.type === 'select' && isHide(item.hide)"
                  :disabled="modelConfig.isAdd ? false : item.disabled"
                  filterable
                  clearable
                  class="col-md-7 v-selectss"
                  v-model="modelConfig.addRow[item.value]"
                >
                  <Option
                    v-for="item in modelConfig.v_select_configs[item.option]"
                    :value="item.value"
                    :key="item.value"
                  >{{ item.label }}</Option>
                </Select>

                <slot v-if="item.type === 'slot' && isHide(item.hide) && !item.ishide" :name="item.name"></slot>
                <label class="required-tip  isRequired_s" v-if="isRequired_s(item.v_validate) && isHide(item.hide)">*</label>

                <label class="required-tip" v-if="isRequired(item.v_validate) && isHide(item.hide)">*</label>

                <poptip v-if="item.tips && isHide(item.hide) && item.type !== 'switch'" word-wrap content="" trigger="hover" placement="bottom" :delay="500">
                  <i class="fa fa-question-circle-o question-circle" aria-hidden="true"></i>
                  <div slot="content" class="slot-content-question-circle">
                    <div v-html="item.tips"></div>
                  </div>
                </poptip>
                <poptip v-if="item.tips && isHide(item.hide) && item.type === 'switch'" word-wrap content="" trigger="hover" placement="right" :delay="500" style="transform: translateY(2px);">
                  <i class="fa fa-question-circle-o question-circle" aria-hidden="true"></i>
                  <div slot="content" class="slot-content-question-circle">
                    <div v-html="item.tips"></div>
                  </div>
                </poptip>

                <label v-show="veeErrors.has(item.value) && isHide(item.hide)" class="col-md-7 help is-danger">{{$t(item.label)}} {{veeErrors.first(item.value)}}</label>
                <label v-if="(item.type === 'select' || item.type === 'multiSelect' || item.type === 'textarea') && item.isError" class="col-md-7 help is-danger">{{$t(item.label)}} {{$t('m_tips_required')}}</label>
              </div>
            </form>
          </div>
          <template v-if="!modelConfig.noBtn">
            <div class="model-footer-f c-dark" v-if="!modelConfig.modalFooter">
              <Button data-dismiss="modal">{{$t('m_button_cancel')}}</Button>
              <Button type="primary" v-if="!modelConfig.saveFunc" @click="save(modelConfig.isAdd)">{{$t('m_button_save')}}</Button>
              <Button type="primary" v-if="modelConfig.saveFunc" @click="customFunc(modelConfig.saveFunc)">{{$t('m_button_save')}}</Button>
            </div>
            <div class="model-footer-f c-dark" v-if="modelConfig.modalFooter">
              <Button data-dismiss="modal">{{$t('m_button_cancel')}}</Button>
              <template v-for="(item, index) in modelConfig.modalFooter">
                <Button type="primary" @click="customFunc(item.Func)" v-if='item.name' :key="index">{{item.name}}</Button>
              </template>
            </div>
          </template>

        </div>
      </div>
    </div>
  </div>
</template>
<script>
import {interceptParams} from '@/assets/js/utils'
export default {
  name: 'modal-component',
  data() {
    return {
      configCopy: this.modelConfig,
      FLAG: false,
    }
  },
  props: ['modelConfig'],
  mounted() {
    const _this = this
    const modalId = !this.$root.$validate.isEmpty(this.modelConfig.modalId) ? 'add_edit_Modal':this.modelConfig.modalId
    this.$root.JQ('#' + modalId).on('hidden.bs.modal', () => {
      // 清理表单验证错误信息
      _this.veeErrors.clear()
      // 清除表单缓存内容  下面把清空switch的数据补全
      this.$root.$validate.emptyJson(_this.modelConfig.addRow)

      // 清除表单缓存的selected数据
      for (const p in _this.modelConfig.v_select_configs) {
        if (p.endsWith('selected')) {
          _this.modelConfig.v_select_configs[p] = null
        }
      }
      // 清除表单selected
      for (let i=0; i<this.modelConfig.config.length; i++){
        // 这里不能清空switch绑定的数据，不然会报错
        if (this.modelConfig.config[i].type === 'switch') {
          this.modelConfig.addRow[this.modelConfig.config[i].value] = false
        }
        if (this.modelConfig.config[i].type === 'textarea' && this.modelConfig.config[i].v_validate) {
          this.modelConfig.config[i].isError = false
        }
        if (this.modelConfig.config[i].type === 'select' && this.modelConfig.config[i].v_validate) {
          this.modelConfig.config[i].isError = false
        }
        if (this.modelConfig.config[i].type === 'slot') {
          const arr = this.modelConfig.config[i].v_validate ? this.modelConfig.config[i].v_validate :[]
          for (let j =0; j<arr.length; j++){
            const key = arr[j].isError
            const value = arr[j].value
            if (arr[j].type === 'select' && !this.modelConfig.v_select_configs[value]) {
              this.modelConfig.v_select_configs[key] = false
            }
          }
        }
      }
    })
  },
  watch: {
  },
  filters: {},
  methods: {
    // 处理提示信息过长问题
    interceptParams() {
      // return interceptParams(this.$parent.modelTip.value, 20)
      return interceptParams(this.$parent.modelTip.value, 20)
    },
    formValidate(){
      return this.$validator.validate().then(result => {
        // result 为false插件验证input没有填写完整,true为验证填写完整
        /** 验证 select是否进行了选填 实例可参照 [manage][authorizations]user-authorized.vue **/
        let flag = true
        for (let i=0; i< this.modelConfig.config.length; i++){
          if (!this.isHide(this.modelConfig.config[i].hide)){
            continue
          }
          if (this.modelConfig.config[i].type === 'textarea' && this.modelConfig.config[i].v_validate){
            const obj = this.modelConfig.config[i]
            if (!this.modelConfig.addRow[obj.value]){
              obj.isError = true
              flag = false
            } else {
              obj.isError = false
            }
          }
          /* ****** 配置里面的select ***** */
          // 配置规则为：在配置type:selcet时，如需校验则添加v_validate: 'required:true',isError: false==>控制错误提示label显示
          // 如果无需校验，则不添加
          if (this.modelConfig.config[i].type === 'select' && this.modelConfig.config[i].v_validate) {
            const obj = this.modelConfig.config[i]
            if (!this.modelConfig.addRow[obj.value]){
              this.modelConfig.config[i].isError = true
              flag = false
            } else {
              this.modelConfig.config[i].isError = false
            }
          }
          if (this.modelConfig.config[i].type === 'multiSelect' && this.modelConfig.config[i].v_validate) {
            const obj = this.modelConfig.config[i]
            if (this.modelConfig.addRow[obj.value] && this.modelConfig.addRow[obj.value].length === 0){
              this.modelConfig.config[i].isError = true
              flag = false
            } else {
              this.modelConfig.config[i].isError = false
            }
          }
          /* ******  slot里面的select  ****** */
          // 配置规则为：在配置type:slot 时，添加 v_validate:[],数组里面存放需要校验的select的配置信息
          // value:绑定值,isError:错误标签显示绑定值，type:select  ===> 如果以后再校验其他类型，再增加判断逻辑
          // {name:'xxxx',type:'slot',v_validate:[{value: 'v_xxx_selected', isError: 'v_xxx_isError', type: 'select'}]}
          // 同时在this.modelConfig.v_select_configs里面定义v_xxx_isError：false
          // slot里面错误提示label显示用 v-if="modelConfig.v_select_configs.v_xxx_isError" 搭配其他具体规则进行组合
          // if(this.modelConfig.config[i].type === 'slot' && !this.modelConfig.config[i].ishide) {
          //   let arr = this.modelConfig.config[i].v_validate ? this.modelConfig.config[i].v_validate :[]
          //   for(let j =0;j<arr.length;j++){
          //     let key = arr[j].isError
          //     let value = arr[j].value
          //     let isMutile = Array.isArray(this.modelConfig.v_select_configs[value]) && this.modelConfig.v_select_configs[value].length<1
          //     if (arr[j].type === 'select' && (isMutile ||!this.modelConfig.v_select_configs[value])) {
          //       this.modelConfig.v_select_configs[key] = true
          //       flag = false
          //     } else {
          //       this.modelConfig.v_select_configs[key] = false
          //     }
          //   }
          // }
        }
        this.modelConfig.config = JSON.parse(JSON.stringify(this.modelConfig.config))
        return result && flag
      })
    },
    save(val) {
      const resultPromise = this.formValidate()
      resultPromise.then(result => {
        if (result){
          if (val) {
            this.$parent.addPost()
          } else {
            this.$parent.editPost()
          }
        }
      })
    },
    // 自定义模态框保存响应函数
    customFunc(func) {
      const val = this.formValidate()
      val.then(result => {
        if (result){
          this.$parent[func]()
        }
      })
    },
    // 控制是否显示必填'*'
    isRequired_s(item){
      if (typeof item==='object') {

        for (let i=0; i<item.length; i++){
          if (item[i].v_validate) {
            if (item[i].v_validate.indexOf('required') > -1) {
              return true
            }
            return false

          }

        }
      }
    },
    isRequired(item) {

      if (!this.$root.$validate.isEmpty(item)) {
        return false
      }

      if (item.indexOf('required') > -1) {
        return true
      }

      return false

    },
    // 控制表单字段在某种情况下是否显示 //通过设置ishide属性处理普通功能弹窗配置的显示与隐藏,非新建编辑弹窗
    isHide(val) {
      if (val === undefined) {
        return true
      }
      const res = (this.modelConfig.isAdd === true ? 'ADD':'EDIT')
      if (res === String(val).toUpperCase() || val === true) {// val,适用于配置项显示隐藏不依据新建/编辑状态，依赖其他条件，配置boolean值：false显示，true不显示
        return false
      }
      return true
    }
  },
  components: {}
}
</script>

<style lang="less" scoped>
 .isRequired_s{
    position: relative;
    right: -425px;
    top: -35px;
  }
  .ui-select {
    width: 70%;
  }
  // .modal{
  //   z-index: 2050;
  // }
  .modal-dialog {
    top:15%;
    font-family: MicrosoftYaHei;
  }
  .modal-body {
    position: initial;
    padding: 16px 8px;
    .params-each {
      display: block;
      margin-bottom: 6px;
      .v-selectss {
        vertical-align: middle;
        min-width: 70% !important;
        display: inline-block;
        padding: 0px;
        word-break: break-all;
      }
      .checkbox {
        margin-top: 11px;
      }
    }
  }
  .modal-header {
    height: 55px;
    align-items: center;
    .modal-title{
      font-size: 16px;
      // color: rgba(0,0,0,0.85);
    }
    i{
      font-size: 8px;
      color: #A7B3BD;
    }
  }
  .label-name {
     text-align: right;
     margin-top: 8px;
     padding-left: 0px;
     padding-right: 2px;
   }
  /*控制v-select时label样式*/
  .lable-name-select {
    // position: inherit;
    // bottom: 14px;
    vertical-align: middle;
  }
  .model-input{
    min-width: 70%;
    font-size: 12px;
    display: inline-block;
  }
  /*取消选中样式*/
  .form-control:focus {
    box-shadow: none;
  }
  .red-border {
    border: 1px solid red !important;
  }
  .is-danger {
    color: red;
    margin-left: 70px;
    margin-bottom: 0px;
  }

  .question-circle {
    font-size: 14px;
    color: gray;
    cursor: pointer;
  }
  .slot-content-question-circle{
    width:200px;
    padding:6px;
    white-space: normal;
  }

  button:focus {
    box-shadow: none;
  }
  .btn {
    width: 65px;
    height: 32px;
    padding: 4px 0px;
  }

  /*placeholer样式--开始 */
  ::-webkit-input-placeholder { /* WebKit, Blink, Edge */
    color: #c0c1c0;
  }
  :-moz-placeholder { /* Mozilla Firefox 4 to 18 */
    color: #c0c1c0;
  }
  ::-moz-placeholder { /* Mozilla Firefox 19+ */
    color: #c0c1c0;
  }
  :-ms-input-placeholder { /* Internet Explorer 10-11 */
    color: #c0c1c0;
  }
  /*placeholer样式--结束 */
  .tipSty{
    color: red;
    margin-left: 8%;
  }
  .textareaSty{
    display: inline-block;
    vertical-align: top;
    width: 70%;
    border-radius: 4px;
    border-color: #dddee1;
    height: 100px;
    padding: 3px;
  }
  .textareaSty:focus {
    outline: none !important;
    border-color: #719ECE;
  }
  textarea:disabled {
    background-color: #fff!important;
  }
</style>
